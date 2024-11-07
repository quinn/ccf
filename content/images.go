package content

import (
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
}
type markdownImagesRenderer struct {
	html.Config
	parentPath string
}

func (r *markdownImagesRenderer) encodeImage(src []byte) ([]byte, error) {
	s := string(src)

	// hotlink
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return src, nil
	}

	// data url
	if strings.HasPrefix(s, "data:") {
		return src, nil
	}

	// absolute path
	if filepath.IsAbs(s) {
		return src, nil
	}

	return []byte(r.parentPath + "/" + s), nil
}

// ALL THE STUFF BELOW IS BOILERPLATE COPIED FROM
// github.com/tenkoh/goldmark-img64@v0.1.1
// I HAVE NO IDEA WHAT IT DOES

func (r *markdownImagesRenderer) renderImage(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	n := node.(*ast.Image)
	_, _ = w.WriteString(`<img src="`)
	if r.Unsafe || !html.IsDangerousURL(n.Destination) {
		s, err := r.encodeImage(n.Destination)
		if err != nil || s == nil {
			_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
		} else {
			_, _ = w.Write(s)
		}
	}
	_, _ = w.WriteString(`" alt="`)
	_, _ = w.Write(nodeToHTMLText(n, source))
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		r.Writer.Write(w, n.Title)
		_ = w.WriteByte('"')
	}
	if n.Attributes() != nil {
		html.RenderAttributes(w, n, html.ImageAttributeFilter)
	}
	if r.XHTML {
		_, _ = w.WriteString(" />")
	} else {
		_, _ = w.WriteString(">")
	}
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

func newMarkdownImagesRenderer(parentPath string) renderer.NodeRenderer {
	return &markdownImagesRenderer{
		parentPath: parentPath,
	}
}

// Extend implements goldmark.Extender.
func (e *markdownImages) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(newMarkdownImagesRenderer(e.parentPath), 500),
	))
}
