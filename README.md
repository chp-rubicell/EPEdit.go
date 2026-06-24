<p align="center">
    <a href="https://github.com/chp-rubicell/EPEdit.go/releases/latest">
        <img src="https://github.com/chp-rubicell/EPEdit.go/blob/main/_assets/epeditgo.svg" width="256" alt="EPEdit.go"><br/>
    </a>
    <!-- <img src="doc/epedit.svg" width="256" alt="EPEdit.go"><br/> -->
    <a href="https://github.com/chp-rubicell/EPEdit.go/releases/latest"><img src="https://img.shields.io/github/release/chp-rubicell/EPEdit.go.svg?style=flat-square&maxAge=600" alt="Downloads"></a>
</p>

**EPEdit.go** is a Go library for parsing, editing, and formatting EnergyPlus Input Data Files (`.idf`).

## Features

- **Parse IDF Files**: Load `.idf` file content into a structured object model.
- **Modify IDF**: Create, update, or delete any object within the IDF model.
- **Find IDF Objects**: Easily find and retrieve objects by their type (e.g., `Building`, `Material`) and name.
- **Modify Fields**: Get and set values for any field of an IDF object.
- **Export to IDF**: Serialize the modified model back into a valid `.idf` file string.


## Usage

### Open IDD and IDF files

```go
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
```

### Find and update objects

```go
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
```
```
Building,
    My Building,              !- Name
    0,                        !- North Axis {deg}
    City,                     !- Terrain
    0.0400,                   !- Loads Convergence Tolerance Value {W}
    0.2000,                   !- Temperature Convergence Tolerance Value {deltaC}
    FullInteriorAndExterior,  !- Solar Distribution
    50,                       !- Maximum Number of Warmup Days
    5;                        !- Minimum Number of Warmup Days
```

### Add new objects

```go
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
```

### Remove objects

```go
// Remove an object
obj, err := idf.GetObjectByName("RunPeriod", "annual")
if err != nil {
	log.Fatal(err)
}
if err := idf.RemoveObject(obj); err != nil {
	log.Fatal(err)
}
```

### Save IDF files
```go
// Save the modified IDF file.
if err := idf.Save("output.idf"); err != nil {
	log.Fatal(err)
}
```


## Related projects

- [EPEdit.js](https://github.com/chp-rubicell/EPEdit.js)


## License

Distributed under the [MIT License](https://github.com/chp-rubicell/EPEdit.go/blob/main/LICENSE).
