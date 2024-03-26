package component

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
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

func String(root *lex.LLHead) string {
	scopes := &Scopes{}
	scopes.Push(beFuncs)
	/*blog, */err := eval(scopes, root)
	name := "Entry"
	data := blog
	if err != nil {
		panic(err)
	}

	bs := &bytes.Buffer{}
	err = pages.Render(bs, name, data)
	if err != nil {
		panic(err)
	}
	return bs.String()
}

func Handler(root *lex.LLHead) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scopes := &Scopes{}
		scopes.Push(beFuncs)
		/*name, data, */err := eval(scopes, root)
		name := "Entry"
		data := blog
		if err != nil {
			panic(err)
		}

		err = pages.Render(w, name, data)
		if err != nil {
			panic(err)
		}
	}
}

func Evaluate(root *lex.LLHead) (template.HTML, error) {
	buf := bytes.NewBuffer([]byte{})

	scopes := &Scopes{}
	scopes.Push(beFuncs)
	/*name, data, */err := eval(scopes, root)
	name := "Entry"
	data := blog

	if err != nil {
		var zero template.HTML
		return zero, err
	}
	err = pages.Render(buf, name, data)
	html := template.HTML(buf.String())
	return html, err
}

var blog = EntryData{
	// set defaults (@todo: read in from config file?)
	BlogName: "save-lisp-and-die",
	Author: Author{
		Name: "cvl",
		EMail: "",
	},
}

type (
	BeFunc func(scopes *Scopes, args Args) error
	Scope map[string]BeFunc
	Scopes struct {
		scopes []Scope
	}
	Args struct {
		args []string
		next int
		finished bool
		errs []error
	}
)

func NewArgs(args ...string) Args {
	return Args{
		args: args,
	}
}

func (a Args) Next(name string) string {
	if a.finished {
		panic("invalid usage: all mandatory arguments must appear before optional ones")
	}
	if a.next < len(a.args) {
		i := a.next
		a.next++
		return a.args[i]
	}
	a.errs = append(a.errs, fmt.Errorf("missing argument: %s", name))
	return "<value missing>"
}

func (a Args) Optional(name string) string {
	a.finished = true
	if a.next < len(a.args) {
		i := a.next
		a.next++
		return a.args[i]
	}
	return ""
}

func (a Args) Finished() error {
	return errors.Join(a.errs...)
}

func (sc *Scopes) Push(scope Scope) {
	sc.scopes = append(sc.scopes, scope)
}

func (sc *Scopes) AddToTopScope(scope Scope) {
	maps.Copy(sc.scopes[len(sc.scopes)-1], scope)
}

func (sc *Scopes) Pop() {
	sc.scopes = sc.scopes[:len(sc.scopes)-1]
}

func (sc *Scopes) Resolve(name string) (fun BeFunc, err error) {
	for i := len(sc.scopes)-1; i >= 0; i-- {
		if fun, ok := sc.scopes[i][name]; ok {
			return fun, nil
		}
	}
	return nil, fmt.Errorf("cannot resolve function: %s", name)
}

var beFuncs = Scope {
	"root": func(scopes *Scopes, args Args) error {
		// @todo: default initialize blog based from config values here?
		return args.Finished()
	},
	"title": func(scopes *Scopes, args Args) error {
		blog.Title = args.Next("title")
		blog.AltTitle = args.Optional("alternative title")
		return args.Finished()
	},
	"author": func(scopes *Scopes, args Args) error {
		blog.Author = Author{}
		scopes.AddToTopScope(Scope{
			"name": func(scopes *Scopes, args Args) error {
				blog.Author.Name = args.Next("author name")
				return args.Finished()
			},
			"email": func(scopes *Scopes, args Args) error {
				blog.Author.EMail = args.Next("author email")
				return args.Finished()
			},
		})
		return args.Finished()
	},
}

// scopes := Scopes{beFuncs}
func eval(scopes *Scopes, head *lex.LLHead) (err error) {
	var fun BeFunc
	args := []string{}
	for c := head.First; c != nil; c = c.Next {
		n := c.El
		switch n.Type {
		case lex.TypeForm: // evaluate recursively
			scopes.Push(Scope{})
			err = eval(scopes, n.Form)
			scopes.Pop()
			if err != nil {
				return err
			}
		case lex.TypeAtom:
			fun, err = scopes.Resolve(string(n.Atom))
			if err != nil {
				return err
			}
		case lex.TypeText:
			args = append(args, string(n.Text))
		default:
			panic(fmt.Errorf("unknown node type: %#v", n))
		}
	}
	err = fun(scopes, NewArgs(args...))
	return err
}
