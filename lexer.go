package epedit

import (
	"bufio"
	"io"
	"strings"
)

// Token type
type TokenType int

const (
	TokenText      TokenType = iota // regular text (ex. Zone, \type)
	TokenComma                      // comma (,)
	TokenSemicolon                  // semicolon (;)
	TokenEOF                        // end of file
	TokenError                      // error
)

// smallest unit of meaning sent from lexer to parser
type Token struct {
	Type  TokenType
	Value string // if text, store real values, otherwise store ",", ";", etc.
}

// lexer
type Lexer struct {
	scanner *bufio.Scanner
	buffer  []Token // queue for storing multiple tokens in one line
	LineNum int     // current line number (for debugging)
	IsIDD   bool    //  for IDD parsing
}

// lexer constructor
func NewLexer(r io.Reader, isIDD bool) *Lexer {
	return &Lexer{
		scanner: bufio.NewScanner(r),
		buffer:  make([]Token, 0),
		LineNum: 0,
		IsIDD:   isIDD,
	}
}

// return next token when parser requests
func (l *Lexer) NextToken() Token {
	// 1. if there are tokens left in the buffer, return first
	if len(l.buffer) > 0 {
		tok := l.buffer[0]
		l.buffer = l.buffer[1:] // remove first from queue (slicing)
		return tok
	}

	// 2. buffer is empty, read next line
	for l.scanner.Scan() {
		l.LineNum++
		line := l.scanner.Text()

		// tokenize line and add to buffer
		l.tokenizeLine(line)

		// if tokens are added to buffer, return first token
		// if empty, continue to next line
		if len(l.buffer) > 0 {
			tok := l.buffer[0]
			l.buffer = l.buffer[1:]
			return tok
		}
	}

	// 3. after loop, check if err or EOF
	if err := l.scanner.Err(); err != nil {
		return Token{Type: TokenError, Value: err.Error()}
	}

	return Token{Type: TokenEOF}
}

// read line and split by commas and semicolons
func (l *Lexer) tokenizeLine(line string) {
	// for joining strings
	var textBuilder strings.Builder

	// remove comments
	if idx := strings.IndexByte(line, '!'); idx >= 0 {
		line = line[:idx]
	}

	for byteIndex, char := range line {
		// if IDD mode and starts with \
		if l.IsIDD && char == '\\' {
			l.pushTextToken(&textBuilder)
			l.buffer = append(l.buffer, Token{
				Type:  TokenText,
				Value: strings.TrimSpace(line[byteIndex:]),
			})
			break
		}

		// parse regular letters
		switch char {
		case ',':
			l.pushTextToken(&textBuilder)
			l.buffer = append(l.buffer, Token{Type: TokenComma, Value: ","})
		case ';':
			l.pushTextToken(&textBuilder)
			l.buffer = append(l.buffer, Token{Type: TokenSemicolon, Value: ";"})
		default:
			textBuilder.WriteRune(char) // add letter to textBuilder
		}
	}

	// if line ended without comma or semicolon (ex. \group Name)
	l.pushTextToken(&textBuilder)
}

// helper function for creating a trimmed string from textBuilder
func (l *Lexer) pushTextToken(b *strings.Builder) {
	text := strings.TrimSpace(b.String())
	if text != "" {
		l.buffer = append(l.buffer, Token{Type: TokenText, Value: text})
	}
	b.Reset() // reset builder for next letters
}
