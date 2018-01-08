_mongoanonymize_ export mongodb documents into bson files by anonymizing some fields.
Then these files can be imported into another collection thanks to the standard command _mongoimport_. It's really useful for exporting a production database into a testing database.

# Run

```
10:11 $ mongoanonymize --help
Version 0.0.0 for git 2d1d7e8/master/dirty [v0.0.0-17-g2d1d7e8-dirty] at 2017-06-20T08:10:30Z
Window of the oplog (time of last oplog - time of first oplog).

Anonymize command fields to files.
Usage:
$ mongoanonymize [options]
Given arguments: mongoanonymize
  -batch int
    	batch size for mongo requests (default 100)
  -collection string
    	collection name (default "test")
  -compressed
    	compress exported files to gzip
  -config string
    	config file (default "config.toml")
  -database string
    	database name (default "test")
  -help
    	help
  -ids string
    	list of mongo ids for starting each pages. If empty, will compute them. If not empty, pages parameter is ignored.
  -limit int
    	number max of documents. By default export all documents (default -1)
  -monitor-port string
    	monitor port for go runtime metrics (ex: localhost:50080/debug/charts) (default "50080")
  -monitored
    	display current status each second
  -noanonymized
    	document won't be anonymize. For testing purpose.
  -output string
    	path to output directory (default "./results")
  -pages int
    	number of parallel mongo queries. For each page query, limit will be 'limit/pages'. If undefined, pages=number of cpus (default -1)
  -prefetch float
    	prefetch ratio from batch for mongo requests (default 0.5)
  -uri string
    	mongo uri (default "mongodb://localhost:27017")

```

## configuration

A toml configuration file allows to set fields to anonymize (_./config.toml_ by default). This file has the format below:

```toml
["field.to.anonymize"]
"method"="set"
"args"=["new value"]

["field.to.set.at.null"]
"method"="nil"

["field.date.to.change"]
"method"="date"
# first argument is date format, second argument is forced date
"args"=["Jan 2 15:04:05 -0700 MST 2006", "Jan 1 00:00:00 -0100 MST 1900"] 
```


We can find some examples [here](examples/config-order.toml) and [here](examples/config-serviceitem.toml)



## simple example

     $ ./mongoanonymize --limit 10000 --config config.toml  --uri mongodb://user:password@servername:27017 -database database --output results --collection col1 --pages 4 --monitored


## import example into new database

An example if we want import bson files with anonymized fields into a new database/collection:

    $ for i in {0..3}; do echo "import export-$i.bson"; mongorestore --host servername --port 27017 -u user -p pwd -d basepreprod -c collection "./export-$i.bson" & done

