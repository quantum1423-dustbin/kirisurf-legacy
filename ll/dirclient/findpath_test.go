package dirclient

import (
	"fmt"
	"testing"
)

func TestPathGroup(t *testing.T) {
	RefreshDirectory()
	fmt.Println(FindPathGroup(5))
}
