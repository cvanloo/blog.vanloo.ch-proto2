package main

import (
	"fmt"
)

func panicIf[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	tokenizer := NewTokenizer([]rune(input))
	for _, t := range panicIf(tokenizer.Tokenize()) {
		fmt.Println(t)
	}
}

const input = `
(title こんにちは、日本)
(tags :clojure :asm :lisp :fp)
(brief Some summary)
(hidden)
(pinned)

(body

(comment This text does not appear in the output)

(table-of-contents)

Lorem ipsum dolor sit amet.

(bold This text is bold)
(b Same as bold but shorter)

(image someimage.png)


Within this text there lays (sidenote (text hidden) This is a sidenote) a sidenote.
Within this text there lays (sidenote (t hidden) This is a sidenote) a sidenote.

(italic This text is italic)
(i Same as italic but shorter)

(enquote This text is in quotes)
(q Same as enquote but shorter)


(code \+
func pointOfNoReturn(n int) (r int) {
	defer func() {
		recover()
		r = n + 1 // calculate result
	}()
	panic("no return")
	unquote? \\+
}
\+)


Here I'm escaping parentheses: \( hello world \).
And here I'm escaping a backslash \(reverse solidus\): \\.


(q I am quoting someone (cite Someone))

(footnotes)

)
`

const test = `
--- internal representation:
(title (text Hello World))
(tags (text :clojure :asm :lisp :fp))
(brief (text Some summary))

--- equivalent to:
(title "Hello World")
(tags ":clojure :asm :lisp :fp")
(brief "Some summary")

--- syntactic sugar:
(title Hello World)
(tags :clojure :asm :lisp :fp)
(brief Some summary)

(body Within this text there lays (sidenote (text hidden) This is a sidenote) a sidenote.)
(body Within this text there lays (sidenote "hidden" This is a sidenote) a sidenote.)

(p Hello World, Hello Moon, Goodbye World, Goodbye Moon)
=> (p (text Hello World, Hello Moon, Goodbye World, Goodbye Moon))

(p Hello World, (bold Hello) Moon, (italic Goodbye) World, Goodbye Moon)
=> (p (text Hello World, (bold (text Hello)) Moon, (italic (text Goodbye)) World, Goodbye Moon))

(p Hello World, Hello Moon, Goodbye World, Goodbye Moon (bold !))
=> (p (text Hello World, Hello Moon, Goodbye World, Goodbye Moon (bold !)))
`
