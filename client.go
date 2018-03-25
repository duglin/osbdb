package main

import (
	"fmt"
	"os"

	"./dbclient"
)

var host = "localhost"
var port = 80

func main() {
	db, _ := dbclient.NewDB("http://localhost/db", "", "")

	err := db.Set("foo", "hello")
	fmt.Printf("Setting 'foo' to 'hello' Err: %s\n", err)
	vb, err := db.GetAsBytes("foo")
	fmt.Printf("Getting 'foo': %#v(%s) Err: %s\n", vb, string(vb), err)

	err = db.Set("foo", "")
	fmt.Printf("Setting 'foo' to '' Err: %s\n", err)
	vb, err = db.GetAsBytes("foo")
	fmt.Printf("Getting 'foo': %#v(%s) Err: %s\n", vb, string(vb), err)

	err = db.SetAsBytes("foo", nil)
	fmt.Printf("Setting 'foo' to nil Err: %s\n", err)
	vb, err = db.GetAsBytes("foo")
	fmt.Printf("Getting 'foo': %#v(%s) Err: %s\n", vb, string(vb), err)

	os.Exit(0)

	v, err := db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)

	err = db.SetAsBytes("foo", nil)
	fmt.Printf("Setting 'foo' to nil Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)

	err = db.DeleteKey("foo")
	fmt.Printf("Deleting 'foo' Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)

	err = db.SetAsBytes("foo", nil)
	fmt.Printf("Setting 'foo' to nil Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %#v Err: %s\n", v, err)

	err = db.Set("foo", "")
	fmt.Printf("Setting 'foo' to '' Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %#v Err: %s\n", v, err)

	err = db.DeleteDB()
	fmt.Printf("Deleting DB: %v Err: %s\n", err)

	v, err = db.Get("foo")
	fmt.Printf("Getting 'foo': %v Err: %s\n", v, err)

	err = dbclient.DeleteDB(db.URL, "", "")
	fmt.Printf("Deleting DB: %v Err: %s\n", err)

}
