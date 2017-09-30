package main

import (
	"log"
	"os"
	"sync"
)

var (
	wg sync.WaitGroup

	// Logger is used to print all NSQ messages
	Logger = log.New(os.Stderr, "", 0)
)
