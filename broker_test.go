package main

import (
	// "context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"./dbclient"
)

var testHost = "localhost:80"
var testUser = ""
var testPassword = ""

func TestMain(m *testing.M) {
	rc := 0

	verbose = 0
	testUser = brokerUser
	testPassword = brokerPassword

	go StartServer()
	time.Sleep(1 * time.Second)

	// Run first w/o any auth
	fmt.Printf("Running w/o any auth\n")
	disableAuth = true
	rc += m.Run()

	// Now run all tests again with auth
	fmt.Printf("\nRunning with auth\n")
	disableAuth = false
	rc += m.Run()

	os.Exit(rc)
}

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
	testURL := fmt.Sprintf("http://%s/db", testHost)

	CleanDBs(t, testURL, testUser, testPassword)
	defer CleanDBs(t, testURL, testUser, testPassword)

	db, err := dbclient.NewDB(testURL, testUser, testPassword)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")

	db1, err := dbclient.GetDB(testURL, db.GetID(), db.User, db.Password)
	Assert(t, err == nil, "Error getting DB: %s", err)
	Assert(t, db1.URL == db.URL, "URLs should match(%q,%q)", db1.URL, db.URL)
	Assert(t, db1.User == db.User, "Users should match(%q,%q)", db1.User, db.User)
	Assert(t, db1.Password == db.Password, "Passwords should match(%q,%q)", db1.Password, db.Password)

	db2, err := dbclient.NewDB(testURL, testUser, testPassword)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != db2.URL, "DB URLs should not match")
	Assert(t, db.Password != db2.Password, "DB passwords should not match")
	Assert(t, db2.URL != "", "Bad db.URL from create")
	Assert(t, db2.User != "", "Bad db.User from create")
	Assert(t, db2.Password != "", "Bad db.Password from create")

}

func TestCreates2(t *testing.T) {
	testURL := fmt.Sprintf("http://%s/db", testHost)

	CleanDBs(t, testURL, testUser, testPassword)
	defer CleanDBs(t, testURL, testUser, testPassword)

	db, err := dbclient.NewDBByID(testURL, "100", testUser, testPassword)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")

	db2, err := dbclient.NewDBByID(testURL, "100", testUser, testPassword)
	Assert(t, err != nil, "Creating DB should have failed")
	Assert(t, db2 == nil, "Failed DB should be nil")

	db2, err = dbclient.NewDBByID(testURL, "101", testUser, testPassword)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")
}

func TestGetDB(t *testing.T) {
	testURL := fmt.Sprintf("http://%s/db", testHost)

	CleanDBs(t, testURL, testUser, testPassword)
	defer CleanDBs(t, testURL, testUser, testPassword)

	db, err := dbclient.NewDB(testURL, testUser, testPassword)
	Assert(t, err == nil, "Error creating DB: %s", err)
	Assert(t, db.URL != "", "Bad db.URL from create")
	Assert(t, db.User != "", "Bad db.User from create")
	Assert(t, db.Password != "", "Bad db.Password from create")

	db2, err := dbclient.GetDB(testURL, db.GetID(), brokerUser, brokerPassword)
	Assert(t, err == nil, "Error getting DB: %s", err)
	Assert(t, db.URL == db2.URL, "URLs should match %q,%q", db.URL, db2.URL)
	Assert(t, db.User == db2.User, "Users should match %q,%q", db.User, db2.User)
	Assert(t, db.Password == db2.Password, "Passwords should match %q,%q", db.Password, db.Password)
}

func TestGetSetString(t *testing.T) {
	testURL := fmt.Sprintf("http://%s/db", testHost)

	CleanDBs(t, testURL, testUser, testPassword)
	defer CleanDBs(t, testURL, testUser, testPassword)

	db, err := dbclient.NewDB(testURL, testUser, testPassword)
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

func TestAuth(t *testing.T) {
	if disableAuth == true {
		t.SkipNow()
	}

	testURL := fmt.Sprintf("http://%s/db", testHost)

	CleanDBs(t, testURL, testUser, testPassword)
	defer CleanDBs(t, testURL, testUser, testPassword)

	_, err := dbclient.NewDB(testURL, "badUser", testPassword)
	Assert(t, err != nil, "Create should have failed")

	_, err = dbclient.NewDB(testURL, testUser, "badPassword")
	Assert(t, err != nil, "Create should have failed")

	db, err := dbclient.NewDB(testURL, testUser, testPassword)
	Assert(t, err == nil, "Create should have worked")

	saveUser := db.User
	savePassword := db.Password

	db.User = "badUser"
	err = db.DeleteDB()
	Assert(t, err != nil, "Delete should have failed")

	db.User = saveUser
	db.Password = "badPassword"
	err = db.DeleteDB()
	Assert(t, err != nil, "Delete should have failed")

	db.Password = savePassword
}
