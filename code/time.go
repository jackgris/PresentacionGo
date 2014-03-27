// +build OMIT

package main

import (
	"fmt"
	"time"
)

func main() {
	// START OMIT
	if time.Now().Hour() < 12 {
		fmt.Println("Buen dia.")
	} else {
		fmt.Println("Buenas tardes (o noches).")
	}
	// END OMIT
}
