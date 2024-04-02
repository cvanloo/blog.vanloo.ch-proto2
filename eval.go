package be

import (
	"fmt"
	"log"
	"strings"
	"time"

	"be/lex"
	. "be/internal/debug"
)

const (
	TypeForm = lex.TypeForm
	TypeAtom = lex.TypeAtom
	TypeText = lex.TypeText
	TypeAny = TypeForm | TypeAtom | TypeText
)

type (
	Head = lex.Head
	Node = lex.Node
)

type (
	FunMap map[string]beFun
	Context struct {
		Parent CompositeRenderable
	}
	Scope struct {
		funs FunMap
		Context *Context
	}
	Scopes struct {
		scopes []Scope
	}
	Args struct {
		finished bool
		next *Node
	}
	beFun func(blog *Blog, scopes *Scopes, args *Args) error
)

func NewScope(parent CompositeRenderable) Scope {
	return Scope{
		funs: FunMap{},
		Context: &Context{
			Parent: parent,
		},
	}
}

func InitScopes(blog *Blog) *Scopes {
	return &Scopes{
		scopes: []Scope{
			Scope{
				funs: rootFuns,
				Context: &Context{
					Parent: blog,
				},
			},
		},
	}
}

func (s *Scopes) Push(scope Scope) {
	s.scopes = append(s.scopes, scope)
}

func (s *Scopes) Pop() {
	s.scopes = s.scopes[:len(s.scopes)-1]
}

func (s *Scopes) Top() Scope {
	return s.scopes[len(s.scopes)-1]
}

func (s *Scopes) RegisterFun(name string, fun beFun) {
	s.Top().funs[name] = fun
}

func (s *Scopes) Resolve(name string) (beFun, error) {
	for i := len(s.scopes)-1; i >= 0; i-- {
		if fun, ok := s.scopes[i].funs[name]; ok {
			return fun, nil
		}
	}
	return nil, fmt.Errorf("function not in scope: %s", name)
}

func (s *Scopes) Parent() CompositeRenderable {
	return s.Top().Context.Parent
}

func NewArgs(node *Node) *Args {
	return &Args{
		next: node,
	}
}

func (a *Args) Next(name string, type_ lex.FormType) (*Node, error) {
	Assert(!a.finished, "all mandatory arguments must appear before optional ones")
	if a.next == nil {
		return nil, fmt.Errorf("missing argument: %s", name)
	}
	arg := a.next
	a.next = a.next.Next
	if (arg.Type & type_) == 0 {
		return arg, fmt.Errorf("argument of incorrect type, want: %+v, got: %+v", type_, arg)
	}
	return arg, nil
}

func (a *Args) Optional(name string, type_ lex.FormType) (*Node, error) {
	a.finished = true
	if a.next == nil {
		return nil, nil
	}
	arg := a.next
	a.next = a.next.Next
	if (arg.Type & type_) == 0 {
		return arg, fmt.Errorf("argument of incorrect type, want: %+v, got: %+v", type_, arg)
	}
	return arg, nil
}

func (a *Args) IsFinished() bool {
	return a.next == nil
}

func (a *Args) Finished() error {
	if !a.IsFinished() {
		return fmt.Errorf("superfluous arguments: %+v", a.next)
	}
	return nil
}

