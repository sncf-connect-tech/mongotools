package main

import (
	. "com/vsct/dt/mongotools/Godeps/_workspace/src/github.com/smartystreets/goconvey/convey"
	"fmt"
	"testing"
)


func TestParseStats(t *testing.T) {

	Convey("Test the parsing of node roles", t, func() {

		fmt.Println("testing node roles")

		var test NodeType
		So(test.ToInt(), ShouldEqual, 5)

		test = ""
		So(test.ToInt(), ShouldEqual, 5)

		test = "PRI"
		So(test.ToInt(), ShouldEqual, 1)

		test = "M"
		So(test.ToInt(), ShouldEqual, 2)

		test = "SEC"
		So(test.ToInt(), ShouldEqual, 3)

		test = "REC"
		So(test.ToInt(), ShouldEqual, 4)

		test = "UNK"
		So(test.ToInt(), ShouldEqual, 5)

		test = "SLV"
		So(test.ToInt(), ShouldEqual, 6)

		test = "RTR"
		So(test.ToInt(), ShouldEqual, 7)

		test = "ARB"
		So(test.ToInt(), ShouldEqual, 8)

		test = "toto"
		So(test.ToInt(), ShouldEqual, 5)

	})

}