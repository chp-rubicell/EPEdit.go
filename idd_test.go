package epedit

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseIDD(t *testing.T) {
	// filepath := "testdata/V9-0-0-Energy+Test.idd"
	// filepath := "testdata/V24-2-0-Energy+Test.idd"
	filepath := "testdata/V24-2-0-Energy+.idd" // class count: 848

	idd, err := ParseIDDFile(filepath)
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDD: %v", err)
	}

	// formattedJSON, err := json.MarshalIndent(idd.OrderedClasses, "", "  ")
	// if err != nil {
	// 	t.Fatalf("Conversion error: %v", err)
	// }
	// fmt.Println(string(formattedJSON))

	fmt.Println(len(idd.OrderedClasses))
	for _, class := range idd.OrderedClasses {
		fmt.Println(class.Name)
	}
}

func TestIDDAPI(t *testing.T) {
	idd, err := ParseIDDFile("testdata/V24-2-0-Energy+.idd")
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDD: %v", err)
	}
	classDef := idd.Classes[strings.ToUpper("BuildingSurface:Detailed")]
	for _, i := range []int{0, 4, 20} {
		fieldName := classDef.GetFieldName(i)
		fieldIdx, _ := classDef.FindFieldIndex(fieldName)
		fmt.Printf("%d -> %s -> %d\n", i, fieldName, fieldIdx)
	}
}
