package component

import (
	"bytes"
	"html/template"

	//"be/lex"
)

type ContentElement interface {
	Render() (template.HTML, error)
}

type Section struct {
	ID string
	Title string
	Content []ContentElement
	Level int
}

var _ ContentElement = (*Section)(nil)

func (s Section) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	templName := "Section"
	switch s.Level {
	case 0:
		templName = "Section"
	default: fallthrough
	case 1:
		templName = "Subsection"
	}
	err := pages.Render(buf, templName, s)
	return template.HTML(buf.String()), err
}

const HtmlSection = `
{{ define "Section" }}
<section id="{{.ID}}">
	<h2><a href="#{{.ID}}">{{.Title}}</a></h2>
	{{ range .Content }}
		{{ Render . }}
	{{ end }}
</section>
{{ end }}
`

const HtmlSubsection = `
{{ define "Subsection" }}
<section id="{{.ID}}">
	<h3><a href="#{{.ID}}">{{.Title}}</a></h3>
	{{ range .Content }}
		{{ Render . }}
	{{ end }}
</section>
{{ end }}
`

type Text string

var _ ContentElement = (*Text)(nil)

func (t Text) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Render(buf, "Paragraph", t)
	return template.HTML(buf.String()), err
}

const HtmlParagraph = `
{{ define "Paragraph" }}
<p>
{{ . }}
</p>
{{ end }}
`

type Link struct {
	Link string
	External bool
}

var _ ContentElement = (*Link)(nil)

func (l Link) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Render(buf, "Link", l)
	return template.HTML(buf.String()), err
}

const HtmlLink = `
{{ define "Link" }}
<a href="{{.Link}}" {{ if .External }} target="_blank" {{ end }}>{{.Text}}</a>
{{ end }}
`

const HtmlAside = `
{{ define "Aside" }}
<aside>
	{{ range .Content }}
		{{ Render . }}
	{{ end }}
</aside>
{{ end }}
`

