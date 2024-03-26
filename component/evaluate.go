package component

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	//"log"
	"net/http"
	"strings"

	"be/lex"
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

func String(root *lex.LLHead) string {
	data, err := eval(nil, nil, root)
	if err != nil {
		panic(err)
	}

	bs := &bytes.Buffer{}
	err = pages.Render(bs, "Entry", data)
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

		err = pages.Render(w, "Entry", data)
		if err != nil {
			panic(err)
		}
	}
}

func Render(element ContentElement) (template.HTML, error) {
	return element.Render()
}

type (
	BeFunc func(blog *EntryData, scope Scope, args *Args) error
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

func NewArgs(args *lex.LLNode) *Args {
	return &Args{
		next: args,
	}
}

func (a *Args) Next(name string) string {
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

func (a *Args) Optional(name string) string {
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

func (a *Args) Finished() error {
	return errors.Join(a.errs...)
}

func (sc *Scopes) Push(scope Scope) {
	sc.scopes = append(sc.scopes, scope)
}

func (sc *Scopes) Top() Scope {
	return sc.scopes[len(sc.scopes)-1]
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
	"root": func(blog *EntryData, scope Scope, args *Args) error {
		// @todo: read defaults from config file?
		blog.BlogName = "save-lisp-and-die"
		blog.Author = Author{
			Name: "cvl",
		}
		return args.Finished()
	},
	"eof": func(blog *EntryData, scope Scope, args *Args) error {
		// @todo: fill in blog.Meta ?
		return args.Finished()
	},
	"title": func(blog *EntryData, scope Scope, args *Args) error {
		blog.Title = args.Next("title")
		blog.AltTitle = args.Optional("alternative title")
		return args.Finished()
	},
	"author": func(blog *EntryData, scope Scope, args *Args) error {
		blog.Author = Author{}
		scope["name"] = func(blog *EntryData, scope Scope, args *Args) error {
			blog.Author.Name = args.Next("author name")
			return args.Finished()
		}
		scope["email"] = func(blog *EntryData, scope Scope, args *Args) error {
			blog.Author.EMail = args.Next("author email")
			return args.Finished()
		}
		return args.Finished()
	},
	"tags": func(blog *EntryData, scope Scope, args *Args) error {
		tagStrs := strings.Split(args.Next("space separated tag list"), " ")
		blog.Tags = make(Tags, len(tagStrs))
		for i, t := range tagStrs {
			blog.Tags[i] = Tag(t)
		}
		return args.Finished()
	},
	"body": func(blog *EntryData, scope Scope, args *Args) error {
		for a := args.Next("content"); a != ""; a = args.Optional("additional content") {
			blog.Content = append(blog.Content, Text(a))
		}
		return args.Finished()
	},
}

/*
func eval(blog *EntryData, scopes *Scopes, node *lex.LLNode) (*EntryData, error) {
	el := nodes.El
	switch el.Type {
	case lex.TypeAtom:
		fun, err := scopes.Resolve(string(el.Atom))
		if err != nil {
			return blog, err
		}
		err = fun(blog, scopes.Top(), el.Next)
		return blog, err
	case lex.TypeForm:
	case lex.TypeText:
	default:
		panic(fmt.Errorf("unknown node type: %#v", n))
	}
}
*/

func eval(blog *EntryData, scopes *Scopes, head *lex.LLHead) (nblog *EntryData, err error) {
	if blog == nil {
		blog = &EntryData{}
	}
	if scopes == nil {
		scopes = &Scopes{}
		scopes.Push(beFuncs)
	}
	var fun BeFunc
	for c := head.First; c != nil; c = c.Next {
		n := c.El
		switch n.Type {
		case lex.TypeForm:
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
			err = fun(blog, scopes.Top(), NewArgs(c.Next))
			if err != nil {
				return blog, err
			}
		case lex.TypeText:
			//blog.Content = append(blog.Content, Text(n.Text))
		default:
			panic(fmt.Errorf("unknown node type: %#v", n))
		}
	}
	return blog, nil
}
