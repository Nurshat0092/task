package main

import (
	"errors"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	log "github.com/sirupsen/logrus"
)

var errEmptyInput = errors.New("no input provided in buffer")

type application struct {
	stats  map[int]int
	numCPU int
}

func main() {
	app := application{
		stats: make(map[int]int),
	}
	// Channel to handle CTRL-C, and print existing statistics
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		app.printStats()
		os.Exit(1)
	}()

	numCPU := runtime.NumCPU()
	log.Infof("number of cpu: %d\n", numCPU)
	runtime.GOMAXPROCS(numCPU)
	urls, err := readInput()
	if err != nil {
		log.Fatal(err)
	}
	if numCPU > len(urls) {
		numCPU = len(urls)
	}
	app.numCPU = numCPU
	app.distributedRequests(urls)
	app.printStats()
}
