package main

import (
	"log"
	"os"
	"sync"
)

var (
	wg sync.WaitGroup

	Version        = "0.1"
	CommitRevision = ""
	ShowVersion    = false

	// Logger is used to print all NSQ messages
	Logger = log.New(os.Stderr, "", 0)
)
