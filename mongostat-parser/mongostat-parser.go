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

var help = flag.Bool("help", false, "help")

var tmpl = flag.String("template", "", "use a template for the output. Available information are in Output struct {Stats,Now}. Type of Now is time.Time. Stats is a custom type which contains all mongostats fields and helpers for parsing these fields. For instance '{{.Now.Unix}} {{.Stats.ArAw.Left}}' displays unix time of now and left part of 'ar|aw' mongostat metric (See README.md for more details)")

/* Size type */

type Size string

func (s Size) Kb() int64 {

	sizeSlice := []byte(string(s))

	switch sizeSlice[len(sizeSlice)-1] {

	case 'G', 'g':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgBbMmKk"), 64)

		if err == nil {

			return int64(res * 1024 * 1024)

		} else {

			fmt.Println(err)

		}

	case 'M', 'm':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgbBMmKk"), 64)

		if err == nil {

			return int64(res * 1024)

		} else {

			fmt.Println(err)

		}

	case 'K', 'k':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgbBMmKk"), 64)

		if err == nil {

			return int64(res)

		} else {

			fmt.Println(err)

		}

	case 'B', 'b':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgbBMmKk"), 64)

		if err == nil {

			return int64(res / 1024)

		} else {

			fmt.Println(err)

		}

	}

	return -1

}

func (s Size) B() int64 {

	sizeSlice := []byte(string(s))

	switch sizeSlice[len(sizeSlice)-1] {

	case 'G', 'g':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgBbMmKk"), 64)

		if err == nil {

			return int64(res * 1024 * 1024 * 1024)

		} else {

			fmt.Println(err)

		}

	case 'M', 'm':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgbBMmKk"), 64)

		if err == nil {

			return int64(res * 1024 * 1024)

		} else {

			fmt.Println(err)

		}

	case 'K', 'k':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgbBMmKk"), 64)

		if err == nil {

			return int64(res * 1024)

		} else {

			fmt.Println(err)

		}

	case 'B', 'b':

		res, err := strconv.ParseFloat(strings.Trim(string(sizeSlice[:len(sizeSlice)]), "GgbBMmKk"), 64)

		if err == nil {

			return int64(res)

		} else {

			fmt.Println(err)

		}

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

/* Stats structure */

type Stats struct {
	ArAw Piped `json:"ar|aw"`

	Command Piped `json:command`

	Conn Size `json:conn`

	Delete Starred `json:delete`

	Faults string `json:faults`

	Flushes string `json:flushes`

	Getmore Starred `json:getmore`

	Host string `json:host`

	Insert Starred `json:insert`

	Locked string `json:locked`

	Mapped string `json:mapped`

	NetIn Size `json:netIn`

	NetOut Size `json:netOut`

	NonMapped Size `json:"non-mapped"`

	QrQw string `json:"qr|qw"`

	Query Starred `json:"query"`

	Res Size `json:"res"`

	Time string `json:"time"`

	Update Starred `json:"update"`
	Vsize  Size    `json:"vsize"`
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
