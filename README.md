[![Build Status](https://travis-ci.org/voyages-sncf-technologies/mongotools.svg?branch=ci%2Frefacto)](https://travis-ci.org/voyages-sncf-technologies/mongotools)
[![codecov](https://codecov.io/gh/voyages-sncf-technologies/mongotools/branch/master/graph/badge.svg)](https://codecov.io/gh/voyages-sncf-technologies/mongotools)

# Mongo tools

Mongo tools developped and used at VSCT.

## Guidelines

* [Unix philosophy](http://www.catb.org/esr/writings/taoup/html/ch01s06.html)
* [Go](http://golang.org/) has development language. The main reasons are: 
  * use [templating](http://golang.org/pkg/text/template/) for output allows to be agnostic on the format
  * [mgo](https://labix.org/mgo), the go driver, is popular and offer high and low level accesses to mongo api
  * ease to deploy on any platform

## Requirements

* [go installation](https://golang.org/doc/install)


## Downloads

You can find releases [here](https://github.com/voyages-sncf-technologies/mongotools/releases).


### Build 

    $ go build github.com/voyages-sncf-technologies/mongotools/cmd/mongoanonymize

## Tools

### mongoanonymize

Anonymize documents and export them to file or other collection (cf. [README](cmd/mongoanonymize/README.md)).

## Tools (Work In Progress)

### mongostat-lag

Compute the lag between master and slaves (cf. [README](cmd/mongostat-lag/README.md)).

### mongostat-parser

Parses json output of mongostat and returns in according to a template (cf. [README](cmd/mongostat-parser/README.md)).

### mongooplog-window

Give the oplog time window between first and last oplog (cf. [README](cmd/mongooplog-window/README.md)).

### mongooplog-tail

A tail on oplog collection with common feature of a `tail` as filtering (cf. [README](cmd/mongooplog-tail/README.md)).
