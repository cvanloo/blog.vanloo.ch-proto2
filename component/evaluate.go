package component

import (
	"bytes"
	"html/template"
	"io"
	"net/http"

	"be/lex"
)

var pages Template = Template{template.New("")}

func init() {
	pages.Funcs(template.FuncMap{
		"Evaluate": Evaluate,
	})

	template.Must(pages.Parse(HtmlCodeBlock))
	template.Must(pages.Parse(HtmlEntry))
	template.Must(pages.Parse(HtmlSection))
	template.Must(pages.Parse(HtmlSubsection))
	template.Must(pages.Parse(HtmlParagraph))
	template.Must(pages.Parse(HtmlLink))
	template.Must(pages.Parse(HtmlAside))
	template.Must(pages.Parse(HtmlSidenote))
}

type Template struct {
	*template.Template
}

func (t *Template) Render(w io.Writer, name string, data any) error {
	return t.Template.ExecuteTemplate(w, name, data)
}

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := pages.Render(w, "Entry", nil)
		if err != nil {
			panic(err)
		}
	}
}

func Evaluate(root *lex.Node) (html template.HTML, err error) {
	buf := bytes.NewBuffer([]byte{})
	name, data := eval(root)
	err = pages.Render(buf, name, data)
	html = template.HTML(buf.String())
	return
}

func eval(node *lex.Node) (name string, data any) {
	return
}
