// A lexer is a software program that performs lexical analysis.
// Lexical analysis is the process of separating a stream of 
// characters into different words, which in computer science we call 'tokens'.
// When you read my answer you are performing the lexical operation 
// of breaking the string of text at the space characters into multiple words.
//
// A parser goes one level further than the lexer and takes the tokens
// produced by the lexer and tries to determine if proper sentences have 
// been formed. Parsers work at the grammatical level,
// lexers work at the word level.
//
// This lexer does not take the responsibility to parse the frontmatter or the 
// markdown content. It just split it.
//
// TODO: take example on hugo parser/pageparser
package main

import "fmt"

// itemType identifies the type of lex items.
type itemType int

const eof = -1

const (
	itemError itemType = iota // error occured, value is text of error

	itemOpenJson    // an opening '{'
	itemCloseJson   // a closing '}'
	itemYamlDelim   // '---'
	itemText        // plain text
	itemEOF
)

// item represents a token returned from the scanner
type item struct {
	typ itemType // Type, such as itemText.
	val string   // Value, such as "23.2".
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	if len(i.val) > 10 {
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// stateFn represents the state of the scanner
// as a function that returns the next state.
type stateFn func(*lexer) stateFn

// run lexes the input by executing state functions
// until the state is nil (representing "done").
func run() {
	for state := startState; state != nil; {
		state = state(lexer)
	}
}

// lexer holds the state of the scanner.
type lexer struct {
	name  string    // used only for error reports.
	input string    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

// starting the lexer
func lex(name, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run() // Concurrently run state machine.
	return l, l.items
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

const openJson = "{"
const yamlDelim = "---"

func lexText(l *lexer) stateFn {
    for {
        if strings.HasPrefix(l.input[l.pos:], yamlDelim) {
           return lexYamlDelim 
        }
    }
}

func lexOpenYaml(l *lexer) stateFn {
    l.pos += len(yamlDelim)
    l.emit(itemYamlDelim)
    return lexInsideYaml
}

func lexInsideYaml(l *lexer) stateFn {
    for {
        if strings.HasPrefix(l.input[l.pos:], yamlDelim) {
            return lexCloseYaml
        }
        switch r := l.next(); {
        case r == eof:
            return l.errorf("unclosed yaml frontmatter")

        }

    }
}

func lexCloseYaml(l *lexer) stateFn {
    l.pos += len(yamlDelim)
    l.emit(itemYamlDelim)
    return lexContent
}
