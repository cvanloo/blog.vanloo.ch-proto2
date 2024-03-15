package component

type CodeLine string

type CodeBlock struct {
	Lines []CodeLine
}

const HtmlCodeBlock = `
{{ define "CodeBlock" }}
<pre><code>
{{ range .Lines }}
<span class="line-number">{{ .Line }}</span>
{{ end }}
</pre></code>
{{ end }}
`

//`<span class="comment">{{ .Comment }}</span>`
