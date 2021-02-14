package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"os"
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

func doConfig() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		serverPort = ":8080"
	} else {
		serverPort = ":" + port
	}
}

func (d *DDVelocity) topPlain() string {
	var ret string

	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		fmt.Println(err)
		return ""
	}
	ret = ret + fmt.Sprintf("<html><body>")
	ret = ret + fmt.Sprintf("Server started on: %s<br>\n", fmt.Sprintf((serverStartTime.In(loc)).String()))
	for i, p := range d.RecordList {

		ret = ret + fmt.Sprintf("--->  %d <--- <br>", i)
		ret = ret + fmt.Sprintf("<a href=\"http://www.reddit.com%s\">%s</a><br>", p.Url, p.Url)
		ret = ret + fmt.Sprintf("LastUp[%d] LastDown[%d] LastRatio[%f]<br>",
			p.LastUp, p.LastDown, p.LastRatio)
		ret = ret + fmt.Sprintf("\tU Velocity:\t")
		for _, val := range p.UpV {
			ret = ret + fmt.Sprintf("[%f] ", val)
		}
		ret = ret + fmt.Sprintf("<br>")

		ret = ret + fmt.Sprintf("\tD Velocity:\t")
		for _, val := range p.DownV {
			ret = ret + fmt.Sprintf("[%f] ", val)
		}
		ret = ret + fmt.Sprintf("<br><br>")
	}
	ret = ret + fmt.Sprintf("</html></body>")
	return ret
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

func main() {
	doConfig()
	r := gin.Default()

	r.GET("/topPlain", func(c *gin.Context) {
		d, _ := getLatestDDVelocity()
		response := d.topPlain()
		c.Data(200, "text/html", []byte(response))
	})

	r.Run(serverPort)

	fmt.Println("Main: Completed, exit")
}
