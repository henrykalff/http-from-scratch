package http1

import (
	"bufio"
	"bytes"
	"errors"
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

	parseBody(&req, reader)
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
	method, exists := VALIDHTTPMethods[decodedSplittedLine[0]]
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
	majorS, minorS, sepExists := strings.Cut(protoVersion[1], ".")
	major, err1 := strconv.ParseInt(majorS, 10, 64)
	minor, err2 := strconv.ParseInt(minorS, 10, 64)
	if !sepExists || err1 != nil || err2 != nil {
		return errors.New("Invalid Protocol Version")
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
		log.Printf("FieldName: %s, FieldValue: %s\n", fieldName, fieldValue)
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

	return nil
}

func parseBody(req *http.Request, reader *bufio.Reader) error {
	if req.ContentLength == 0 {
		return nil
	}
	limited := io.LimitReader(reader, req.ContentLength)
	req.Body = io.NopCloser(limited)
	return nil
}
