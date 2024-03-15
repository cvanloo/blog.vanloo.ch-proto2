package component

type CodeLine string

type CodeBlock struct {
	Lines []CodeLine
}

const HtmlCodeBlock = `
<pre><code>
{{ range .Lines }}
<span class="line-number">{{ .Line }}</span>
{{ end }}
</pre></code>
`

//`<span class="comment">{{ .Comment }}</span>`
