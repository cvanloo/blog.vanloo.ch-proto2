package component

type Sidenote struct {
	ID int
	ShortText, ExpandedText string
}

const HtmlSidenote = `
{{ define "Sidenote" }}
<span class="sidenote">
	<input type="checkbox"
		   id="sidenote__checkbox--{{.ID}}"
		   class="sidenote__checkbox"
		   aria-label="show sidenote" />
	<label for="sidenote__checkbox--{{.ID}}"
		   aria-describedby="sidenote-{{.ID}}"
		   title="{{.ExpandedText}}"
		   class="sidenote__button">{{.ShortText}}
	</label>
	<small id="sidenote-{{.ID}}"
		   class="sidenote__content">
		<span class="sidenote__content-parenthesis">(sidenote:</span>
		{{.ExpandedText}}
		<span class="sidenote__content-parenthesis">)</span>
	</small>
</span>
{{ end }}
`
