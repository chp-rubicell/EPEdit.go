package epedit

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// * Object definition

type Fields map[string]any // for ease of use for field input

type IDFObject struct {
	Class  *ClassDef
	Values []string // field values as slices
}

type IDF struct {
	IDD     *IDD
	Objects map[string][]*IDFObject // CLASSNAME: IDFObject slice
}

func NewIDF(idd *IDD) *IDF {
	return &IDF{
		IDD:     idd,
		Objects: make(map[string][]*IDFObject),
	}
}

// * Parse IDF file into IDF struct

func ParseIDF(r io.Reader, idd *IDD) (*IDF, error) {
	lexer := NewLexer(r, false) // turn of IDD mode
	idf := NewIDF(idd)

	var currentValues []string
	var lastText string

TokenLoop:
	for {
		tok := lexer.NextToken()

		switch tok.Type {
		case TokenEOF:
			break TokenLoop

		case TokenError:
			return nil, fmt.Errorf(`IDF parsing error (Line %d): %s`, lexer.LineNum, tok.Value)

		case TokenText:
			lastText = tok.Value

		case TokenComma:
			currentValues = append(currentValues, lastText)
			lastText = "" // reset for continuous commas (,,)

		case TokenSemicolon:
			currentValues = append(currentValues, lastText)

			if len(currentValues) > 0 {
				className := currentValues[0]
				searchKey := strings.ToUpper(className)

				classDef, exists := idd.Classes[searchKey]
				if !exists {
					return nil, fmt.Errorf(`Line %d: Class name "%s" not defined in IDD`, lexer.LineNum, className)
				}

				obj := &IDFObject{
					Class:  classDef,
					Values: currentValues[1:], // excluding first field (class name)
				}

				idf.Objects[searchKey] = append(idf.Objects[searchKey], obj)
			}

			currentValues = nil
			lastText = ""

		default:
			return nil, fmt.Errorf(`Line %d: Unrecognized token type (%d)`, lexer.LineNum, tok.Type)
		}
	}

	// TODO: version check

	return idf, nil
}

// * IDF manipulation API (Read)

// get object by class name
func (idf *IDF) GetObjects(className string) []*IDFObject {
	return idf.Objects[strings.ToUpper(className)]
}

// get object by first field (likely name)
func (idf *IDF) GetObjectByName(className string, objectName string) *IDFObject {
	candidates := idf.GetObjects(className)
	for _, obj := range candidates {
		if len(obj.Values) < 1 {
			continue
		}
		if strings.EqualFold(obj.Values[0], objectName) {
			return obj
		}
	}
	return nil
}

// get value of field as string (case-insensitive). "" if empty
func (obj *IDFObject) GetString(fieldName string) (string, error) {
	idx, err := obj.Class.FindFieldIndex(fieldName)
	if err != nil {
		return "", err
	}
	if idx >= len(obj.Values) {
		return "", nil // empty value
	}
	return obj.Values[idx], nil
}

// get value of field as string (case-insensitive, ignores error). "" if empty
func (obj *IDFObject) String(fieldName string) string {
	val, _ := obj.GetString(fieldName)
	return val
}

// get value of field as float (case-insensitive). math.NaN() if empty
func (obj *IDFObject) GetFloat(fieldName string) (float64, error) {
	valStr, err := obj.GetString(fieldName)
	if err != nil {
		return math.NaN(), err
	}
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return math.NaN(), err
	}
	return val, nil
}

// get value of field as float (case-insensitive, ignores error). math.NaN() if empty
func (obj *IDFObject) Float(fieldName string) float64 {
	val, _ := obj.GetFloat(fieldName)
	return val
}

// get value of field as int (case-insensitive). -1 if empty
func (obj *IDFObject) GetInt(fieldName string) (int, error) {
	valStr, err := obj.GetString(fieldName)
	if err != nil {
		return -1, err
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return -1, err
	}
	return val, nil
}

// get value of field as int (case-insensitive, ignores error). -1 if empty
func (obj *IDFObject) Int(fieldName string) int {
	val, _ := obj.GetInt(fieldName)
	return val
}

// * IDF manipulation API (Create, Update, Delete)

// set value of field (case-insensitive)
func (obj *IDFObject) Set(fieldName string, value any) error {
	targetIndex, err := obj.Class.FindFieldIndex(fieldName)
	if err != nil {
		return err
	}

	// 1. if targetIndex extends beyond current length
	if targetIndex >= len(obj.Values) {
		if targetIndex < cap(obj.Values) {
			// extend slice's length by re-slicing it
			obj.Values = obj.Values[:targetIndex+1]
		} else {
			// increase capacity
			needed := targetIndex + 1 - len(obj.Values)
			obj.Values = append(obj.Values, make([]string, needed)...)
		}
	}

	// 2. set value
	obj.Values[targetIndex] = AnyToString(value)
	return nil
}

// add object to IDF
func (idf *IDF) AddObject(className string, initialValues Fields) (*IDFObject, error) {
	searchKey := strings.ToUpper(className)
	classDef, exists := idf.IDD.Classes[searchKey]
	if !exists {
		return nil, fmt.Errorf("Unknown class: %s", className)
	}

	// minimum number of fields to preallocate
	minFields := len(classDef.Fields)
	if classDef.MinFields > 0 {
		minFields = classDef.MinFields
	} else if classDef.Extensible != nil && classDef.Extensible.BeginIndex > -1 {
		minFields = classDef.Extensible.BeginIndex + classDef.Extensible.Size
	}
	// TODO: preallocate based on maximum index in initialValues?

	newObj := &IDFObject{
		Class:  classDef,
		Values: make([]string, 0, minFields),
	}

	for key, val := range initialValues {
		if err := newObj.Set(key, val); err != nil {
			return nil, fmt.Errorf("Error while setting initial values: %v", err)
		}
	}

	idf.Objects[searchKey] = append(idf.Objects[searchKey], newObj)

	return newObj, nil
}

// remove object from IDF
func (idf *IDF) RemoveObject(target *IDFObject) error {
	searchKey := strings.ToUpper(target.Class.Name)

	list, exists := idf.Objects[searchKey]
	if !exists {
		return fmt.Errorf(`Failed to remove object: class "%s" does not exists`, target.Class.Name)
	}

	for i, obj := range list {
		if obj == target {
			copy(list[i:], list[i+1:]) // offset objects after target by -1
			list[len(list)-1] = nil    // remove last pointer
			idf.Objects[searchKey] = list[:len(list)-1]
			return nil
		}
	}

	return fmt.Errorf(`Failed to remove object: can't find object in class "%s" does not exists`, target.Class.Name)
}

// * Open and parse IDD file

func NewIDFFromFile(filepath string, idd *IDD) (*IDF, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	idf, err := ParseIDF(file, idd)
	if err != nil {
		return nil, err
	}

	return idf, nil
}
