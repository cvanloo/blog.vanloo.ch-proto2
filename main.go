package main

import (
	"fmt"
	"net/http"

	"be/component"
	"be/tok"
	"be/lex"
)

func panicIf[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	tokenizer := tok.NewTokenizer([]rune("(title Hello, World!)"))
	//tokenizer := tok.NewTokenizer([]rune(input))
	tokens := panicIf(tokenizer.Tokenize())
	for _, t := range tokens {
		fmt.Println(t)
	}
	fmt.Println("---------------")
	root := lex.Lex(tokens)
	fmt.Printf("%s\n", root)

	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	http.HandleFunc("/", component.Handler(root.First))
	http.ListenAndServe(":8080", nil)
}

const input = `
(title こんにちは、日本)
(tags clojure asm lisp fp)
(brief Some summary)
(hidden)
(pinned)

(body

(comment This text does not appear in the output)

(table-of-contents)

(section First Section)

Lorem ipsum dolor sit amet.

(subsection First Subsection)

(bold This text is bold)
(b Same as bold but shorter)

(section Second Section)

(image someimage.png)

Within this text there lays (sidenote (text hidden) This is a sidenote) a sidenote.
Within this text there lays (sidenote (t hidden) This is a sidenote) a sidenote.

(italic This text is italic)
(i Same as italic but shorter)

(subsection Links)

(link (text https://barcode.vanloo.ch/) Code-128 Generator)
(link (extern) (text https://barcode.vanloo.ch/) Code-128 Generator \(external link\))

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
