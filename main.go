package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"sync"
)

var serverPort string

func doConfig() {
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
	rbot.interval = 5 * 60
	//how many velocity history to keep
	rbot.maxIntervals = 6
	//max number of links to track
	rbot.maxRecords = 100

	//ths subreddit
	rbot.subreddit = "/r/TrailerParkBets/new"

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
		response := (&rbot).topPlane()

		c.Data(200, "text/plain", []byte(response))
	})

	r.StaticFile("/top.htm", "./www/top.htm")

	r.Run(serverPort)

	fmt.Println("Main: Waiting for rbot & server to finish")
	wg.Wait()
	fmt.Println("Main: Completed, exit")
}
