package component

import "be/lex"

type Section struct {
	ID string
	Title string
	Content *lex.Node
}

const HtmlSection = `
<section id="{{.ID}}">
	<h2><a href="#{{.ID}}">{{.Title}}</a></h2>
	{{ range .Content }}
		{{ Evaluate . }}
	{{ end }}
</section>
`

const HtmlSubsection = `
<section id="{{.ID}}">
	<h3><a href="#{{.ID}}">{{.Title}}</a></h3>
	{{ range .Content }}
		{{ Evaluate . }}
	{{ end }}
</section>
`

type Text string

const HtmlParagraph = `
<p>
{{ . }}
</p>
`

type Link struct {
	Link string
	External bool
}

const HtmlLink = `
<a href="{{.Link}}" {{ if .External }} target="_blank" {{ end }}>{{.Text}}</a>
`

const HtmlAside = `
<aside>
	{{ range .Content }}
		{{ Evaluate . }}
	{{ end }}
</aside>
`

