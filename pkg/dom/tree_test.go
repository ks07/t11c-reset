package dom

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func nodeFromString(src string) (*html.Node, error) {
	r := strings.NewReader(src)
	// html.Parse will fix broken documents, i.e. inject missing document tag structure
	// Prevent this by always parsing test data in the context of a fake body tag
	fakeParent := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Body,
		Data:     "body",
	}
	nodes, err := html.ParseFragment(r, fakeParent)
	if err != nil {
		return nil, err
	}
	if len(nodes) != 1 {
		return nil, errors.New("got multiple nodes from test src, should only have a single top-level node")
	}
	return nodes[0], nil
}

func docFromString(src string) (*html.Node, error) {
	r := strings.NewReader(src)
	return html.Parse(r)
}

func TestGetID(t *testing.T) {
	const srcBasicID = `<div id="mydiv">content</div>`
	n, err := nodeFromString(srcBasicID)
	if err != nil {
		t.Error(err)
	}
	ok, id := GetID(n)
	assert.True(t, ok, "Should find an ID in the basic case")
	assert.Equal(t, "mydiv", id, "Should extract the ID string")

	const srcNoID = `<p>Content</p>`
	n, err = nodeFromString(srcNoID)
	if err != nil {
		t.Error(err)
	}
	ok, _ = GetID(n)
	assert.False(t, ok, "Should not find an ID if one does not exist")

	const srcNoIDNested = `<p><span id="slogan_text">foo</span> is the slogan</p>`
	n, err = nodeFromString(srcNoIDNested)
	if err != nil {
		t.Error(err)
	}
	ok, _ = GetID(n)
	assert.False(t, ok, "Should not find an ID from a child node")

	const srcIDNested = `<ul id="opt-list"> <li id="opt-main">Foo</li> <li>Bar</li> </ul>`
	n, err = nodeFromString(srcIDNested)
	if err != nil {
		t.Error(err)
	}
	ok, id = GetID(n)
	assert.True(t, ok, "Should find an ID with nested elements")
	assert.Equal(t, "opt-list", id, "Should extract the ID of the parent")

	const srcIDAttrs = `<img src="foo.png" class="bar id" alt="the id of foo" id="mainImage">`
	n, err = nodeFromString(srcIDAttrs)
	if err != nil {
		t.Error(err)
	}
	ok, id = GetID(n)
	assert.True(t, ok, "Should find an ID amongst other attributes")
	assert.Equal(t, "mainImage", id, "Should extract the ID amongst other attributes")
}

func TestFindBodyElement(t *testing.T) {
	const srcBasic = `
		<html>
		<head><title>Hello World</title></head>
		<body>
			<div id="otherdiv"></div>
			<div id="mydiv"><br id="pass"></div>
        </body>
        </html>
	`
	doc, err := docFromString(srcBasic)
	if err != nil {
		t.Error(err)
	}
	n := FindBodyElement("mydiv", doc)
	assert.NotNil(t, n, "Should find a node")
	// The following checks determine whether the node identified is correct, using the br element as a marker
	assert.NotNil(t, n.FirstChild, "Should find the node with a child")
	ok, id := GetID(n.FirstChild)
	assert.True(t, ok, "Should find the node with the child marker node")
	assert.Equal(t, "pass", id, "Should find the correct node as marked by the pass br")

	const srcNone = `
		<html>
		<head><title>Hello World</title></head>
		<body>
			<div id="otherdiv"><br></div>
        </body>
        </html>
	`
	doc, err = docFromString(srcNone)
	if err != nil {
		t.Error(err)
	}
	n = FindBodyElement("mydiv", doc)
	assert.Nil(t, n, "Should not find a node if the id doesn't exist")

	const srcNested = `
		<html>
		<head><title>Hello World</title></head>
		<body>
			<div class="mydiv">
				<p>Foo</p>
				<div id="mydiv"><br id="pass"></div>
			</div>
        </body>
        </html>
	`
	doc, err = docFromString(srcNested)
	if err != nil {
		t.Error(err)
	}
	n = FindBodyElement("mydiv", doc)
	assert.NotNil(t, n, "Should find a node when nested")
	assert.NotNil(t, n.FirstChild, "Should find the node with a child")
	ok, id = GetID(n.FirstChild)
	assert.True(t, ok, "Should find the node with the child marker node")
	assert.Equal(t, "pass", id, "Should find the correct node as marked by the pass br")

	const srcInHead = `
		<html>
		<head><title id="targetid">Hello World</title></head>
		<body>
			<div id="otherdiv"><br></div>
        </body>
        </html>
	`
	doc, err = docFromString(srcInHead)
	if err != nil {
		t.Error(err)
	}
	n = FindBodyElement("targetid", doc)
	assert.Nil(t, n, "Should not find a node in the head section")
}
