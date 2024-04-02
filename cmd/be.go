package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	. "be"
	. "be/internal/debug"
	"be/lex"
	"be/tok"
)

var shouldServe = flag.Bool("serve", false, "serve generated output on :8080")

func init() {
	flag.Parse()
}

func main() {
	//tokenizer := tok.NewTokenizer([]rune(testInput2))
	content := Must(os.ReadFile("blog_example.be"))
	_ = content
	tokenizer := tok.NewTokenizer([]rune(string(content)))
	tokens := Must(tokenizer.Tokenize())
	for _, t := range tokens {
		fmt.Println(t)
	}
	fmt.Println("---------------")
	root := lex.Lex(tokens)
	fmt.Printf("%s\n", root)

	blog := &Blog{}
	scopes := InitScopes(blog)
	if err := blog.Eval(scopes, root.First); err != nil {
		fmt.Printf("error evaluating blog: %v", err)
		return
	}
	blogHtml, err := String(blog)
	if err != nil {
		fmt.Printf("error generating html: %v", err)
		return
	}
	fmt.Println(blogHtml)
	PanicIf(os.WriteFile("out.html", []byte(blogHtml), 0644))

	if *shouldServe {
		http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))
		http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
		http.HandleFunc("/", Handler(blog, PanicIf))
		http.ListenAndServe(":8080", nil)
	}
}
