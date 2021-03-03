package util

import (
	"fmt"
	"reflect"
)

const tagName = "tag"

type A struct {
	A1 string `json:"a1" tag:"t1"`
	A2 string `json:"a2" tag:"t2"`
	A3 string `json:"a3" tag:"t3"`
}

type B struct {
	B1 string `json:"b1" tag:"t1"`
	B2 string `json:"b2" tag:"t2"`
	B3 string `json:"b3" tag:"t3"`
}

func convertAtoB(a A, b B) {
	t := reflect.TypeOf(a)

	fmt.Println("Type:", t.Name())
	fmt.Println("Kind:", t.Kind())

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagName)
		fmt.Printf("%d. %v (%v), tag: '%v'\n", i+1, field.Name, field.Type.Name(), tag)
	}
}

func main() {
	fmt.Println("Hello, playground")
	a, b := A{A1: "1", A2: "2", A3: "3"}, B{}
	convertAtoB(a, b)

}
