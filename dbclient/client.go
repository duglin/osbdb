package dbclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type DBConnection struct {
	URL      string
	User     string
	Password string
}

// Takes DB user/password
func NewDBConnection(url string, user, password string) *DBConnection {
	return &DBConnection{
		URL:      url,
		User:     user,
		Password: password,
	}
}

func (dbc *DBConnection) GetID() string {
	tmp := strings.TrimRight(dbc.URL, "/")
	if i := strings.LastIndex(tmp, "/"); i > 0 {
		return tmp[i+1:]
	}
	return ""
}

// Take admin user/password
func GetDBs(url string, u, p string) ([]*DBConnection, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}

	if u != "" {
		req.SetBasicAuth(u, p)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Can't get DBs: %s\n%s\n", res.Status,
			string(body))
	}

	dbs := []*DBConnection{}
	err = json.Unmarshal(body, &dbs)
	if err != nil {
		return nil, fmt.Errorf("Can't parse the DBs: %s", err)
	}

	return dbs, nil
}

// Take admin user/password
func GetDB(url string, id, u, p string) (*DBConnection, error) {
	req, err := http.NewRequest("GET", url+"/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}

	if u != "" {
		req.SetBasicAuth(u, p)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Can't get DBs: %s\n%s\n", res.Status,
			string(body))
	}

	db := DBConnection{}
	err = json.Unmarshal(body, &db)
	if err != nil {
		return nil, fmt.Errorf("Can't parse the DBs: %s", err)
	}

	return &db, nil
}

// Take admin user/password
func NewDB(url string, u, p string) (*DBConnection, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}

	if u != "" {
		req.SetBasicAuth(u, p)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Can't create a DB: %s\n%s\n", res.Status,
			string(body))
	}

	dbc := DBConnection{}
	err = json.Unmarshal(body, &dbc)
	if dbc.URL == "" {
		return nil, fmt.Errorf("Missing DB URL in response: %s\n", body)
	}

	return &DBConnection{
		URL:      dbc.URL,
		User:     dbc.User,
		Password: dbc.Password,
	}, nil
}

// Take admin user/password
func NewDBByID(url string, id, u, p string) (*DBConnection, error) {
	req, err := http.NewRequest("PUT", url+"/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}

	if u != "" {
		req.SetBasicAuth(u, p)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Can't create a DB: %s\n%s\n", res.Status,
			string(body))
	}

	dbc := DBConnection{}
	err = json.Unmarshal(body, &dbc)
	if dbc.URL == "" {
		return nil, fmt.Errorf("Missing DB URL in response: %s\n", body)
	}

	return &DBConnection{
		URL:      dbc.URL,
		User:     dbc.User,
		Password: dbc.Password,
	}, nil
}

// Take admin user/password
func DeleteDB(url string, u, p string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		fmt.Printf("Can't create connection: %s\n", err)
		return nil
	}

	if u != "" {
		req.SetBasicAuth(u, p)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Can't delete DB: %s(%s)", string(body), res.Status)
	}
	return nil
}

func (db *DBConnection) Get(key string) (string, error) {
	v, err := db.GetAsBytes(key)
	return string(v), err
}

func (db *DBConnection) GetAsBytes(key string) ([]byte, error) {
	req, err := http.NewRequest("GET", db.URL+"/"+key, nil)
	if err != nil {
		return nil, fmt.Errorf("Can't create http request: %s\n", err)
	}

	if db.User != "" {
		req.SetBasicAuth(db.User, db.Password)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}

	if res.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Can't get data: %s(%s)", body, res.Status)
	}

	return body, nil
}

func (db *DBConnection) Set(key string, value string) error {
	return db.SetAsBytes(key, []byte(value))
}

func (db *DBConnection) SetAsBytes(key string, value []byte) error {
	req, err := http.NewRequest("PUT", db.URL+"/"+key, bytes.NewReader(value))
	if err != nil {
		return fmt.Errorf("Can't create http request: %s\n", err)
	}

	if db.User != "" {
		req.SetBasicAuth(db.User, db.Password)
	}

	if value == nil {
		req.Header.Add("X-NULL", "true")
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Can't set key(%s): %s(%s)", key,
			string(body), res.Status)
	}
	return nil
}

func (db *DBConnection) DeleteKey(key string) error {
	req, err := http.NewRequest("DELETE", db.URL+"/"+key, nil)
	if err != nil {
		return fmt.Errorf("Can't create http request: %s\n", err)
	}

	if db.User != "" {
		req.SetBasicAuth(db.User, db.Password)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Can't delete key: %s(%s)\n", res.Status,
			string(body))
	}
	return nil
}

func (db *DBConnection) DeleteDB() error {
	req, err := http.NewRequest("DELETE", db.URL, nil)
	if err != nil {
		return fmt.Errorf("Can't create http request: %s\n", err)
	}

	if db.User != "" {
		req.SetBasicAuth(db.User, db.Password)
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Can't create connection: %s\n", err)
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Can't delete DB: %s(%s)", string(body), res.Status)
	}
	return nil
}
