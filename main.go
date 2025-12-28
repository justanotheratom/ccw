package main

import (
	"log"

	"github.com/ccw/ccw/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("ccw: %v", err)
	}
}
