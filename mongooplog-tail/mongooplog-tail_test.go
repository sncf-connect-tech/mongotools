package main

import (
	. "com/vsct/dt/mongotools/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"com/vsct/dt/mongotools/Godeps/_workspace/src/gopkg.in/mgo.v2/bson"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"
)

var currentTime time.Time

func TestBuildQuery(t *testing.T) {

	Convey("Test the query build", t, func() {

		fmt.Println("test")

		query := buildQuery()

		fmt.Printf("%v", query)

		if (query["ts"].(bson.M))["$gte"] != bson.MongoTimestamp(6215538544622960640) {

			t.Errorf("ts: %v", (query["ts"].(bson.M))["$gte"])

		}

		So((query["ts"].(bson.M))["$gte"], ShouldEqual, bson.MongoTimestamp(6215538544622960640))

	})

}

func TestNop(t *testing.T) {

	Convey("do nothing", t, func() {

		fmt.Println("do really nothing")

		Convey("what?", func() {

			So(true, ShouldBeTrue)

		})

	})

}

func TestMain(m *testing.M) {

	fmt.Println("main")

	flag.Parse()

	currentTime, _ = time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Nov 10, 2015 at 3:04pm (MST)")

	*op = "i"

	*ns = "test.collection"

	*sd = int(currentTime.Unix())

	fmt.Println(*sd)

	os.Exit(m.Run())

}
