package main

import (
	"fmt"
	"log"

	epedit "github.com/chp-rubicell/EPEdit.go"
)

func main() {
	// Load the EnergyPlus IDD schema.
	idd, err := epedit.ParseIDDFile("Energy+.idd")
	if err != nil {
		log.Fatal(err)
	}

	// Load an existing IDF file using the parsed IDD.
	idf, err := epedit.ParseIDFFile("input.idf", idd)
	if err != nil {
		log.Fatal(err)
	}

	// Find an object by class name and object name.
	building, err := idf.GetObjectByName("Building", "My Building")
	if err != nil {
		log.Fatal(err)
	}

	// Update fields by their IDD field names.
	if err := building.Set("North Axis", 0.0); err != nil {
		log.Fatal(err)
	}
	if err := building.Set("Terrain", "City"); err != nil {
		log.Fatal(err)
	}

	// Update multiple fields
	err = building.Update(epedit.Fields{
		"Maximum Number of Warmup Days": 50,
		"Minimum Number of Warmup Days": 5,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(building)

	// Add a new object with initial field values.
	_, err = idf.AddObject("RunPeriod", epedit.Fields{
		"Name":                      "New Annual Run",
		"Begin Month":               1,
		"Begin Day of Month":        1,
		"End Month":                 12,
		"End Day of Month":          31,
		"Day of Week for Start Day": "Monday",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Add a new object without default values.
	_, err = idf.AddObject(
		"Material",
		epedit.Fields{
			"Name":          "New Insulation",
			"Thickness":     0.05,
			"Conductivity":  0.0314,
			"Density":       265,
			"Specific Heat": 836.8,
		},
		false,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Remove an object
	obj, err := idf.GetObjectByName("RunPeriod", "annual")
	if err != nil {
		log.Fatal(err)
	}
	if err := idf.RemoveObject(obj); err != nil {
		log.Fatal(err)
	}

	// Save the modified IDF file.
	if err := idf.Save("output.idf"); err != nil {
		log.Fatal(err)
	}
}
