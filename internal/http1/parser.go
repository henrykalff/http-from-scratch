package http1

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type HTTPMethod string

var VALIDHTTPMethods = map[string]string{
	"GET":    "GET",
	"HEAD":   "HEAD",
	"POST":   "POST",
	"PUT":    "PUT",
	"DELETE": "DELETE"}

var ErrUnsupportedVersion = errors.New("505 HTTP Version Not Supported")
var ErrBadRequest = errors.New("HTTP is malformed")

// const requestLine = 114 101 113 117 101 115 116 45 108 105 110 101
func ParseRequest(reader *bufio.Reader) (*http.Request, error) {
	req := http.Request{
		Header: make(http.Header),
	}

	if err := parseRequestLine(&req, reader); err != nil {
		return nil, err
	}

	if err := parseHeader(&req, reader); err != nil {
		return nil, err
	}

	if err := parseBody(&req, reader); err != nil {
		return nil, err
	}
	return &req, nil
}

func parseRequestLine(req *http.Request, reader *bufio.Reader) error {
	line, _, err := reader.ReadLine()
	if err != nil {
		return err
	}
	decodedLine := string(line)
	decodedSplittedLine := strings.Split(decodedLine, " ")

	if len(decodedSplittedLine) != 3 {
		return errors.New("BadRequest")
	}

	// Parse HTTPMethod to request
	method, exists := VALIDHTTPMethods[strings.ToUpper(decodedSplittedLine[0])]
	if !exists {
		return errors.New("MethodNotFound")
	}
	req.Method = method

	// Parse Url
	parsedURL, errParseUrl := url.Parse(decodedSplittedLine[1])
	if nil != errParseUrl {
		return errParseUrl
	}
	req.URL = parsedURL

	req.Proto = decodedSplittedLine[2]
	protoVersion := strings.Split(decodedSplittedLine[2], "/")
	if protoVersion[0] != "HTTP" {
		return ErrBadRequest
	}
	majorS, minorS, sepExists := strings.Cut(protoVersion[1], ".")
	major, err1 := strconv.ParseInt(majorS, 10, 64)
	minor, err2 := strconv.ParseInt(minorS, 10, 64)

	if !sepExists || err1 != nil || err2 != nil {
		return errors.New("Invalid Protocol Version")
	}
	if major != 1 || minor > 1 {
		return fmt.Errorf("%w: received HTTP/%d.%d", ErrUnsupportedVersion, major, minor)
	}
	req.ProtoMajor = int(major)
	req.ProtoMinor = int(minor)
	return nil
}

var CRLF = []byte("\r\n")

func parseHeader(req *http.Request, reader *bufio.Reader) error {

	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return err
		}

		if bytes.Equal(line, CRLF) || len(line) == 0 {
			break
		}

		fieldName, fieldValue, found := strings.Cut(string(line), ":")

		if !found {
			return errors.New("MethodNotFound")
		}
		fieldName = strings.TrimSpace(fieldName)
		fieldValue = strings.TrimSpace(fieldValue)
		// log.Printf("FieldName: %s, FieldValue: %s\n", fieldName, fieldValue)
		req.Header.Add(fieldName, fieldValue)
	}

	host := req.Header.Get("Host")
	if host == "" {
		return errors.New("HTTP/1.1 requires Host field")
	}
	req.Host = host
	req.Header.Del("Host")

	if transferEncoding := req.Header.Get("Transfer-Encoding"); transferEncoding != "" {
		req.TransferEncoding = strings.Split(transferEncoding, ",")
		for i := range req.TransferEncoding {
			req.TransferEncoding[i] = strings.TrimSpace(req.TransferEncoding[i])
		}
		req.Header.Del("Transfer-Encoding")
	}

	if contentLength := req.Header.Get("Content-Length"); contentLength != "" {
		val, err := strconv.ParseInt(strings.TrimSpace(contentLength), 10, 64)
		if err != nil || val < 0 {
			return errors.New("invalid Content Length")
		}
		req.ContentLength = val
		req.Header.Del("Content-Length")
	} else {
		req.ContentLength = 0
	}

	parseClose(req)

	return nil
}

func parseClose(req *http.Request) {
	connectionHeader := strings.ToLower(req.Header.Get("Connection"))
	isHTTP10 := req.ProtoMajor == 1 && req.ProtoMinor == 0

	req.Close = strings.Contains(connectionHeader, "close") || (isHTTP10 && !strings.Contains(connectionHeader, "keep-alive"))
}

func parseBody(req *http.Request, reader *bufio.Reader) error {
	for _, enc := range req.TransferEncoding {
		if strings.EqualFold(enc, "chunked") {
			cr := &chunkedReader{reader: reader}
			body, err := io.ReadAll(cr)
			if err != nil {
				return err
			}
			req.Body = io.NopCloser(bytes.NewReader(body))
			return nil
		}
	}

	if req.ContentLength == 0 {
		return nil
	}
	_, err := reader.Peek(int(req.ContentLength))
	if err != nil {
		log.Println("Too short.")
		return fmt.Errorf("%w: Content-Length is too short.", ErrBadRequest)
	}
	limited := io.LimitReader(reader, req.ContentLength)
	req.Body = io.NopCloser(limited)
	return nil
}

func readCRLFLine(reader *bufio.Reader) (string, error) {
	var line []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return "", err
		}
		if b == '\r' {
			next, err := reader.ReadByte()
			if err != nil {
				return "", err
			}
			if next == '\n' {
				return string(line), nil
			}
			return "", fmt.Errorf("%w: bare CR", ErrBadRequest)
		}
		if b == '\n' {
			return "", fmt.Errorf("%w: bare CR", ErrBadRequest)
		}
		line = append(line, b)
	}
}

type chunkedReader struct {
	reader    *bufio.Reader
	remaining int64
	done      bool
}

func (cr *chunkedReader) Read(p []byte) (int, error) {
	if cr.done {
		return 0, io.EOF
	}

	if cr.remaining == 0 {
		chunkSizeS, err := readCRLFLine(cr.reader)
		if err != nil {
			return 0, fmt.Errorf("%w: truncated chunk header", ErrBadRequest)
		}
		chunkSize, err := strconv.ParseInt(string(chunkSizeS), 16, 64)
		log.Println(chunkSize)
		if err != nil || chunkSize < 0 {
			return 0, fmt.Errorf("%w: Could Not Parse Chunked Size", ErrBadRequest)
		}
		if chunkSize == 0 {
			cr.done = true
			emptyChunkTrailer, err := readCRLFLine(cr.reader)
			if err != nil {
				return 0, fmt.Errorf("%w: truncated chunk header", ErrBadRequest)
			}
			if len(emptyChunkTrailer) != 0 {
				return 0, fmt.Errorf("%w: Last Trailer is not empty", ErrBadRequest)
			}
			return 0, io.EOF
		}
		cr.remaining = chunkSize
	}

	toRead := int64(len(p))
	if toRead > cr.remaining {
		toRead = cr.remaining
	}

	n, err := cr.reader.Read(p[:toRead])
	cr.remaining -= int64(n)
	if cr.remaining == 0 && err == nil {
		crlf, err := readCRLFLine(cr.reader)
		if err != nil || crlf != "" {
			return n, fmt.Errorf("%w: missing CRLF after chunk data", ErrBadRequest)
		}
	}

	if err != nil {
		if errors.Is(err, io.EOF) && cr.remaining > 0 {
			return n, fmt.Errorf("%w: chunk body truncted", ErrBadRequest)
		}
		return n, err
	}
	return n, nil

}

func (cr *chunkedReader) Close() error {
	return nil
}
