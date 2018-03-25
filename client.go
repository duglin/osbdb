package main

import (
	"fmt"

	"./dbclient"
)

var host = "localhost"
var port = 80

func main() {
	db := dbclient.NewDB("http://localhost/db", "", "")
	err := db.Set("foo", "hello")
	fmt.Printf("Setting 'foo' to 'hello' Err: %s\n", err)

	v, err := db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)

	err = db.SetAsBytes("foo", nil)
	fmt.Printf("Setting 'foo' to nil Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)

	err = db.DelKey("foo")
	fmt.Printf("Deleting 'foo' Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)

	err = db.SetAsBytes("foo", nil)
	fmt.Printf("Setting 'foo' to nil Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)
}
