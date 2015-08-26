# Mongo tools

Ensemble d'outils de la DT pour mongo.

## Cahier des charges

* Unix-style: chaque outil doit faire une et une seule chose, les outils pouvant se composer entre eux ou avec les outils mongo existants (mongostat, mongotop, mongo etc...) à l'aide de pipe unix
* Output-agnostic: on doit facilement customiser la sortie (csv, format graphite, etc...)
* No runtime dependency: on ne doit pas dépendre de l'installation d'un runtime (jdk, python etc...)

[Go](http://golang.org/) semble répondre à la plupart des critères:

* Gestion du [templating](http://golang.org/pkg/text/template/)
* Librairie mongo ([mgo](https://labix.org/mgo)) offrant beaucoup de possibilités bas et haut-niveau
* Produit des exécutables autonomes qu'on peut ajouter facilement dans le _PATH_
* les [mongotools](https://github.com/mongodb/mongo-tools) officiels de MongoInc sont aussi en go et utilise aussi la librairie mgo
## Tools

### mongostat-lag

Mesure le lag entre le master et ses slaves (cf. [README](mongostat-lag/README.md)).

### mongostat-parser

Parse la sortie json de mongostat et la restitue selon un template (cf. [README](mongostat-parser/README.md)).

### mongooplog-window

Donne la différence entre le premier et le dernier oplog. (cf. [README](mongooplog-window/README.md)).

### mongooplog-tail

Permet de faire un tail sur les oplogs en filtrant sur certains champs et par date. (cf. [README](mongooplog-tail/README.md)).