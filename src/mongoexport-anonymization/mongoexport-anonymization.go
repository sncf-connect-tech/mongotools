package main

import (
	"time"
	"strconv"
	"flag"
	"gopkg.in/mgo.v2/bson"
	"fmt"
	"os"
	"gopkg.in/mgo.v2"
	"log"
	"encoding/json"
	"path/filepath"
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
}

var (
	uri           = flag.String("uri", "mongodb://localhost:27017", "mongo uri")
	dbName        = flag.String("database", "test", "database name")
	coll          = flag.String("collection", "test", "collection name")
	iterators     = flag.Int("iterators", 20, "number of iterators")
	unmarshallers = flag.Int("unmarshallers", 20, "number of unmarshallers")
	writers       = flag.Int("writers", 40, "number of writers")
	limit         = flag.Int("limit", 1000, "number max of documents")
	pages         = flag.Int("pages", 1, "number of parallel mongo queries. For each page query, limit will be 'limit/pages'.")
	batch         = flag.Int("batch", 100, "batch size for mongo requests")
	prefetch      = flag.Float64("prefetch", 0.5, "prefetch ratio from batch for mongo requests")
	output        = flag.String("output", "./results/", "path to output directory")
	prefix        = flag.String("prefix", "export-", "prefix of result files")
	help          = flag.Bool("help", false, "help")
	counters      = &counter{iterations:0, unmarshalledDocs:0, writedDocs:0}
	stopped       = false
	allChannels   = &channels{}
)

func replaceIfPresent(element bson.M, field string, value string) {
	if _, ok := element[field]; ok {
		element[field] = value
	}
}

func filter() {

	for unfilteredDoc := range allChannels.inUnmarshalledDocs {
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

		allChannels.filters <- unfilteredDoc
	}
}

func unmarshall(index int) {
	fmt.Printf("start unmarshalling step: %v\n", index)
	for raw := range allChannels.outRaws {
		var unmarshalledDoc bson.M
		raw.Unmarshal(&unmarshalledDoc)
		counters.unmarshalledDocs++
		allChannels.inUnmarshalledDocs <- unmarshalledDoc
		allChannels.inRaws <- raw
	}
	fmt.Printf("end of unmarshalling docs: %v\n", index)
	panic("shouldn't stop to unmarshall documents")
}

func write(index int) {
	fmt.Printf("start writing step: %v\n", index)
	os.Mkdir(*output, os.ModeDir)
	out, _ := os.Create(*output + string(filepath.Separator) + *prefix + strconv.Itoa(index) + ".json")
	defer out.Close()
	enc := json.NewEncoder(out)
	for d := range allChannels.filters {
		enc.Encode(&d)
		counters.writedDocs++
	}
	fmt.Printf("encoded %v\n", index)
	out.Sync()
}

func display() {
	oldCounter := &counter{writedDocs: counters.writedDocs, iterations:counters.iterations, unmarshalledDocs:counters.unmarshalledDocs}
	for {
		currentCounter := &counter{writedDocs: counters.writedDocs, iterations:counters.iterations, unmarshalledDocs:counters.unmarshalledDocs}
		fmt.Printf("read docs: %d, unmarshalled docs: %d, writed docs: %d, ", currentCounter.iterations-oldCounter.iterations, currentCounter.unmarshalledDocs-oldCounter.unmarshalledDocs, currentCounter.writedDocs-oldCounter.writedDocs)
		fmt.Printf("waiting docs: %d, waiting to unmarshall docs: %d, waiting to write docs: %d\n", len(allChannels.inRaws), len(allChannels.outRaws), len(allChannels.inUnmarshalledDocs))
		if stopped {
			break
		}
		time.Sleep(1 * time.Second)
		oldCounter = currentCounter
	}
}

func iterate(iter *mgo.Iter, page int) {
	var currentRaw *bson.Raw
	currentRaw = <-allChannels.inRaws
	for iter.Next(currentRaw) {
		counters.iterations++
		allChannels.outRaws <- currentRaw
		currentRaw = <-allChannels.inRaws
	}
	allChannels.pages <- true
	fmt.Printf("end of mongo page %d \n", page)
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

	for index := 0; index < *unmarshallers; index++ {
		go unmarshall(index)
		go filter()
	}
	for index := 0; index < *writers; index++ {
		go write(index)
	}

	// feed raw pointers for mongo iteration
	for index := 0; index < *iterators; index++ {
		var currentRaw bson.Raw
		allChannels.inRaws <- &currentRaw
	}

	go display()

	if *pages > 1 {
		pageLimit := *limit / *pages
		for page := 0; page < *pages; page++ {
			go iterate(db.C(*coll).Find(nil).Prefetch(*prefetch).Batch(*batch).Skip(page * pageLimit).Limit(pageLimit).Iter(), page)
		}
	} else {
		go iterate(db.C(*coll).Find(nil).Prefetch(*prefetch).Batch(*batch).Limit(*limit).Iter(), 0)
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
