[![build status](http://gitlab.socrate.vsct.fr/dt/mongoanonymize/badges/master/build.svg)](http://gitlab.socrate.vsct.fr/dt/mongoanonymize/commits/master)
[![coverage report](http://gitlab.socrate.vsct.fr/dt/mongoanonymize/badges/master/coverage.svg)](http://gitlab.socrate.vsct.fr/dt/mongoanonymize/commits/master)

_mongoanonymize_ exporte dans des fichiers json les commandes PAO en anonymizant certains champs.
Ces fichiers peuvent être utilisés pour être importées avec la commande _mongoimport_.

# Download

Les exécutables se trouvent sur nexus:

* [0.1.0](http://nexus/index.html#view-repositories;dt-releases~browsestorage~/com/vsct/dt/mongoanonymize/0.1.0/mongoanonymize-0.1.0)
* [0.0.4](http://nexus/index.html#view-repositories;dt-releases~browsestorage~/com/vsct/dt/mongoanonymize/0.0.4/mongoanonymize-0.0.4)
* [0.0.3](http://nexus/index.html#view-repositories;dt-releases~browsestorage~/com/vsct/dt/mongoanonymize/0.0.3/mongoanonymize-0.0.3)
* [0.0.2](http://nexus/index.html#view-repositories;dt-releases~browsestorage~/com/vsct/dt/mongoanonymize/0.0.2/mongoanonymize-0.0.2)
* [0.0.1](http://nexus/index.html#view-repositories;dt-releases~browsestorage~/com/vsct/dt/mongoanonymize/0.0.1/mongoanonymize-0.0.1)
* [0.0.0](http://nexus/index.html#view-repositories;dt-releases~browsestorage~/com/vsct/dt/mongoanonymize/0.0.0/mongoanonymize-0.0.0)

  [CHANGELOG](CHANGELOG.md)

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

Pour spécifier les champs à anonymiser, il faut le spécificer dans le fichier configuration toml (_config.toml_ par défaut). Ce fichier a ce format:

```toml
["champ.a.anonymiser"]
"method"="set"
"args"=["nouvelle valeur"]

["champ.a.mettre.a.null"]
"method"="nil"

["champ.date.a.changer"]
"method"="date"
"args"=["Jan 2 15:04:05 -0700 MST 2006", "Jan 1 00:00:00 -0100 MST 1900"]
```

On peut trouver des exemples pour [order](examples/config-order.toml) et [service item](examples/config-serviceitem.toml)



## exemple simple

     $ ./mongoanonymize --limit 10000 --config config.toml  --uri mongodb://user:password@servername:27017 -database database --output results --collection col1 --pages 4 --monitored


## exemple d'import dans la nouvelle base

Si on veut importer les documents dans la nouvelle base à partir des fichiers exportés:

    $ for i in {0..3}; do echo "import export-$i.bson"; mongorestore --host servername --port 27017 -u user -p pwd -d basepreprod -c collection "./export-$i.bson" & done


# Récupération des sources

Pré-requis : Définir un _GOSPACE_, répertoire où mettre ses projets go

    export GOPATH={PATH_TO_GOSPACE}
    $ go get --insecure gitlab.socrate.vsct.fr/dt/mongoanonymize

Les sources sont dans _$GOPATH/src/gitlab.socrate.vsct.fr/dt/mongoanonymize/_.

# build

## Build Local

    go test gitlab.socrate.vsct.fr/dt/mongoanonymize/test -v

Va lancer les tests du projet                       

    go build gitlab.socrate.vsct.fr/dt/mongoanonymize

Va générer un binaire dans le répertoire courant
    
    go install gitlab.socrate.vsct.fr/dt/mongoanonymize

Va générer un binaire dans _$GOPATH/bin_    

## Build sur Gitlab CI

A chaque _git push_, gitlab va pousser le nouveau build sur nexus (cf [.gitlab-ci.yml](.gitlab-ci.yml)). C'est la version contenue dans le fichier [VERSION](VERSION) qui sera utilisée.

# Release

Par exemple pour releaser la version courante _0.0.1_ et passer ensuite en _0.0.2_:

     $ ./release.sh 0.0.2

 Le script va mettre à jour [VERSION](VERSION), poser un nouveau tag _v0.0.1_ et mettre à jour le [CHANGELOG](CHANGELOG.md) en fonction des commentaires git.