var rootFuns = FunMap {
	"root": func(blog *Blog, scopes *Scopes, args *Args) error {
		// @todo: read defaults from config file?
		blog.BlogName = "save-lisp-and-die"
		blog.Author = Author{}
		blog.Author.Name = "cvl"

		for !args.IsFinished() {
			content, err := args.Optional("root content", TypeForm)
			if err != nil {
				return fmt.Errorf("root: %w", err)
			}
			if err := blog.Apply(blog, scopes, content.Form.First); err != nil {
				return err
			}
		}
		return args.Finished()
	},
	"eof": func(blog *Blog, scopes *Scopes, args *Args) error {
		// @todo: fill in blog.Meta?
		blog.Meta = Meta{
			Language: "en",
			//CanonicalURL string
			//Description string
			Published: time.Now(),
			//Revisions []time.Time
			//Topic string
			//EstReadingTime ReadingTime
		}
		return args.Finished()
	},
	"html-comment": func(blog *Blog, scopes *Scopes, args *Args) error {
		content, err := args.Optional("html-comment text", TypeText)
		if err != nil {
			return fmt.Errorf("html-comment: %w", err)
		}
		comment := Comment(content.Text)
		scopes.Parent().Append(comment)
		return args.Finished()
	},
	"comment": func(blog *Blog, scopes *Scopes, args *Args) error {
		for !args.IsFinished() {
			content, err := args.Optional("comment content (will be ignored)", TypeAny)
			if err != nil {
				return fmt.Errorf("comment: %w", err)
			}
			_ = content
		}
		return args.Finished()
	},
	"title": func(blog *Blog, scopes *Scopes, args *Args) error {
		title, err := args.Next("title", TypeText)
		if err != nil {
			return err
		}
		blog.Title = string(title.Text)
		altTitle, err := args.Optional("alternative title", TypeText)
		if err != nil {
			return err
		}
		if altTitle != nil {
			blog.AltTitle = string(altTitle.Text)
		}
		return args.Finished()
	},
	"author": func(blog *Blog, scopes *Scopes, args *Args) error {
		blog.Author = Author{} // ensure author is initialized and zeroed
		scopes.RegisterFun("name", func(blog *Blog, scopes *Scopes, args *Args) error {
			name, err := args.Next("author name", TypeText)
			if err != nil {
				return fmt.Errorf("author-name: %w", err)
			}
			blog.Author.Name = string(name.Text)
			return args.Finished()
		})
		scopes.RegisterFun("email", func(blog *Blog, scopes *Scopes, args *Args) error {
			email, err := args.Next("author email", TypeText)
			if err != nil {
				return fmt.Errorf("author-email: %w", err)
			}
			blog.Author.EMail = string(email.Text)
			return args.Finished()
		})
		for _ = range len(scopes.Top().funs) {
			nextArgs, err := args.Optional("author args", TypeForm)
			if err != nil {
				return fmt.Errorf("author: %w", err)
			}
			err = blog.Apply(scopes.Parent(), scopes, nextArgs.Form.First)
			if err != nil {
				return err
			}
		}
		return args.Finished()
	},
	"tags": func(blog *Blog, scopes *Scopes, args *Args) error {
		if len(blog.Tags) > 0 {
			log.Printf("tags: already set, overwriting")
		}
		tags := Tags{}
		for !args.IsFinished() {
			tagList, err := args.Optional("space separated tag list", TypeText)
			if err != nil {
				return fmt.Errorf("tags: %w", err)
			}
			for _, tagStr := range strings.Split(string(tagList.Text), " ") {
				tags = append(tags, Tag(tagStr))
			}
		}
		blog.Tags = tags
		return args.Finished()
	},
	"body": func(blog *Blog, scopes *Scopes, args *Args) error {
		for !args.IsFinished() {
			content, err := args.Optional("body content", TypeAny)
			if err != nil {
				return fmt.Errorf("body: %w", err)
			}
			err = blog.Apply(blog, scopes, content)
			if err != nil {
				return fmt.Errorf("body: %w", err)
			}
		}
		return args.Finished()
	},
	"paragraph": func(blog *Blog, scopes *Scopes, args *Args) error {
		pg := &Paragraph{}
		scopes.Parent().Append(pg)
		for !args.IsFinished() {
			content, err := args.Optional("paragraph content", TypeAny)
			if err != nil {
				return fmt.Errorf("paragraph: %w", err)
			}
			err = blog.Apply(pg, scopes, content)
			if err != nil {
				return fmt.Errorf("paragraph: %w", err)
			}
		}
		return args.Finished()
	},
	"section": func(blog *Blog, scopes *Scopes, args *Args) error {
		scopes.RegisterFun("subsection", func(blog *Blog, scopes *Scopes, args *Args) error {
			heading, err := args.Next("subsection heading", TypeText)
			if err != nil {
				return fmt.Errorf("subsection: %w", err)
			}
			subsection := NewSubsection(string(heading.Text))
			scopes.Parent().Append(subsection)
			for !args.IsFinished() {
				content, err := args.Optional("subsection content", TypeAny)
				if err != nil {
					return fmt.Errorf("subsection: %w", err)
				}
				err = blog.Apply(subsection, scopes, content)
				if err != nil {
					return fmt.Errorf("subsection: %w", err)
				}
			}
			return args.Finished()
		})
		heading, err := args.Next("section heading", TypeText)
		if err != nil {
			return fmt.Errorf("section: %w", err)
		}
		section := NewSection(string(heading.Text))
		scopes.Parent().Append(section)
		for !args.IsFinished() {
			content, err := args.Optional("section content", TypeAny)
			if err != nil {
				return fmt.Errorf("section: %w", err)
			}
			err = blog.Apply(section, scopes, content)
			if err != nil {
				return fmt.Errorf("section: %w", err)
			}
		}
		return args.Finished()
	},
	"abstract": func(blog *Blog, scopes *Scopes, args *Args) error {
		// @fixme: implement correctly
		for !args.IsFinished() {
			content, err := args.Optional("abstract content", TypeAny)
			if err != nil {
				return fmt.Errorf("abstract: %w", err)
			}
			_ = content
		}
		return args.Finished()
	},
	"enquote": func(blog *Blog, scopes *Scopes, args *Args) error {
		text, err := args.Next("enquote text", TypeText)
		if err != nil {
			return fmt.Errorf("enquote: %w", err)
		}
		// @fixme: implement correctly
		scopes.Parent().Append(Text(text.Text))
		return args.Finished()
	},
	"sidenote": func(blog *Blog, scopes *Scopes, args *Args) error {
		short, err := args.Next("sidenote short text", TypeText)
		if err != nil {
			return fmt.Errorf("sidenote: %w", err)
		}
		full, err := args.Next("sidenote content", TypeText)
		if err != nil {
			return fmt.Errorf("sidenote: %w", err)
		}
		sidenote := NewSidenote(string(short.Text), string(full.Text))
		scopes.Parent().Append(sidenote)
		return args.Finished()
	},
	"mono": func(blog *Blog, scopes *Scopes, args *Args) error {
		text, err := args.Next("monospace text", TypeText)
		if err != nil {
			return fmt.Errorf("mono: %w", err)
		}
		// @fixme: implement correctly
		scopes.Parent().Append(Text(text.Text))
		return args.Finished()
	},
	"code": func(blog *Blog, scopes *Scopes, args *Args) error {
		text, err := args.Next("code text", TypeText)
		if err != nil {
			return fmt.Errorf("code: %w", err)
		}
		var code CodeBlock
		for _, line := range strings.Split(string(text.Text), "\n") {
			code.Lines = append(code.Lines, CodeLine(line))
		}
		scopes.Parent().Append(code)
		return args.Finished()
	},
}

func (blog *Blog) Eval(scopes *Scopes, el *Node) error {
	switch el.Type {
	case TypeAtom:
		fun, err := scopes.Resolve(string(el.Atom))
		if err != nil {
			return err
		}
		return fun(blog, scopes, NewArgs(el.Next))
	case TypeForm:
		Unreachable()
	case TypeText:
		scopes.Parent().Append(Text(el.Text))
	default:
		Unreachable()
	}
	return nil
}

func (blog *Blog) Apply(parent CompositeRenderable, scopes *Scopes, node *Node) (err error) {
	scopes.Push(NewScope(parent))
	defer scopes.Pop()
	switch node.Type {
	case TypeForm:
		err = blog.Eval(scopes, node.Form.First)
	case TypeText: fallthrough
	case TypeAtom:
		err = blog.Eval(scopes, node)
	}
	return err
}
