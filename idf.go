package epedit

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// TODO: create empty IDF
// TODO: default values

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

// * Open and parse IDD file

func ParseIDFFile(filename string, idd *IDD) (*IDF, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("Failed to open IDF file (%s): %w", filename, err)
	}
	defer file.Close()

	idf, err := ParseIDF(file, idd)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse IDF: %w", err)
	}

	return idf, nil
}

// * IDF manipulation API (Read)

// get object by class name
func (idf *IDF) GetObjects(className string) []*IDFObject {
	return idf.Objects[strings.ToUpper(className)]
}

// get object by first field (likely name)
func (idf *IDF) GetObjectByName(className string, objectName string) (*IDFObject, error) {
	candidates := idf.GetObjects(className)
	for _, obj := range candidates {
		if len(obj.Values) < 1 {
			continue
		}
		if strings.EqualFold(obj.Values[0], objectName) {
			return obj, nil
		}
	}
	return nil, fmt.Errorf(`Failed to find object "%s" in class "%s"`, objectName, className)
}

// get value of field as string (case-insensitive, also returns error). "" if empty
func (obj *IDFObject) GetStringErr(fieldName string) (string, error) {
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
func (obj *IDFObject) GetString(fieldName string) string {
	val, _ := obj.GetStringErr(fieldName)
	return val
}

// get value of field as float (case-insensitive, also returns error). math.NaN() if empty
func (obj *IDFObject) GetFloatErr(fieldName string) (float64, error) {
	valStr, err := obj.GetStringErr(fieldName)
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
func (obj *IDFObject) GetFloat(fieldName string) float64 {
	val, _ := obj.GetFloatErr(fieldName)
	return val
}

// get value of field as int (case-insensitive, also returns error). -1 if empty
func (obj *IDFObject) GetIntErr(fieldName string) (int, error) {
	valStr, err := obj.GetStringErr(fieldName)
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
func (obj *IDFObject) GetInt(fieldName string) int {
	val, _ := obj.GetIntErr(fieldName)
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
	obj.Values[targetIndex] = strings.TrimSpace(AnyToString(value))
	return nil
}

// update object from map
func (obj *IDFObject) Update(values Fields) error {
	for key, val := range values {
		if err := obj.Set(key, val); err != nil {
			return err
		}
	}
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

	err := newObj.Update(initialValues)
	if err != nil {
		return nil, fmt.Errorf("Error while setting initial values: %w", err)
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

// * Export

// format setting
type formatConfig struct {
	classIndent string // indent for class names
	fieldIndent string // indent for fields
	fieldSize   int    // minimum size for field values
	compact     bool   // compact mode
}

// generate formatConfig
func NewFormatConfig(classIndentSize int, fieldIndentSize int, fieldSize int) formatConfig {
	return formatConfig{
		classIndent: strings.Repeat(" ", classIndentSize),
		fieldIndent: strings.Repeat(" ", fieldIndentSize),
		fieldSize:   fieldSize,
		compact:     false,
	}
}

// default value
var defaultFormatConfig = NewFormatConfig(0, 4, 25)

// minimal format
var MinimalFormatConfig = formatConfig{"", "", 0, true}

// write IDFObject to io.Writer with formatConfig
func (obj *IDFObject) writeWithFormat(w io.Writer, cfg formatConfig) (int64, error) {
	var totalWritten int64

	// closure helper function
	writeStr := func(s string) error {
		n, err := io.WriteString(w, s)
		totalWritten += int64(n)
		return err
	}

	linebreak := "\n"
	if cfg.compact {
		linebreak = ""
	}

	// 1. print class name
	if err := writeStr(cfg.classIndent + obj.Class.Name); err != nil {
		return totalWritten, err
	}

	// 2. find last field with non-empty value
	lastIdx := -1
	for i := len(obj.Values) - 1; i >= 0; i-- {
		if strings.TrimSpace(obj.Values[i]) != "" {
			lastIdx = i
			break
		}
	}

	// 3. if all fields are empty, print ; and return
	if lastIdx == -1 {
		err := writeStr(";" + linebreak)
		return totalWritten, err
	}

	// 4. if not, print , after class name
	if err := writeStr("," + linebreak); err != nil {
		return totalWritten, err
	}

	// 5. print until lastIdx
	for i := 0; i <= lastIdx; i++ {
		val := obj.Values[i]

		// field value string
		fieldValString := val + ","
		if i == lastIdx {
			// finish with semicolon and an extra newline
			fieldValString = val + ";"
		}

		// add padding to field value string
		if cfg.fieldSize > 0 {
			fieldValString = fmt.Sprintf("%-*s", cfg.fieldSize, fieldValString)
		}

		// comment string
		commentString := ""
		if !cfg.compact {
			commentString = " !- " + obj.Class.GetFieldName(i, true)
		}

		// final line
		if err := writeStr(cfg.fieldIndent + fieldValString + commentString + linebreak); err != nil {
			return totalWritten, err
		}
	}

	// if compact mode, add final linebreak
	if cfg.compact {
		if err := writeStr("\n"); err != nil {
			return totalWritten, err
		}
	}

	return totalWritten, nil
}

// write IDFObject to io.Writer
func (obj *IDFObject) WriteTo(w io.Writer) (int64, error) {
	return obj.writeWithFormat(w, defaultFormatConfig)
}

// write IDF to io.Writer with formatConfig
func (idf *IDF) writeWithFormat(w io.Writer, cfg formatConfig) (int64, error) {
	var totalWritten int64
	var currentGroup string

	// iterate through IDD's ordered class list
	for _, classDef := range idf.IDD.OrderedClasses {
		searchKey := strings.ToUpper(classDef.Name)
		objects, exists := idf.Objects[searchKey]

		// if class is not in IDF
		if !exists || len(objects) == 0 {
			continue
		}

		// add group separator if changed
		if !cfg.compact && currentGroup != classDef.Group {
			currentGroup = classDef.Group
			n, err := fmt.Fprintf(w, "\n! ***%s***\n", strings.ToUpper(currentGroup))
			totalWritten += int64(n)
			if err != nil {
				return totalWritten, err
			}
		}

		// write objects
		for _, obj := range objects {
			// add newline
			n1, err := w.Write([]byte("\n"))
			totalWritten += int64(n1)
			if err != nil {
				return totalWritten, err
			}
			// write IDFObject
			n2, err := obj.writeWithFormat(w, cfg)
			totalWritten += n2
			if err != nil {
				return totalWritten, err
			}
		}
	}

	return totalWritten, nil
}

// write IDF to io.Writer
func (idf *IDF) WriteTo(w io.Writer) (int64, error) {
	return idf.writeWithFormat(w, defaultFormatConfig)
}

// * Convert to string

// convert IDFObject to string (for debugging)
func (obj *IDFObject) Format(cfg formatConfig) string {
	var builder strings.Builder // for efficient string building
	// use WriteTo to buffer (Builder) instead of file
	if _, err := obj.writeWithFormat(&builder, cfg); err != nil {
		return ""
	}
	return builder.String()
}

// convert IDFObject to string with default format config
func (obj *IDFObject) String() string {
	return obj.Format(defaultFormatConfig)
}

// convert IDF to string (for debugging)
func (idf *IDF) Format(cfg formatConfig) string {
	var builder strings.Builder
	if _, err := idf.writeWithFormat(&builder, cfg); err != nil {
		return ""
	}
	return strings.TrimSpace(builder.String())
}

// convert IDFObject to string with default format config
func (idf *IDF) String() string {
	return idf.Format(defaultFormatConfig)
}

// * Save to file

func (idf *IDF) Save(filename string, cfg ...formatConfig) error {
	// format config
	activeCfg := defaultFormatConfig
	if len(cfg) > 0 {
		activeCfg = cfg[0]
	}

	// create file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Failed to create file (%s): %w", filename, err)
	}
	defer file.Close()

	// create buffer writer
	bufferedWriter := bufio.NewWriter(file)

	// add header
	header := fmt.Sprintf("! Generated using EPEdit.go\n! Saved at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	if _, err := bufferedWriter.WriteString(header); err != nil {
		return fmt.Errorf("Failed to write header: %w", err)
	}

	// write idf to buffer
	if _, err := idf.writeWithFormat(bufferedWriter, activeCfg); err != nil {
		return fmt.Errorf("Failed to write IDF: %w", err)
	}
	// flush remaining data
	if err := bufferedWriter.Flush(); err != nil {
		return fmt.Errorf("Failed to write IDF: %w", err)
	}

	return nil
}
