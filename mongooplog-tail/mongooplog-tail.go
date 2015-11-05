package main

import "bytes"

import "encoding/binary"

import "com/vsct/dt/mongotools/Godeps/_workspace/src/gopkg.in/mgo.v2"

import "com/vsct/dt/mongotools/Godeps/_workspace/src/gopkg.in/mgo.v2/bson"

import "fmt"

import "time"

import "flag"

import "log"

import "os"

import "text/template"

// CLI configuration

var uri = flag.String("uri", "mongodb://localhost:27017", "mongo uri")

var ns = flag.String("namespace", "", "namespace of oplog (for instance 'mydb.mycoll'). By default is empty, so there is no filtering.")

var op = flag.String("operation", "", "show only these operations of oplog (for instance 'c','i','u' etc...). By default is empty, so there is no filtering.")

var sd = flag.Int("startdate", -1, "timestamp of the start date. Timestamp in seconds as stored in oplog.")

var si = flag.Int("startincr", 0, "timestamp of the start date. Increment part of the timestamp as stored in oplog. By default is 0. If stardate is not setted, this parameter is mute.")

var timeout = flag.Int64("timeout", -1, "timeout in seconds. Beyond this timeout without new oplog, the process returns. By default -1 (disable timeout).")

var help = flag.Bool("help", false, "help")

var tmpl = flag.String("template", "", "use a template for the output. Available information are in a struct. For instance for graphite, it could be: 'DT.my.measure {{.Ts}}  {{.Timestamp.Unix}} '. Type struct is: TsRaw bson.Raw, Ns string, H int64, V int, Op string, O map[string]interface{}, TsDateTime time.Time, TsIncr int32, CurrentTime time.Time")

// Oplog struct

type Oplog struct {
	TsRaw bson.Raw "ts"

	Ns string "ns"

	H int64 "h"

	V int "v"

	Op string "op"

	O map[string]interface{} "o"

	TsDateTime time.Time

	TsIncr int32

	CurrentTime time.Time
}

func buildQuery() bson.M {

	fmt.Println("buildQuery")

	query := bson.M{}

	if *ns != "" {

		query["ns"] = *ns

	}

	if *op != "" {

		query["op"] = *op

	}

	if *sd > 0 {

		// byte buffer for storing timestamp and inc parts of special Mongo timestamp => for ex. Timestamp(1437386925, 1)

		// see http://bsonspec.org/spec.html

		buf := new(bytes.Buffer)

		var ts int32 = int32(*sd) // timestamp part

		var inc int32 = int32(*si) // inc part

		binary.Write(buf, binary.LittleEndian, inc)

		binary.Write(buf, binary.LittleEndian, ts)

		// bson.MongoTimestamp has a uint64 representation

		var time uint64 = binary.LittleEndian.Uint64(buf.Bytes())

		log.Printf("ask timestamp >= %v\n", time)

		query["ts"] = bson.M{"$gte": bson.MongoTimestamp(time)}

	}

	return query

}

func main() {

	flag.Parse()

	if *help {

		fmt.Fprintf(os.Stderr, "Tail on oplog.\nUsage: $ mongooplogtail\n%v\n", os.Args[0])

		flag.PrintDefaults()

		return

	}

	log.Printf("starting...\n")

	session, _ := mgo.Dial(*uri)

	c := session.DB("local").C("oplog.rs")

	count, _ := c.Count()

	log.Printf("oplog size: %v\n", count)

	query := buildQuery()

	iter := c.Find(query).Sort("$natural").Tail(time.Duration(*timeout) * time.Second)

	for {

		result := Oplog{}

		for iter.Next(&result) {

			result.CurrentTime = time.Now()

			tsSlice := []byte(result.TsRaw.Data)

			var tsPart, incPart int32

			buff := bytes.NewReader(tsSlice[4:])

			binary.Read(buff, binary.LittleEndian, &tsPart)

			buff = bytes.NewReader(tsSlice[:4])

			binary.Read(buff, binary.LittleEndian, &incPart)

			result.TsDateTime = time.Unix(int64(tsPart), 0)

			result.TsIncr = incPart

			if *tmpl == "" {

				fmt.Printf("%+v\n", result)

			} else {

				tmplParsed, _ := template.New("output").Parse(*tmpl)

				err := tmplParsed.Execute(os.Stdout, result)

				if err != nil {

					panic(err)

				} else {

					fmt.Fprintf(os.Stdout, "\n")

				}

			}

		}

		if iter.Err() != nil {

			log.Printf("error: %v", iter.Err())

			return

		}

		if iter.Timeout() {

			log.Printf("timeout reached (%v seconds)\n", *timeout)

			return

		}

	}

	iter.Close()

}
