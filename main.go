package main

import (
	"log"
	"os"

	"github.com/gonejack/tumblr-likes/cmd"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	err := cmd.New().Run()
	if err != nil {
		log.Fatal(err)
	}
}
