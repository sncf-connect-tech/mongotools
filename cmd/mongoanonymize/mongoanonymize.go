package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"net/http"
	_ "net/http/pprof"

	"github.com/gorilla/handlers"
	_ "github.com/mkevac/debugcharts"
	"github.com/voyages-sncf-technologies/mongotools/doccleaner"
)

type counter struct {
	iterations int32
}

type channels struct {
	inRaws   chan *bson.Raw // channel from mongo iteration to unmarshalling step
	outRaws  chan *bson.Raw // channel from unmarshalling steps to writing step
	pages    chan bool
	filters  chan bson.M
	firstIds chan interface{} // channel for last ids
}

// Closeabe type for easily close all resources
type Closeable interface {
	Close() (err error)
}

var (
	// arguments
	config       = flag.String("config", "config.toml", "config file")
	uri          = flag.String("uri", "mongodb://localhost:27017", "mongo uri")
	dbName       = flag.String("database", "test", "database name")
	coll         = flag.String("collection", "test", "collection name")
	limit        = flag.Int("limit", -1, "number max of documents. By default export all documents")
	pages        = flag.Int("pages", -1, "number of parallel mongo queries. For each page query, limit will be 'limit/pages'. If undefined, pages=number of cpus")
	batch        = flag.Int("batch", 100, "batch size for mongo requests")
	prefetch     = flag.Float64("prefetch", 0.5, "prefetch ratio from batch for mongo requests")
	output       = flag.String("output", "./results", "path to output directory")
	compressed   = flag.Bool("compressed", false, "compress exported files to gzip")
	monitored    = flag.Bool("monitored", false, "display current status each second")
	noanonymized = flag.Bool("noanonymized", false, "document won't be anonymize. For testing purpose.")
	monitorPort  = flag.String("monitor-port", "50080", "monitor port for go runtime metrics (ex: localhost:50080/debug/charts)")
	ids          = flag.String("ids", "", "list of mongo ids for starting each pages. If empty, will compute them. If not empty, pages parameter is ignored.")
	help         = flag.Bool("help", false, "help")
	// variables
	counters        = &counter{iterations: 0}
	stopped         = false
	allChannels     = &channels{}
	anonymizedRules = map[string]bool{"order": true, "service": true}

	// govvv variables
	GitBranch  string
	GitState   string
	GitSummary string
	BuildDate  string
	Version    string
	GitCommit  string

	cleaner *doccleaner.DocCleaner
)

