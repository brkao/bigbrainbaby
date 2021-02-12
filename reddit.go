package main

import (
	"encoding/json"
	"fmt"
	"github.com/turnage/graw/reddit"
	"sort"
	"time"
)

type RedditBot struct {
	bot          reddit.Bot
	subreddit    string
	interval     int32
	urlMap       map[string]*PostRecord
	recordList   []*PostRecord
	maxRecords   int32
	maxIntervals int32
}

type PostRecord struct {
	Url       string    `json:"url"`
	UpV       []float32 `json:"UpVelocity"`
	DownV     []float32 `json:"DownVelocity"`
	LastUp    int32     `json:"LastUpVote"`
	LastDown  int32     `json:"LastDownVote"`
	LastRatio float32   `json:"LastVoteRatio"`
}

func (p *PostRecord) printout() {
	fmt.Printf("\tURL[%s] LastUp[%d] LastDown[%d] LastRatio[%f]\n",
		p.Url, p.LastUp, p.LastDown, p.LastRatio)
	fmt.Printf("\tUp Velocity:\t")
	for _, val := range p.UpV {
		fmt.Printf("[%f] ", val)
	}
	fmt.Printf("\n")

	fmt.Printf("\tDown Velocity:\t")
	for _, val := range p.DownV {
		fmt.Printf("[%f] ", val)
	}
	fmt.Printf("\n")

}

func (p *PostRecord) MarshalJSON() ([]byte, error) {
	po := *p
	str, _ := json.Marshal(po)
	return str, nil
}

func (r *RedditBot) top() []byte {
	ret, _ := json.Marshal(r.recordList)
	return ret
}

func (r *RedditBot) topPlane() string {

	var ret string
	for i, p := range r.recordList {
		ret = ret + fmt.Sprintf("%d<br>", i)
		ret = ret + fmt.Sprintf("http://www.reddit.com%s<br>", p.Url)
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
		ret = ret + fmt.Sprintf("<br>")
	}
	return ret
}

func (r *RedditBot) getFreeRecord() *PostRecord {
	if len(r.recordList) >= int(r.maxRecords) {
		fmt.Printf("Max record numbers, deleting one\n")
		r.recordList = r.recordList[:len(r.recordList)-1]
	}
	var ret PostRecord
	ret.UpV = make([]float32, r.maxIntervals)
	ret.DownV = make([]float32, r.maxIntervals)
	return &ret
}

func (r *RedditBot) addNewRecord(po *PostRecord) {
	r.urlMap[po.Url] = po
	for i, record := range r.recordList {
		if record.UpV[0] <= po.UpV[0] {
			fmt.Printf("Inserting record into index[%d]\n", i)
			r.recordList = append(r.recordList, nil)
			copy(r.recordList[i+1:], r.recordList[i:])
			r.recordList[i] = po
			fmt.Printf("Added record, len %d\n", len(r.recordList))
			return
		} else {
			fmt.Printf("Record in array[%d] has higher velocity\n", i)
		}
	}
	r.recordList = append(r.recordList, po)
	fmt.Printf("Adding record to tail, len %d\n", len(r.recordList))
}

func (r *RedditBot) start() {
	b, err := reddit.NewBotFromAgentFile(".profile", 0)
	if err != nil {
		fmt.Println("Failed to create bot handle:  ", err)
		return
	}
	r.bot = b
	r.urlMap = make(map[string]*PostRecord)
	fmt.Printf("Bot Start: maxIntervals %d interval %d maxRecords %d\n",
		r.maxIntervals, r.maxRecords)
	r.harvest()

	ticker := time.NewTicker(time.Duration(r.interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			fmt.Println("Timer, harvesting")
			r.harvest()
			fmt.Println("Harvest done")
			for _, val := range r.recordList {
				val.printout()
			}
			fmt.Println("\n")
		}
	}
}

func (r *RedditBot) harvest() {
	//harvest, err := r.bot.Listing(r.subreddit, "")
	var params map[string]string
	params = make(map[string]string)
	//params["sort"] = "new"

	harvest, err := r.bot.ListingWithParams(r.subreddit, params)
	if err != nil {
		fmt.Println("Failed to fetch ", err)
		return
	}

	for _, post := range harvest.Posts {
		if post.LinkFlairText != "DD" {
			continue
		}

		var down int32
		var up int32
		down = post.Score - int32(float64(post.Score)*float64(post.UpvoteRatio))
		up = int32(float64(post.Score) * float64(post.UpvoteRatio))
		fmt.Printf("Permalink[%s] Score[%d] Ratio [%f] Ups[%d] Downs[%d]\n",
			post.Permalink, post.Score, post.UpvoteRatio, up, down)

		if val, ok := r.urlMap[post.Permalink]; ok {
			UpV := float32((up - val.LastUp))
			if UpV < 0 {
				UpV = 0
			} else {
				UpV = UpV / float32(r.interval)
			}

			DownV := float32((down - val.LastDown))
			if DownV <= 0 {
				DownV = 0
			} else {
				DownV = DownV / float32(r.interval)
			}

			fmt.Printf("Found Permalink in map UpV[%f] DownV[%f] LastRatio[%f]\n",
				UpV, DownV, post.UpvoteRatio)
			val.LastRatio = post.UpvoteRatio
			val.LastDown = down
			val.LastUp = up

			copy(val.UpV[1:], val.UpV[0:len(val.UpV)-2])
			copy(val.DownV[1:], val.DownV[0:len(val.DownV)-2])

			val.UpV[0] = UpV
			val.DownV[0] = DownV

			sort.Slice(r.recordList[:], func(i, j int) bool {
				return r.recordList[i].UpV[0] >= r.recordList[j].UpV[0]
			})

		} else {
			po := r.getFreeRecord()
			po.Url = post.Permalink
			po.LastUp = up
			po.LastDown = down
			po.LastRatio = post.UpvoteRatio

			r.addNewRecord(po)
		}
	}
}
