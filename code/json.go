// +build OMIT

package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const blob = `[
	{"Titulo":"Øredev", "URL":"http://oredev.org"},
	{"Titulo":"Strange Loop", "URL":"http://thestrangeloop.com"}
]`

type Item struct {
	Titulo string
	URL    string
}

func main() {
	var items []*Item
	json.NewDecoder(strings.NewReader(blob)).Decode(&items)
	for _, item := range items {
		fmt.Printf("Titulo: %v URL: %v\n", item.Titulo, item.URL)
	}
}
