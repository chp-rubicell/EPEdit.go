package epedit

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// * Field and class definition

// IDD field definition (ex. Outside_Boundary_Condition)
type FieldDef struct {
	Name             string
	Required         bool
	Units            string
	Default          string // TODO: deal with type conversion later...
	Autosizable      bool
	Autocalculatable bool
	Type             string   // alpha, real, integer, choice, etc.
	Choices          []string // possible values for "\type choice"
	// TODO: add more later
}

// IDD \extensible field properties (used in ClassDef)
type ExtensibleDef struct {
	BeginIndex int // start index of the extensible fields
	Size       int // size of the extensible fields (ex. X, Y, Z coords -> 3)
}

// IDD class definition (ex. Building, Zone)
type ClassDef struct {
	Name       string     // original name with capitalization
	Group      string     // \group
	Fields     []FieldDef // array of FieldDefs
	MinFields  int
	Extensible *ExtensibleDef // nil if empty
}

// IDD object that contains all of the definitions
type IDD struct {
	Version        string
	Classes        map[string]*ClassDef // map for fast search (without capitalization)
	OrderedClasses []*ClassDef          // for preserving order during export
}

func NewIDD() *IDD {
	return &IDD{
		Classes:        make(map[string]*ClassDef),
		OrderedClasses: make([]*ClassDef, 0),
	}
}

// * Parse IDD file into IDD struct

// state for tracking current parser mode
type parseState int

const (
	stateLookingForClass parseState = iota
	stateInClass
)

// return pointer for parsed IDD struct using Lexer
func ParseIDD(r io.Reader) (*IDD, error) {
	lexer := NewLexer(r)
	idd := NewIDD()

	// state machine
	state := stateLookingForClass
	var currentGroup string
	var currentClass *ClassDef
	var currentField *FieldDef

	// temporary text until comma or semicolon token
	var lastText string

	for {
		tok := lexer.NextToken()

		if tok.Type == TokenEOF {
			break
		}
		if tok.Type == TokenError {
			return nil, fmt.Errorf("Parsing error (Line %d): %s", lexer.LineNum, tok.Value)
		}

		// 1. text token
		if tok.Type == TokenText {
			if strings.HasPrefix(tok.Value, `\`) {
				// property
				if after, found := strings.CutPrefix(tok.Value, `\group`); found {
					currentGroup = strings.TrimSpace(after)
				} else if currentField != nil {
					// if there is active field, add as field property
					parseFieldProperty(currentClass, currentField, tok.Value)
				} else if currentClass != nil {
					// if no active field and active class, add as class property
					parseClassProperty(currentClass, tok.Value)
				}
			} else {
				// does not starts with \ (ex. "Zone", "A1")
				// add to temporary text
				lastText = tok.Value
			}
			continue
		}

		// 2. comma (,) token
		if tok.Type == TokenComma {
			switch state {
			case stateLookingForClass:
				// lastText is the new class name
				currentClass = &ClassDef{
					Name:  lastText,
					Group: currentGroup,
				}
				// add to map (for fast searching)
				idd.Classes[strings.ToUpper(lastText)] = currentClass
				// add to slice (for preserving order)
				idd.OrderedClasses = append(idd.OrderedClasses, currentClass)
				currentField = nil
				lastText = ""
				state = stateInClass
			case stateInClass:
				//lastText is new field name (ex. A1)
				newField := FieldDef{Name: lastText}
				lastText = ""
				currentClass.Fields = append(currentClass.Fields, newField)
				currentField = &currentClass.Fields[len(currentClass.Fields)-1]
			}
			continue
		}

		// 3. semicolon (;) token
		if tok.Type == TokenSemicolon {
			if state == stateInClass {
				// last field of class
				newField := FieldDef{Name: lastText}
				lastText = ""
				currentClass.Fields = append(currentClass.Fields, newField)
				currentField = &currentClass.Fields[len(currentClass.Fields)-1]

				// finished inputting class, looking for new class
				state = stateLookingForClass

				// don't reset currentClass and currentField to nil
				// more field info can follow after ;
			}
		}
	}

	return idd, nil
}

// * Helper functions for parsing class and field property

func parseClassProperty(class *ClassDef, val string) {
	if strings.HasPrefix(val, `\extensible`) {
		class.Extensible = &ExtensibleDef{}
		// parse \extensible:# info
		parts := strings.Split(val, ":")
		if len(parts) == 2 {
			if size, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				class.Extensible.Size = size
			}
		}
	} else if strings.HasPrefix(val, `\min-fields`) {
		parts := strings.Split(val, " ")
		if len(parts) >= 2 {
			if minFields, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				class.MinFields = minFields
			}
		}
	}
	// TODO: \memo, etc.
}

func parseFieldProperty(class *ClassDef, field *FieldDef, val string) {
	if after, found := strings.CutPrefix(val, `\field`); found {
		// replace temporary names (ex. A1, N1)
		field.Name = strings.TrimSpace(after)
	} else if after, found := strings.CutPrefix(val, `\type`); found {
		field.Type = strings.TrimSpace(after)
	} else if val == `\required-field` {
		field.Required = true
	} else if val == `\autosizable` {
		field.Autosizable = true
	} else if val == `\autocalculatable` {
		field.Autocalculatable = true
	} else if val == `\begin-extensible` {
		// current field is the starting field of extensibles
		if class.Extensible != nil {
			class.Extensible.BeginIndex = len(class.Fields) - 1
			// TODO: add extensible field name patterns
		}
	}
	// TODO: \default, \key, etc.
}
