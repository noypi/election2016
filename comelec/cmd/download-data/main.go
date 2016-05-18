package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"

	"log"

	"github.com/barakmich/glog"
	"github.com/joinhack/fqueue"
	"github.com/noypi/fp"

	"github.com/noypi/kv"
	"github.com/noypi/kv/leveldb"
)

type JsRegion struct {
	CategoryName string
	Name         string
	Level        int
	Contests     []*JsContest
	SubRegions   map[string]*JsSubRegion
}

type JsContest struct {
	Url string
}

type JsSubRegion struct {
	CategoryName string
	Name         string
	Level        int
	Url          string
}

const g_baseurl = "https://www.pilipinaselectionresults2016.com/"

var g_dbpath = "/d/dev/res/election2016/www.pilipinaselectionresults2016.com"
var g_clearance = "2644c83aa405098ff0687c847c730e87b02b472f-1463305203-3600"
var g_useragent = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36"
var g_gatheronly = false

var g_store kv.KVStore
var g_wrtr kv.KVWriter
var g_rdr kv.KVReader

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.StringVar(&g_clearance, "clearance", "", "cloudflare clearance")
	flag.StringVar(&g_dbpath, "path", "", "path to comelect db")
	flag.StringVar(&g_useragent, "user_agent", "", "set HTTP client's user agent")
	flag.BoolVar(&g_gatheronly, "gather_only", false, "gather only urls and save to error queue")

	flag.Parse()

	if 0 == len(g_dbpath) {
		log.Fatal("invalid path")
		flag.PrintDefaults()
	}
	if 0 == len(g_useragent) {
		log.Fatal("invalid user agent")
		flag.PrintDefaults()
	}
	g_dbpath = filepath.Clean(g_dbpath) + "/db"
	log.Println("using clearance=", g_clearance)
	log.Println("using path=", g_dbpath)

	g_store, _ = leveldb.GetDefault(g_dbpath)
	g_wrtr, _ = g_store.Writer()
	g_rdr, _ = g_store.Reader()
}

var g_client *http.Client
var g_fqueueErr *fqueue.FQueue

func main() {
	var urlpath = "https://www.pilipinaselectionresults2016.com/data/regions/0.json"

	loadclient()
	fqueue.QueueLimit = 600 * fqueue.M

	var err error
	qcurls, _ := fqueue.NewFQueue("./q.fqueue")
	defer qcurls.Close()
	qcurls.Limit = 600 * fqueue.M

	g_fqueueErr, err = fqueue.NewFQueue("./qerr.fqueue")
	defer g_fqueueErr.Close()
	if nil != err {
		glog.Fatal("err creating g_fqueueErr err=", err)
	}

	if nil == g_fqueueErr {
		glog.Fatal("g_fqueueErr is nil")
	}

	retryErrs()

	bb, err := qcurls.Pop()
	if nil == err && 0 < len(bb) {
		urlpath = string(bb)
	}

	cUrl := make(chan interface{}, 100)
	cUrl <- urlpath

	c := make(chan interface{}, 400)
	qc := fp.DistributeWorkCh(c, func(a interface{}) (interface{}, error) {
		fmt.Print(".")

		bb := get(a.(string))
		if 0 == len(bb) || bb[0] == '<' {
			return nil, nil
		}
		var oregion JsRegion
		err := json.Unmarshal(bb, &oregion)
		if nil != err {
			log.Println("__Error Then err=", err)
			return nil, nil
		}

		for _, region := range oregion.SubRegions {
			if err = qcurls.Push([]byte(g_baseurl + region.Url)); nil != err {
				log.Fatal("qcurls push err=", err)
			}
		}

		for _, contest := range oregion.Contests {
			if err = qcurls.Push([]byte(g_baseurl + contest.Url)); nil != err {
				log.Fatal("qcurls push err=", err)
			}
		}

		return true, nil

	}, 800)

	q := fp.DistributeWorkCh(cUrl, func(a interface{}) (interface{}, error) {
		//log.Println(">>> len curl=", len(cUrl))

		contests, regions := begin(a.(string))
		if 0 == len(regions) && 0 == len(cUrl) && 0 == qcurls.Contents {
			log.Println("+++++DONE close.")
			log.Println("+++++DONE close.")
			log.Println("+++++DONE close.")
			log.Println("+++++DONE close.")
			close(cUrl)
			close(c)
		}

		for _, u := range append(regions, contests...) {
			//log.Println(">>>____ len c=", len(c))
			c <- u
			//log.Println("<<<<____ len c=", len(c))
		}

		return true, nil

	}, 200)

	fnAddCurl := func(_ interface{}) (interface{}, error) {
		fmt.Print(".")

		for i := len(cUrl); i <= cap(cUrl); i++ {
			bb, err := qcurls.Pop()
			if nil != err {
				log.Println("pop err=", err)
				return nil, nil
			}
			//log.Println(">>> len cUrl=", len(cUrl))
			cUrl <- string(bb)
		}
		return nil, nil
	}

	q = q.Then(fnAddCurl)
	qc = qc.Then(fnAddCurl)

	//TODO: fix this
	//go func() { fp.Flush(qc) }()
	//fp.Flush(q)

	var wg fp.WaitGroup
	wg.Add(qc, q)
	wg.Wait()
}

