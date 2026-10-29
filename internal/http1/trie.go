package http1

import (
	"net/http"
	"strings"
)

type SegmentTrie struct {
	root *TrieNode
}

func NewSegmentTrie() SegmentTrie {
	return SegmentTrie{root: &TrieNode{map[string]*TrieNode{}, nil, map[string]Handler{}}}
}

func (t SegmentTrie) SetRoute(pattern string, handler Handler) {
	method, uri, hasMethod := strings.Cut(pattern, " ")
	if !hasMethod {
		uri = method
		method = "all"
	}

	t.root.insert(method, uri, handler)

}

func (t SegmentTrie) Handler(r *http.Request) Handler {
	return search(t.root, r)
}

type TrieNode struct {
	children map[string]*TrieNode
	param    *TrieNode
	handlers map[string]Handler
}

func (node *TrieNode) insert(method string, path string, handler Handler) {
	if path == "/" {
		node.setMethodForNode(method, handler)
		return
	}
	segments := strings.Split(path, "/")

	if segments[0] != "" {
		panic("Specified URI must start with \"/\"")
	}

	for _, seg := range segments[1:] {
		if param := parseParameter(seg); param != "" {
			if node.param == nil {
				node.param = &TrieNode{
					children: map[string]*TrieNode{},
					param:    nil,
					handlers: map[string]Handler{},
				}
			}
			node = node.param
			continue
		}
		if child := node.children[seg]; child != nil {
			node = child
		} else {
			tmp := TrieNode{
				children: map[string]*TrieNode{},
				handlers: map[string]Handler{},
			}
			node.children[seg] = &tmp
			node = &tmp
		}
	}
	node.setMethodForNode(method, handler)
}

func (node *TrieNode) setMethodForNode(method string, handler Handler) {
	if method == "all" {
		for _, m := range VALIDHTTPMethods {
			node.handlers[m] = handler
		}
	} else {
		node.handlers[method] = handler
	}
}

func parseParameter(seg string) string {
	if seg[0] == '{' && seg[len(seg)-1] == '}' {
		return seg[1 : len(seg)-1]
	}
	return ""
}

func search(node *TrieNode, req *http.Request) Handler {
	segments := strings.Split(req.URL.Path, "/")

	for _, seg := range segments[1:] {
		if child := node.children[seg]; child != nil {
			node = child
		} else if node.param != nil {
			node = node.param
		} else {
			return nil
		}
	}
	return node.handlers[req.Method]
}
