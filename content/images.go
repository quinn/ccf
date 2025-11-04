package content

import (
	"bufio"
	"bytes"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

type markdownImages struct {
	parentPath string
	callback   func(imageTag string) string
}

// Extend implements goldmark.Extender.
func (e *markdownImages) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newMarkdownImagesRenderer(e.parentPath, e.callback), 500),
	))
}

type markdownImagesRenderer struct {
	html.Config
	parentPath string
	callback   func(imageTag string) string
}

func (r *markdownImagesRenderer) encodeImage(src []byte) string {
	s := string(src)

	// hotlink
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	}

	// data url
	if strings.HasPrefix(s, "data:") {
		return s
	}

	// absolute path
	if filepath.IsAbs(s) {
		return s
	}

	return filepath.Join(r.parentPath, s)
}

// ALL THE STUFF BELOW IS BOILERPLATE COPIED FROM
// github.com/tenkoh/goldmark-img64@v0.1.1
// I HAVE NO IDEA WHAT IT DOES

func (r *markdownImagesRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*ast.Image)
	elt := ""
	elt += `<img src="`

	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		elt += r.encodeImage(n.Destination)
	}
	elt += `" alt="`
	elt += string(nodeToHTMLText(n, source))
	elt += `"`
	if n.Title != nil {
		elt += ` title="`
		elt += string(n.Title)
		elt += `"`
	}

	buf := &bytes.Buffer{}
	if n.Attributes() != nil {
		html.RenderAttributes(bufio.NewWriter(buf), n, html.ImageAttributeFilter)
	}
	elt += buf.String()

	if r.XHTML {
		elt += " />"
	} else {
		elt += ">"
	}

	elt = r.callback(elt)
	_, _ = w.WriteString(elt)
	return ast.WalkSkipChildren, nil
}

func nodeToHTMLText(n ast.Node, source []byte) []byte {
	var buf bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if s, ok := c.(*ast.String); ok && s.IsCode() {
			buf.Write(s.Text(source))
		} else if !c.HasChildren() {
			buf.Write(util.EscapeHTML(c.Text(source)))
		} else {
			buf.Write(nodeToHTMLText(c, source))
		}
	}
	return buf.Bytes()
}

// RegisterFuncs implements renderer.NodeRenderer.
func (r *markdownImagesRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindImage, r.renderImage)
}

func newMarkdownImagesRenderer(parentPath string, callback func(imageTag string) string) renderer.NodeRenderer {
	return &markdownImagesRenderer{
		parentPath: parentPath,
		callback:   callback,
	}
}
