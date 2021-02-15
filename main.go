package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"os"
	"sort"
	"time"
)

var serverPort string
var serverStartTime time.Time

type DDVelocity struct {
	RecordList []*PostRecord `json:"RecordList"`
	LastUpdate time.Time     `json:"Timestamp"`
}
type PostRecord struct {
	Url       string    `json:"url"`
	UpV       []float32 `json:"UpVelocity"`
	DownV     []float32 `json:"DownVelocity"`
	LastUp    int32     `json:"LastUpVote"`
	LastDown  int32     `json:"LastDownVote"`
	LastRatio float32   `json:"LastVoteRatio"`
}

type SentimentScore struct {
	Neg      string `json:"neg"`
	Neu      string `json:"neu"`
	Pos      string `json:"pos"`
	Compound string `json:"compound"`
	Count    int    `json:"count"`
}

type SentimentMap struct {
	Timestamp string                    `json:"timestamp"`
	Tickers   map[string]SentimentScore `json:"tickers"`
}

type Sentiment struct {
	Ticker string
	Score  SentimentScore
}
type SentimentByCount []Sentiment

func (a SentimentByCount) Len() int {
	return len(a)
}

func (a SentimentByCount) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a SentimentByCount) Less(i, j int) bool {
	return a[i].Score.Count > a[j].Score.Count
}

func doConfig() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		serverPort = ":8080"
	} else {
		serverPort = ":" + port
	}
}

func getLatestSentiment() (*SentimentMap, error) {
	var s SentimentMap

	c, err := redis.DialURL(os.Getenv("REDIS_URL"), redis.DialTLSSkipVerify(true))
	if err != nil {
		fmt.Printf("Error connecting to REDIS\n")
		return nil, err
	}
	defer c.Close()

	value, err := redis.String(c.Do("LINDEX", "sentiment_scores", "0"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = json.Unmarshal([]byte(value), &s)
	if err != nil {
		fmt.Printf("Unmarshal err: %v\n", err)
		return nil, err
	}

	return &s, nil

}
func getLatestDDVelocity() (*DDVelocity, error) {
	var v DDVelocity

	c, err := redis.DialURL(os.Getenv("REDIS_URL"), redis.DialTLSSkipVerify(true))
	if err != nil {
		fmt.Printf("Error connecting to REDIS\n")
		return nil, err
	}
	defer c.Close()

	value, err := redis.String(c.Do("LINDEX", "dd_velocities", "0"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	err = json.Unmarshal([]byte(value), &v)
	if err != nil {
		fmt.Printf("Unmarshal err\n")
		return nil, err
	}

	return &v, nil
}

func showIndexPage(c *gin.Context) {
	// Call the HTML method of the Context to render a template
	c.HTML(
		// Set the HTTP status to 200 (OK)
		http.StatusOK,
		// Use the index.html template
		"index.html",
		// Pass the data that the page uses
		gin.H{
			"title":   "Index Page",
			"payload": "Index",
		},
	)
}

func showSentimentPage(c *gin.Context) {
	var sentimentArr []Sentiment

	s, err := getLatestSentiment()
	if err == nil {
		for k, v := range s.Tickers {
			sentimentArr = append(sentimentArr, Sentiment{k, v})
		}
		sort.Sort(SentimentByCount(sentimentArr))
		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"sentiment.html",
			// Pass the data that the page uses
			gin.H{
				"title":     "Sentiment Page",
				"timestamp": s.Timestamp,
				"payload":   sentimentArr,
			},
		)
	} else {
		c.JSON(200, gin.H{
			"status":  "Error",
			"message": "Sentiment Not Available",
		})
	}
}

func showVelocityPage(c *gin.Context) {
	ddv, _ := getLatestDDVelocity()
	// Call the HTML method of the Context to render a template
	c.HTML(
		// Set the HTTP status to 200 (OK)
		http.StatusOK,
		// Use the index.html template
		"velocity.html",
		// Pass the data that the page uses
		gin.H{
			"title":   "Velocity Page",
			"payload": ddv,
		},
	)
}

func initRoutes(r *gin.Engine) {

	r.GET("/", showIndexPage)
	r.GET("/velocity", showVelocityPage)
	r.GET("/sentiment", showSentimentPage)
}

func main() {
	doConfig()
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")

	initRoutes(r)

	r.Run(serverPort)

	fmt.Println("Main: Completed, exit")
}
