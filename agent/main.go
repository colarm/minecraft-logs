package main

import (
	"log"
	"minecraft-log-agent/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
