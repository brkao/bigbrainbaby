package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

var timeZone string = "America/New_York"

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
	Neg           float32 `json:"neg"`
	Neu           float32 `json:"neu"`
	Pos           float32 `json:"pos"`
	Compound      float32 `json:"compound"`
	Count         int     `json:"count"`
	CompoundDelta float32
	CountDelta    int
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

func (a SentimentByCount) Len() int           { return len(a) }
func (a SentimentByCount) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SentimentByCount) Less(i, j int) bool { return a[i].Score.Count > a[j].Score.Count }

type SentimentByCountDelta []Sentiment

func (a SentimentByCountDelta) Len() int      { return len(a) }
func (a SentimentByCountDelta) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SentimentByCountDelta) Less(i, j int) bool {
	return a[i].Score.CountDelta > a[j].Score.CountDelta
}

type SentimentByCompoundDelta []Sentiment

func (a SentimentByCompoundDelta) Len() int      { return len(a) }
func (a SentimentByCompoundDelta) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SentimentByCompoundDelta) Less(i, j int) bool {
	return a[i].Score.CompoundDelta > a[j].Score.CompoundDelta
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
	var s1 SentimentMap

	c, err := redis.DialURL(os.Getenv("REDIS_URL"), redis.DialTLSSkipVerify(true))
	if err != nil {
		fmt.Printf("Error connecting to REDIS\n")
		return nil, err
	}
	defer c.Close()

	entries, err := redis.Int(c.Do("LLEN", "sentiment_scores"))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	//if the db doesn't have any entries, return blank
	if entries == 0 {
		return &s, nil
	} else if entries == 1 {
		//if there is one entry, the do not compute delta
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
	} else {
		v1, err := redis.String(c.Do("LINDEX", "sentiment_scores", "0"))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		err = json.Unmarshal([]byte(v1), &s)
		if err != nil {
			fmt.Printf("Unmarshal err: %v\n", err)
			return nil, err
		}

		v2, err := redis.String(c.Do("LINDEX", "sentiment_scores", "1"))
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		err = json.Unmarshal([]byte(v2), &s1)
		if err != nil {
			fmt.Printf("Unmarshal err: %v\n", err)
			return nil, err
		}
		for symbol, sentiment := range s.Tickers {
			oldSentiment, ok := s1.Tickers[symbol]
			if ok {
				sentiment.CountDelta = sentiment.Count - oldSentiment.Count
				sentiment.CompoundDelta = sentiment.Compound - oldSentiment.Compound
				s.Tickers[symbol] = sentiment
			}
		}
	}
	return &s, nil
}

func getLatestDDVelocity() (*DDVelocity, error) {
	var v DDVelocity
	loc, _ := time.LoadLocation(timeZone)

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

	v.LastUpdate = v.LastUpdate.In(loc)
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

func showTopSentimentPage(c *gin.Context) {
	var byCountArr []Sentiment
	p1 := c.Query("count")
	count, err := strconv.Atoi(p1)
	if err != nil {
		count = 5
	}

	s, err := getLatestSentiment()
	if err == nil {
		for k, v := range s.Tickers {
			byCountArr = append(byCountArr, Sentiment{k, v})
		}
		byCompoundArr := make([]Sentiment, len(byCountArr))
		copy(byCompoundArr, byCountArr)
		sort.Sort(SentimentByCountDelta(byCountArr))
		sort.Sort(SentimentByCompoundDelta(byCompoundArr))

		byCountArr = byCountArr[:count]
		byCompoundArr = byCompoundArr[:count]

		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"topSentiment.html",
			// Pass the data that the page uses
			gin.H{
				"title":           fmt.Sprintf("Top %d Sentiment Page", count),
				"timestamp":       s.Timestamp,
				"byCountDelta":    byCountArr,
				"byCompoundDelta": byCompoundArr,
			},
		)
	} else {
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"error.html",
			// Pass the data that the page uses
			gin.H{
				"title":   "Error Page",
				"message": err.Error(),
			},
		)
	}
}

func getSentiment(c *gin.Context) {
	var sentimentArr []Sentiment

	s, err := getLatestSentiment()
	if err == nil {
		for k, v := range s.Tickers {
			sentimentArr = append(sentimentArr, Sentiment{k, v})
		}
		sort.Sort(SentimentByCount(sentimentArr))
		c.JSON(
			http.StatusOK,
			gin.H{
				"status":    "success",
				"timestamp": s.Timestamp,
				"message":   sentimentArr,
			},
		)
	} else {
		c.JSON(
			http.StatusOK,
			gin.H{
				"title":   "Error Page",
				"message": err.Error(),
			},
		)
	}
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
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"error.html",
			// Pass the data that the page uses
			gin.H{
				"title":   "Error Page",
				"message": err.Error(),
			},
		)
	}
}

func showVelocityPage(c *gin.Context) {
	ddv, err := getLatestDDVelocity()
	if err != nil {
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"error.html",
			// Pass the data that the page uses
			gin.H{
				"title":   "Error Page",
				"message": err.Error(),
			},
		)
		return
	}
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
	r.GET("/sentiment/raw", getSentiment)

	r.GET("/topSentiment", showTopSentimentPage)

}

func main() {
	doConfig()
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://tpb-admin.netlify.app", "http://localhost:3000"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://tpb-admin.netlify.app"
	    	},
		MaxAge: 12 * time.Hour,
	}))
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./assets")

	initRoutes(r)

	r.Run(serverPort)

	fmt.Println("Main: Completed, exit")
}
