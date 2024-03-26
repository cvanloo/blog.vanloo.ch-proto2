package component

import (
	"fmt"
	"time"

	//"be/lex"
)

type Tag string

func (t Tag) String() string {
	return ":" + string(t)
}

type Tags []Tag

func (ts Tags) KeywordList() (s string) {
	if ts == nil {
		return ""
	}
	s = string(ts[0])
	if len(ts) > 1 {
		for _, t := range ts[1:] {
			s += ", " + string(t)
		}
	}
	return
}

type ReadingTime struct {
	time.Duration
}

func (rt ReadingTime) String() string {
	return fmt.Sprintf("~%s", rt.Duration.String())
}

type Meta struct {
	// https://en.wikipedia.org/wiki/List_of_ISO_639_language_codes
	Language string
	CanonicalURL string
	Description string
	Published time.Time
	Revisions []time.Time
	Topic string
	EstReadingTime ReadingTime
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

type Author struct {
	Name string
	EMail string
}

type Language struct {
	Link string
	Language string
}

type EntryData struct {
	BlogName string
	Title, AltTitle string
	Author Author
	Tags Tags
	Meta Meta
	Abstract string
	Languages []Language
	Content []ContentElement
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
		<meta property="article:published_time" content="{{.Meta.Published}}" />
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
							<p class="published-date"><small>{{.Meta.Published}}</small></p>
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
			<address>&copy; {{.Meta.CopyYear}} <a href="mailto:{{.Author.EMail}}?subject=RE: {{.Title}}">{{.Author.Name}}</a></address>
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
