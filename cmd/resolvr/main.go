package main

import (
	"fmt"
	"github.com/RileyMichael/resolvr/internal/resolvr"
	"log"
)

func main() {
	config, err := resolvr.LoadConfig()

	if err != nil {
		log.Fatal("error loading config", err)
	}
	fmt.Println(config)
}
