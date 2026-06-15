package epedit

// get start and end indices of continous digits (first appearance)
func GetContinuousDigitsIndices(name string) (startIndex int, endIndex int) {
	startIndex, endIndex = -1, -1

	for byteIndex, char := range name {
		if char >= '0' && char <= '9' {
			if startIndex == -1 {
				startIndex = byteIndex
			}
		} else if startIndex != -1 {
			endIndex = byteIndex
			break
		}
	}

	if startIndex != -1 {
		if endIndex == -1 {
			endIndex = len(name) // if ends with number (ex. "Vertex 1")
		}
	}

	return startIndex, endIndex
}
