// +build OMIT

package main

import (
	"fmt"
	"time"
)

func main() {
	// START OMIT
	t := time.Now()
	fmt.Println(t.In(time.UTC))
	home, _ := time.LoadLocation("America/Buenos_Aires")
	fmt.Println(t.In(home))
	// END OMIT
}
