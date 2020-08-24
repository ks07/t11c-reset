/*
Copyright © 2020 George Field <george@cucurbit.dev>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package t11c

import (
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
)

var errWANIPElementNotFound = errors.New("no WAN IP element found")
var errWANIPTextNotFound = errors.New("no WAN IP text found")

func extractWANIP(body io.Reader) (string, error) {
	dom, err := html.Parse(body)
	if err != nil {
		return "", err
	}

	n := findNode("DeviceInfo_WanIP", dom)
	if n == nil {
		return "", errWANIPElementNotFound
	}

	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.TextNode {
			return strings.TrimSpace(child.Data), nil
		}
	}

	return "", errWANIPTextNotFound
}

func getID(n *html.Node) (bool, string) {
	for _, attr := range n.Attr {
		if attr.Key == "id" {
			return true, attr.Val
		}
	}
	return false, ""
}

func findNode(id string, n *html.Node) *html.Node {
	// Quit early if the node isn't an interesting type or one that can have interesting children
	if n.Type != html.ElementNode && n.Type != html.DocumentNode {
		return nil
	}

	if n.Type == html.ElementNode {
		// Don't traverse the head section
		if n.Data == "head" {
			return nil
		}

		ok, nID := getID(n)
		if ok && nID == id {
			return n
		}
	}

	// Search nodes depth-first
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		matched := findNode(id, child)
		if matched != nil {
			return matched
		}
	}

	return nil
}
