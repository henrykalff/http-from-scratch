package http1

import (
	"log"
	"net/http"
	"testing"
)

type InsertTestCase struct {
	behavior int // 0=success, 1=error, 2=panic
	url      string
}
type SearchTestCase struct {
	url string
	exp string
}

func TestInsert_SingleTrie(t *testing.T) {
	tcs := []InsertTestCase{
		{0, "/"},
		{0, "/hello"},
		{0, "/hellohere"},
		{0, "/hello/{id}"},
	}
	trie := NewSegmentTrie()
	c := make(chan string)
	for _, tc := range tcs {
		log.Println("Test: ", tc.url)
		trie.SetRoute(tc.url, handlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c <- tc.url
		}))
	}
}
