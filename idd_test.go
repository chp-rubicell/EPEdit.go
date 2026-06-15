package epedit

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestParseIDD(t *testing.T) {
	// filepath := "testdata/V9-0-0-Energy+Test.idd"
	filepath := "testdata/V24-2-0-Energy+Test.idd"
	// filepath := "testdata/V24-2-0-Energy+.idd"
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Printf("Failed to open file: %v\n", err)
		return
	}
	defer file.Close()

	idd, err := ParseIDD(file)
	if err != nil {
		t.Fatalf("Error occurred while parsing IDD: %v", err)
	}

	formattedJSON, err := json.MarshalIndent(idd.OrderedClasses, "", "  ")
	if err != nil {
		t.Fatalf("Conversion error: %v", err)
	}
	fmt.Println(string(formattedJSON))

	// fmt.Println(len(idd.OrderedClasses))
	// for _, class := range idd.OrderedClasses {
	// 	fmt.Println(class.Name)
	// }
}
