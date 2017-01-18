package main

import "flag"
import "fmt"
import "log"
import "os"
import "text/template"
import "encoding/json"
import "io"
import "math"
import "strconv"
import "strings"
import "time"

/* CLI configuration */

var help = flag.Bool("help", false, "help")

var tmpl = flag.String("template", "", "use a template for the output. Available information are in Output struct {Stats,Now}. Type of Now is time.Time. Stats is a custom type which contains all mongostats fields and helpers for parsing these fields. For instance '{{.Now.Unix}} {{.Stats.ArAw.Left}}' displays unix time of now and left part of 'ar|aw' mongostat metric (See README.md for more details)")

/* Size type */

type Size string

// convert Size to KB
func (s Size) KB() int64 {
	return int64(s.ToInt64() / 1024)
}

// convert Size to MB
func (s Size) MB() int64 {
	return int64(s.ToInt64() / (1024 * 1024))
}

// convert Size to GB
func (s Size) GB() int64 {
	return int64(s.ToInt64() / (1024 * 1024 * 1024))
}

// convert string of Size to the unitary value (Size("1k").ToInt64() => 1024)
func (s Size) ToInt64() int64 {
	sizeSlice := []byte(string(s))
	power := 0.0
	res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgBbMmKk"), 64)
	switch sizeSlice[len(sizeSlice)-1] {
	case 'G', 'g':
		power = 3.0
	case 'M', 'm':
		power = 2.0
	case 'K', 'k':
		power = 1.0
	case 'B', 'b':
		power = 0.0
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		power = 0.0
	default:
		err = fmt.Errorf("can't find type of Size %v", s)
	}
	if err == nil {
		return int64(res * math.Pow(1024.0, float64(power)))
	} else {
		fmt.Println(err)
	}
	return -1
}

/* Data with a pipe separator */
type Piped string

func (p Piped) Left() string {
	return strings.Split(string(p), "|")[0]
}

func (p Piped) Right() string {
	return strings.Split(string(p), "|")[1]
}

/* Data with potentially a star */

type Starred string

func (s Starred) Unstarred() string {

	return strings.Trim(string(s), "*")

}

type NodeType string

func (s NodeType) ToInt() int {
	nodeType := 0
	switch s {
	case "PRI": // Value retrieved from real life, not included in doc
		nodeType = 1
	case "M": // Values from doc
		nodeType = 2
	case "SEC":
		nodeType = 3
	case "REC":
		nodeType = 4
	case "UNK":
		nodeType = 5
	case "SLV":
		nodeType = 6
	case "RTR":
		nodeType = 7
	case "ARB":
		nodeType = 8
	default:
		nodeType = 5
	}

	return nodeType
}

type Float string

func (s Float) ToInt() int {
	r, err := strconv.ParseFloat(string(s), 32)
	if err != nil {
		panic(err)
	}
	return int(r)
}

/* Stats structure */

type Stats struct {
	ArAw      Piped    `json:"ar|aw"`
	Command   Piped    `json:"command"`
	Conn      Size     `json:"conn"`
	Delete    Starred  `json:"delete"`
	Faults    string   `json:"faults"`
	Flushes   string   `json:"flushes"`
	Getmore   Starred  `json:"getmore"`
	Host      string   `json:"host"`
	Insert    Starred  `json:"insert"`
	Locked    string   `json:"locked"`
	Mapped    Size     `json:"mapped"`
	NetIn     Size     `json:"netIn"`
	NetOut    Size     `json:"netOut"`
	NonMapped Size     `json:"non-mapped"`
	QrQw      Piped    `json:"qr|qw"`
	Query     Starred  `json:"query"`
	Res       Size     `json:"res"`
	Time      string   `json:"time"`
	Update    Starred  `json:"update"`
	Vsize     Size     `json:"vsize"`
	NodeType  NodeType `json:"repl"`
	Dirty     Float    `json:"% dirty"`
	Used      Float    `json:"% used"`
}

type Output struct {
	Stats Stats
	Now   time.Time
}

func main() {

	flag.Parse()
	if *help {
		fmt.Fprintf(os.Stderr, "Parse mongostat json format.\nUsage mongostat-parser [options]\nCurrent usage: %v\n", os.Args[0])
		flag.PrintDefaults()
		return
	}

	log.Printf("starting...\n")

	dec := json.NewDecoder(os.Stdin)
	for {
		var stats map[string]Stats
		if err := dec.Decode(&stats); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		if *tmpl == "" {
			fmt.Printf("%+v\n", stats)
		} else {
			var err error
			var tmplParsed *template.Template
			if strings.HasPrefix(*tmpl, "file://") {
				log.Printf("read template from file: %s", *tmpl)
				tmplParsed, err = template.ParseFiles(strings.TrimPrefix(*tmpl, "file://"))
			} else {
				log.Printf("read template from argument: %s", *tmpl)
				tmplParsed, err = template.New("output").Parse(*tmpl)
			}
			if err != nil {
				panic(err)
			}

			for k, v := range stats {
				log.Printf("get stat from server %s\n", k)
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
