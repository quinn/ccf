package content

import (
	"bufio"
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
	"go.abhg.dev/goldmark/wikilink"
)

type markdownImages struct {
	parentPath  string
	callback    func(imageTag string) string
	resolveLink func(target string) string
}

// Extend implements goldmark.Extender.
func (e *markdownImages) Extend(m goldmark.Markdown) {
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		// Use priority 100 to override default wikilink renderer (lower number = higher priority)
		util.Prioritized(newMarkdownImagesRenderer(e.parentPath, e.callback, e.resolveLink), 100),
	))
}

type markdownImagesRenderer struct {
	html.Config
	parentPath  string
	callback    func(imageTag string) string
	resolveLink func(target string) string

	// hasDest records whether a node had a destination when we resolved
	// it. This is needed to decide whether a closing </a> must be added
	// when exiting a Node render.
	hasDest sync.Map // *Node => struct{}
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

	if r.callback != nil {
		elt = r.callback(elt)
	}
	_, _ = w.WriteString(elt)
	return ast.WalkSkipChildren, nil
}

// renderWikilink handles Obsidian embed syntax: ![[image.png]]
func (r *markdownImagesRenderer) renderWikilink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	link, ok := node.(*wikilink.Node)
	if !ok {
		return ast.WalkStop, fmt.Errorf("unexpected node %T, expected *wikilink.Node", node)
	}

	if entering {
		return r.enterWikilink(w, link)
	}

	r.exitWikilink(w, link)
	return ast.WalkContinue, nil
}

// renderWikilink handles Obsidian embed syntax: ![[image.png]]
func (r *markdownImagesRenderer) exitWikilink(w util.BufWriter, link *wikilink.Node) {
	if _, ok := r.hasDest.LoadAndDelete(link); ok {
		_, _ = w.WriteString("</a>")
	}
}

func (r *markdownImagesRenderer) enterWikilink(w util.BufWriter, link *wikilink.Node) (ast.WalkStatus, error) {
	if !link.Embed {
		r.hasDest.Store(link, struct{}{})
		_, _ = w.WriteString(`<a href="`)
		_, _ = w.Write(util.URLEscape([]byte(r.resolveLink(string(link.Target))), true /* resolve references */))
		_, _ = w.WriteString(`">`)
		return ast.WalkContinue, nil
	}

	target := string(link.Target)
	basename := filepath.Base(target)

	// Check if this is an image file
	ext := strings.ToLower(filepath.Ext(basename))
	imageExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
		".webp": true, ".svg": true, ".bmp": true, ".mp4": true,
	}

	if !imageExts[ext] {
		// Not an image, let default renderer handle it
		return ast.WalkContinue, nil
	}

	// Render as image
	elt := ""
	elt += `<img src="`
	elt += r.encodeImage([]byte(basename))
	elt += `" alt="`
	elt += basename
	elt += `"`

	if r.XHTML {
		elt += " />"
	} else {
		elt += ">"
	}

	if r.callback != nil {
		elt = r.callback(elt)
	}
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
	reg.Register(wikilink.Kind, r.renderWikilink)
}

func newMarkdownImagesRenderer(parentPath string, callback func(imageTag string) string, resolveLink func(target string) string) renderer.NodeRenderer {
	return &markdownImagesRenderer{
		parentPath:  parentPath,
		callback:    callback,
		resolveLink: resolveLink,
	}
}
