package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"sync"
	"time"
)

var serverPort string

func doConfig() {
	argv := os.Args
	argc := len(argv)
	fmt.Printf("argc %d\n", argc)
	if argc != 1 {
		for true {
			fmt.Printf("SLEEPING 60 Min\n")
			time.Sleep(60 * time.Minute)
		}
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		serverPort = ":8080"
	} else {
		serverPort = ":" + port
	}
}

func main() {
	doConfig()

	var rbot RedditBot
	//this is in seconds
	rbot.interval = 10 * 60
	//how many velocity history to keep
	rbot.maxIntervals = 8
	//max number of links to track
	rbot.maxRecords = 100

	//ths subreddit
	rbot.subreddit = "/r/WallStreetBets/new"

	var wg sync.WaitGroup
	fmt.Println("Main: starting redditBot")
	wg.Add(1)
	go (&rbot).start()

	r := gin.Default()
	r.GET("/top", func(c *gin.Context) {
		jsonResponse := (&rbot).top()

		c.Data(200, "application/json", jsonResponse)
	})
	r.GET("/topPlain", func(c *gin.Context) {
		response := (&rbot).topPlain()

		c.Data(200, "text/html", []byte(response))
	})

	r.Run(serverPort)

	fmt.Println("Main: Waiting for rbot & server to finish")
	wg.Wait()
	fmt.Println("Main: Completed, exit")
}
