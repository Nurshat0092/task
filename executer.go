package main

import (
	"bufio"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// prints statistics of completed requests
func (app *application) printStats() {
	log.Infof("\nStatistics:\n")
	if len(app.stats) == 0 {
		log.Infof("No request completed")
	}
	for key, val := range app.stats {
		log.Infof("Thread '%d' performed %d requests\n", key, val)
	}
}

// reads from os.Stdin buffer. While reading if error is io.EOF return nil
// if no input provided return ErrEmptyInput
// to prevent reading press CTRL-D
func readInput() ([]string, error) {
	urls := []string{}
	buff := bufio.NewReader(os.Stdin)
	for {
		url, err := buff.ReadString('\n')
		if err == io.EOF {
			if len(urls) == 0 {
				return urls, errEmptyInput
			}
			return urls, nil
		} else if err != nil {
			return urls, err
		}
		url = strings.TrimSuffix(url, "\n")
		urls = append(urls, url)
	}
}

// divides array of urls according to Core number.
// and perform GET-requests to given URLs
func (app *application) distributedRequests(urls []string) {
	wg := sync.WaitGroup{}
	chunk := len(urls) / app.numCPU
	remainder := len(urls) % app.numCPU
	if chunk == 0 {
		app.numCPU = len(urls)
	}
	wg.Add(app.numCPU)
	start, end := 0, chunk
	for i := 0; i < app.numCPU; i++ {
		if remainder > 0 {
			end++
			remainder--
		}
		go func(urls []string, threadID int) {
			defer wg.Done()
			for _, u := range urls {
				app.doRequest(u, threadID)
			}
		}(urls[start:end], i)
		start = end
		end += chunk
	}
	wg.Wait()
}

// makes http GET request to given URL, and logs response information
// format: "url;statusCode;size;RequestCompletedTime"
func (app *application) doRequest(url string, threadID int) {
	logErrFields := log.Fields{
		"url":       url,
		"processID": threadID,
	}
	reqTime := time.Now()
	res, err := http.Get(url)
	if err != nil {
		logErrFields["func"] = "http.Get()"
		log.WithFields(logErrFields).Error(err)
		return
	}
	resTime := time.Now()
	diff := resTime.Sub(reqTime).Milliseconds()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logErrFields["func"] = "ioutil.ReadAll()"
		log.WithFields(logErrFields).Error(err)
		return
	}
	size := len(body)
	app.stats[threadID]++
	log.Infof("%s;%d;%d;%dms", url, res.StatusCode, size, diff)
}
