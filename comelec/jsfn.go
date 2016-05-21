package comelec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"unsafe"

	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/parser"
	_ "github.com/robertkrimen/otto/underscore"

	"time"

	"github.com/blevesearch/bleve"
	"github.com/noypi/fp"
	"github.com/noypi/kv"
	"github.com/noypi/kv/leveldb"
)

type jsNamed struct {
	Name         string
	CategoryName string
	Url          string
}
type indexMsg struct {
	Id   string
	Name string
}

type _timer struct {
	timer    *time.Timer
	duration time.Duration
	interval bool
	call     otto.FunctionCall
}

type JsVmComelec struct {
	vm           *otto.Otto
	store        kv.KVStore
	dbwrtr       kv.KVWriter
	dbrdr        kv.KVReader
	includes     map[string]struct{}
	libpaths     []string
	comelec_path string
	db_name      string
	qMap         map[string]*fp.Promise
}

func (this *JsVmComelec) dbname() string {
	if 0 < len(this.db_name) {
		return this.db_name
	}
	v, err := this.vm.Get("$dbname")
	if nil != err || 0 == len(v.String()) {
		this.db_name = "db"
	}
	this.db_name = v.String()
	return this.db_name

}

func (this *JsVmComelec) mustHaveDb() {
	dbpath := this.path(this.dbname())
	if 0 == len(dbpath) {
		log.Fatal("Comelec Data path is not specified, specify $path.")
	}
	if nil == this.store {
		store, err := leveldb.GetDefault(dbpath)
		if nil != err {
			log.Fatalln("Cannot open Comelec Data, err=", err, ";dbpath=", dbpath)
		}
		this.store = store
	}
	if nil == this.dbrdr {
		this.dbrdr, _ = this.store.Reader()
	}
	if nil == this.dbwrtr {
		this.dbwrtr, _ = this.store.Writer()
	}
}

func (this *JsVmComelec) jsFn_DbgRaw(call otto.FunctionCall) otto.Value {
	this.mustHaveDb()

	if 0 == len(call.ArgumentList) {
		return otto.UndefinedValue()
	}

	bb, _ := this.dbrdr.Get([]byte(call.Argument(0).String()))
	fmt.Println(string(bb))
	return otto.UndefinedValue()

}

func (this JsVmComelec) jsFn_Print(call otto.FunctionCall) otto.Value {
	for _, o := range call.ArgumentList {
		fmt.Print(o)
	}
	fmt.Print("\n")
	return otto.UndefinedValue()
}

func (this *JsVmComelec) jsFn_PrettyPrn(call otto.FunctionCall) otto.Value {
	if 0 == len(call.ArgumentList) {
		return otto.UndefinedValue()
	}

	o, _ := this.vm.Object("JSON")
	v, err := o.Call("stringify", call.Argument(0))
	if nil == err {
		buf := bytes.NewBufferString("")
		json.Indent(buf, []byte(v.String()), "", "\t")
		fmt.Println(buf.String())
	} else {
		fmt.Println("")
	}
	return otto.UndefinedValue()

}

func (this *JsVmComelec) VM() *otto.Otto {
	return this.vm
}

func (this *JsVmComelec) indexpath() string {
	return this.path("") + "/index.db"
}

func (this *JsVmComelec) jsFn_UpdateIndex(call otto.FunctionCall) otto.Value {

	this.mustHaveDb()

	var indexpath = this.indexpath()
	var index bleve.Index
	var err error
	if f, err := os.Stat(indexpath); nil == f || os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(indexpath, mapping)
	} else {
		index, err = bleve.Open(indexpath)
	}

	if nil != err {
		log.Fatal("Error db:", err)
	}
	defer index.Close()

	fmt.Println("indexing...")
	var n = 0
	batch := index.NewBatch()

	it := this.dbrdr.PrefixIterator([]byte("data/regions/"))
	defer it.Close()
	for ; it.Valid(); it.Next() {
		//k := it.Key()
		v := it.Value()

		var o = new(jsNamed)
		err = json.Unmarshal(v, o)
		if nil != err ||
			0 == len(o.Name) ||
			0 == len(o.Url) {
			continue
		}

		d, _ := index.Document(o.Url)
		if nil != d {
			continue
		}

		if _, err := strconv.Atoi(o.Name); nil == err {
			// skip numbers
			continue
		}

		fmt.Println(o.Url)

		batch.Index(o.Url, o)
		n++
		if n <= 10000 {
			index.Batch(batch)
			batch = index.NewBatch()
			n = 0
		}

	}

	if 0 < n {
		index.Batch(batch)
	}

	return otto.UndefinedValue()
}

