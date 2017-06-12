package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type counter struct {
	iterations       int32
	unmarshalledDocs int32
	writedDocs       int32
}

type channels struct {
	inRaws             chan *bson.Raw // channel from mongo iteration to unmarshalling step
	outRaws            chan *bson.Raw // channel from unmarshalling steps to writing step
	inUnmarshalledDocs chan bson.M
	pages              chan bool
	filters            chan bson.M
	lastIds            chan string // channel for last ids
}

var (
	uri           = flag.String("uri", "mongodb://localhost:27017", "mongo uri")
	dbName        = flag.String("database", "test", "database name")
	coll          = flag.String("collection", "test", "collection name")
	iterators     = flag.Int("iterators", 20, "number of iterators")
	unmarshallers = flag.Int("unmarshallers", 20, "number of unmarshallers")
	writers       = flag.Int("writers", 40, "number of writers")
	limit         = flag.Int("limit", 100000, "number max of documents")
	pages         = flag.Int("pages", 10, "number of parallel mongo queries. For each page query, limit will be 'limit/pages'.")
	batch         = flag.Int("batch", 100, "batch size for mongo requests")
	prefetch      = flag.Float64("prefetch", 0.5, "prefetch ratio from batch for mongo requests")
	output        = flag.String("output", "./results/", "path to output directory")
	prefix        = flag.String("prefix", "export-", "prefix of result files")
	help          = flag.Bool("help", false, "help")
	counters      = &counter{iterations: 0, unmarshalledDocs: 0, writedDocs: 0}
	stopped       = false
	allChannels   = &channels{}
)

func replaceIfPresent(element bson.M, field string, value string) {
	if _, ok := element[field]; ok {
		element[field] = value
	}
}

func filter(unfilteredDoc bson.M) {

	if customers, ok := unfilteredDoc["customers"]; ok {
		for _, customer := range customers.([]interface{}) {
			replaceIfPresent(customer.(bson.M), "iuc", "xxx")
			replaceIfPresent(customer.(bson.M), "firstname", "prenom")
			replaceIfPresent(customer.(bson.M), "name", "nom")
		}
	}

	for _, serviceItem := range unfilteredDoc["serviceItems"].([]interface{}) {

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

func unmarshall(index int) {
	fmt.Printf("start unmarshalling step: %v\n", index)
	for raw := range allChannels.outRaws {
		var unmarshalledDoc bson.M
		raw.Unmarshal(&unmarshalledDoc)
		counters.unmarshalledDocs++
		allChannels.inUnmarshalledDocs <- unmarshalledDoc
	}
	fmt.Printf("end of unmarshalling docs: %v\n", index)
	panic("shouldn't stop to unmarshall documents")
}

func monitor() {
	oldCounter := &counter{writedDocs: counters.writedDocs, iterations: counters.iterations, unmarshalledDocs: counters.unmarshalledDocs}
	for {
		currentCounter := &counter{writedDocs: counters.writedDocs, iterations: counters.iterations, unmarshalledDocs: counters.unmarshalledDocs}
		fmt.Printf("read docs: %d, unmarshalled docs: %d, writed docs: %d, ", currentCounter.iterations-oldCounter.iterations, currentCounter.unmarshalledDocs-oldCounter.unmarshalledDocs, currentCounter.writedDocs-oldCounter.writedDocs)
		fmt.Printf("waiting to unmarshall docs: %d, waiting to write docs: %d\n", len(allChannels.outRaws), len(allChannels.inUnmarshalledDocs))
		if stopped {
			break
		}
		time.Sleep(1 * time.Second)
		oldCounter = currentCounter
	}
}

func iterate(page int) {
	pageLimit := *limit / *pages
	lastId := <-allChannels.lastIds
	fmt.Printf("start iteration from mongo _id: %s\n", lastId)
	db := createDB()
	iter := db.C(*coll).Find(bson.M{"_id": bson.M{"$gte": lastId}}).Prefetch(*prefetch).Batch(*batch).Sort("_id").Limit(pageLimit).Iter()

	// create output file for this index
	fileName := *output + string(filepath.Separator) + *prefix + strconv.Itoa(page) + ".json"
	fmt.Printf("will write to file %s\n", fileName)
	os.Mkdir(*output, os.ModeDir)
	out, _ := os.Create(fileName)
	enc := json.NewEncoder(out)
	defer out.Sync()
	defer out.Close()

	var currentRaw bson.Raw
	var unmarshalledDoc bson.M
	itCnt := 0
	for iter.Next(&currentRaw) {
		itCnt++
		counters.iterations++
		// unmarshal
		currentRaw.Unmarshal(&unmarshalledDoc)
		filter(unmarshalledDoc)
		enc.Encode(&unmarshalledDoc)
	}

	allChannels.pages <- true
	fmt.Printf("%d iteration(s) from id %s \n", itCnt, lastId)

}

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

	go monitor()

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

	fmt.Println("start iteration from mongo")
	for i := 0; i < *pages; i++ {
		<-allChannels.pages
	}
	fmt.Println("pages finished\n ")
	stopped = true
	endTime := time.Now()
	fmt.Printf("end: %v\n", endTime.Sub(queryTime))

}
