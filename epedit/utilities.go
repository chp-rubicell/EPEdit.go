package epedit

import "os"

// readFileToString reads a whole file into a string.
// It returns the content as a string and an error if one occurred.
func ReadFileToString(filePath string) (string, error) {
	// Use os.ReadFile to get the content as a byte slice.
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		// If an error occurred, return an empty string and the error.
		return "", err
	}

	// Convert the byte slice to a string and return it with a nil error.
	return string(contentBytes), nil
}
