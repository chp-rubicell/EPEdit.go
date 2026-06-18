package epedit

import (
	"fmt"
	"io"
	"os"
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

// store extensible field names as Prefix + # + Suffix
// ex. "Vertex 1 X-coordinate" -> "Vertex ", " X-coordinate"
type ExtPattern struct {
	Prefix       string
	Suffix       string
	SearchPrefix string // for searching (lowercase)
	SearchSuffix string // for searching (lowercase)
}

// IDD extensible field properties (used in ClassDef)
type ExtensibleDef struct {
	BeginIndex int // start index of the extensible fields
	Size       int // size of the extensible fields (ex. X, Y, Z coords -> 3)
	Patterns   []ExtPattern
}

// IDD class definition (ex. Building, Zone)
type ClassDef struct {
	Name                string     // original name with capitalization
	Group               string     // \group
	Fields              []FieldDef // array of FieldDefs
	MinFields           int
	BaseFieldIndexMap   map[string]int // for fast indexing of fields (lowercase)
	Extensible          *ExtensibleDef // nil if empty
	FieldIdxWithDefault []int          // field indices with default value
}

// IDD object that contains all of the definitions
type IDD struct {
	Version        string
	Classes        map[string]*ClassDef // map for fast search (uppercase)
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
	lexer := NewLexer(r, true)
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
					err := parseClassProperty(currentClass, tok.Value, lexer.LineNum)
					if err != nil {
						return nil, err
					}
				}
			} else {
				// does not starts with \ (ex. "Zone", "A1")
				// add to temporary text
				lastText = tok.Value
			}

		} else if tok.Type == TokenComma {
			// 2. comma (,) token
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
				// if has extensible, skip after first set of extensible fields
				limit := -1
				if currentClass.Extensible != nil && currentClass.Extensible.BeginIndex >= 0 {
					// first set of extensible fields
					limit = currentClass.Extensible.BeginIndex + currentClass.Extensible.Size
				}
				if limit >= 0 && len(currentClass.Fields) >= limit {
					currentField = nil
				} else {
					// lastText is new field name (ex. A1)
					newField := FieldDef{Name: lastText}
					currentClass.Fields = append(currentClass.Fields, newField)
					currentField = &currentClass.Fields[len(currentClass.Fields)-1]
				}
				lastText = ""
			}

		} else if tok.Type == TokenSemicolon {
			// 3. semicolon (;) token
			if state == stateInClass {
				// if has extensible, skip after first set of extensible fields
				limit := -1
				if currentClass.Extensible != nil && currentClass.Extensible.BeginIndex >= 0 {
					// first set of extensible fields
					limit = currentClass.Extensible.BeginIndex + currentClass.Extensible.Size
				}
				if limit >= 0 && len(currentClass.Fields) >= limit {
					currentField = nil
				} else {
					// last field of class
					newField := FieldDef{Name: lastText}
					currentClass.Fields = append(currentClass.Fields, newField)
					currentField = &currentClass.Fields[len(currentClass.Fields)-1]
				}

				lastText = ""
				// finished inputting class, looking for new class
				state = stateLookingForClass

				// don't reset currentClass and currentField to nil
				// more field info can follow after ;
			}
		}
	}

	// after parsing, build indices for fast searching
	for _, class := range idd.OrderedClasses {
		// fix Extensible.BeginIndex if broken
		if err := class.fixMissingBeginIndex(); err != nil {
			return nil, err
		}
		class.buildIndices()
	}

	return idd, nil
}

// * Helper functions for parsing class and field property

