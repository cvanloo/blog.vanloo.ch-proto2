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
	LLHead = lex.LLHead
	LLNode = lex.LLNode
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
		next *LLNode
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

func NewArgs(node *LLNode) *Args {
	return &Args{
		next: node,
	}
}

func (a *Args) Next(name string, type_ lex.FormType) (*LLNode, error) {
	Assert(!a.finished, "all mandatory arguments must appear before optional ones")
	if a.next == nil {
		return nil, fmt.Errorf("missing argument: %s", name)
	}
	arg := a.next
	a.next = a.next.Next
	if (arg.El.Type & type_) == 0 {
		return arg, fmt.Errorf("argument of incorrect type, want: %+v, got: %+v", type_, arg.El)
	}
	return arg, nil
}

func (a *Args) Optional(name string, type_ lex.FormType) (*LLNode, error) {
	a.finished = true
	if a.next == nil {
		return nil, nil
	}
	arg := a.next
	a.next = a.next.Next
	if (arg.El.Type & type_) == 0 {
		return arg, fmt.Errorf("argument of incorrect type, want: %+v, got: %+v", type_, arg.El)
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
			if err := blog.Apply(blog, scopes, content.El.Form.First); err != nil {
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
		comment := Comment(content.El.Text)
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
		blog.Title = string(title.El.Text)
		altTitle, err := args.Optional("alternative title", TypeText)
		if err != nil {
			return err
		}
		if altTitle != nil {
			blog.AltTitle = string(altTitle.El.Text)
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
			blog.Author.Name = string(name.El.Text)
			return args.Finished()
		})
		scopes.RegisterFun("email", func(blog *Blog, scopes *Scopes, args *Args) error {
			email, err := args.Next("author email", TypeText)
			if err != nil {
				return fmt.Errorf("author-email: %w", err)
			}
			blog.Author.EMail = string(email.El.Text)
			return args.Finished()
		})
		for _ = range len(scopes.Top().funs) {
			nextArgs, err := args.Optional("author args", TypeForm)
			if err != nil {
				return fmt.Errorf("author: %w", err)
			}
			err = blog.Apply(scopes.Parent(), scopes, nextArgs.El.Form.First)
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
			for _, tagStr := range strings.Split(string(tagList.El.Text), " ") {
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
	"section": func(blog *Blog, scopes *Scopes, args *Args) error {
		scopes.RegisterFun("subsection", func(blog *Blog, scopes *Scopes, args *Args) error {
			heading, err := args.Next("subsection heading", TypeText)
			if err != nil {
				return fmt.Errorf("subsection: %w", err)
			}
			subsection := NewSubsection(string(heading.El.Text))
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
		section := NewSection(string(heading.El.Text))
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
}

func (blog *Blog) Eval(scopes *Scopes, node *LLNode) error {
	el := node.El
	switch el.Type {
	case TypeAtom:
		fun, err := scopes.Resolve(string(el.Atom))
		if err != nil {
			return err
		}
		return fun(blog, scopes, NewArgs(node.Next))
	case TypeForm:
		Assert(false, "unreachable")
		//err := blog.Apply(scopes.Parent(), scopes, el.Form.First)
		//if err != nil {
		//	return err
		//}
	case TypeText:
		scopes.Parent().Append(Text(el.Text))
	default:
		Unreachable()
	}
	return nil
}

func (blog *Blog) Apply(parent CompositeRenderable, scopes *Scopes, node *LLNode) (err error) {
	scopes.Push(NewScope(parent))
	defer scopes.Pop()
	switch node.El.Type {
	case TypeForm:
		err = blog.Eval(scopes, node.El.Form.First)
	case TypeText: fallthrough
	case TypeAtom:
		err = blog.Eval(scopes, node)
	}
	return err
}
