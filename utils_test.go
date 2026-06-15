package epedit

import (
	"fmt"
	"testing"
)

func TestGetContinuousDigitsIndices(t *testing.T) {
	str := "asdf 12 asddf3"
	s, e := GetContinuousDigitsIndices(str)
	fmt.Printf(`"%s" "%s" "%s"\n`, str[:s], str[s:e], str[e:])
}
