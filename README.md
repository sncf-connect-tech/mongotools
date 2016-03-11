# Mongo tools

Mongo tools used at VSCT.

## Guidelines

* [Unix philosophy](http://www.catb.org/esr/writings/taoup/html/ch01s06.html)
* [Go](http://golang.org/) has development language. The main reasons are: 
  * use [templating](http://golang.org/pkg/text/template/) for output allows to be agnostic on the format
  * [mgo](https://labix.org/mgo), the go driver, is popular and offer high and low level accesses to mongo api
  * ease to deploy on any platform

## Requirements

* [installation de go](https://golang.org/doc/install) et du `$GOPATH`, configure the [workspace](https://golang.org/doc/code.html)
* use [godep](http://github.com/tools/godep) for dependencies management:

```
$ go get github.com/tools/godep
```


## Build

### Using current $GOPATH
```
$ git clone git@gitlab.socrate.vsct.fr:dt/mongotools.git $GOPATH/src/com/vsct/dt/mongotools
$ cd $GOPATH
$ go install com/vsct/dt/mongotools/...
$ ls bin/ 
```

### Build a future release

```
$ ./build.sh
$ tree build/
build/
├── git-hash
├── linux
│   └── 386
│       ├── mongooplog-tail
│       ├── mongooplog-window
│       ├── mongostat-lag
│       └── mongostat-parser
└── mongotools.tar.gz

2 directories, 6 files 
```

### Build and deploy to nexus

```
$ ./build.sh -r $VERSION -n http://nexus/service/local/artifact/maven/content -nu <user> -np <password> -nr dt-releases
```

## Tools

### mongostat-lag

Compute the lag between master and slaves (cf. [README](mongostat-lag/README.md)).

### mongostat-parser

Parses json output of mongostat and returns in according to a template (cf. [README](mongostat-parser/README.md)).

### mongooplog-window

Give the oplog time window between first and last oplog (cf. [README](mongooplog-window/README.md)).

### mongooplog-tail

A tail on oplog collection with common feature of a `tail` as filtering (cf. [README](mongooplog-tail/README.md)).
