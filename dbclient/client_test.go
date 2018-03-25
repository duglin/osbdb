package dbclient_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	dbclient "."
)

var host = "localhost"
var port = 80
var url = "http://lostlcaho:80/db/"
var user = ""
var password = ""

func Assert(t *testing.T, b bool, errStr string, args ...interface{}) {
	if !b {
		str := ""

		fpcs := make([]uintptr, 1)
		n := runtime.Callers(2, fpcs)
		if n != 0 {
			fun := runtime.FuncForPC(fpcs[0] - 1)
			f, l := fun.FileLine(fpcs[0] - 1)
			if i := strings.LastIndex(f, "/"); i > 0 {
				f = f[i+1:]
			}
			str = fmt.Sprintf("%s:%d", f, l)
		}
		fmt.Printf(str+":"+errStr+"\n", args...)
		t.FailNow()
	}
}

// Note: this implicity tests GetDBs on each test
func CleanDBs(t *testing.T, url string, user, password string) {
	dbs, err := dbclient.GetDBs(url, user, password)
	Assert(t, err == nil, "Can't get the DBs: %s", err)

	for _, db := range dbs {
		db.DeleteDB()
	}

	// Verify we're clean
	dbs, err = dbclient.GetDBs(url, user, password)
	Assert(t, err == nil, "Error getting DBs: %s", err)
	Assert(t, len(dbs) == 0, "Len of DBs should be zero, its %d", len(dbs))
}

func TestCreates1(t *testing.T) {
	testURL := fmt.Sprintf("http://%s:%d/db", host, port)

	CleanDBs(t, testURL, user, password)
	defer CleanDBs(t, testURL, user, password)

	db, err := dbclient.NewDB(testURL, user, password)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")

	db1, err := dbclient.GetDB(testURL, db.GetID(), user, password)
	Assert(t, db1.URL == db.URL, "URLs should match(%q,%q)", db1.URL, db.URL)
	Assert(t, db1.User == db.User, "Users should match(%q,%q)", db1.User, db.User)
	Assert(t, db1.Password == db.Password, "Passwords should match(%q,%q)", db1.Password, db.Password)

	db2, err := dbclient.NewDB(testURL, user, password)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != db2.URL, "DB URLs should not match")
	Assert(t, db.Password != db2.Password, "DB passwords should not match")
	Assert(t, db2.URL != "", "Bad db.URL from create")
	Assert(t, db2.User != "", "Bad db.User from create")
	Assert(t, db2.Password != "", "Bad db.Password from create")

}

func TestCreates2(t *testing.T) {
	testURL := fmt.Sprintf("http://%s:%d/db", host, port)

	CleanDBs(t, testURL, user, password)
	defer CleanDBs(t, testURL, user, password)

	db, err := dbclient.NewDBByID(testURL, "100", user, password)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")

	db2, err := dbclient.NewDBByID(testURL, "100", user, password)
	Assert(t, err != nil, "Creating DB should have failed")
	Assert(t, db2 == nil, "Failed DB should be nil")

	db2, err = dbclient.NewDBByID(testURL, "101", user, password)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")
}

func TestGetDB(t *testing.T) {
	testURL := fmt.Sprintf("http://%s:%d/db", host, port)

	CleanDBs(t, testURL, user, password)
	defer CleanDBs(t, testURL, user, password)

	db, err := dbclient.NewDB(testURL, user, password)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")

	db2, err := dbclient.GetDB(testURL, db.GetID(), user, password)
	Assert(t, err == nil, "Error getting DB: %s", err)
	Assert(t, db.URL == db2.URL, "URLs should match %q,%q", db.URL, db2.URL)
	Assert(t, db.User == db2.User, "Users should match %q,%q", db.User, db2.User)
	Assert(t, db.Password == db2.Password, "Passwords should match %q,%q", db.Password, db.Password)
}

func TestGetSetString(t *testing.T) {
	testURL := fmt.Sprintf("http://%s:%d/db", host, port)

	CleanDBs(t, testURL, user, password)
	defer CleanDBs(t, testURL, user, password)

	db, err := dbclient.NewDB(testURL, user, password)
	Assert(t, err == nil, "Error creating DB: %s", err)

	val, err := db.Get("prop1")
	Assert(t, err != nil, "Should fail to get prop")
	Assert(t, val == "", "Val should not be set: %v", val)

	err = db.Set("prop2", "hello")
	Assert(t, err == nil, "Failed to set prop2: %s", err)

	val, err = db.Get("prop2")
	Assert(t, err == nil, "Failed getting value: %s", err)
	Assert(t, val == "hello", "Val has wrong value: %q", val)

	// Make sure prop1 is still missing
	_, err = db.Get("prop1")
	Assert(t, err != nil, "Should fail to get prop")

	err = db.Set("prop2", "goodbye")
	Assert(t, err == nil, "Failed to set prop2: %s", err)

	val, err = db.Get("prop2")
	Assert(t, err == nil, "Failed getting value: %s", err)
	Assert(t, val == "goodbye", "Val has wrong value: %q", val)

	err = db.Set("prop2", "")
	Assert(t, err == nil, "Failed to set prop2: %s", err)

	val, err = db.Get("prop2")
	Assert(t, err == nil, "Failed getting value: %s", err)
	Assert(t, val == "", "Val has wrong value: %q", val)

	err = db.DeleteKey("prop2")
	Assert(t, err == nil, "Failed to delete prop2: %s", err)

	val, err = db.Get("prop2")
	Assert(t, err != nil, "Should fail to get prop")
	Assert(t, val == "", "Val should not be set: %v", val)
}
