package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

/* Misc Stuff */
/**************/
var verbose int = 3
var hostString string = ""
var brokerUser string = "user"
var brokerPassword string = "passw0rd"
var disableAuth bool = false

func Debug(level int, format string, args ...interface{}) {
	if verbose < level {
		return
	}
	fmt.Printf(format, args...)
}

func WriteJSON(w http.ResponseWriter, obj interface{}) {
	b, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		b = []byte(err.Error())
	}
	w.Write(b)
	w.Write([]byte("\n"))
}

func GeneratePassword() string {
	len := 6 + rand.Int()%6
	var password [12]byte

	for i := 0; i < len; i++ {
		r := rand.Int() % 62
		if r < 26 {
			password[i] = byte('a' + r)
		} else if r < 52 {
			password[i] = byte('A' + (r - 26))
		} else {
			password[i] = byte('0' + (r - 52))
		}
	}
	return string(password[:len])
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf(
		"OSB API Sample DB Broker\n"+
			"------------------------\n"+
			"User: %s\n"+
			"Password: %s\n"+
			"DBs: %d\n"+
			"Services: %d\n"+
			"Instances: %d\n",
		brokerUser, brokerPassword, len(DBs), len(catalog.Services),
		len(broker.Instances))
	w.Write([]byte(str))
}

func VerifyBasicAuth(w http.ResponseWriter, r *http.Request,
	user, password string) bool {

	if disableAuth || user == "" {
		return true
	}

	u, p, ok := r.BasicAuth()
	if ok && user == u && password == p {
		return true
	}
	w.WriteHeader(http.StatusUnauthorized)
	return false
}

/* Service(DB) Stuff */
/*********************/
var lastID int = 0
var DBs map[string]*DB = map[string]*DB{} // DBid -> DB

type DB struct {
	ID       string
	User     string
	Password string
	Data     map[string][]byte // key -> value
	URL      string            // Access URL
	mutex    sync.Mutex
}

var newDBIDmutex = &sync.Mutex{}
var DBMapmutex = &sync.Mutex{}

func NewDB(r *http.Request) *DB {
	strID := ""

	newDBIDmutex.Lock()
	for {
		lastID++
		strID = strconv.Itoa(lastID)
		if DBs[strID] == nil {
			break
		}
	}
	newDBIDmutex.Unlock()

	host := r.Host
	if hostString != "" {
		host = hostString
	}

	db := &DB{
		ID:       strID,
		User:     "user",
		Password: GeneratePassword(),
		Data:     map[string][]byte{},
		URL:      fmt.Sprintf("http://%s/db/"+strID, host),
		mutex:    sync.Mutex{},
	}

	DBMapmutex.Lock()
	DBs[db.ID] = db
	DBMapmutex.Unlock()

	Debug(2, "DB %s: created\n", db.ID)
	return db
}

func DeleteDB(db *DB) {
	DBMapmutex.Lock()
	delete(DBs, db.ID)
	DBMapmutex.Unlock()
	Debug(2, "DB %s: deleted\n", db.ID)
}

func DBAllHandler(w http.ResponseWriter, r *http.Request) {
	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	tmpDBs := []*DB{}
	for _, db := range DBs {
		tmpDBs = append(tmpDBs, db)
	}
	WriteJSON(w, tmpDBs)
}

func DBHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if dbID := vars["dbID"]; dbID != "" {
		if db := DBs[dbID]; db != nil {
			if !VerifyBasicAuth(w, r, db.User, db.Password) {
				return
			}
			WriteJSON(w, db)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func DBCreateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbID := vars["dbID"]

	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	if dbID == "" {
		db := NewDB(r)
		DBs[db.ID] = db
		w.Header().Add("Location", "/db/"+db.ID)
		w.WriteHeader(http.StatusCreated)
		WriteJSON(w, db)
		return
	}

	if DBs[dbID] != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	db := NewDB(r)
	DBs[dbID] = db

	// Add host:port
	w.Header().Add("Location", "/db/"+dbID)

	w.WriteHeader(http.StatusCreated)
	WriteJSON(w, db)
}

func DBDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if dbID := vars["dbID"]; dbID != "" {
		if db := DBs[dbID]; db != nil {
			if !VerifyBasicAuth(w, r, db.User, db.Password) &&
				!VerifyBasicAuth(w, r, brokerUser, brokerPassword) {

				return
			}
			DeleteDB(db)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func DBGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if dbID := vars["dbID"]; dbID != "" {
		if db := DBs[dbID]; db != nil {
			if !VerifyBasicAuth(w, r, db.User, db.Password) {
				return
			}
			if key := vars["key"]; key != "" {
				w.Write(db.Data[key])
				return
			}
		}
	}

	w.WriteHeader(http.StatusNotFound)
}

func DBSetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if db := DBs[vars["dbID"]]; db != nil {
		if !VerifyBasicAuth(w, r, db.User, db.Password) {
			return
		}
		if key := vars["key"]; key != "" {
			value, _ := ioutil.ReadAll(r.Body)
			db.Data[key] = value
			Debug(3, "DB %s: Set %q to %q\n", db.ID, key, string(value))
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func DBRemoveHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if db := DBs[vars["dbID"]]; db != nil {
		if !VerifyBasicAuth(w, r, db.User, db.Password) {
			return
		}
		if key := vars["key"]; key != "" {
			if _, ok := db.Data[key]; ok {
				delete(db.Data, key)
				Debug(3, "DB %s: Removed %q\n", db.ID, key)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

/* OSB APIs */
/************/

type Broker struct {
	Instances map[string]*Instance // InstanceID -> Instance
}

var broker = Broker{
	Instances: map[string]*Instance{},
}

type Instance struct {
	DB       *DB
	Request  ProvisionRequest
	Bindings map[string]interface{} // BindingID -> struct
}

type Catalog struct {
	Services []Service `json:"services"`
}

type Service struct {
	Name            string                 `json:"name"`
	ID              string                 `json:"id"`
	Description     string                 `json:"description"`
	Tags            []string               `json:"tags,omitempty"`
	Requires        []string               `json:"requires,omitempty"`
	Bindable        bool                   `json:"bindable"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	DashboardClient interface{}            `json:"dashboard_client,omitempty"`
	PlanUpdateable  bool                   `json:"plan_updateable,omitempty"`
	Plans           []Plan                 `json:"plans"`
}

type Plan struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Free        bool                   `json:"free,omitempty"`
	Bindable    bool                   `json:"bindable,omitempty"`
	Schemas     interface{}            `json:"schemas,omitempty"`
}

var catalog = Catalog{
	Services: []Service{
		Service{
			Name:        "db",
			ID:          "service-1-id",
			Description: "In-memory DB for demos",
			Bindable:    true,
			Plans: []Plan{
				Plan{
					ID:          "plan-1-id",
					Name:        "free",
					Description: "Totally free usage",
				},
			},
		},
	},
}

func WriteOSBError(w http.ResponseWriter, err, description string) {
	OSBError := struct {
		Error       string `json:"error"`
		Description string `json:"description,omitempty"`
	}{
		Error:       err,
		Description: description,
	}
	WriteJSON(w, OSBError)
}

func CatalogHandler(w http.ResponseWriter, r *http.Request) {
	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	WriteJSON(w, catalog)
}

type Context map[string]interface{}

type ProvisionRequest struct {
	ServiceID  string            `json:"service_id"`
	PlanID     string            `json:"plan_id"`
	Content    Context           `json:"context,omitempty"`
	OrgID      string            `json:"organization_guid"`
	SpaceID    string            `json:"space_guid"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type ProvisonResponse struct {
	DashboardURL string `json:"dashboard_url,omitempty"`
	Operation    string `json:"operation,omitempty"`
}

func ProvisionHandler(w http.ResponseWriter, r *http.Request) {
	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	vars := mux.Vars(r)

	instanceID := vars["iID"]
	if instanceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing InstanceID", "")
		return
	}

	pReq := ProvisionRequest{}
	body, _ := ioutil.ReadAll(r.Body)
	if err := json.Unmarshal(body, &pReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, err.Error(), "")
		return
	}

	if pReq.ServiceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing service_id", "")
		return
	}

	if pReq.PlanID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing plan_id", "")
		return
	}

	foundOne := false
	for _, service := range catalog.Services {
		if service.ID == pReq.ServiceID {
			for _, plan := range service.Plans {
				if plan.ID == pReq.PlanID {
					foundOne = true
					break
				}
			}
			break
		}
	}
	if !foundOne {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, fmt.Sprintf("Can't find service/plan %s/%s",
			pReq.ServiceID, pReq.PlanID), "")
		return
	}

	if i := broker.Instances[instanceID]; i != nil {
		if reflect.DeepEqual(i.Request.Parameters, pReq.Parameters) {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusConflict)
		WriteOSBError(w, fmt.Sprintf("Instance with that ID(%s) already exists",
			instanceID), "")
		return
	}

	Debug(2, "Instance %s: created\n", instanceID)
	broker.Instances[instanceID] = &Instance{
		DB:       NewDB(r),
		Request:  pReq,
		Bindings: map[string]interface{}{},
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{}"))
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	w.Write([]byte("Hello\n"))
}

func DeprovisionHandler(w http.ResponseWriter, r *http.Request) {
	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	vars := mux.Vars(r)

	instanceID := vars["iID"]
	if instanceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing InstanceID", "")
		return
	}

	params := r.URL.Query()

	serviceID := ""
	if params["service_id"] != nil {
		serviceID = params["service_id"][0]
	}
	if serviceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing ServiceID", "")
		return
	}

	planID := ""
	if params["plan_id"] != nil {
		planID = params["plan_id"][0]
	}
	if planID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing PlanID", "")
		return
	}

	instance := broker.Instances[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		WriteOSBError(w, "Can't find instance with id: "+instanceID, "")
		return
	}

	DeleteDB(instance.DB)
	delete(broker.Instances, instanceID)

	w.WriteHeader(http.StatusOK)
	Debug(2, "Instance %s: deleted\n", instanceID)
}

func BindHandler(w http.ResponseWriter, r *http.Request) {
	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	vars := mux.Vars(r)

	instanceID := vars["iID"]
	if instanceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing InstanceID", "")
		return
	}

	bindingID := vars["bID"]
	if bindingID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing BindingID", "")
		return
	}

	instance := broker.Instances[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		WriteOSBError(w, "Can't find instance with id: "+instanceID, "")
		return
	}

	binding := instance.Bindings[bindingID]
	if binding != nil {
		// Not sure if we should fail yet

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	instance.Bindings[bindingID] = struct{}{}

	creds := struct {
		User     string `json:"user,omitempty"`
		Password string `json:"password,omitempty"`
		URL      string `json:"url,omitempty"`
	}{
		User:     instance.DB.User,
		Password: instance.DB.Password,
		URL:      instance.DB.URL,
	}

	w.WriteHeader(http.StatusCreated)
	WriteJSON(w, creds)
	Debug(2, "Instance %s: Binding %q created\n", instanceID, bindingID)
}

