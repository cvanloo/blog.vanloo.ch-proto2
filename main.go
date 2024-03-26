package main

import (
	"fmt"
	//"net/http"

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
	tokenizer := tok.NewTokenizer([]rune(testInput))
	//tokenizer := tok.NewTokenizer([]rune(remarkableReviewBlogPostSource))
	tokens := panicIf(tokenizer.Tokenize())
	for _, t := range tokens {
		fmt.Println(t)
	}
	fmt.Println("---------------")
	root := lex.Lex(tokens)
	fmt.Printf("%s\n", root)

	fmt.Println(component.String(root))

	//http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("fonts"))))
	//http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	//http.HandleFunc("/", component.Handler(root.First))
	//http.ListenAndServe(":8080", nil)
}

const testInput = `
(author (name Colin van~Loo) (email I'm not going to tell you....))
(title Reviewing the reMarkable)

(tags reMarkable review technology proprietary)

(body

This is text. 
This text will    be  joined with the previous line.  


This however, is a new text element \(because there are two \(!\) newlines in-between\).
)

Text outside the body! What might happen to it?
`

const remarkableReviewBlogPostSource = `
(author (name Colin van~Loo) (email colin@vanloo.ch)) 

(set :reMarkable (stylize :keep reMarkable))

(title My review of the :reMarkable)

(tags reMarkable review technology proprietary)

(brief
The :reMarkable is a paper tablet that advertises note taking without any distractions.
Its high price point certainly makes one think twice before buying.
Pros: quality, battery for weeks, feels like writing on real paper, open access to underlying Linux system, intuitive interface, tags.
Cons: disposable tips (have to buy refills), subscription service, pencil breaks easily when dropped, bad touch recognition in the corners/borders, proprietary file format, vendor lock-in (not a safe space for your notes!).
)

(abstract (copy :brief))

(comment Blog will be pinned on top of page, but will be hidden from timeline.)
(pinned)
(hidden)

(body

(print-abstract)
(print-table-of-contents)

(section The good parts)

Because of its high price, I would not have bought the (link :extern (url remarkable.com) :reMarkable) if it were not for the 30-day money-back guarantee.
After a week of daily usage, returning it was the last thing on my mind.
Consistent with their advertised claims, it really feels like writing on real paper.
The :reMarkable is handy, light-weight, and has lots of storage for my notes.
No more difficult to search through piles of paper up to the ceiling.
Tagging notebooks and single pages within a notebook makes them easily discoverable.

(subsection It runs Linux)

I did not want to end on a negative note, so I kept the best part for last:
The :reMarkable runs on Linux (which makes total sense if you (sidenote (t easier to extend, needs few resources, does not deplete battery quickly) think about it)).
When plugged into a computer, the :reMarkable automatically opens an SSH port.
It takes a bit of rummaging through settings to find the IP address(es) and root password.

(image (path settings-show-ssh-creds.png) SSH login credentials hidden under (q Compliance (@todo look up what it was called)).)

To make things simpler, I recommend setting up an SSH config...

(code (file ~/.ssh/config) \+
host remarkable
	Hostname 10.11.99.1
	User root
\+)

...and copy over your pubkey:

(code ~$ ssh-copy-id)

I use this to backup my data:

(code \+
~$ mkdir -p "backup-`+"`"+`date +%F`+"`"+`/files" && cd $_ && cd ..
~$ scp remarkable:~/.config/remarkable/xochitl.conf . # backup config
~$ scp remarkable:/usr/bin/xochitl                    # backup binary
~$ rsync -avAX remarkable:~/.local/share/remarkable/xochitl/ files/ # backup files
\+)

(q \+
Backups are for wimps. Real men upload their data to an FTP site and have everyone else mirror it.
\+ (cite Linus Torvalds))

(print-footnotes)

(subsection Some tests, not part of the blog post)

\+
\\+ <- this should render as a reverse solidus followed by a plus sign.
raw text block ends after this: <here>
\+

Here I'm escaping parentheses: \( hello world \).
And here I'm escaping a backslash \(reverse solidus\): \\.

)
`
