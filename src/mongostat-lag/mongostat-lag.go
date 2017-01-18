package main

import "gopkg.in/mgo.v2"
import "gopkg.in/mgo.v2/bson"
import "fmt"
import "flag"
import "log"
import "os"
import "text/template"
import "time"

/* CLI configuration */

var uri = flag.String("uri", "mongodb://localhost:27017", "mongo uri")
var serverName = flag.String("servername", "", "the name of the server to compute the lag. If empty, the lag is computed for all secondaries")
var verbose = flag.Bool("verbose", false, "verbose output")
var unitParam = flag.String("unit", "second", "unit of time (nanosecond,microsecond,second,minute,hour)")
var tmpl = flag.String("template", "", "use a template for the output. Available information are in ServerInfo struct (Lag,Timestamp,Name,Optime,OptimeDate,State,StateStr,PingMs,ElectionDate,Health,Uptime). For instance: \"DT.my.measure {{.Lag}}  {{.Timestamp.Unix()}} \"")
var help = flag.Bool("help", false, "help")

/* ReplicaInfo structure */
type ServerInfo struct {
	Name         string              `bson:"name"`
	Optime       bson.MongoTimestamp `bson:"optime"`
	OptimeDate   time.Time           `bson:"optimeDate"`
	State        int                 `bson:"state"`
	StateStr     string              `bson:"stateStr"`
	PingMs       int                 `bson:"pingMs"`
	ElectionDate time.Time           `bson:"electionDate"`
	Health       int                 `bson:"health"`
	Uptime       int64               `bson:"uptime"`
	Lag          time.Duration
	Timestamp    time.Time
}

type ReplicaInfo struct {
	Date    time.Time    `json:"date"`
	MyState int          `json:"myState"`
	Members []ServerInfo `json:"members"`
}

func main() {

	flag.Parse()
	if *help {
		fmt.Fprintf(os.Stderr, "Compute the lag in nanoseconds between a given secondary or all secondaries and the primary.\n\nUsage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	unit := time.Second
	switch {
	case *unitParam == "nanosecond":
		unit = time.Nanosecond
	case *unitParam == "microsecond":
		unit = time.Microsecond
	case *unitParam == "second":
		unit = time.Second
	case *unitParam == "minute":
		unit = time.Minute
	case *unitParam == "hour":
		unit = time.Hour
	}
	log.Printf("starting...\n")

	session, _ := mgo.Dial(*uri)
	db := session.DB("admin")
	result := ReplicaInfo{}

	// Execute the command mongo
	currentTime := time.Now()
	err := db.Run(bson.D{{`bson:"replSetGetStatus"`, 1}}, &result)
	if err != nil {
		log.Printf("err %v\n", err)
		return
	} else if *verbose {
		log.Printf("command succeed. Found %v member(s) in replica\n", len(result.Members))
	}

	// Analyze result
	var lastOptime time.Time
	var secondaries []ServerInfo
	for i, m := range result.Members {
		m.Lag = time.Duration(0)
		m.Timestamp = currentTime
		if *verbose {
			log.Printf("member %v: %+v\n", i, m)
		}
		if m.State == 1 {
			// it's the master
			lastOptime = m.OptimeDate
		} else if m.State == 2 {
			// it's the secondary
			if *serverName == "" {
				secondaries = append(secondaries, m)
			} else if m.Name == *serverName {
				secondaries = append(secondaries, m)
			}
		}
	}
	// TODO could be clever with one pass with precedent iteration (goroutine+channel ?)
	for _, s := range secondaries {
		s.Lag = lastOptime.Sub(s.OptimeDate) / unit
		if *verbose {
			log.Printf("secondary %v has %v delay\n", s.Name, s.Lag)
		}
		// TODO display content could be a method of ServerInfo
		if *tmpl == "" {
			fmt.Printf("%d\n", s.Lag)
		} else {
			tmplParsed, _ := template.New("output").Parse(*tmpl)
			err := tmplParsed.Execute(os.Stdout, s)
			if err != nil {
				panic(err)
			} else {
				fmt.Fprintf(os.Stdout, "\n")
			}
		}
	}
}
