package main

import (
	"github.com/mdfriday/hugoverse/cmd"
	"log"
)

func main() {
	log.SetFlags(0)
	err := cmd.New()
	if err != nil {
		log.Fatalf("\nError: %s", err)
	}
}
