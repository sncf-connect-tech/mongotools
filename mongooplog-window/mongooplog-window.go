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

/* CLI configuration */

var uri = flag.String("uri", "mongodb://localhost:27017", "mongo uri")
var help = flag.Bool("help", false, "help")
var verbose = flag.Bool("verbose", false, "verbose output")
var tmpl = flag.String("template", "", "use a template for the output. The type of the output is Output which contains two fields: Duration (window)and Timestamp (current time).the go type time.Duration with additionnal current time function '.Timestamp' which returns the type time.Time (useful for graphite). For instance for graphite, it could be: 'DT.my.measure {{.Seconds}}  {{.Timestamp.Unix}} ' (the duration in secondes)")

/* Oplog structure */
type Oplog struct {
	TsRaw bson.Raw "ts"
}

type Output struct {
	Duration  time.Duration
	Timestamp time.Time
}

/**
Return the date time contained in TS of oplog element.
**/
func (o *Oplog) TsDateTime() time.Time {
	// TODO optimize => cache the result
	tsSlice := []byte(o.TsRaw.Data)
	var tsPart int32
	buff := bytes.NewReader(tsSlice[4:])
	binary.Read(buff, binary.LittleEndian, &tsPart)
	return time.Unix(int64(tsPart), 0)
}

func main() {

	flag.Parse()
	if *help {
		fmt.Fprintf(os.Stderr, "Window of the oplog (time of last oplog - time of first oplog).\nUsage:\n$ mongooplog-window [options]\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	log.Printf("starting...\n")

	session, _ := mgo.Dial(*uri)

	c := session.DB("local").C("oplog.rs")

	var firstOplog Oplog
	var lastOplog Oplog
	c.Find(bson.M{}).Sort("ts").Limit(1).Iter().Next(&firstOplog)
	c.Find(bson.M{}).Sort("-ts").Limit(1).Iter().Next(&lastOplog)

	log.Printf("first oplog %v", firstOplog.TsDateTime())
	log.Printf("last oplog %v", lastOplog.TsDateTime())

	duration := lastOplog.TsDateTime().Sub(firstOplog.TsDateTime())

	if *tmpl == "" {
		fmt.Printf("%+v\n", duration)
	} else {
		tmplParsed, _ := template.New("output").Parse(*tmpl)
		err := tmplParsed.Execute(os.Stdout, duration)
		if err != nil {
			panic(err)
		} else {
			fmt.Fprintf(os.Stdout, "\n")
		}
	}
}
