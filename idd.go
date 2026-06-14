package epedit

import (
	"bufio"
	"os"
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
	Version string
	Classes map[string]*ClassDef // map for fast search (without capitalization)
}

func NewIDD() *IDD {
	return &IDD{
		Classes: make(map[string]*ClassDef),
	}
}

// * Parse IDD file into IDD struct

// open IDD file in filepath and return pointer for parsed IDD struct
func ParseIDD(filepath string) (*IDD, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	idd := NewIDD()
	scanner := bufio.NewScanner(file)

	// state machine
	var currentGroup string
	var currentClass *ClassDef
	var currentField *FieldDef

	for scanner.Scan() {
		line := scanner.Text()

		// 1. remove comments and whitespaces
		if idx := strings.Index(line, "!"); idx != -1 {
			line = line[:idx]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 2. parse \group
		if strings.HasPrefix(line, `\group`) {
			currentGroup = strings.TrimSpace(strings.TrimPrefix(line, `\group`))
			continue
		}

		// 3. parse metadata for class or field level
		if strings.HasPrefix(line, `\`) {
			if currentField != nil {
				// currently parsing a field
				parseFieldProperty(currentField, line)
			} else if currentClass != nil {
				// no current field, but parsing a class (ex. \extensible, \memo)
				parseClassProperty(currentClass, line)
			}
			continue
		}

		// 4. start new class or field (ends with `,` or `;`)
	}

	return idd, scanner.Err()
}

// * Helper functions for parsing class and field property

func parseClassProperty(class *ClassDef, line string) {
	if strings.HasPrefix(line, `\extensible`) {
		class.Extensible = &ExtensibleDef{}
		// TODO: parse other info such as \extensible:#
	}
	// TODO: \memo, \min-fields, etc.
}

func parseFieldProperty(field *FieldDef, line string) {
	if strings.HasPrefix(line, `\type`) {
		field.Type = strings.TrimSpace(strings.TrimPrefix(line, `\type`))
	} else if line == `\required-field` {
		field.Required = true
	}
	// TODO: \default, \key, etc.
}
