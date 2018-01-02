# mongostat-parser

Parse la sortie json de mongostat et la restitue selon un template.

## usage

Mode par défaut (pas de template):

    $ mongostat --json | mongostat-parser 
    map[bimonga:{ArAw:0|0 Command:26|0 Conn:379 Delete:*0 Faults:0 Flushes:0 Getmore:0 Host:bimonga Insert:*0 Locked: Mapped:16.6G NetIn:2k NetOut:16k NonMapped: QrQw:0|0 Query:*0 Res:1.3G Time:15:41:06 Update:*0 Vsize:34.5G}]

Les logs sont dans la sortie stderr, donc si on ne veut pas les afficher:

    $ mongostat --json | mongostat-parser --template="graphite.tpl" 2>/dev/null

Exemple d'envoie de métrique dans graphite à l'aide d'un template

    $ mongostat --json | mongostat-parser --template="DT.TDC.collectd.bimonga.mongostats.gauge-mdbpaoh1_insert {{.Stats.Insert.Unstarred}} {{.Now.Unix}}" 2>/dev/null | nc $GRAPHITE_HOST $GRAPHITE_PORT


Ce template peut être contenu dans un fichier:

    $ mongostat --json | mongostat-parser --template="graphite.tpl" 2>/dev/null | nc $GRAPHITE_HOST $GRAPHITE_PORT
    
## output et template

En sortie ```mongostat-parser``` fournit un struct:

    type Output struct {
       Stats Stats
       Now time.Time
    }
    
```Now``` représente le moment de la requête. Son type est 
[time.Time](http://golang.org/pkg/time/#Time) donc on peut appliquer dans le 
template les méthodes de conversion comme ```{{.Now.Unix}}``` pour avoir le 
temps unix.

```Stats``` est lui-même un struct qui contient les champs récupérer dans 
mongostat au format json:

    type Stats struct {
        ArAw      Piped   `json:"ar|aw"`
        Command   Piped   `json:command`
        Conn      Size    `json:conn`
        Delete    Starred `json:delete`
        Faults    string  `json:faults`
        Flushes   string  `json:flushes`
        Getmore   Starred `json:getmore`
        Host      string  `json:host`
        Insert    Starred `json:insert`
        Locked    string  `json:locked`
        Mapped    string  `json:mapped`
        NetIn     Size    `json:netIn`
        NetOut    Size    `json:netOut`
        NonMapped Size    `json:"non-mapped"`
        QrQw      Piped   `json:"qr|qw"`
        Query     Starred `json:"query"`
        Res       Size    `json:"res"`
        Time      string  `json:"time"`
        Update    Starred `json:"update"`
        Vsize     Size    `json:"vsize"`
    }

On peut voir qu'il existe 3 sous-types, en plus de ```string```:
* Piped: quand 2 valeurs sont séparées par un pipe (par ex 'ar|aw'). On peut 
alors appliquer les méthodes ```Left``` et ```Right``` pour récupérer l'un ou 
l'autre. Par exemple: "{{.Stats.ArAw.Left}}".
* Starred: quand 1 valeur peut être précédé d'une étoile pour indiquer que 
l'opération s'est fait sur le master. Dans ce cas on peut vouloir supprimer 
l'étoile: "{{.Stats.Insert.Unstarred}}"
* Size: quand une valeur contient une unité de taille (G,g,M,m,K,k,B,b). Dans ce
cas on peut lui demander de le convertir au format qu'on veut. Par exemple:
"{{.Stats.Res.Kb}}" convertira la taille mémoire résident en ```kb```.


 
