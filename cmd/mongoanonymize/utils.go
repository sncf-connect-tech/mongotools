package main

import "gopkg.in/mgo.v2/bson"

// replaceIfPresent replace field of the document if exists by given value
func replaceIfPresent(element bson.M, field string, value interface{}) {
	if _, ok := element[field]; ok {
		element[field] = value
	}
}