func (this *JsVmComelec) jsFn_Search(call otto.FunctionCall) otto.Value {
	if 0 == len(call.ArgumentList) {
		return otto.UndefinedValue()
	}

	var indexpath = this.indexpath()
	q := call.Argument(0).String()
	index, err := bleve.Open(indexpath)
	if nil != err || 0 == len(q) {
		return otto.UndefinedValue()
	}
	defer index.Close()

	var arr, _ = this.vm.Object("([])")

	query := bleve.NewQueryStringQuery(q)
	req := bleve.NewSearchRequest(query)
	res, _ := index.Search(req)
	for i, hit := range res.Hits {
		arr.Set(strconv.Itoa(i), hit.ID)
	}

	return arr.Value()

}

func (this *JsVmComelec) jsFn_FromPath(call otto.FunctionCall) otto.Value {
	this.mustHaveDb()

	if 0 == len(call.ArgumentList) {
		return otto.UndefinedValue()
	}

	k := []byte(call.Argument(0).String())
	if 0 == len(k) {
		return otto.UndefinedValue()
	}

	bb, err := this.dbrdr.Get(k)
	if nil != err || 0 == len(bb) {
		return otto.UndefinedValue()
	}

	v, err := this.vm.Object("(" + string(bb) + ")")
	if nil != err {
		log.Printf("Error: %s: %v\n", string(k), err)
	}
	if nil == v {
		return otto.UndefinedValue()
	}
	return v.Value()
}

func (this *JsVmComelec) jsFn_Include(call otto.FunctionCall) otto.Value {
	if 0 == len(call.ArgumentList) {
		return otto.UndefinedValue()
	}

	v, _ := this.Include(call.Argument(0).String())
	return v

}

func (this *JsVmComelec) eachls(path string, cb func(string) bool) {
	this.mustHaveDb()

	if 0 < len(path) {
		path = filepath.Clean(path) + "/"
	}

	it := this.dbrdr.PrefixIterator([]byte(path))
	defer it.Close()
	for ; it.Valid(); it.Next() {
		k := it.Key()

		if bytes.Contains(k[len(path):], []byte{'/'}) {
			break
		}

		if !cb(string(k)) {
			break
		}
	}
}

func (this *JsVmComelec) jsFn_List(call otto.FunctionCall) otto.Value {
	var sub string
	var limit int
	if 0 < len(call.ArgumentList) {
		sub = call.Argument(0).String()
	}
	if 1 < len(call.ArgumentList) {
		n, _ := call.Argument(1).ToInteger()
		limit = int(n)
	}

	var arr, _ = this.vm.Object("([])")
	var i int
	this.eachls(sub, func(sk string) bool {
		err := arr.Set(strconv.Itoa(i), sk)
		if nil != err {
			log.Println("Error:", err)
		}
		i++
		if 0 < limit {
			return i < limit
		}
		return true
	})

	return arr.Value()
}

func (this *JsVmComelec) jsFn_EachList(call otto.FunctionCall) otto.Value {

	if 2 > len(call.ArgumentList) ||
		!call.Argument(0).IsString() ||
		!call.Argument(1).IsFunction() {
		return otto.UndefinedValue()
	}

	var sub string = call.Argument(0).String()
	fn := call.Argument(1)

	var limit int
	var scope otto.Value
	if 2 < len(call.ArgumentList) {
		scope = call.Argument(2)
	}
	if scope.IsNumber() {
		n, _ := scope.ToInteger()
		limit = int(n)
	}

	var n int
	this.eachls(sub, func(sk string) bool {
		v, err := fn.Call(scope, sk)
		n++
		if nil != err {
			log.Println("Error:", err)
		} else if v.IsBoolean() {
			if b, err := v.ToBoolean(); nil == err && !b {
				return false
			}
		}

		if 0 < limit {
			return n < limit
		}
		return true

	})

	return otto.UndefinedValue()
}

func (this *JsVmComelec) path(sub string) string {
	if 0 == len(this.comelec_path) {
		vpath, err := this.vm.Get("$path")
		if nil != err {
			log.Println("$path is not set")
			return ""
		}
		this.comelec_path = vpath.String()
	}
	return filepath.Clean(this.comelec_path + "/" + sub)

}

func (this *JsVmComelec) AddLibPath(fpath string) {

	fpath = filepath.Clean(fpath)
	for _, s := range this.libpaths {
		if s == fpath {
			return
		}
	}

	if 0 == len(this.libpaths) {
		this.libpaths = []string{"."}
	}

	this.libpaths = append(this.libpaths, fpath)
}

func (this *JsVmComelec) jsFn_AddLibPath(call otto.FunctionCall) otto.Value {
	if 0 < len(call.ArgumentList) {
		this.AddLibPath(call.Argument(0).String())
	}

	return otto.UndefinedValue()
}

func (this *JsVmComelec) jsFn_ImportLib(call otto.FunctionCall) otto.Value {
	if 0 < len(call.ArgumentList) {
		this.ImportLib(call.Argument(0).String())
	}
	return otto.UndefinedValue()
}

