package main

import (
	"fmt"
	"log"
	"time"

	. "github.com/noypi/election2016/rappler"
	//"time"
)

func main() {
	for loc := 1; loc <= 1735; loc++ {
		check(2, loc)
	}
}

type CheatDetected struct {
	Prev *JsBallotHistory
	Curr *JsBallotHistory
}

func check(id, loc int) {
	//log.Println("================ location =------------", loc)
	data := GetHistLog("../save-histlog/histlog", id, loc)
	if nil == data {
		log.Fatal("histlog not found id=", id, ";loc=", loc)
	}

	for _, candidateData := range data.Contest_data {
		var cheats []*CheatDetected
		var prev *JsBallotHistory
		var prevDate, prevVoteCount int64
		var bFirst = true
		//log.Println("processing", candidateData.Ballot_name)
		for _, histlog := range candidateData.Values {
			if bFirst {
				prevDate = histlog.Date
				prevVoteCount = histlog.Vote_count
				bFirst = false
				continue
			}
			//log.Printf("prevDate=%v, currDate=%v\n", time.Unix(prevDate/1000, 0), time.Unix(histlog.Date/1000, 0))
			if prevDate > histlog.Date {
				log.Printf("invalid data detected prevDate=%v, currDate=%v\n", prevDate, histlog.Date)
			}
			//log.Printf("prevVoteCount=%v, Vote_count=%v\n", prevVoteCount, histlog.Vote_count)
			if prevVoteCount > histlog.Vote_count {
				//log.Printf("Cheat Detected [%s:loc=%d] prevVoteCount=%d, currDate=%v\n", candidateData.Ballot_name, loc, prevVoteCount, histlog.Date)
				cheats = append(cheats, &CheatDetected{
					Prev: prev,
					Curr: histlog,
				})
			}

			prevDate = histlog.Date
			prevVoteCount = histlog.Vote_count
			prev = histlog
		}
		if 0 < len(cheats) {
			dumpCheats(id, loc, cheats)
		}
	}

}

var g_locph, _ = time.LoadLocation("Asia/Manila")

func formatDate(n int64) string {
	if nil == g_locph {
		log.Fatal("timezone manila not loaded")
	}
	t := time.Unix(n/1000, 0).In(g_locph)
	return t.Format("'1/2 03:04 pm'")
}
func dumpCheats(id, loc int, cheats []*CheatDetected) {
	oloc := GetLocation("../save-histlog/loc", loc)
	if nil == oloc {
		log.Fatal("location not found=", loc)
	}

	url := fmt.Sprintf("http://ph.rappler.com/api/Votes/live/history/contest/id/%d/location/id/%d", id, loc)

	fmt.Println("\n=============================")
	fmt.Printf("cheat detected @%s\n", oloc.Location_name)
	fmt.Println("reference: ", url)

	for _, cheat := range cheats {
		fmt.Println("candidate:\t", cheat.Curr.Slug)
		fmt.Printf("Before:\t%d votes \t%v (raw:%v)\n", cheat.Prev.Vote_count, formatDate(cheat.Prev.Date), cheat.Prev.Date)
		fmt.Printf("After:\t%d votes \t%v (raw:%v)\n", cheat.Curr.Vote_count, formatDate(cheat.Curr.Date), cheat.Curr.Date)
		fmt.Println("===> difference of ", cheat.Prev.Vote_count-cheat.Curr.Vote_count, " votes decreased.")
		fmt.Println("___________________________")
	}

}
