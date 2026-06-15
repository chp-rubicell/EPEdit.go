package epedit

import (
	"fmt"
	"os"
	"testing"
)

func TestLexer(t *testing.T) {
	filepath := "testdata/V24-2-0-Energy+Test.idd"
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer file.Close()

	lexer := NewLexer(file)

	// EOF를 만날 때까지 토큰 뽑아내기
	for {
		tok := lexer.NextToken()
		fmt.Printf("Type: %-2d | Value: %s\n", tok.Type, tok.Value)
		if tok.Type == TokenEOF {
			break
		}
	}
}
