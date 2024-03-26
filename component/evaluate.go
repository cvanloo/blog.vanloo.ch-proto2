package component

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"maps"
	"net/http"
	"strings"

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
	data, err := eval(nil, nil, root)
	if err != nil {
		panic(err)
	}
	name := "Entry"

	bs := &bytes.Buffer{}
	err = pages.Render(bs, name, data)
	if err != nil {
		panic(err)
	}
	return bs.String()
}

func Handler(root *lex.LLHead) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := eval(nil, nil, root)
		if err != nil {
			panic(err)
		}
		name := "Entry"

		err = pages.Render(w, name, data)
		if err != nil {
			panic(err)
		}
	}
}

func Evaluate(root *lex.LLHead) (template.HTML, error) {
	buf := bytes.NewBuffer([]byte{})

	data, err := eval(nil, nil, root)
	if err != nil {
		var zero template.HTML
		return zero, err
	}
	name := "Entry"

	err = pages.Render(buf, name, data)
	html := template.HTML(buf.String())
	return html, err
}

type (
	BeFunc func(blog *EntryData, scopes *Scopes, args Args) error
	Scope map[string]BeFunc
	Scopes struct {
		scopes []Scope
	}
	Args struct {
		next *lex.LLNode
		finished bool
		errs []error
	}
)

func NewArgs(args *lex.LLNode) Args {
	return Args{
		next: args,
	}
}

func (a Args) Next(name string) string {
	if a.finished {
		panic("invalid usage: all mandatory arguments must appear before optional ones")
	}
	if a.next == nil {
		a.errs = append(a.errs, fmt.Errorf("missing argument: %s", name))
		return "<value missing>"
	}
	n := a.next.El
	a.next = a.next.Next

	if n.Type != lex.TypeText {
		panic("arg must be of type text")
	}
	return string(n.Text)
}

func (a Args) Optional(name string) string {
	a.finished = true
	if a.next == nil {
		return ""
	}
	n := a.next.El
	a.next = a.next.Next

	if n.Type != lex.TypeText {
		panic("arg must be of type text")
	}
	return string(n.Text)
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
	"root": func(blog *EntryData, scopes *Scopes, args Args) error {
		blog.BlogName = "save-lisp-and-die"
		blog.Author = Author{
			Name: "cvl",
		}

		return args.Finished()
	},
	"eof": func(blog *EntryData, scopes *Scopes, args Args) error {
		return args.Finished()
	},
	"title": func(blog *EntryData, scopes *Scopes, args Args) error {
		blog.Title = args.Next("title")
		blog.AltTitle = args.Optional("alternative title")
		return args.Finished()
	},
	"author": func(blog *EntryData, scopes *Scopes, args Args) error {
		blog.Author = Author{}
		scopes.AddToTopScope(Scope{
			"name": func(blog *EntryData, scopes *Scopes, args Args) error {
				blog.Author.Name = args.Next("author name")
				return args.Finished()
			},
			"email": func(blog *EntryData, scopes *Scopes, args Args) error {
				blog.Author.EMail = args.Next("author email")
				return args.Finished()
			},
		})
		return args.Finished()
	},
	"tags": func(blog *EntryData, scopes *Scopes, args Args) error {
		tagStrs := strings.Split(args.Next("space separated tag list"), " ")
		blog.Tags = make(Tags, len(tagStrs))
		for i, t := range tagStrs {
			blog.Tags[i] = Tag(t)
		}
		return args.Finished()
	},
}

func eval(blog *EntryData, scopes *Scopes, head *lex.LLHead) (nblog *EntryData, err error) {
	if blog == nil {
		blog = &EntryData{}
	}
	if scopes == nil {
		scopes = &Scopes{}
		scopes.Push(beFuncs)
	}
	var fun BeFunc
	//args := []string{}
	for c := head.First; c != nil; c = c.Next {
		n := c.El
		switch n.Type {
		case lex.TypeForm: // evaluate recursively
			scopes.Push(Scope{})
			blog, err = eval(blog, scopes, n.Form)
			scopes.Pop()
			if err != nil {
				return blog, err
			}
		case lex.TypeAtom:
			fun, err = scopes.Resolve(string(n.Atom))
			if err != nil {
				return blog, err
			}
			err = fun(blog, scopes, NewArgs(c.Next))
		case lex.TypeText:
			log.Printf("unhandled: %#v", n)
			//panic("type text invalid in this position")
			//args = append(args, string(n.Text))
		default:
			panic(fmt.Errorf("unknown node type: %#v", n))
		}
	}
	//err = fun(scopes, NewArgs(args...))
	//return err
	return blog, nil
}
