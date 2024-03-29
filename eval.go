package main

import (
	"fmt"
	"strings"
	"log"

	"be/lex"
	//"be/tok"
	"be/component"
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
	Blog = component.EntryData
)

type (
	Scope map[string]beFun
	Scopes struct {
		scopes []Scope
	}
	Arg = Node
	Args struct {
		finished bool
		next *LLNode
	}
	beFun func(blog *Blog, scopes *Scopes, args *Args) error
)

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
	s.Top()[name] = fun
}

func (s *Scopes) Resolve(name string) (beFun, error) {
	for i := len(s.scopes)-1; i >= 0; i-- {
		if fun, ok := s.scopes[i][name]; ok {
			return fun, nil
		}
	}
	return nil, fmt.Errorf("function not in scope: %s", name)
}

func NewArgs(node *LLNode) *Args {
    return &Args{
        next: node,
    }
}

func (a *Args) Next(name string, type_ lex.FormType) (*Arg, error) {
	assert(!a.finished, "all mandatory arguments must appear before optional ones")
	if a.next == nil {
		return nil, fmt.Errorf("missing argument: %s", name)
	}
	arg := a.next.El
	a.next = a.next.Next
	if (arg.Type & type_) == 0 {
		return arg, fmt.Errorf("argument of incorrect type, want: %#v, got: %#v", type_, arg)
	}
	return arg, nil
}

func (a *Args) Optional(name string, type_ lex.FormType) (*Arg, error) {
	a.finished = true
	if a.next == nil {
		return nil, nil
	}
	arg := a.next.El
	a.next = a.next.Next
	if (arg.Type & type_) == 0 {
		return arg, fmt.Errorf("argument of incorrect type, want: %#v, got: %#v", type_, arg)
	}
	return arg, nil
}

func (a *Args) IsFinished() bool {
	return a.next == nil
}

func (a *Args) Finished() error {
	if !a.IsFinished() {
		return fmt.Errorf("superfluous arguments: %#v", a.next)
	}
	return nil
}

var rootFuns = Scope {
	"root": func(blog *Blog, scopes *Scopes, args *Args) error {
		// @todo: read defaults from config file?
		blog.BlogName = "save-lisp-and-die"
		blog.Author = component.Author{}
		blog.Author.Name = "cvl"

		for !args.IsFinished() {
			nextArgs, err := args.Optional("root args", TypeForm)
			if err != nil {
				return fmt.Errorf("root: %w", err)
			}
			log.Printf("nextArgs: %#v, %#v", nextArgs, nextArgs.Form.First.El)
			scopes.Push(Scope{})
			err = Eval(blog, scopes, NewArgs(nextArgs.Form.First))
			scopes.Pop()
			if err != nil {
				return err
			}
		}
		return args.Finished()
	},
	"eof": func(blog *Blog, scopes *Scopes, args *Args) error {
		// @todo: fill in blog.Meta?
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
		blog.Author = component.Author{} // ensure author is initialized and zeroed
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
		//for !args.IsFinished() {
		for _ = range 2 {
			nextArgs, err := args.Optional("author args", TypeForm)
			if err != nil {
				return fmt.Errorf("author: %w", err)
			}
			err = Eval(blog, scopes, NewArgs(nextArgs.Form.First))
			if err != nil {
				return err
			}
		}
		return args.Finished()
	},
	"tags": func(blog *Blog, scopes *Scopes, args *Args) error {
		tags := component.Tags{}
		for !args.IsFinished() {
			tagList, err := args.Optional("space separated tag list", TypeText)
			if err != nil {
				return fmt.Errorf("tags: %w", err)
			}
			for _, tagStr := range strings.Split(string(tagList.Text), " ") {
				tags = append(tags, component.Tag(tagStr))
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
			if content.Type == TypeText {
				blog.Content = append(blog.Content, component.Text(content.Text))
			} else {
				return fmt.Errorf("body: unhandled argument type: %#v", content)
			}
		}
		return args.Finished()
	},
}

func Eval(blog *Blog, scopes *Scopes, args *Args) error {
	arg, err := args.Next("eval", TypeAny)
	if err != nil {
		return fmt.Errorf("eval: %w", err)
	}
	switch arg.Type {
	case TypeAtom:
		fun, err := scopes.Resolve(string(arg.Atom))
		if err != nil {
			return fmt.Errorf("eval: %w", err)
		}
		return fun(blog, scopes, args)
	case TypeForm:
		assert(false, "@todo: unhandled")
		// @fixme: doesn't work (here)!
		//scopes.Push(Scope{})
		//defer scopes.Pop()
		//return Eval(blog, scopes, NewArgs(arg.Form.First))
	case TypeText:
		assert(false, "@todo: unhandled")
	default:
		unreachable()
	}
	return nil
}

func Apply(blog *Blog, scopes *Scopes, fun beFun, arg *Arg) error {
	return nil
}
