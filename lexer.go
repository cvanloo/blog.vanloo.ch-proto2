package main

// Input:
// (title Hello World)

// Tokenizer:
// TokenFormStart{"("}
// TokenAtom{"title"}
// TokenText{"Hello World"}
// TokenFormEnd{")"}

// Lexer:
// Form {
//   Atom {
//     "title"
//   },
//   Form {
//     Atom {
//       "text"
//     },
//     Text {
//       "Hello World"
//     }
//   }
// }

