package main

import (
	"fmt"
	"reflect"
)

type Cat struct {
	Color string
}

func main() {
	cat := &Cat{"red"}

	fmt.Println("值", reflect.ValueOf(cat))
	fmt.Println("类型", reflect.TypeOf(cat))
}