func parseClassProperty(class *ClassDef, val string, lineNum int) error {
	if after, found := strings.CutPrefix(val, `\extensible`); found {
		class.Extensible = &ExtensibleDef{
			BeginIndex: -1, // -1 indicates before parsing the value
			Size:       -1,
		}
		// parse \extensible:# info
		startIndex, endIndex := GetContinuousDigitsIndices(after)
		if startIndex > -1 {
			size, err := strconv.Atoi(after[startIndex:endIndex])
			if err != nil {
				return fmt.Errorf(`Line %d: Can't convert \extensible size to number (%s)`, lineNum, after[startIndex:endIndex])
			}
			class.Extensible.Size = size
		} else {
			return fmt.Errorf(`Line %d: No number found after \extensible (%s)`, lineNum, val)
		}
	} else if strings.HasPrefix(val, `\min-fields`) {
		parts := strings.Split(val, " ")
		if len(parts) >= 2 {
			minFieldsString := strings.TrimSpace(parts[1])
			minFields, err := strconv.Atoi(minFieldsString)
			if err != nil {
				return fmt.Errorf(`Line %d: Can't convert \min-fields to number (%s)`, lineNum, minFieldsString)
			}
			class.MinFields = minFields
		} else {
			return fmt.Errorf(`Line %d: No number found after \min-fields (%s)`, lineNum, val)
		}
	}
	// TODO: \memo, etc.

	return nil
}

func parseFieldProperty(class *ClassDef, field *FieldDef, val string) {
	if after, found := strings.CutPrefix(val, `\field`); found {
		// replace temporary names (ex. A1, N1)
		field.Name = strings.TrimSpace(after)
	} else if val == `\required-field` {
		field.Required = true
	} else if val == `\begin-extensible` {
		// current field is the starting field of extensibles
		if class.Extensible != nil {
			class.Extensible.BeginIndex = len(class.Fields) - 1
			// TODO: add extensible field name patterns
		}
	} else if after, found := strings.CutPrefix(val, `\units`); found {
		field.Units = strings.TrimSpace(after)
	} else if after, found := strings.CutPrefix(val, `\default`); found {
		defaultValue := strings.TrimSpace(after)
		field.Default = defaultValue
		// if field has default value, add index to cache
		if defaultValue != "" {
			class.FieldIdxWithDefault = append(class.FieldIdxWithDefault, len(class.Fields)-1)
		}
	} else if val == `\autosizable` {
		field.Autosizable = true
	} else if val == `\autocalculatable` {
		field.Autocalculatable = true
	} else if after, found := strings.CutPrefix(val, `\type`); found {
		field.Type = strings.TrimSpace(after)
	} else if after, found := strings.CutPrefix(val, `\key`); found {
		field.Choices = append(field.Choices, strings.TrimSpace(after))
	}
	// TODO: \default, \key, etc.
}

// * Helper functions for building IDD

// run after IDD parsing to build indices
func (class *ClassDef) buildIndices() {
	class.BaseFieldIndexMap = make(map[string]int)

	limit := len(class.Fields)
	if class.Extensible != nil {
		limit = class.Extensible.BeginIndex
	}

	// add non-extensible fields to BaseFieldIndexMap
	for i := 0; i < limit; i++ {
		lowerName := strings.ToLower(class.Fields[i].Name)
		class.BaseFieldIndexMap[lowerName] = i
	}

	// add extensible fields to Extensible.Patterns
	if class.Extensible != nil && limit < len(class.Fields) {
		class.Extensible.Patterns = make([]ExtPattern, class.Extensible.Size)

		for i := 0; i < class.Extensible.Size; i++ {
			if limit+i < len(class.Fields) {
				prefix, suffix := extractPrefixSuffix(class.Fields[limit+i].Name)
				class.Extensible.Patterns[i] = ExtPattern{
					Prefix:       prefix,
					Suffix:       suffix,
					SearchPrefix: strings.ToLower(prefix),
					SearchSuffix: strings.ToLower(suffix),
				}
			}
		}
	}
}

