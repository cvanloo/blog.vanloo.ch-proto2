package be

import (
	"bytes"
	"time"
	"fmt"
	"html/template"
	"io"
	"net/http"
)

var (
	pages Template = Template{template.New("")}
	generatedIDs = map[string]struct{}{}
)

func init() {
	pages.Funcs(template.FuncMap{
		"Render": Render,
	})

	template.Must(pages.Parse(HtmlCodeBlock))
	template.Must(pages.Parse(HtmlEntry))
	template.Must(pages.Parse(HtmlSection))
	template.Must(pages.Parse(HtmlSubsection))
	template.Must(pages.Parse(HtmlText))
	template.Must(pages.Parse(HtmlParagraph))
	template.Must(pages.Parse(HtmlLink))
	template.Must(pages.Parse(HtmlAside))
	template.Must(pages.Parse(HtmlSidenote))
	template.Must(pages.Parse(HtmlEnquote))
	template.Must(pages.Parse(HtmlMono))
	template.Must(pages.Parse(HtmlEm))
}

func Render(element Renderable) (template.HTML, error) {
	return element.Render()
}

type Template struct {
	*template.Template
}

func (t *Template) Execute(w io.Writer, name string, data any) error {
	return t.Template.ExecuteTemplate(w, name, data)
}

func String(blog *Blog) (string, error) {
	bs := &bytes.Buffer{}
	err := pages.Execute(bs, "Entry", blog)
	return bs.String(), err
}

func Handler(blog *Blog, onError func(error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := pages.Execute(w, "Entry", blog)
		if err != nil {
			onError(err)
		}
	}
}

func GenerateID(prefix string) (id string) {
	for _, r := range prefix {
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
	Renderable interface {
		Render() (template.HTML, error)
	}
	CompositeRenderable interface {
		Renderable
		Append(child Renderable)
	}
	TextRenderable interface {
		Renderable
		Text() string
	}
)

type (
	Blog struct {
		BlogName string
		Title, AltTitle string
		Author Author
		Tags Tags
		Meta Meta
		Abstract string
		Languages []Language
		Content []Renderable
	}
	Author struct {
		Name string
		EMail string
	}
	Language struct {
		Link string
		Language string
	}
	Meta struct {
		// https://en.wikipedia.org/wiki/List_of_ISO_639_language_codes
		Language string
		CanonicalURL string
		Description string
		Published time.Time
		Revisions []time.Time
		Topic string
		EstReadingTime ReadingTime
	}
	Tag string
	Tags []Tag
	ReadingTime struct {
		time.Duration
	}
)

var _ CompositeRenderable = (*Blog)(nil)

func (blog *Blog) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Entry", blog)
	return template.HTML(buf.String()), err
}

func (blog *Blog) Append(child Renderable) {
	blog.Content = append(blog.Content, child)
}

func (t Tag) String() string {
	return ":" + string(t)
}

func (ts Tags) KeywordList() (s string) {
	if ts == nil || len(ts) == 0 {
		return ""
	}
	s = string(ts[0])
	if len(ts) > 1 {
		for _, t := range ts[1:] {
			s += ", " + string(t)
		}
	}
	return s
}

func (m Meta) IsRevised() bool {
	return len(m.Revisions) > 0
}

func (m Meta) LastRevised() time.Time {
	if len(m.Revisions) > 0 {
		return m.Revisions[len(m.Revisions)-1]
	}
	return time.Time{}
}

func (m Meta) CopyYear() int {
	if m.IsRevised() {
		return m.LastRevised().Year()
	}
	return m.Published.Year()
}

func (rt ReadingTime) String() string {
	return fmt.Sprintf("~%s", rt.Duration.String())
}

