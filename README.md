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

## Downloads

You can find releases [here](https://github.com/voyages-sncf-technologies/mongotools/releases).


### Build from source

#### Requirements


You need a [go installation](https://golang.org/doc/install) (>= 1.9).

We use the official go package manager [dep](https://github.com/golang/dep). For installing it:

    $ go get -u github.com/golang/dep/cmd/dep

You need too to add $GOPATH/bin in your binary path (certainly _$PATH_ variable).

#### Downloading and building mongotools

For downloading the project, you can use _go get_:

    $ go get github.com/voyages-sncf-technologies/mongotools/cmd/...

With _go get_, as tools are _applications_ (main packages in _cmd_), go will automatically compile and add mongotools in your directory _GOPATH/bin_:

    $ ls $GOPATH/bin/
    mongooplog-tail    mongostat-collection  mongoanonymize  mongooplog-window  mongostat-parser

Now your sources are in _$GOPATH/src/github.com/voyages-sncf-technologies/mongotools_. For building the project you need now to get dependencies and then you can build or install the project:

    $ cd $GOPATH/src/github.com/voyages-sncf-technologies/mongotools
    $ ls 
    CHANGELOG.md  cmd  doccleaner  Gopkg.lock  Gopkg.toml  README.md
    $ dep ensure # dep command comes from go dep project, you need to add $GOPATH/bin in your _PATH_.
    $ go install github.com/voyages-sncf-technologies/mongotools/cmd/... # will replace binaries in $GOPATH/bin if there are changes

If you want (re)install one application:

    $ go install github.com/voyages-sncf-technologies/mongotools/cmd/mongoanonymize
    

## Tools

### mongoanonymize

Anonymize documents and export them to file or other collection (cf. [README](cmd/mongoanonymize/README.md)).

## Tools (Work In Progress)

Currently, these applications need to be improved or re-validated.

### mongostat-lag

Compute the lag between master and slaves (cf. [README](cmd/mongostat-lag/README.md)).

### mongostat-parser

Parses json output of mongostat and returns in according to a template (cf. [README](cmd/mongostat-parser/README.md)).

### mongooplog-window

Give the oplog time window between first and last oplog (cf. [README](cmd/mongooplog-window/README.md)).

### mongooplog-tail

A tail on oplog collection with common feature of a `tail` as filtering (cf. [README](cmd/mongooplog-tail/README.md)).
