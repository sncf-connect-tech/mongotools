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



