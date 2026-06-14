package epedit

import (
	"bufio"
	"os"
)

// IDD field definition (ex. Outside_Boundary_Condition)
type FieldDef struct {
	Name             string
	Required         bool
	Units            string
	Default          string //TODO deal with type conversion later...
	Autosizable      bool
	Autocalculatable bool
	Type             string   // alpha, real, integer, choice, etc.
	Choices          []string // possible values for "\type choice"
	//TODO add more later
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
	Classes map[string]*ClassDef // Map for fast search (without capitalization)
}

func NewIDD() *IDD {
	return &IDD{
		Classes: make(map[string]*ClassDef),
	}
}

// open IDD file in filepath and return pointer for parsed IDD struct
func ParseIDD(filepath string) (*IDD, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	idd := NewIDD()
	scanner := bufio.NewScanner(file)

	return idd, err
}
