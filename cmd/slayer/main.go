package main

import (
	"log"

	"github.com/prabalesh/slayer/internal/shell"
	"github.com/prabalesh/slayer/internal/store"
)

func main() {
	s, err := store.NewStore()
	if err != nil {
		log.Fatal("[ERROR]: unable to initilaize store, error : ", err)
	}
	shellSession := shell.NewShell(*s)
	shellSession.Start()
}
