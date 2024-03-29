package component

import (
	"bytes"
	"fmt"
	"html/template"

	//"be/lex"
)

var generatedIDs = map[string]struct{}{}

func GenerateID(name string) (id string) {
	for _, r := range name {
		if 'a' <= r && r <= 'z' {
			id += string(r)
		} else if r == ' ' {
			id += "-"
		} else if 'A' <= r && r <= 'Z' {
			id += string(r + ('a' - 'A')) // to lower case
		} else {
			id += "_"
		}
	}
	n, alreadyExists, ext := 1, true, ""
	for alreadyExists {
		if _, alreadyExists = generatedIDs[id + ext]; alreadyExists {
			ext = fmt.Sprintf("-%d", n)
			n++
		}
	}
	generatedIDs[id + ext] = struct{}{}
	return id + ext
}

type (
	ContentElement interface {
		Render() (template.HTML, error)
	}
	Renderable interface {
		ContentElement
		Append(child ContentElement)
	}
)

type (
	Section struct {
		ID string
		Title string
		Content []ContentElement
		Level SectionLevel
	}
	SectionLevel int
)

const (
	SectionLevelSection SectionLevel = iota
	SectionLevelSubsection
)

var _ Renderable = (*Section)(nil)

func NewSection(title string) *Section {
	return &Section{
		ID: GenerateID(title),
		Title: title,
		Content: []ContentElement{},
		Level: SectionLevelSection,
	}
}

func NewSubsection(title string) *Section {
	return &Section{
		ID: GenerateID(title),
		Title: title,
		Content: []ContentElement{},
		Level: SectionLevelSubsection,
	}
}

func (s *Section) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	templName := "Section"
	switch s.Level {
	case SectionLevelSection:
		templName = "Section"
	default: fallthrough
	case SectionLevelSubsection:
		templName = "Subsection"
	}
	err := pages.Render(buf, templName, s)
	return template.HTML(buf.String()), err
}

func (s *Section) Append(child ContentElement) {
	s.Content = append(s.Content, child)
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

type Comment string

var _ ContentElement = (*Comment)(nil)

func (c Comment) Render() (template.HTML, error) {
	// ExecuteTemplate would remove the entire thing from the output; don't call it
	return template.HTML("<!-- " + string(c) + " -->"), nil
}

type ActualComment struct {
	Content []ContentElement
}

var _ Renderable = (*ActualComment)(nil)

func (c *ActualComment) Render() (template.HTML, error) {
	// @fixme: well then, ... we could just skip the rendering right away...
	buf := &bytes.Buffer{}
	err := pages.Render(buf, "ActualComment", c)
	return template.HTML(buf.String()), err
}

func (c *ActualComment) Append(child ContentElement) {
	c.Content = append(c.Content, child)
}

const HtmlActualComment = `
{{ define "ActualComment" }}
<!--
{{ range .Content }}
	{{ Render . }}
{{ end }}
-->
{{ end }}
`
