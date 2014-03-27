// +build OMIT

package main

import (
	"fmt"
	"time"
)

func main() {
	// START OMIT
	birthday, _ := time.Parse("Mar 2 2014", "Nov 10 2009") // time.Time
	age := time.Since(birthday)                            // time.Duration
	fmt.Printf("Go tiene %d dias desde su nacimiento\n", age/(time.Hour*24))
	// END OMIT
}
