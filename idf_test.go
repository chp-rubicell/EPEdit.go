package epedit

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestParseIDF(t *testing.T) {
	// filepath := "testdata/RefBldgMediumOfficeNew2004_Chicago.idf"
	filepath := "testdata/RefBldgMediumOfficeNew2004_Chicago_Test.idf"

	idd, err := NewIDDFromFile("testdata/V24-2-0-Energy+.idd")
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDD: %v\n", err)
	}

	idf, err := NewIDFFromFile(filepath, idd)
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDF: %v\n", err)
	}

	formattedJSON, err := json.MarshalIndent(idf.Objects, "", "  ")
	if err != nil {
		t.Fatalf("Conversion error: %v\n", err)
	}
	fmt.Println(string(formattedJSON))
}

func TestUsage(t *testing.T) {
	filepath := "testdata/RefBldgMediumOfficeNew2004_Chicago_Test.idf"

	idd, err := NewIDDFromFile("testdata/V24-2-0-Energy+.idd")
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDD: %v\n", err)
	}

	idf, err := NewIDFFromFile(filepath, idd)
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDF: %v\n", err)
	}

	surfaces, _ := idf.Objects["BUILDINGSURFACE:DETAILED"]
	for i, surf := range surfaces {
		fmt.Printf("[%d] Surface Name: %s\n", i+1, surf.Values[0])
	}

	example_surf := idf.Objects["BUILDINGSURFACE:DETAILED"][0]

	fmt.Println(example_surf)

	formattedJSON, err := json.MarshalIndent(example_surf.Values, "", "  ")
	if err != nil {
		fmt.Printf("Conversion error: %v", err)
	}
	fmt.Println(string(formattedJSON))

	err = example_surf.Set("Vertex 10 Z-coordinate", -10)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(example_surf)

	formattedJSON, err = json.MarshalIndent(example_surf.Values, "", "  ")
	if err != nil {
		fmt.Printf("Conversion error: %v", err)
	}
	fmt.Println(string(formattedJSON))

	fmt.Println(idf)
}
