package election2016

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func GetHistLog(path string, id, loc int) *JsData {
	bb, err := ioutil.ReadFile(fmt.Sprintf("%s/%d-%d.log", path, id, loc))
	if os.IsNotExist(err) {
		log.Fatal("GetHistLog err=", err)
	}
	if nil != err {
		log.Println("getCurrLog ReadFile err=", err)
		return nil
	}

	var histlogs []*JsData
	if err = json.Unmarshal(bb, &histlogs); nil != err || 0 == len(histlogs) {
		log.Println("getCurrLog unmarshall err=", err)
		return nil
	}

	var histlog = histlogs[0]

	if 0 >= histlog.Timestamp {
		log.Println("invalid data id=", id, ";loc=", loc)
		return nil
	}

	return histlog
}

func GetLocation(path string, loc int) *JsLoc {
	bb, err := ioutil.ReadFile(fmt.Sprintf("%s/loc-%d.log", path, loc))
	if nil != err {
		log.Fatal("GetLocation ReadFile err=", err)
	}

	var olocs []*JsLoc
	if err = json.Unmarshal(bb, &olocs); nil != err || 0 == len(olocs) {
		log.Fatal("GetLocation unmarshall err=", err)
	}

	var oloc = olocs[0]
	if 0 == len(oloc.Location_name) || 0 >= oloc.Location_id {
		log.Fatal("GetLocation invalid data loc=", loc)
	}

	return oloc
}
