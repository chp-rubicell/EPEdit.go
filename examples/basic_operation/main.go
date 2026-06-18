package main

import (
	"fmt"
	"log"

	epedit "github.com/chp-rubicell/EPEdit.go"
)

func Check(err error) {
	if err != nil {
		log.Fatalln("Fatal error:", err)
	}
}

func Must[T any](result T, err error) T {
	if err != nil {
		log.Fatalln("Fatal error:", err)
	}
	return result
}

func main() {

	filepath := "../../testdata/RefBldgMediumOfficeNew2004_Chicago.idf"

	idd, err := epedit.ParseIDDFile("../../testdata/V24-2-0-Energy+.idd")
	if err != nil {
		fmt.Printf("Error occurred while opening and parsing IDD: %v\n", err)
	}

	idf, err := epedit.ParseIDFFile(filepath, idd)
	if err != nil {
		fmt.Printf("Error occurred while opening and parsing IDF: %v\n", err)
	}

	// ? Read

	// 1. get objects of specific class
	zones := idf.GetObjects("ZONE")
	fmt.Println("Total zone count:", len(zones))

	// 2. get object by first field (name)
	zn := Must(idf.GetObjectByName("ZONE", "Perimeter_mid_ZN_1"))

	// 3. get value
	yorigin := zn.GetFloat("Y Origin")
	name := zn.GetString("Name")
	fmt.Println(yorigin, name)

	// ? Update

	surf := Must(idf.GetObjectByName("BUILDINGSURFACE:DETAILED", "Core_bot_ZN_5_Wall_South"))

	// modify field
	Check(surf.Set("Sun Exposure", "SunExposed"))

	// modify extensible field
	Check(surf.Set("Vertex 20 X-coordinate", 15.5))

	// remove data from field
	Check(surf.Set("Vertex 20 X-coordinate", ""))

	// ? Create / Delete

	// add new object
	newMat, err := idf.AddObject("MATERIAL", epedit.Fields{"Name": "MyNewInsulation"})
	newMat.Set("Thickness", 0.05)
	newMat.Set("Conductivity", 0.0314)

	// remove object
	idf.RemoveObject(zn)

	// ? Export

	err = idf.Save("basic_operation_example.idf")
}