const HtmlEntry = `
{{ define "Entry" }}
<!DOCTYPE html>
<html lang="{{.Meta.Language}}">
	<head>
		<meta charset="utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<link rel="stylesheet" href="/public/styles.css" />
		<link rel="icon" type="image/png" href="/public/favicon.png" />
		<link rel="canonical" href="{{.Meta.CanonicalURL}}" />
		<title>{{.Title}} &mdash; ({{.BlogName}})</title>
		<meta name="author" content="{{.Author.Name}}" />
		<meta name="keywords" content="{{.Tags.KeywordList}}"/>
		<meta name="description" content="{{.Meta.Description}}"/>
		{{ if .Meta.IsRevised }}
		<meta name="revised" content="{{.Meta.LastRevised}}" />
		{{ end }}
		<meta name="topic" content="{{.Meta.Topic}}">
		<meta name="subject" content="{{.Meta.Topic}}">
		<meta name="language" content="{{.Meta.Language}}">
		<meta name="abstract" content="{{.Abstract}}">
		<meta name="summary" content="{{.Abstract}}">
		<meta name="url" content="{{.Meta.CanonicalURL}}">
		<meta name="og:title" content="{{.Title}}"/>
		<meta name="og:type" content="article"/>
		<meta property="article:published_time" content="{{.Meta.Published.Format "2006-01-02"}}" />
		{{ if .Meta.IsRevised }}
		<meta property="article:revised_time" content="{{.Meta.LastRevised}}" />
		{{ end }}
		<meta name="og:url" content="{{.Meta.CanonicalURL}}"/>
		<meta name="og:site_name" content="{{.BlogName}}"/>
		<meta name="og:description" content="{{.Meta.Description}}"/>
	</head>
	<body>
		<div class="scroll-progress">
			<div id="scroll-progress"></div>
		</div>
		<header>
			<nav>
				<p class="fill">
				<!-- 2^7633587786 -->
				<code>({{.BlogName}}</code>
				<span class="keywords">
					<code><a href="/index.html">:home</a></code>
					<code><a href="/about.html">:about</a></code>
					<code><a href="/rss.xml">:rss</a></code>
				</span>
				<code>)</code>
				</p>
			</nav>
		</header>
		<main>
			<article>
				<div class="title">
					<h1>{{.Title}}</h1>
					<aside class="content-info">
						<div class="info">
							<p class="published-date"><small>{{.Meta.Published.Format "2006-01-02"}}</small></p>
							<p class="time-est-reading"><small>{{.Meta.EstReadingTime}}</small></p>
						</div>
						<div class="taglist">
							{{ range .Tags }}
							<p><a href="/search?tags={{.}}">{{.}}</a></p>
							{{ end}}
						</div>
					</aside>
				</div>
				<ul class="language-selection">
					<li>English
						<ul class="dropdown">
							{{ range .Languages }}
							<li><a href="{{.Link}}">{{.Language}}</a></li>
							{{ end }}
						</ul>
					</li>
				</ul>
				{{ range .Content }}
					{{ Render . }}
				{{ end }}
			</article>
		</main>
		<footer>
			<p id="eof">STOP)))))</p>
			<address>&copy; {{.Meta.CopyYear}} <a href="mailto:{{.Author.EMail}}?subject=RE:%20{{.Title}}">{{.Author.Name}}</a></address>
			<span class="credits">
				<a href="/about.html#credits">Font Licenses</a>
				<a href="/about.html">About</a>
				<a href="/rss.xml">RSS Feed</a>
			</span>
		</footer>
		<script>
			function calculateProgress() {
				const winScroll = document.body.scrollTop || document.documentElement.scrollTop;
				const height = document.documentElement.scrollHeight - document.documentElement.clientHeight;
				const scrolled = (winScroll / height) * 100;
				document.getElementById('scroll-progress').style.width = scrolled + "%";
			}
			window.onscroll = function() {
				calculateProgress();
			};
		</script>
	</body>
</html>
{{ end }}
`

type (
	Section struct {
		ID string
		Title string
		Content []Renderable
		Level SectionLevel
	}
	SectionLevel int
)

const (
	SectionLevelSection SectionLevel = iota
	SectionLevelSubsection
)

var _ CompositeRenderable = (*Section)(nil)

func NewSection(title string) *Section {
	return &Section{
		ID: GenerateID(title),
		Title: title,
		Content: []Renderable{},
		Level: SectionLevelSection,
	}
}

func NewSubsection(title string) *Section {
	return &Section{
		ID: GenerateID(title),
		Title: title,
		Content: []Renderable{},
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
	err := pages.Execute(buf, templName, s)
	return template.HTML(buf.String()), err
}

func (s *Section) Append(child Renderable) {
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

type Paragraph struct {
	Content []Renderable
}

var _ CompositeRenderable = (*Paragraph)(nil)

func (p *Paragraph) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Paragraph", p)
	return template.HTML(buf.String()), err
}

func (p *Paragraph) Append(child Renderable) {
	p.Content = append(p.Content, child)
}

const HtmlParagraph = `
{{ define "Paragraph" }}<p>
{{ range .Content }}
{{ Render . }}
{{ end }}
</p>
{{ end }}
`

type Text string

var _ TextRenderable = (*Text)(nil)

func (t Text) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Text", t)
	return template.HTML(buf.String()), err
}

