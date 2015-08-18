package main

import "flag"
import "fmt"
import "log"
import "os"
import "text/template"
import "encoding/json"
import "io"
import "strconv"
import "strings"
import "time"

/* CLI configuration */

var uri = flag.String("uri", "mongodb://localhost:27017", "mongo uri")
var help = flag.Bool("help", false, "help")
var verbose = flag.Bool("verbose", false, "verbose output")
var tmpl = flag.String("template", "", "use a template for the output. Available information are in Oplog struct (Ts,Ns,H,V,Op,O). For instance for graphite, it could be: 'DT.my.measure {{.Ts}}  {{.Timestamp.Unix()}} '")

/* Size type */
type Size string

func (s Size) Ko() int64 {
	if strings.HasSuffix(string(s), "G") {
		res, err := strconv.ParseFloat(strings.Trim(string(s), "G"), 64)
		if err == nil {
			return int64(res * 1024 * 1024)
		}
	}
	return 0
}

func (s Size) O() int64 {
	if strings.HasSuffix(string(s), "G") {
		res, err := strconv.ParseFloat(strings.Trim(string(s), "G"), 64)
		if err == nil {
			return int64(res * 1024 * 1024 * 1024)
		}
	}
	return 0
}

/* Stats structure */
type Stats struct {
	ArAw      string `json:"ar|aw"`
	Command   string `json:command`
	Conn      string `json:conn`
	Delete    string `json:delete`
	Faults    string `json:faults`
	Flushes   string `json:flushes`
	Getmore   string `json:getmore`
	Host      string `json:host`
	Insert    string `json:insert`
	Locked    string `json:locked`
	Mapped    string `json:mapped`
	NetIn     string `json:netIn`
	NetOut    string `json:netOut`
	NonMapped string `json:"non-mapped"`
	QrQw      string `json:"qr|qw"`
	Query     string `json:"query"`
	Res       string `json:"res"`
	Time      string `json:"time"`
	Update    string `json:"update"`
	Vsize     Size   `json:"vsize"`
}

type Output struct {
	Stats Stats
	Now   time.Time
}

func main() {

	flag.Parse()
	if *help {
		fmt.Fprintf(os.Stderr, "Parse mongostat json format.\nUsage mongostat-parser [options]\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	log.Printf("starting...\n")

	dec := json.NewDecoder(os.Stdin)
	for {
		var stats map[string]Stats
		if err := dec.Decode(&stats); err == io.EOF {
			fmt.Printf("ok")
			break
		} else if err != nil {
			log.Fatal(err)
		}
		if *tmpl == "" {
			fmt.Printf("%+v\n", stats)
		} else {
			for k, v := range stats {
				log.Printf("get stat from server %s\n", k)
				tmplParsed, _ := template.New("output").Parse(*tmpl)
				output := Output{v, time.Now()}
				err := tmplParsed.Execute(os.Stdout, output)
				if err != nil {
					panic(err)
				} else {
					fmt.Fprintf(os.Stdout, "\n")
				}
			}
		}
	}

}
