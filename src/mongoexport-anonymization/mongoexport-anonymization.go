package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/handlers"
	_ "github.com/mkevac/debugcharts"
)

type counter struct {
	iterations int32
}

type channels struct {
	inRaws             chan *bson.Raw // channel from mongo iteration to unmarshalling step
	outRaws            chan *bson.Raw // channel from unmarshalling steps to writing step
	inUnmarshalledDocs chan bson.M
	pages              chan bool
	filters            chan bson.M
	lastIds            chan string // channel for last ids
}

// Closeabe type for easily close all resources
type Closeable interface {
	Close() (err error)
}

var (
	uri           = flag.String("uri", "mongodb://localhost:27017", "mongo uri")
	dbName        = flag.String("database", "test", "database name")
	coll          = flag.String("collection", "test", "collection name")
	iterators     = flag.Int("iterators", 20, "number of iterators")
	unmarshallers = flag.Int("unmarshallers", 20, "number of unmarshallers")
	writers       = flag.Int("writers", 40, "number of writers")
	limit         = flag.Int("limit", 100000, "number max of documents")
	pages         = flag.Int("pages", -1, "number of parallel mongo queries. For each page query, limit will be 'limit/pages'. If undefined, pages=GOMAXPROC")
	batch         = flag.Int("batch", 100, "batch size for mongo requests")
	prefetch      = flag.Float64("prefetch", 0.5, "prefetch ratio from batch for mongo requests")
	output        = flag.String("output", "./results", "path to output directory")
	prefix        = flag.String("prefix", "export-", "prefix of result files")
	compressed    = flag.Bool("compressed", false, "compressed")
	monitored     = flag.Bool("monitored", false, "display current status each second")
	monitorPort   = flag.String("monitor-port", "50080", "monitor port for go runtime metrics (ex: localhost:50080/metrics)")
	help          = flag.Bool("help", false, "help")
	counters      = &counter{iterations: 0}
	stopped       = false
	allChannels   = &channels{}
)

// replaceIfPresent replace field of the document if exists by given value
func replaceIfPresent(element bson.M, field string, value string) {
	if _, ok := element[field]; ok {
		element[field] = value
	}
}

// anonymize some fields of the document
func anonymize(document bson.M) {

	if customers, ok := document["customers"]; ok {
		for _, customer := range customers.([]interface{}) {
			replaceIfPresent(customer.(bson.M), "iuc", "xxx")
			replaceIfPresent(customer.(bson.M), "firstname", "prenom")
			replaceIfPresent(customer.(bson.M), "name", "nom")
		}
	}

	for _, serviceItem := range document["serviceItems"].([]interface{}) {

		if contact, ok := serviceItem.(bson.M)["contactInformation"]; ok {
			fields := []string{"firstname", "name", "address1", "address2", "address3", "address4", "city", "zipCode", "country", "mobilePhoneNumber", "emailAddress"}
			for _, field := range fields {
				replaceIfPresent(contact.(bson.M), field, field)
			}
			replaceIfPresent(contact.(bson.M), "mobilePhoneNumber", "0123456789")
			replaceIfPresent(contact.(bson.M), "emailAddress", "toto@toto.fr")
			contact.(bson.M)["landlinePhoneNumbers"] = nil
		}

		if passengers, ok := serviceItem.(bson.M)["passengers"]; ok {
			for _, passenger := range passengers.([]interface{}) {
				fields := []string{"iuc", "firstName", "lastName", "mobilePhoneNumber", "emailAddress", "fceNumber", "mrcNumber"}
				for _, field := range fields {
					replaceIfPresent(passenger.(bson.M), field, "xxx")
				}
				passenger.(bson.M)["emailAddress"] = "toto@toto.fr"
				passenger.(bson.M)["mobilePhoneNumber"] = "0123456789"
				if clientInformation, ok := passenger.(bson.M)["cilentCardNumber"]; ok {
					replaceIfPresent(clientInformation.(bson.M), "fidNumber", "xxx")
				}
				if cilentCardNumber, ok := passenger.(bson.M)["clientCard"]; ok {
					replaceIfPresent(cilentCardNumber.(bson.M), "cardNumber", "xxx")
				}
				if sncfAgentAddritionalProperties, ok := passenger.(bson.M)["sncfAgentAddritionalProperties"]; ok {
					replaceIfPresent(sncfAgentAddritionalProperties.(bson.M), "agentId", "xxx")
				}
			}
		}

		if railTransportationContracts, ok := serviceItem.(bson.M)["railTransportationContracts"]; ok {
			for _, contract := range railTransportationContracts.([]interface{}) {
				if holder, ok := contract.(bson.M)["holder"]; ok {
					replaceIfPresent(holder.(bson.M), "firstName", "prenom")
					replaceIfPresent(holder.(bson.M), "lastName", "nom")
				}
			}
		}
	}
}

// monitor current export status
func monitor() {
	oldCounter := &counter{iterations: counters.iterations}
	for {
		currentCounter := &counter{iterations: counters.iterations}
		fmt.Printf("docs: %d\n", currentCounter.iterations-oldCounter.iterations)
		if stopped {
			break
		}
		time.Sleep(1 * time.Second)
		oldCounter = currentCounter
	}
}

