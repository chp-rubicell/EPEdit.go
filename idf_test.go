package epedit

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestParseIDF(t *testing.T) {
	// filepath := "testdata/RefBldgMediumOfficeNew2004_Chicago.idf"
	filepath := "testdata/RefBldgMediumOfficeNew2004_Chicago_Test.idf"

	idd, err := ParseIDDFile("testdata/V24-2-0-Energy+.idd")
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDD: %v\n", err)
	}

	idf, err := ParseIDFFile(filepath, idd)
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDF: %v\n", err)
	}

	formattedJSON, err := json.MarshalIndent(idf.Objects, "", "  ")
	if err != nil {
		t.Fatalf("Conversion error: %v\n", err)
	}
	fmt.Println(string(formattedJSON))
}

func TestIDFEdit(t *testing.T) {
	filepath := "testdata/RefBldgMediumOfficeNew2004_Chicago_Test.idf"

	idd, err := ParseIDDFile("testdata/V24-2-0-Energy+.idd")
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDD: %v\n", err)
	}

	idf, err := ParseIDFFile(filepath, idd)
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDF: %v\n", err)
	}

	fmt.Println(idf)
	fmt.Println("---")
	v := idf.GetObjectByName("version", "24.2")
	fmt.Println(v)
	fmt.Println("---")
	idf.RemoveObject(v)
	fmt.Println(idf)
	fmt.Println("---")

	ss := idf.GetObjects("simulationcontrol")[0]
	fmt.Println(ss)
	fmt.Println("---")
	ss.Update(Fields{
		"Do System Sizing Calculation": "No",
		"Do Plant Sizing Calculation":  "No",
	})
	fmt.Println(ss)
	fmt.Println("---")

	surfaces := idf.GetObjects("BUILDINGSURFACE:DETAILED")
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

func TestIDFParseAndSave(t *testing.T) {
	filepath := "testdata/RefBldgMediumOfficeNew2004_Chicago.idf"

	idd, err := ParseIDDFile("testdata/V24-2-0-Energy+.idd")
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDD: %v\n", err)
	}

	idf, err := ParseIDFFile(filepath, idd)
	if err != nil {
		t.Fatalf("Error occurred while opening and parsing IDF: %v\n", err)
	}

	err = idf.Save("testdata/ExportTest.idf")
	if err != nil {
		t.Fatalf("Error occurred while saving IDF: %v\n", err)
	}
}
