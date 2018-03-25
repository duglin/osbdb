package dbclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

// Take admin user/password
func NewDB(url string, u, p string) *DBConnection {
	req, err := http.NewRequest("POST", url, nil)
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
		fmt.Printf("Can't create connection: %s\n", err)
		return nil
	}
	body := []byte{}
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = ioutil.ReadAll(res.Body)
	}
	if res.StatusCode != http.StatusCreated {
		fmt.Printf("Can't create a DB: %s\n%s\n", res.Status, string(body))
		return nil
	}

	dbc := DBConnection{}
	err = json.Unmarshal(body, &dbc)
	if dbc.URL == "" {
		fmt.Printf("Missing DB URL in response: %s\n", body)
		return nil
	}

	return &DBConnection{
		URL:      dbc.URL,
		User:     dbc.User,
		Password: dbc.Password,
	}
}

func (db *DBConnection) Get(key string) ([]byte, error) {
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
		return nil, fmt.Errorf("Can't get data: %s\n%s\n", res.Status,
			string(body))
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
		return fmt.Errorf("Can't set key(%s): %s\n%s\n", key,
			res.Status, string(body))
	}
	return nil
}

func (db *DBConnection) DelKey(key string) error {
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
		return fmt.Errorf("Can't delete key: %s\n%s\n", res.Status, string(body))
	}
	return nil
}
