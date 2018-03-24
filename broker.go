package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

/* Service Stuff */
/*****************/

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

	db := &DB{
		ID:       strID,
		User:     "user",
		Password: "passw0rd",
		Data:     map[string][]byte{},
		URL:      fmt.Sprintf("http://%s/db/"+strID, r.Host),
		mutex:    sync.Mutex{},
	}

	DBMapmutex.Lock()
	DBs[db.ID] = db
	DBMapmutex.Unlock()
	return db
}

func DeleteDB(db *DB) {
	DBMapmutex.Lock()
	delete(DBs, db.ID)
	DBMapmutex.Unlock()
}

func DBAllHandler(w http.ResponseWriter, r *http.Request) {
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
			WriteJSON(w, db)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func DBCreateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbID := vars["dbID"]

	if dbID == "" {
		db := NewDB(r)
		DBs[db.ID] = db
		w.Header().Add("Location", "/db/"+db.ID)
		w.WriteHeader(http.StatusCreated)
		return
	}

	if DBs[dbID] != nil {
		w.WriteHeader(http.StatusConflict)
		return
	}

	DBs[dbID] = NewDB(r)

	// Add host:port
	w.Header().Add("Location", "/db/"+dbID)

	w.WriteHeader(http.StatusCreated)
}

func DBDeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dbID := vars["dbID"]
	if dbID != "" {
		delete(DBs, dbID)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func DBGetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if dbID := vars["dbID"]; dbID != "" {
		if db := DBs[dbID]; db != nil {
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
		if key := vars["key"]; key != "" {
			var value []byte

			if valueStr, ok := vars["value"]; ok {
				value = []byte(valueStr)
			} else {
				value, _ = ioutil.ReadAll(r.Body)
			}

			fmt.Printf("value: %v\n", string(value))
			db.Data[key] = value
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

/* Generic Stuff */
/*****************/
var lastID int = 0
var DBs map[string]*DB = map[string]*DB{} // DBid -> DB

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	str := fmt.Sprintf(
		`OSB API Sample DB Broker
------------------------
DBs: %d
`, len(DBs))
	w.Write([]byte(str))
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

func WriteJSON(w http.ResponseWriter, obj interface{}) {
	// Check for err
	b, _ := json.MarshalIndent(obj, "", "  ")
	w.Write(b)
	w.Write([]byte("\n"))
}

func CatalogHandler(w http.ResponseWriter, r *http.Request) {
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
	vars := mux.Vars(r)

	instanceID := vars["instanceID"]
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

	broker.Instances[instanceID] = &Instance{
		DB:       NewDB(r),
		Request:  pReq,
		Bindings: map[string]interface{}{},
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{}"))
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello\n"))
}

func DeprovisionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	instanceID := vars["instanceID"]
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
}

func BindHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	instanceID := vars["instanceID"]
	if instanceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing InstanceID", "")
		return
	}

	bindingID := vars["bindingID"]
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
}

func UnbindHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	instanceID := vars["instanceID"]
	if instanceID == "" {
		w.WriteHeader(http.StatusBadRequest)
		WriteOSBError(w, "Missing InstanceID", "")
		return
	}

	bindingID := vars["bindingID"]
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
}

func main() {
	port := 80
	ip := "0.0.0.0"

	r := mux.NewRouter()
	r.HandleFunc("/info", InfoHandler)
	r.HandleFunc("/", InfoHandler)

	r.HandleFunc("/v2/catalog", CatalogHandler).Methods("GET")
	r.HandleFunc("/v2/service_instances/{instanceID}", ProvisionHandler).
		Methods("PUT")
	r.HandleFunc("/v2/service_instances/{instanceID}", UpdateHandler).
		Methods("PATCH")
	r.HandleFunc("/v2/service_instances/{instanceID}/service_bindings/{bindingID}",
		BindHandler).Methods("PUT")
	r.HandleFunc("/v2/service_instances/{instanceID}/service_bindings/{bindingID}",
		UnbindHandler).Methods("DELETE")
	r.HandleFunc("/v2/service_instances/{instanceID}", DeprovisionHandler).
		Methods("DELETE")

	r.HandleFunc("/db", DBAllHandler).Methods("GET")
	r.HandleFunc("/db", DBCreateHandler).Methods("POST")
	r.HandleFunc("/db/", DBCreateHandler).Methods("POST")
	r.HandleFunc("/db/{dbID}", DBCreateHandler).Methods("PUT")
	r.HandleFunc("/db/{dbID}", DBHandler).Methods("GET")
	r.HandleFunc("/db/{dbID}/", DBHandler).Methods("GET")
	r.HandleFunc("/db/{dbID}", DBDeleteHandler).Methods("DELETE")
	r.HandleFunc("/db/{dbID}/{key}", DBGetHandler).Methods("GET")
	r.HandleFunc("/db/{dbID}/{key}", DBSetHandler).Methods("PUT")
	r.HandleFunc("/db/{dbID}/{key}/", DBSetHandler).Methods("PUT")
	r.HandleFunc("/db/{dbID}/{key}/{value:.*}", DBSetHandler).Methods("PUT")

	fmt.Printf("Server listening on %s:%d\n", ip, port)
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
