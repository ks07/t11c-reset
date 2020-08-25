package dom

import "golang.org/x/net/html"

func GetID(n *html.Node) (bool, string) {
	for _, attr := range n.Attr {
		if attr.Key == "id" {
			return true, attr.Val
		}
	}
	return false, ""
}

func FindBodyElement(id string, n *html.Node) *html.Node {
	// Quit early if the node isn't an interesting type or one that can have interesting children
	if n.Type != html.ElementNode && n.Type != html.DocumentNode {
		return nil
	}

	if n.Type == html.ElementNode {
		// Don't traverse the head section
		if n.Data == "head" {
			return nil
		}

		if ok, nID := GetID(n); ok && nID == id {
			return n
		}
	}

	// Search nodes depth-first
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		if matched := FindBodyElement(id, child); matched != nil {
			return matched
		}
	}

	return nil
}
