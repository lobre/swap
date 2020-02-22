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
package main

import (
    "bytes"
    "fmt"
    "unicode/utf8"
)

// itemType identifies the type of lex items.
type itemType int

const eof = -1

const (
	itemError itemType = iota // error occured, value is text of error

    itemFrontmatterTOML
    itemFrontmatterYAML
    itemFrontmatterJSON
    itemMarkdown
	itemEOF
)

// item represents a token returned from the scanner
type item struct {
	typ itemType // Type, such as itemContent.
	val []byte   // Value, such as "23.2".
}

func (i item) String() string {
	switch i.typ {
	case itemEOF:
		return "EOF"
	case itemError:
		return string(i.val)
	}
	if len(i.val) > 20 {
		return fmt.Sprintf("%.20q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

// stateFunc represents the state of the scanner
// as a function that returns the next state.
type stateFunc func(*lexer) stateFunc

// lexer holds the state of the scanner.
type lexer struct {
	input []byte    // the string being scanned.
	start int       // start position of this item.
	pos   int       // current position in the input.
	width int       // width of last rune read from input.
	items chan item // channel of scanned items.
}

// starting the lexer
func lex(input []byte) (*lexer, chan item) {
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run() // Concurrently run state machine.
	return l, l.items
}

// run lexes the input by executing state functions until
// the state is nil.
func (l *lexer) run() {
	for state := lexStart; state != nil; {
		state = state(l)
	}
	close(l.items) // No more tokens will be delivered.
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
	l.start = l.pos
}

// error returns an error token and terminates the scan
// by passing back a nil pointer that will be the next
// state, terminating l.run.
func (l *lexer) errorf(format string, args ...interface{}) stateFunc {
    l.items <- item{
        itemError,
        []byte(fmt.Sprintf(format, args...)),
    }
    return nil
}

// next returns the next rune in the input.
func (l *lexer) next() (r rune) {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRune(l.input[l.pos:])
	l.pos += l.width
	return r 
}

// backup steps back one rune.
func (l *lexer) backup() {
    l.pos -= l.width
}

// peek returns but does not consume
// the next rune in the input.
func (l *lexer) peek() rune {
    r := l.next()
    l.backup()
    return r
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
    l.start = l.pos
}

// consumeCRLF consumes next runes if end of line.
func (l *lexer) consumeCRLF() bool {
	var consumed bool

	for _, r := range []rune{'\r', '\n'} {
		if l.next() != r {
			l.backup()
		} else {
			consumed = true
		}
	}
	return consumed
}

// hasPrefix check if the next string matches the prefix.
func (l *lexer) hasPrefix(prefix []byte) bool {
	return bytes.HasPrefix(l.input[l.pos:], prefix)
}

func lexStart(l *lexer) stateFunc {
    r := l.peek()

    switch {
    case r == '+':
        return lexFrontmatterTOML
    case r == '-':
        return lexFrontmatterYAML
    case r == '{':
        return lexFrontmatterJSON
    default:
        return lexMarkdown
    }
}

func lexFrontmatterTOML(l *lexer) stateFunc {
	for i := 0; i < 3; i++ {
		if r := l.next(); r != '+' {
			return l.errorf("invalid TOML delimiter")
		}
	}

    // ignore starting delimiter
    l.consumeCRLF()
    l.ignore()

    for {
        r := l.next()
        if r == eof {
            return l.errorf("EOF looking for end TOML front matter delimiter")
        }

        if isEndOfLine(r) && l.hasPrefix([]byte("+++")) {
            l.emit(itemFrontmatterTOML)
            l.pos += 3
            l.consumeCRLF()
            // ignore ending delimiter
            l.ignore()
            break
        }
    }

    return lexMarkdown
}

func lexFrontmatterYAML(l *lexer) stateFunc {
	for i := 0; i < 3; i++ {
		if r := l.next(); r != '-' {
			return l.errorf("invalid YAML delimiter")
		}
	}

    // ignore starting delimiter
    l.consumeCRLF()
    l.ignore()

    for {
        r := l.next()
        if r == eof {
            return l.errorf("EOF looking for end YAML front matter delimiter")
        }

        if isEndOfLine(r) && l.hasPrefix([]byte("---")) {
            l.emit(itemFrontmatterYAML)
            l.pos += 3
            l.consumeCRLF()
            // ignore ending delimiter
            l.ignore()
            break
        }
    }

    return lexMarkdown
}

func lexFrontmatterJSON(l *lexer) stateFunc {
	var (
		inQuote bool
		level   int
	)

	for {

		r := l.next()

		switch {
		case r == eof:
			return l.errorf("unexpected EOF parsing JSON front matter")
		case r == '{':
			if !inQuote {
				level++
			}
		case r == '}':
			if !inQuote {
				level--
			}
		case r == '"':
			inQuote = !inQuote
		case r == '\\':
			// This may be an escaped quote. Make sure it's not marked as a
			// real one.
			l.next()
		}

		if level == 0 {
			break
		}
	}

	l.consumeCRLF()
	l.emit(itemFrontmatterJSON)

	return lexMarkdown
}

func lexMarkdown(l *lexer) stateFunc {
    for {
        r := l.next()
        if r == eof {
            l.emit(itemMarkdown)
            return lexDone
        }
    }
}

func lexDone(l *lexer) stateFunc {
	l.emit(itemEOF)
	return nil
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
