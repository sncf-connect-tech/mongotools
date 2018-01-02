# mongooplog-tail

## features

* tail on oplog collection of a cluster

## usage

```
$  mongooplog-tail -h
Usage of mongooplog-tail:
  -help=false: help
  -namespace="": namespace of oplog (for instance 'mydb.mycoll'). By default is empty, so there is no filtering.
  -operation="": show only these operations of oplog (for instance 'c','i','u' etc...). By default is empty, so there is no filtering.
  -startdate=-1: timestamp of the start date. Timestamp in seconds as stored in oplog.
  -startincr=0: timestamp of the start date. Increment part of the timestamp as stored in oplog. By default is 0. If stardate is not setted, this parameter is mute.
  -template="": use a template for the output. Available information are in a struct. For instance for graphite, it could be: 'DT.my.measure {{.Ts}}  {{.Timestamp.Unix}} '. Type struct is: TsRaw bson.Raw, Ns string, H int64, V int, Op string, O map[string]interface{}, TsDateTime time.Time, TsIncr int32, CurrentTime time.Time
  -timeout=-1: timeout in seconds. Beyond this timeout without new oplog, the process returns. By default -1 (disable timeout).
  -uri="mongodb://localhost:27017": mongo uri
```
## output

Output struct:
```
// Oplog struct
type Oplog struct {
	TsRaw       bson.Raw               "ts"
	Ns          string                 "ns"
	H           int64                  "h"
	V           int                    "v"
	Op          string                 "op"
	O           map[string]interface{} "o"
	TsDateTime  time.Time
	TsIncr      int32
	CurrentTime time.Time
}

```

| Name | Description | Example |
| ----- | --------------------------------------------------------------------------------------------------- | --------------------------------------- |
| TsRaw | raw format of the oplog timestamp. The timestamp contains the timestamp by second and an increment | `{Kind:17 Data:[1 0 0 0 34 127 19 86]}` |
| Ns | the namespace of the operation | `test.test` |
| V | version of the operation |  |
| Op | type of operation (i for insert, u for update ...) |  |
| O | operation content | `map[test:test _id:ObjectIdHex("56137f39b37ea91354657c4a")]` | 
| TsDateTime | conversion of TsRaw in `time.Time` | | 
| TsIncr | conversion of the incrment in `int32` | |  
| CurrentTime | current time in `time.Time` | |


## examples

### by default

```
23:16 $ mongooplog-tail
2015/10/27 23:16:20 starting...
2015/10/27 23:16:20 oplog size: 4
{TsRaw:{Kind:17 Data:[1 0 0 0 34 127 19 86]} Ns: H:0 V:2 Op:n O:map[msg:Reconfig set version:2] TsDateTime:2015-10-06 09:58:26 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:16:20.837202584 +0100 CET}
{TsRaw:{Kind:17 Data:[1 0 0 0 57 127 19 86]} Ns:test.$cmd H:-5238967533901447151 V:2 Op:c O:map[create:test] TsDateTime:2015-10-06 09:58:49 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:16:20.83726855 +0100 CET}
{TsRaw:{Kind:17 Data:[2 0 0 0 57 127 19 86]} Ns:test.test H:4369333125721920517 V:2 Op:i O:map[test:test _id:ObjectIdHex("56137f39b37ea91354657c4a")] TsDateTime:2015-10-06 09:58:49 +0200 CEST TsIncr:2 CurrentTime:2015-10-27 23:16:20.837289814 +0100 CET}
{TsRaw:{Kind:17 Data:[1 0 0 0 207 128 19 86]} Ns:test.test H:3831806247746119568 V:2 Op:i O:map[_id:ObjectIdHex("561380cfb37ea91354657c4b") test:test] TsDateTime:2015-10-06 10:05:35 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:16:20.837334158 +0100 CET}
```

### filter by operation (`-operation=i`)