func (t Text) Text() string {
	return string(t)
}

const HtmlText = `{{ define "Text" }}{{ . }}{{ end }}`

type Link struct {
	Link string
	External bool
}

var _ Renderable = (*Link)(nil)

func (l Link) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Link", l)
	return template.HTML(buf.String()), err
}

const HtmlLink = `
{{ define "Link" }}
<a href="{{.Link}}" {{ if .External }} target="_blank" {{ end }}>{{.Text}}</a>
{{ end }}
`

type Aside struct {
	Content []Renderable
}

var _ CompositeRenderable = (*Aside)(nil)

func (a *Aside) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Aside", a)
	return template.HTML(buf.String()), err
}

func (a *Aside) Append(child Renderable) {
	a.Content = append(a.Content, child)
}

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

var _ Renderable = (*Comment)(nil)

func (c Comment) Render() (template.HTML, error) {
	// ExecuteTemplate would remove the entire thing from the output; don't call it
	return template.HTML("<!-- " + string(c) + " -->"), nil
}

type Sidenote struct {
	ID string
	ShortText string
	Expanded []Renderable
}

var _ CompositeRenderable = (*Sidenote)(nil)

func NewSidenote(short string) *Sidenote {
	return &Sidenote{
		ID: GenerateID("sn"),
		ShortText: short,
	}
}

func (s *Sidenote) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Sidenote", s)
	return template.HTML(buf.String()), err
}

func (s *Sidenote) Append(child Renderable) {
	s.Expanded = append(s.Expanded, child)
}

func (s *Sidenote) ExpandedTextOnly() string {
	text := ""
	for _, r := range s.Expanded {
		if str, ok := r.(TextRenderable); ok {
			text += str.Text()
		}
	}
	return text
}

// Adapted @from: https://github.com/kslstn/sidenotes
const HtmlSidenote = `
{{ define "Sidenote" }}<span class="sidenote">
	<input type="checkbox"
		   id="sidenote__checkbox--{{.ID}}"
		   class="sidenote__checkbox"
		   aria-label="show sidenote" />
	<label for="sidenote__checkbox--{{.ID}}"
		   aria-describedby="sidenote-{{.ID}}"
		   title="{{.ExpandedTextOnly}}"
		   class="sidenote__button">{{.ShortText}}
	</label>
	<small id="sidenote-{{.ID}}"
		   class="sidenote__content">
		<span class="sidenote__content-parenthesis">(sidenote:</span>
		{{ range .Expanded }}
		{{ Render . }}
		{{ end }}
		<span class="sidenote__content-parenthesis">)</span>
	</small>
</span>{{ end }}
`

type CodeLine string

type CodeBlock struct {
	Lines []CodeLine
}

var _ Renderable = (*CodeBlock)(nil)

func (c CodeBlock) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "CodeBlock", c)
	return template.HTML(buf.String()), err
}

const HtmlCodeBlock = `
{{ define "CodeBlock" }}
<pre><code>
{{ range .Lines }} <span class="line-number">{{ . }}</span> {{ end }}
</pre></code>
{{ end }}
`

//`<span class="comment">{{ .Comment }}</span>`

type Enquote string

var _ TextRenderable = (*Enquote)(nil)

func (e Enquote) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Enquote", e)
	return template.HTML(buf.String()), err
}

func (e Enquote) Text() string {
	return string(e)
}

const HtmlEnquote = `
{{ define "Enquote" }}
<q>{{ . }}</q>
{{ end}}
`

type Mono string

var _ TextRenderable = (*Mono)(nil)

func (m Mono) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Mono", m)
	return template.HTML(buf.String()), err
}

func (m Mono) Text() string {
	return string(m)
}

const HtmlMono = `
{{ define "Mono" }}
<code>{{ . }}</code>
{{ end }}
`

type Em string

var _ TextRenderable = (*Em)(nil)

func (e Em) Render() (template.HTML, error) {
	buf := &bytes.Buffer{}
	err := pages.Execute(buf, "Em", e)
	return template.HTML(buf.String()), err
}

func (e Em) Text() string {
	return string(e)
}

const HtmlEm = `
{{ define "Em" }}
<em>{{ . }}</em>
{{ end }}
`