// helper function for fixing classes with missing \begin-extensible tag
func (class *ClassDef) fixMissingBeginIndex() error {
	ext := class.Extensible

	// filter when size is defined but BeginIndex is -1
	if ext == nil || ext.BeginIndex != -1 || ext.Size <= 0 {
		return nil
	}

	numFields := len(class.Fields)
	if numFields < ext.Size {
		return fmt.Errorf(`IDD schema error: class "%s" have less fields (%d) than \extensible size (%d)`, class.Name, numFields, ext.Size) // broken beyond repair
	}

	// 1. extract ExtPatterns from back
	patterns := make([]ExtPattern, ext.Size) // temporary patterns for matching
	for i := 0; i < ext.Size; i++ {
		fieldName := class.Fields[numFields-ext.Size+i].Name
		prefix, suffix := extractPrefixSuffix(fieldName)
		patterns[i] = ExtPattern{Prefix: prefix, Suffix: suffix}
	}

	// 2. match pattern from back
	beginIdx := numFields
	for i := numFields - 1; i >= 0; i-- {
		// get offset within extensible group
		offset := ext.Size - 1 - ((numFields - 1 - i) % ext.Size)

		pattern := patterns[offset]
		name := class.Fields[i].Name

		// if pattern does not match, that is the beginning index
		if !strings.HasPrefix(name, pattern.Prefix) || !strings.HasSuffix(name, pattern.Suffix) {
			break
		}

		// if matched, move the beginning index
		beginIdx = i
	}

	// 3. apply beginIdx
	class.Extensible.BeginIndex = beginIdx
	return nil
}

// helper function for extracting prefix and suffix from extensible field name
func extractPrefixSuffix(name string) (prefix string, suffix string) {
	startIndex, endIndex := GetContinuousDigitsIndices(name)

	if startIndex > -1 {
		return name[:startIndex], name[endIndex:]
	} else {
		// if number is not found (ex. "Wavelength"), add space at the end (ex. "Wavelength 1")
		return name + " ", ""
	}
}

// * Open and parse IDD file

func ParseIDDFile(filename string) (*IDD, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to open IDD file (%s): %w", filename, err)
	}
	defer file.Close()

	idd, err := ParseIDD(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse IDD: %w", err)
	}

	return idd, nil
}

// * API (Read)

// get index from field name (case-insensitive)
func (class *ClassDef) FindFieldIndex(fieldName string) (int, error) {
	// convert to lowercase
	searchKey := strings.ToLower(fieldName)

	// 1. base fields
	if idx, exists := class.BaseFieldIndexMap[searchKey]; exists {
		return idx, nil
	}

	// 2. extensible fields
	if class.Extensible != nil {
		ext := class.Extensible
		for offset, pat := range ext.Patterns {
			if after, found := strings.CutPrefix(searchKey, pat.SearchPrefix); found {
				if numStr, found := strings.CutSuffix(after, pat.SearchSuffix); found {
					groupNum, err := strconv.Atoi(numStr)
					if err == nil && groupNum > 0 {
						return ext.BeginIndex + (groupNum-1)*ext.Size + offset, nil
					}
				}
			}
		}
	}

	return -1, fmt.Errorf(`Unknown field name in class "%s": %s`, class.Name, fieldName)
}

// get field name from index
func (class *ClassDef) GetFieldName(index int, addUnits bool) string {
	ext := class.Extensible

	// base fields
	if ext == nil || ext.BeginIndex == -1 || index < ext.BeginIndex {
		if index >= len(class.Fields) {
			return ""
		}
		field := class.Fields[index]
		if addUnits && field.Units != "" {
			return field.Name + " {" + field.Units + "}"
		}
		return field.Name
	}

	// extensible fields
	groupNum := (index-ext.BeginIndex)/ext.Size + 1
	offset := (index - ext.BeginIndex) % ext.Size
	if offset >= len(ext.Patterns) {
		return ""
	}
	pat := ext.Patterns[offset]
	fieldName := fmt.Sprintf("%s%d%s", pat.Prefix, groupNum, pat.Suffix)
	if units := class.Fields[ext.BeginIndex+offset].Units; addUnits && units != "" {
		return fieldName + " {" + units + "}"
	}
	return fieldName
}
