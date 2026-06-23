package epedit

import (
	"fmt"
	"strconv"
)

// convert any type to string
func AnyToString(value any) string {
	switch v := value.(type) {
	case string:
		return v

	case bool:
		// EnergyPlus Yes/No convention
		if v {
			return "Yes"
		}
		return "No"

	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)

	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)

	// TODO: consider using 'g' for extremely small or big values
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		// prec -1: "15.0" -> "15", "15.123" -> "15.123"
		return strconv.FormatFloat(v, 'f', -1, 64)

	case nil:
		return ""

	default:
		// use default formatter
		return fmt.Sprintf("%v", v)
	}
}

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
