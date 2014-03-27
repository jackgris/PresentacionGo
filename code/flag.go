// +build OMIT

package main

import (
	"flag"
	"fmt"
	"time"
)

var (
	message = flag.String("message", "Hola!", "que hay para decir")
	delay   = flag.Duration("delay", 2*time.Second, "cuanto tiempo debo esperar")
)

func main() {
	flag.Parse()
	fmt.Println(*message)
	time.Sleep(*delay)
}