func UnbindHandler(w http.ResponseWriter, r *http.Request) {
	if !VerifyBasicAuth(w, r, brokerUser, brokerPassword) {
		return
	}

	vars := mux.Vars(r)

	instanceID := vars["iID"]
	if instanceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing InstanceID", "")
		return
	}

	bindingID := vars["bID"]
	if bindingID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing BindingID", "")
		return
	}

	params := r.URL.Query()

	serviceID := ""
	if params["service_id"] != nil {
		serviceID = params["service_id"][0]
	}
	if serviceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing ServiceID", "")
		return
	}

	planID := ""
	if params["plan_id"] != nil {
		planID = params["plan_id"][0]
	}
	if planID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing PlanID", "")
		return
	}

	instance := broker.Instances[instanceID]
	if instance == nil {
		w.WriteHeader(http.StatusGone)
		WriteOSBError(w, "Can't find instance with id: "+instanceID, "")
		return
	}

	binding := instance.Bindings[bindingID]
	if binding == nil {
		w.WriteHeader(http.StatusGone)
		WriteOSBError(w, "Can't find binding with id: "+bindingID, "")
		return
	}

	delete(instance.Bindings, bindingID)

	w.WriteHeader(http.StatusOK)
	Debug(2, "Instance %s: Binding %q deleted\n", instanceID, bindingID)
}

func main() {
	port := 80
	ip := "0.0.0.0"

	if v := os.Getenv("VERBOSE"); v != "" {
		if vInt, err := strconv.Atoi(v); err == nil {
			verbose = vInt
		}
	}

	flag.IntVar(&verbose, "v", verbose, "Verbosity level")
	flag.IntVar(&port, "p", port, "Listen port")
	flag.StringVar(&ip, "i", ip, "IP/interface to listen on")
	flag.StringVar(&hostString, "h", "", "Host/port string to use for DBs ")
	flag.StringVar(&brokerUser, "u", brokerUser, "Username for broker/DB admin")
	flag.StringVar(&brokerPassword, "w", brokerPassword, "Password for broker/DB admin")
	flag.BoolVar(&disableAuth, "a", false, "Turn off all auth checking")

	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/info", InfoHandler)
	r.HandleFunc("/", InfoHandler)

	r.HandleFunc("/v2/catalog", CatalogHandler).Methods("GET")
	r.HandleFunc("/v2/service_instances/{iID}", ProvisionHandler).
		Methods("PUT")
	r.HandleFunc("/v2/service_instances/{iID}", UpdateHandler).
		Methods("PATCH")
	r.HandleFunc("/v2/service_instances/{iID}/service_bindings/{bID}",
		BindHandler).Methods("PUT")
	r.HandleFunc("/v2/service_instances/{iID}/service_bindings/{bID}",
		UnbindHandler).Methods("DELETE")
	r.HandleFunc("/v2/service_instances/{iID}", DeprovisionHandler).
		Methods("DELETE")

	r.HandleFunc("/db", DBAllHandler).Methods("GET")
	r.HandleFunc("/db/", DBAllHandler).Methods("GET")
	r.HandleFunc("/db", DBCreateHandler).Methods("POST")
	r.HandleFunc("/db/", DBCreateHandler).Methods("POST")

	r.HandleFunc("/db/{dbID}", DBCreateHandler).Methods("PUT")
	r.HandleFunc("/db/{dbID}/", DBCreateHandler).Methods("PUT")
	r.HandleFunc("/db/{dbID}", DBHandler).Methods("GET")
	r.HandleFunc("/db/{dbID}/", DBHandler).Methods("GET")
	r.HandleFunc("/db/{dbID}", DBDeleteHandler).Methods("DELETE")
	r.HandleFunc("/db/{dbID}/", DBDeleteHandler).Methods("DELETE")

	r.HandleFunc("/db/{dbID}/{key:.*}", DBGetHandler).Methods("GET")
	r.HandleFunc("/db/{dbID}/{key:.*}", DBSetHandler).Methods("PUT")
	r.HandleFunc("/db/{dbID}/{key:.*}", DBRemoveHandler).Methods("DELETE")

	Debug(1, "Server listening on %s:%d\n", ip, port)
	server := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%d", ip, port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
