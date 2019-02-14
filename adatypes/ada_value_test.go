package adatypes

import (
	"fmt"
	"testing"
)

func TestAdaValue(t *testing.T) {
	value := doubleValue{adaValue: adaValue{}}

	var x IAdaValue
	var a adaValue

	x = &value
	a = value.adaValue
	fmt.Println(x, a)
}