func (this *JsVmComelec) ImportLib(lib string) {
	for _, s := range this.libpaths {
		fpath := s + "/" + lib
		f, err := os.Stat(fpath)
		if nil == f || os.IsNotExist(err) {
			continue
		}
		if f.IsDir() {
			fmt.Println("Warning:", fpath, " is a directory.")
			return
		}
		if nil != err {
			fmt.Println("Error: importing fpath=", fpath)
			return
		}

		_, err = this.Include(fpath)
		if nil != err {
			fmt.Println("Error: importing fpath=", fpath)
			return
		}
	}
}

func (this *JsVmComelec) Include(fpath string) (otto.Value, error) {
	fpath = filepath.Clean(fpath)
	if _, has := this.includes[fpath]; has {
		return otto.UndefinedValue(), fmt.Errorf("redefined file:", fpath)
	}

	bb, err := ioutil.ReadFile(fpath)
	if nil != err {
		log.Fatal(err)
	}
	program, err := parser.ParseFile(nil, filepath.Base(fpath), string(bb), 0)
	if nil != err {
		log.Fatal("Error ParseFile:", err)
	}

	if nil == this.includes {
		this.includes = map[string]struct{}{}
	}

	v, err := this.vm.Run(program)
	if nil != err {
		log.Println("Error:", err)
	} else {
		this.includes[fpath] = struct{}{}
	}

	return v, err
}

func (this *JsVmComelec) jsFn_FindPrefix(call otto.FunctionCall) otto.Value {
	var sub string
	if 0 < len(call.ArgumentList) {
		sub = call.Argument(0).String()
	}

	var fn otto.Value
	if 1 < len(call.ArgumentList) {
		fn = call.Argument(1)
	}

	var scope otto.Value
	if 2 < len(call.ArgumentList) {
		scope = call.Argument(2)
	}

	var limit int
	if scope.IsNumber() {
		n, _ := scope.ToInteger()
		limit = int(n)
		scope = otto.Value{}
	}

	var i int
	this.FindPrefix(sub, func(k string) bool {
		v, err := fn.Call(scope, k)
		if nil != err {
			log.Println("Error:", err)
			return false
		}
		i++
		if limit <= i {
			return false
		}
		if v.IsBoolean() {
			b, _ := v.ToBoolean()
			return b
		}
		return true
	})

	return otto.UndefinedValue()
}

func (this *JsVmComelec) FindPrefix(s string, cb func(string) bool) {
	this.mustHaveDb()
	it := this.dbrdr.PrefixIterator([]byte(s))
	defer it.Close()
	for ; it.Valid(); it.Next() {
		if !cb(string(it.Key())) {
			break
		}
	}
}

func (this *JsVmComelec) jsFn_Q(call otto.FunctionCall) otto.Value {
	fn := call.Argument(0)

	if nil == this.qMap {
		this.qMap = map[string]*fp.Promise{}
	}

	q := fp.Future(func() (interface{}, error) {
		return fn.Call(otto.UndefinedValue())
	})

	obj, _ := this.vm.Object("({})")
	id := fmt.Sprintf("%v", unsafe.Pointer(q))
	this.qMap[id] = q
	obj.Set("$$id", id)
	obj.Set("then", this.jsFn_Qthen)

	return obj.Value()
}

func (this *JsVmComelec) jsFn_Qthen(objcall otto.FunctionCall) otto.Value {
	objparam1 := objcall.Argument(0)
	objparam2 := objcall.Argument(1)
	vid, err := objcall.This.Object().Get("$$id")
	if nil != err {
		log.Println("Warn: unknown Q.")
		return otto.UndefinedValue()
	}
	n, err := vid.ToString()
	if nil != err {
		log.Println("Warn: Q has invalid $$id.")
		return otto.UndefinedValue()
	}

	qCurr, has := this.qMap[n]
	if !has {
		log.Fatal("Err: invalid Q, $$id=", n)
	}
	qNew := qCurr.Then(func(a interface{}) (interface{}, error) {
		if objparam1.IsFunction() {
			return objparam1.Call(objcall.This, a)
		}
		return otto.UndefinedValue(), nil

	}, func(a interface{}) (interface{}, error) {
		if objparam2.IsFunction() {
			return objparam2.Call(objcall.This, a)
		}
		return otto.UndefinedValue(), nil

	})

	delete(this.qMap, fmt.Sprintf("%v", unsafe.Pointer(qCurr)))
	id := fmt.Sprintf("%v", unsafe.Pointer(qNew))
	this.qMap[id] = qNew
	obj, _ := this.vm.Object("({})")
	obj.Set("$$id", id)

	return obj.Value()
}

func (this *JsVmComelec) EnableQSupport() {
	this.vm.Set("$Q", this.jsFn_Q)
}

func (this *JsVmComelec) QDone() {
	var wg fp.WaitGroup
	for _, q := range this.qMap {
		wg.Add(q)
	}
	wg.Wait()
}
