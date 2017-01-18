# Mongo tools

Mongo tools used at VSCT.

## Guidelines

* [Unix philosophy](http://www.catb.org/esr/writings/taoup/html/ch01s06.html)
* [Go](http://golang.org/) has development language. The main reasons are: 
  * use [templating](http://golang.org/pkg/text/template/) for output allows to be agnostic on the format
  * [mgo](https://labix.org/mgo), the go driver, is popular and offer high and low level accesses to mongo api
  * ease to deploy on any platform

## Requirements

* [go installation](https://golang.org/doc/install)


## Build


### Build 
    
    $ GOPATH=$PWD go build mongostat-parser 
    $ GOPATH=$PWD go build mongostat-lag 
    $ GOPATH=$PWD go build mongooplog-window 
    $ GOPATH=$PWD go build mongooplog-tail 
    $ GOPATH=$PWD go build mongoexport-anonymization 

## Tools

### mongostat-lag

Compute the lag between master and slaves (cf. [README](mongostat-lag/README.md)).

### mongostat-parser

Parses json output of mongostat and returns in according to a template (cf. [README](mongostat-parser/README.md)).

### mongooplog-window

Give the oplog time window between first and last oplog (cf. [README](mongooplog-window/README.md)).

### mongooplog-tail

A tail on oplog collection with common feature of a `tail` as filtering (cf. [README](mongooplog-tail/README.md)).

### mongoexport-anonymization

Anonymize documents and export them to file or other collection (cf. [README](mongoexport-anonymization/README.md)).