// writer returns a gzip or bufio writer depending of user parameter
func writer(page int) (w io.Writer, closeables []Closeable) {
	if *compressed {
		fileName := *output + string(filepath.Separator) + *prefix + strconv.Itoa(page) + ".gz"
		file, _ := os.Create(fileName)
		gz := gzip.NewWriter(file)
		gz.Name = *prefix + strconv.Itoa(page)
		w = gz
		closeables = []Closeable{gz, file}
		fmt.Printf("[%d] will write to file %s\n", page, fileName)
	} else {
		fileName := *output + string(filepath.Separator) + *prefix + strconv.Itoa(page) + ".json"
		file, _ := os.Create(fileName)
		w = bufio.NewWriter(file)
		closeables = []Closeable{file}
		fmt.Printf("[%d] will write to file %s\n", page, fileName)
	}
	return w, closeables
}

// iterate on given page. A page is part a of the whole collection.
func iterate(page int) {
	pageLimit := *limit / *pages
	lastId := <-allChannels.lastIds
	fmt.Printf("[%d] start iteration from mongo _id: %s\n", page, lastId)
	db := createDB()
	iter := db.C(*coll).Find(bson.M{"_id": bson.M{"$gte": lastId}}).Prefetch(*prefetch).Batch(*batch).Sort("_id").Limit(pageLimit).Iter()

	// create output file for this index
	os.Mkdir(*output, os.ModeDir)
	zipWriter, closeables := writer(page)
	enc := json.NewEncoder(zipWriter)

	var currentRaw bson.Raw
	var unmarshalledDoc bson.M
	itCnt := 0
	for iter.Next(&currentRaw) {
		itCnt++
		counters.iterations++
		// unmarshal
		currentRaw.Unmarshal(&unmarshalledDoc)
		anonymize(unmarshalledDoc)
		enc.Encode(&unmarshalledDoc)
	}
	for _, closeable := range closeables {
		closeable.Close()
	}

	allChannels.pages <- true
	fmt.Printf("[%d] %d iteration(s) from id %s \n", page, itCnt, lastId)

}

// createDB mongo database
func createDB() *mgo.Database {
	session, err := mgo.Dial(*uri)
	if err != nil {
		panic(err)
	}

	session.SetSafe(nil)
	session.SetBatch(*batch)
	session.SetPrefetch(*prefetch)
	session.SetBypassValidation(true)
	session.SetMode(mgo.SecondaryPreferred, false)
	db := session.DB(*dbName)
	return db
}

func main() {
	flag.Parse()
	if *help {
		fmt.Fprintf(os.Stderr, "Window of the oplog (time of last oplog - time of first oplog).\nUsage:\n$ mongooplog-window [options]\nGiven arguments: %v\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	if *pages == -1 {
		*pages = runtime.NumCPU()
		fmt.Printf("%d detected cpu, so will run with %d pages\n", *pages, *pages)
	}

	log.Println("starting...")

	session, err := mgo.Dial(*uri)
	if err != nil {
		panic(err)
	}

	session.SetSafe(nil)
	session.SetBatch(*batch)
	session.SetPrefetch(*prefetch)
	session.SetBypassValidation(true)
	session.SetMode(mgo.SecondaryPreferred, false)
	db := session.DB(*dbName)
	startTime := time.Now()

	queryTime := time.Now()
	fmt.Printf("query %v\n", queryTime.Sub(startTime))

	allChannels.inRaws = make(chan *bson.Raw, *iterators)      // channel from mongo iteration to unmarshalling step
	allChannels.outRaws = make(chan *bson.Raw, *unmarshallers) // channel from unmarshalling steps to writing step
	allChannels.inUnmarshalledDocs = make(chan bson.M, *writers)
	allChannels.filters = make(chan bson.M, 10)
	allChannels.pages = make(chan bool)
	allChannels.lastIds = make(chan string, *pages)

	if *monitored {
		go monitor()
	}

	for page := 0; page < *pages; page++ {
		go iterate(page)
	}

	var firstDoc map[string]interface{}

	db.C(*coll).Find(nil).Prefetch(*prefetch).Batch(*batch).Sort("_id").Limit(1).Select(bson.M{"_id": 1}).One(&firstDoc)
	firstId := firstDoc["_id"].(string)
	fmt.Printf("firstId: %s\n", firstId)

	lastId := firstId
	pageLimit := *limit / *pages
	startTimeId := time.Now()
	for i := 0; i < *pages; i++ {
		allChannels.lastIds <- lastId
		iter := db.C(*coll).Find(bson.M{"_id": bson.M{"$gte": lastId}}).Prefetch(*prefetch).Batch(*batch).Sort("_id").Select(bson.M{"_id": 1}).Limit(pageLimit).Iter()
		for iter.Next(&firstDoc) {
			lastId = firstDoc["_id"].(string)
		}
		fmt.Printf("last id=%s\n", lastId)
	}
	durationId := time.Now().Sub(startTimeId)
	fmt.Printf("duration for reading all ids: %s\n", durationId.String())

	go http.ListenAndServe(":8080", handlers.CompressHandler(http.DefaultServeMux))
	fmt.Println("start iteration from mongo")
	for i := 0; i < *pages; i++ {
		<-allChannels.pages
	}
	fmt.Println("pages finished\n ")
	stopped = true
	endTime := time.Now()
	fmt.Printf("end: %v\n", endTime.Sub(queryTime))

}
