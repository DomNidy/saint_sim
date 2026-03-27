package main

import (
	"log"

	api "github.com/DomNidy/saint_sim/apps/api"
)

func main() {
	if err := api.Run(); err != nil {
		log.Fatal(err)
	}
}
