package main

import (
	"fmt"

	"github.com/gowok/gowok"
)

func main() {
	conf, err := gowok.Configure("config.yaml")
	fmt.Println(conf, err)
}
