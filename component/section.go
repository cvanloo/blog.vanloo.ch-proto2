package component

import (
	"html/template"

	"be/lex"
)

type Section struct {
	ID string
	Title string
	Content *lex.Node
}

const HtmlSection = `
{{ define "Section" }}
<section id="{{.ID}}">
	<h2><a href="#{{.ID}}">{{.Title}}</a></h2>
	{{ range .Content }}
		{{ Evaluate . }}
	{{ end }}
</section>
{{ end }}
`

const HtmlSubsection = `
{{ define "Subsection" }}
<section id="{{.ID}}">
	<h3><a href="#{{.ID}}">{{.Title}}</a></h3>
	{{ range .Content }}
		{{ Evaluate . }}
	{{ end }}
</section>
{{ end }}
`

type Text string

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

const HtmlLink = `
{{ define "Link" }}
<a href="{{.Link}}" {{ if .External }} target="_blank" {{ end }}>{{.Text}}</a>
{{ end }}
`

const HtmlAside = `
{{ define "Aside" }}
<aside>
	{{ range .Content }}
		{{ Evaluate . }}
	{{ end }}
</aside>
{{ end }}
`

