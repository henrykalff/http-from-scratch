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
	req := http.Request{}
	err1 := startLine(&req, reader)
	if err1 != nil {
		return &http.Request{}, err1
	}

	header := http.Header{}
	err := parseHeader(header, reader)
	if err != nil {

	}
	parseHeaderToRequest(header, &req)
	parseBody(&req, reader)
	return &req, nil
}

func startLine(req *http.Request, reader *bufio.Reader) error {
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

func parseHeader(header http.Header, reader *bufio.Reader) error {
	for {
		line, _, err := reader.ReadLine()
		if err != nil {
			return err
		}
		if bytes.Equal(line, CRLF) || len(line) == 0 {
			break
		}
		decoded_line := string(line)
		before, after, found := strings.Cut(decoded_line, ":")
		log.Println(decoded_line)
		log.Println(before, after, found)

		if !found {
			return errors.New("MethodNotFound")
		}
		header.Add(before, after)
	}
	return nil
}

func parseHeaderToRequest(header http.Header, req *http.Request) error {
	if field, exists := header["Content-Length"]; exists {
		contentLength, err := strconv.ParseInt(field[0], 10, 64)
		if err != nil {
			return err
		}
		req.ContentLength = contentLength
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
