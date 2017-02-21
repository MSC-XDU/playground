package server

import (
	"context"
	"testing"
)

const code = `package main
import "fmt"

func main() {
	fmt.Println("Hello")
}`

var url string

func TestSaveCode(t *testing.T) {
	var err error
	url, err = SaveCode(context.Background(), code, TypeGo)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetCode(t *testing.T) {
	c, tp, err := GetCode(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}

	if tp != TypeGo {
		t.Fatalf("got incorrent type %d excepted %d", tp, TypeGo)
	}

	if c != code {
		t.Fatalf("got incorrent code %s excepted %s", c, code)
	}
}
