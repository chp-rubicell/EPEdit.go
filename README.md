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
