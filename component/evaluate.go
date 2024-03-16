package component

import (
	"bytes"
	"html/template"
	"io"
	"log"
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

func Handler(root *lex.LLNode) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name, data, err := eval(root)
		if err != nil {
			panic(err)
		}

		err = pages.Render(w, name, data)
		if err != nil {
			panic(err)
		}
	}
}

func Evaluate(root *lex.LLNode) (template.HTML, error) {
	buf := bytes.NewBuffer([]byte{})
	name, data, err := eval(root)
	if err != nil {
		var zero template.HTML
		return zero, err
	}
	err = pages.Render(buf, name, data)
	html := template.HTML(buf.String())
	return html, err
}

func eval(node *lex.LLNode) (name string, data any, err error) {
	// @todo: implement
	log.Printf("eval start ---:\n%s\n--- eval end\n", node)

	el := node.El
	switch el.Type {
	case lex.TypeForm:
		fun := eval(node.Next)
	case lex.TypeAtom:
		fun := funcMap[el.Atom]
	case lex.TypeText:
	}
	return "Paragraph", node.String(), nil
}