```
23:16 $ mongooplog-tail -operation=i
2015/10/27 23:23:27 starting...
2015/10/27 23:23:28 oplog size: 4
{TsRaw:{Kind:17 Data:[2 0 0 0 57 127 19 86]} Ns:test.test H:4369333125721920517 V:2 Op:i O:map[test:test _id:ObjectIdHex("56137f39b37ea91354657c4a")] TsDateTime:2015-10-06 09:58:49 +0200 CEST TsIncr:2 CurrentTime:2015-10-27 23:23:28.00193202 +0100 CET}
{TsRaw:{Kind:17 Data:[1 0 0 0 207 128 19 86]} Ns:test.test H:3831806247746119568 V:2 Op:i O:map[_id:ObjectIdHex("561380cfb37ea91354657c4b") test:test] TsDateTime:2015-10-06 10:05:35 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:23:28.002006574 +0100 CET}
```

### filter by namespace (`-namespace=test.$cmd`)

```
23:27 $ mongooplog-tail -namespace='test.$cmd'
2015/10/27 23:28:00 starting...
2015/10/27 23:28:00 oplog size: 4
{TsRaw:{Kind:17 Data:[1 0 0 0 57 127 19 86]} Ns:test.$cmd H:-5238967533901447151 V:2 Op:c O:map[create:test] TsDateTime:2015-10-06 09:58:49 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:28:00.94122196 +0100 CET}
```

### display timestamp in second thanks to template

```
23:34 $ mongooplog-tail -template="{{.CurrentTime}} {{.TsDateTime.Unix}} {{.Op}} {{.O}}"
2015/10/27 23:34:33 starting...
2015/10/27 23:34:33 oplog size: 4
2015-10-27 23:34:33.163127871 +0100 CET 1444118306 n map[msg:Reconfig set version:2]
2015-10-27 23:34:33.163303357 +0100 CET 1444118329 c map[create:test]
2015-10-27 23:34:33.16335678 +0100 CET 1444118329 i map[_id:ObjectIdHex("56137f39b37ea91354657c4a") test:test]
2015-10-27 23:34:33.163429734 +0100 CET 1444118735 i map[test:test _id:ObjectIdHex("561380cfb37ea91354657c4b")]
```

### filter by start date

```
23:38 $ mongooplog-tail -startdate 1444118329 -template="{{.CurrentTime}} {{.TsDateTime.Unix}} {{.Op}} {{.O}}"
2015/10/27 23:38:44 starting...
2015/10/27 23:38:44 oplog size: 4
2015/10/27 23:38:44 ask timestamp >= 6202440994609168384
2015-10-27 23:38:44.171588278 +0100 CET 1444118329 c map[create:test]
2015-10-27 23:38:44.171755638 +0100 CET 1444118329 i map[_id:ObjectIdHex("56137f39b37ea91354657c4a") test:test]
2015-10-27 23:38:44.171816615 +0100 CET 1444118735 i map[test:test _id:ObjectIdHex("561380cfb37ea91354657c4b")]
```

### timeout

```
23:42 $ mongooplog-tail -timeout 3
2015/10/27 23:42:28 starting...
2015/10/27 23:42:28 oplog size: 4
{TsRaw:{Kind:17 Data:[1 0 0 0 34 127 19 86]} Ns: H:0 V:2 Op:n O:map[msg:Reconfig set version:2] TsDateTime:2015-10-06 09:58:26 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:42:28.595410162 +0100 CET}
{TsRaw:{Kind:17 Data:[1 0 0 0 57 127 19 86]} Ns:test.$cmd H:-5238967533901447151 V:2 Op:c O:map[create:test] TsDateTime:2015-10-06 09:58:49 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:42:28.595479109 +0100 CET}
{TsRaw:{Kind:17 Data:[2 0 0 0 57 127 19 86]} Ns:test.test H:4369333125721920517 V:2 Op:i O:map[_id:ObjectIdHex("56137f39b37ea91354657c4a") test:test] TsDateTime:2015-10-06 09:58:49 +0200 CEST TsIncr:2 CurrentTime:2015-10-27 23:42:28.595500486 +0100 CET}
{TsRaw:{Kind:17 Data:[1 0 0 0 207 128 19 86]} Ns:test.test H:3831806247746119568 V:2 Op:i O:map[_id:ObjectIdHex("561380cfb37ea91354657c4b") test:test] TsDateTime:2015-10-06 10:05:35 +0200 CEST TsIncr:1 CurrentTime:2015-10-27 23:42:28.595543985 +0100 CET}
2015/10/27 23:42:38 timeout reached (3 seconds)
```