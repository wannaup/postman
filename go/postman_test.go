
package main

import (
	"bytes"
	"testing"
	"github.com/stretchr/testify/require"
	"encoding/json"
	"net/http"
    "net/http/httptest"
    "reflect"
    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

const TestAuthHeader string = "Basic NTE4Y2JiMTM4OWRhNzlkM2EyNTQ1M2Y5Om5vcGFzc3c="
var TestOwnerId string = "518cbb1389da79d3a25453f9"
var TestMsg = []byte(`{"from": "pinco@random.com","to": "pinco@random.com","msg": "hello!"}`)
var TestThread = []byte(`{
	"id":"",
    "owner": {
        "id": "518cbb1389da79d3a25453f9"
    },
    "messages": [{
        "from": "pinco@random.com",
        "to": "pinco@random.com",
        "msg": "hello!"
    }]
}`)
var TestBadJSON = []byte(`{"from": "pinco@random.com","to": `)

var negro = negroni.Classic()
var createdThreadId bson.ObjectId

func init() {
	router := mux.NewRouter().StrictSlash(true)
    //setup routing and middleware     
    PrepareRouting(router, negro)
    //clean db
    ResetDB()
}

func ResetDB() {
	database := "wure" 
    session, err := mgo.Dial("")
    if err != nil {
        panic(err)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)
    err = session.DB(database).C("message_threads").DropCollection()
    if err != nil {
        panic(err)
    }
}

//test not auth header and invalid JSON
func TestBadRequest(t *testing.T) {
	//no auth
	request := BuildJSONReq("POST", "/threads", TestMsg)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require := require.New(t)
    require.Equal(response.Code, http.StatusBadRequest)
    //bad json
    request = BuildJSONReq("POST", "/threads", TestBadJSON)
    response = httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require.Equal(response.Code, http.StatusBadRequest)
}

func TestCreateThread(t *testing.T) {
	request := BuildJSONReq("POST", "/threads", TestMsg)
	AuthRequest(request, TestAuthHeader)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require := require.New(t)
    require.Equal(response.Code, http.StatusOK)
    //check thread is created correctly
    var nt Thread
    err := UnmarshalObject(response.Body, &nt)
    require.Nil(err)
    //check thread is ok
    var tt Thread
    err = json.NewDecoder(bytes.NewBuffer(TestThread)).Decode(&tt)
    //save created thread so we can test later
    createdThreadId = nt.Id
    nt.Id = ""
    require.Equal(reflect.DeepEqual(tt, nt), true)
}

func TestGetAllThreads(t *testing.T) {
	request, _ := http.NewRequest("GET", "/threads", nil)
	AuthRequest(request, TestAuthHeader)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require := require.New(t)
    require.Equal(response.Code, http.StatusOK)
    var threadList []Thread
    UnmarshalObject(response.Body, &threadList)
    require.Equal(len(threadList), 1)
	//build the truth struct
    var tt Thread
    err := json.NewDecoder(bytes.NewBuffer(TestThread)).Decode(&tt)
    require.Nil(err)
    //set the correct Id
    tt.Id = createdThreadId
    require.Equal(reflect.DeepEqual(threadList[0], tt), true)
}

//test we can correctly get a specific thread
func TestGetOneThread(t *testing.T) {
	request, _ := http.NewRequest("GET", "/threads/" + createdThreadId.Hex(), nil)
    AuthRequest(request, TestAuthHeader)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require := require.New(t)
    require.Equal(response.Code, http.StatusOK)
    var nt Thread
    UnmarshalObject(response.Body, &nt)
    //build the truth struct
    var tt Thread
    err := json.NewDecoder(bytes.NewBuffer(TestThread)).Decode(&tt)
    require.Nil(err)
    //set the correct Id
    tt.Id = createdThreadId
    require.Equal(reflect.DeepEqual(tt, nt), true)

}

func BuildJSONReq(method string, url string, mJson []byte) *http.Request{
	contentReader := bytes.NewBuffer(mJson)
	req, _ := http.NewRequest(method, url, contentReader)
	req.Header.Add("Content-Type", "application/json")
	return req
}

func AuthRequest(req *http.Request, authHeader string) {
		req.Header.Set("Authorization", authHeader)
}


