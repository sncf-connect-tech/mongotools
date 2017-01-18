# build

    $ go get gopkg.in/mgo.v2
    $ go get gopkg.in/mgo.v2/bson
    $ go build mongo-export.go
    
    
# run

    $ ./mongo-export -h
      Usage of ./mongo-export:
        -batch int
          	batch size for mongo requests (default 100)
        -collection string
          	collection name (default "test")
        -database string
          	database name (default "test")
        -help
          	help
        -iterators int
          	number of iterators (default 20)
        -limit int
          	number max of documents (default 1000)
        -output string
          	path to output directory (default "./")
        -pages int
          	number of parallel mongo queries. For each page query, limit will be 'limit/pages'. (default 1)
        -prefetch float
          	prefetch ratio from batch for mongo requests (default 0.5)
        -unmarshallers int
          	number of unmarshallers (default 20)
        -uri string
          	mongo uri (default "mongodb://localhost:27017")
        -writers int
          	number of writers (default 40)


# example simple


     $ ./mongo-export --limit 10000 --uri mongodb://mdbrooe1:mdbrooe1@quadrelle:27017/BEROO01 -database BEROO01 --output results --collection order --pages 4
     

