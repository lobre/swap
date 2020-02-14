package main

import "fmt"

// itemType identifies the type of lex items.
type itemType int

// TODO: should I use blackfriday to parse markdown?
const (
	itemError itemType = iota // error occured, value is text of error

	itemLeftJson    // an opening '{'
	itemLeftYaml    // a opening '---'
	itemFrontmatter // content of frontmatter
	itemRightJson   // a closing '}'
	itemRightYaml   // a closing '---'
	itemH1          // markdown '#' title
	itemH2          // markdown '##' title
	itemH3          // markdown '###' title
	itemH4          // markdown '####' title
	itemH5          // markdown '#####' title
	itemH6          // markdown '######' title
	itemLink        // markdown link
	itemText        // plain text
	itemEOF
)

// item represents a token returned from the scanner
type item struct {
	typ itemType // Type, such as itemNumber.
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

// emit pases an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}
