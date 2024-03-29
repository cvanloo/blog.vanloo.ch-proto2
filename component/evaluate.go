package component

import (
	"bytes"
	//"errors"
	//"fmt"
	"html/template"
	"io"
	//"log"
	//"net/http"
	//"strings"

	//"be/lex"
)

var pages Template = Template{template.New("")}

func init() {
	pages.Funcs(template.FuncMap{
		"Render": Render,
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

func String(blog *EntryData) string {
	bs := &bytes.Buffer{}
	err := pages.Render(bs, "Entry", blog)
	if err != nil {
		panic(err)
	}
	return bs.String()
}

/*
func Handler(root *lex.LLHead) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := eval(nil, nil, root.First)
		if err != nil {
			panic(err)
		}

		err = pages.Render(w, "Entry", data)
		if err != nil {
			panic(err)
		}
	}
}
*/

func Render(element ContentElement) (template.HTML, error) {
	return element.Render()
}