// anonymize some fields of the document
func anonymizeDocument(document bson.M) {
	_, err := cleaner.Clean(document)
	if err != nil {
		panic(err)
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

// iterate on given page. A page is part a of the whole collection.
func iterate(page int, firstId interface{}) {
	pageLimit := int(math.Ceil(float64(*limit) / float64(*pages)))
	fmt.Printf("[%d] start iteration from mongo _id: %s\n", page, firstId)
	db := createDB()
	iter := db.C(*coll).Find(bson.M{"_id": bson.M{"$gte": firstId}}).Prefetch(*prefetch).Batch(*batch).Sort("_id").Limit(pageLimit).Iter()

	// create output file for this index
	os.Mkdir(*output, os.ModeDir|0777)
	prefix := "export-"
	var writer io.Writer
	// TODO refactor for factorizing
	if *compressed {
		fileName := *output + string(filepath.Separator) + prefix + strconv.Itoa(page) + ".bson.gz"
		file, _ := os.Create(fileName)
		defer file.Close() // warn: defer order is important here

		zip := gzip.NewWriter(file)
		zip.Name = prefix + strconv.Itoa(page)
		defer zip.Close()
		writer = zip
		fmt.Printf("[%d] will write to file %s\n", page, fileName)
	} else {
		fileName := *output + string(filepath.Separator) + prefix + strconv.Itoa(page) + ".bson"
		file, _ := os.Create(fileName)
		defer file.Close() // warn: defer order is important here

		buffer := bufio.NewWriter(file)
		defer buffer.Flush()
		writer = buffer
		fmt.Printf("[%d] will write to file %s\n", page, fileName)
	}

	// start process: query -> unmarshall -> anonymize -> bson marshal
	var currentRaw bson.Raw
	var unmarshalledDoc bson.M
	itCnt := 0
	for iter.Next(&currentRaw) {
		itCnt++
		counters.iterations++
		// unmarshal
		if err := currentRaw.Unmarshal(&unmarshalledDoc); err != nil {
			panic(err)
		}
		if !*noanonymized {
			anonymizeDocument(unmarshalledDoc)
		}
		if out, err := bson.Marshal(unmarshalledDoc); err != nil {
			panic(err)
		} else if _, err := writer.Write(out); err != nil {
			panic(err)
		}
	}
	if iter.Err() != nil {
		panic(iter.Err())
	}

	allChannels.pages <- true
	fmt.Printf("[%d] %d iteration(s) from id %s \n", page, itCnt, firstId)

}

// createDB mongo database
func createDB() *mgo.Database {
	session, err := mgo.Dial(*uri + "/" + *dbName)
	if err != nil {
		panic(err)
	}

	session.SetSafe(nil)
	session.SetBatch(*batch)
	session.SetPrefetch(*prefetch)
	session.SetBypassValidation(true)
	session.SetMode(mgo.SecondaryPreferred, true)
	session.SetSocketTimeout(1 * time.Hour)
	db := session.DB(*dbName)
	return db
}

func main() {
	fmt.Printf("Version %s for git %s/%s/%s [%s] at %s\n", Version, GitCommit, GitBranch, GitState, GitSummary, BuildDate)

	flag.Parse()
	if *help {
		fmt.Fprintf(os.Stderr, "Anonymize command fields to files.\nUsage:\n$ mongoanonymize [options]\nGiven arguments: %v\n", os.Args[0])
		flag.PrintDefaults()
		return
	}
	go http.ListenAndServe(":"+*monitorPort, handlers.CompressHandler(http.DefaultServeMux))

	// load config file and create document cleaner
	configFile, err := os.Open(*config)
	if err != nil {
		panic(err)
	}
	configReader := bufio.NewReader(configFile)
	cleaner = doccleaner.NewDocCleaner(configReader)

	if *pages == -1 {
		*pages = runtime.NumCPU()
		fmt.Printf("%d detected cpu, so will run with %d pages\n", *pages, *pages)
	}

	log.Println("starting...")

	db := createDB()
	startTime := time.Now()

	queryTime := time.Now()
	fmt.Printf("query %v\n", queryTime.Sub(startTime))

	allChannels.filters = make(chan bson.M, 10)
	allChannels.pages = make(chan bool)
	allChannels.firstIds = make(chan interface{}, *pages)

	if *monitored {
		go monitor()
	}

	if cnt, err := db.C(*coll).Count(); err != nil {
		panic(err)
	} else {
		fmt.Printf("there are %d document on collection %s\n", cnt, *coll)

		if *limit == -1 {
			*limit = cnt
		}
	}

	if !anonymizedRules[*coll] {
		*noanonymized = false
		fmt.Printf("No anonymization rule for collection %s \n", *coll)
	}

	if *ids == "" {
		pageLimit := int(math.Ceil(float64(*limit) / float64(*pages)))
		fmt.Printf("Page Limit : %d\n", pageLimit)
		startTimeId := time.Now()
		var firstDoc bson.M
		iter := db.C(*coll).Find(nil).Sort("_id").Select(bson.M{"_id": 1}).Prefetch(*prefetch).Batch(*batch).Limit(*limit).Iter()
		idx := 0
		page := 0
		firstIds := make([]interface{}, *pages)
		for iter.Next(&firstDoc) {
			if firstDoc != nil && math.Mod(float64(idx), float64(pageLimit)) == 0.0 {

				switch firstDoc["_id"].(type) {
				case string:
					firstIds[page] = firstDoc["_id"].(string)
				case bson.ObjectId:
					firstIds[page] = firstDoc["_id"].(bson.ObjectId)
				default:
					fmt.Errorf("object id is not a string or bson.ObjectId. Type: %T, Value: %+v", firstDoc["_id"], firstDoc["_id"])
					panic("can't go on")
				}
				page = page + 1
			}
			if idx >= *limit {
				break
			} else {
				idx = idx + 1
			}
		}
		durationId := time.Now().Sub(startTimeId)
		fmt.Printf("duration for reading all ids: %s\n", durationId.String())

		fileName := *output + string(filepath.Separator) + "ids"
		file, _ := os.Create(fileName)
		writer := bufio.NewWriter(file)

		for page, firstId := range firstIds {
			writer.WriteString(fmt.Sprintf("%d-%s\n", page, firstId))
			go iterate(page, firstId)
		}
		writer.Flush()
		file.Close()
	} else {
		*pages = len(strings.Split(*ids, ","))
		for page, firstId := range strings.Split(*ids, ",") {
			go iterate(page, firstId)
		}

	}
	fmt.Println("start iteration from mongo")
	for i := 0; i < *pages; i++ {
		<-allChannels.pages
	}
	fmt.Println("pages finished\n ")
	stopped = true
	endTime := time.Now()
	fmt.Printf("end: %v\n", endTime.Sub(queryTime))

}
