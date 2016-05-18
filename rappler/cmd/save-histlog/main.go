package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	. "github.com/noypi/election2016/rappler"
)

func main() {
	//save2loc()
	savehist()
}

func save2loc() {
	c := make(chan bool, 3)

	var ndone int32
	for loc := 1736; loc <= 1740; loc++ {
		c <- true
		go func(_loc int) {
			saveLoc("loc", _loc)
			<-c
			atomic.AddInt32(&ndone, 1)
			if 1735 == ndone {
				close(c)
			}

		}(loc)
	}

}

func savehist() {
	c := make(chan bool, 3)

	for id := 1; id <= 3; id++ {
		// 1735 max
		var ndone int32
		for loc := 1; loc <= 1736; loc++ {

			oloc := GetLocation("loc", 1)
			ohistlog := GetHistLog("histlog", id, loc)
			if nil != ohistlog {
				if ohistlog.Timestamp >= oloc.Timestamp {
					//log.Printf("have latest id=%d, loc=%d\n", id, loc)
					if 3 == id && 1736 == ndone {
						close(c)
						break
					}

					continue
				} else {
					log.Printf("need to update id=%d, loc=%d, ohistlog.timestamp=%d oloc.Timestamp=%d, \n",
						id, loc, ohistlog.Timestamp, oloc.Timestamp)
				}
			}

			c <- true
			go func(_id, _loc int) {
				saveHistlog("histlog", _id, _loc)
				<-c
				atomic.AddInt32(&ndone, 1)
				if 3 == id && 1736 == ndone && 0 == len(c) {
					close(c)
				}

			}(id, loc)
		}
	}

	for {
		time.Sleep(10 * time.Second)
		if 0 == len(c) {
			break
		}
	}
}

func saveHistlog(path string, id, loc int) {
	fpath := fmt.Sprintf("%s/%d-%d.log", path, id, loc)

	log.Println("creating:", fpath)

	url := fmt.Sprintf("http://ph.rappler.com/api/Votes/live/history/contest/id/%d/location/id/%d", id, loc)
	res, err := http.Get(url)
	if nil != err {
		log.Fatal("get err=", err)
	}

	bb, err := ioutil.ReadAll(res.Body)
	if nil != err {
		log.Fatal("read body err=", err)
	}

	log.Println("status=", res.Status, "len=", len(bb), "fpath=", fpath)

	err = ioutil.WriteFile(fpath, bb, os.ModePerm)
	log.Printf("write file=%s err=%v", fpath, err)

}

func saveLoc(path string, loc int) {
	fpath := fmt.Sprintf("%s/loc-%d.log", path, loc)

	f, err := os.Stat(fpath)
	if os.IsExist(err) || (nil != f) {
		return
	}

	log.Println("creating:", fpath)

	url := fmt.Sprintf("http://ph.rappler.com/api/Votes/live/location/id/%d", loc)
	res, err := http.Get(url)
	if nil != err {
		log.Fatal("get err=", err)
	}

	bb, err := ioutil.ReadAll(res.Body)
	if nil != err {
		log.Fatal("read body err=", err)
	}

	log.Println("status=", res.Status, "len=", len(bb))

	ioutil.WriteFile(fpath, bb, os.ModePerm)

}