func begin(url string) (contests, regions []string) {
	fmt.Print("/")
	//defer log.Println("-END url=", url)

	bb := get(url)
	if 0 == len(bb) {
		return
	}
	var oregion JsRegion
	err := json.Unmarshal(bb, &oregion)
	if nil != err {
		log.Println("get country err=", err, ";bb=")
		return
	}

	for _, contest := range oregion.Contests {
		contests = append(contests, g_baseurl+contest.Url)
	}
	for _, region := range oregion.SubRegions {
		regions = append(regions, g_baseurl+region.Url)
	}

	return

}

func loadclient() {
	cookieJar, _ := cookiejar.New(nil)
	g_client = &http.Client{
		Jar: cookieJar,
	}

	// cloudflare captcha clearance
	// -- to update value, download cloudhole, a firefox plugin
	var cookies = []*http.Cookie{
		&http.Cookie{
			Name:  "cf_clearance",
			Value: g_clearance,
		},
	}
	u, err := url.Parse("https://www.pilipinaselectionresults2016.com")
	if nil != err {
		log.Fatal("parse url=", err)
	}
	cookieJar.SetCookies(u, cookies)
}

func validatejs(bb []byte) bool {
	js := map[string]interface{}{}
	err := json.Unmarshal(bb, &js)
	if nil != err {
		if strings.Contains(string(bb), "Checking your browser before accessing") {
			log.Fatal("Checking your browser before accessing")

		} else if strings.Contains(string(bb), "Attention Required!") {
			log.Fatal("Attention Required!")

		} else {
			log.Println("invalid json. err=", err, "bb=", string(bb))
		}
	}
	return nil == err
}

func get(url string) []byte {
	fmt.Print(".")
	//log.Println("get url=", url)
	//defer log.Println("done url=", url)

	k := []byte(strings.Replace(url, g_baseurl, "", 1))
	bb, err := g_rdr.Get(k)
	if nil != err {
		log.Fatal("Error: get db=", err)
	}

	if 0 < len(bb) {
		if !validatejs(bb) {
			batch := g_wrtr.NewBatch()
			batch.Delete(k)
			g_wrtr.ExecuteBatch(batch)
		} else {
			return bb
		}
	}

	if g_gatheronly {
		if err := g_fqueueErr.Push([]byte(url)); nil != err {
			log.Fatal("Error: gather only, push queue, err=", err)
		}
		return nil
	}

	bb = download(url)
	if 0 < len(bb) {
		save(k, bb)
	}
	return bb
}

func retryErrs() {

	ql := fp.Lazy(func() interface{} {
		if nil == g_fqueueErr {
			log.Fatal("g_fqueueErr is nil")
		}
		bb, err := g_fqueueErr.Pop()
		if nil != err {
			log.Println("pop err=", err)
			return nil
		}
		return string(bb)
	})

	type kvpair struct {
		k, v []byte
	}
	q := fp.DistributeWorkL0(ql, func(a interface{}) (interface{}, error) {
		url := a.(string)

		k := []byte(strings.Replace(url, g_baseurl, "", 1))
		if bbExist, _ := g_rdr.Get(k); 0 < len(bbExist) {
			// exists
			fmt.Print("X")
			return nil, nil
		}

		bb := download(url)
		if 0 < len(bb) {
			//save(k, bb)
			return &kvpair{k, bb}, nil

		}
		return nil, nil
	}, 50)

	var batch kv.KVBatch
	var ncount = 0
	q = q.Then(func(a interface{}) (interface{}, error) {
		if nil == a {
			return nil, nil
		}
		if 0 == ncount {
			batch = g_wrtr.NewBatch()
		}

		pair := a.(*kvpair)
		batch.Set(pair.k, pair.v)
		ncount++
		if 50 <= ncount {
			fmt.Print("`")
			g_wrtr.ExecuteBatch(batch)
			ncount = 0
		}

		return nil, nil
	})

	if 0 < ncount {
		fmt.Print("`")
		g_wrtr.ExecuteBatch(batch)
	}

	fp.Flush(q)
}

func save(k, bb []byte) {
	if validatejs(bb) {
		batch := g_wrtr.NewBatch()
		batch.Set(k, bb)

		err := g_wrtr.ExecuteBatch(batch)
		if nil != err {
			log.Fatal("Error: wrtr exec batch=", err)
		}

	}
}

func download(url string) (bb []byte) {
	fmt.Print("#")
	//defer log.Println("-DOWNLOADING url=", url)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", g_useragent)

	resp, err := g_client.Do(req)
	if nil != err {
		glog.Errorln("Get err=", err, ";url=", url)
		g_fqueueErr.Push([]byte(url))
		return nil
	}
	defer resp.Body.Close()

	bb, err = ioutil.ReadAll(resp.Body)
	if nil != err {
		glog.Fatal("read resp.body err=", err)
	}
	if 0 < len(bb) {
		if !validatejs(bb) {
			return nil
		}

	} else {
		fmt.Print("o")
	}

	return
}
