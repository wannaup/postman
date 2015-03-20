
package main

import (
"fmt"
	"bytes"
	"testing"
	"github.com/stretchr/testify/require"
	"encoding/json"
	"net/http"
    "net/http/httptest"
    "net/url"
    "reflect"
    "io/ioutil"
    "strings"
    "github.com/codegangsta/negroni"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)

const TestAuthHeader string = "Basic NTE4Y2JiMTM4OWRhNzlkM2EyNTQ1M2Y5Om5vcGFzc3c="
var TestOwnerId string = "518cbb1389da79d3a25453f9"
var TestMsg = []byte(`{"from": "pinco@random.com","to": "pinco@modnar.com","msg": "hello!"}`)
var TestThread = []byte(`{
	"id":"",
    "owner": {
        "id": "518cbb1389da79d3a25453f9"
    },
    "messages": [{
        "from": "pinco@random.com",
        "to": "pinco@modnar.com",
        "msg": "hello!"
    }]
}`)
var TestReplyMsg = []byte(`{"from": "pinco@modnar.com","msg": "hello too!"}`)
var TestInsertedReplyMsg = []byte(`{"from": "pinco@modnar.com","to": "pinco@random.com","msg": "hello too!"}`)
var TestInsertedInboundMsg = []byte(`{"from": "pinco@test.com","to": "pinco@modnar.com","msg": "A test message"}`)
var TestBadJSON = []byte(`{"from": "pinco@random.com","to": `)
var negro *negroni.Negroni
var createdThreadId bson.ObjectId

func init() {
    PreFlight("conf_debug.json")
    //setup routing and middleware     
    negro = StirNegroni()
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
    if err != nil && err.Error() != "ns not found"{
        panic(err)
    }
}

func TestCreateThread(t *testing.T) {
    proofInvalidJSON("/threads",t)
    proofNoAuth("/threads", "POST", TestMsg, t)
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
    proofNoAuth("/threads", "GET", nil, t)
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
    proofNoAuth("/threads/" + createdThreadId.Hex(), "GET", nil, t)
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

//test we can correctly reply to a thread
func TestReplyThread(t *testing.T) {
    proofInvalidJSON("/threads/" + createdThreadId.Hex() + "/reply", t)
    proofNoAuth("/threads/" + createdThreadId.Hex() + "/reply", "POST", TestReplyMsg, t)
	request := BuildJSONReq("POST", "/threads/" + createdThreadId.Hex() + "/reply", TestReplyMsg)
    AuthRequest(request, TestAuthHeader)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require := require.New(t)
    require.Equal(response.Code, http.StatusOK)
    var nmt Thread
    UnmarshalObject(response.Body, &nmt)
    //build the truth struct
    var tt Thread
    err := json.NewDecoder(bytes.NewBuffer(TestThread)).Decode(&tt)
    require.Nil(err)
    //set the correct Id
    tt.Id = createdThreadId
    //add the addedd msg
    var nm Message
    err = json.NewDecoder(bytes.NewBuffer(TestInsertedReplyMsg)).Decode(&nm)
    require.Nil(err)
    tt.Messages = append(tt.Messages, nm)
    require.Equal(reflect.DeepEqual(tt, nmt), true)
}

//test we can correctly receive inbound requests from mail provider
func TestInbound(t *testing.T) {
    proofInvalidJSON("/inbound", t)
    //replace the inbound email for the existing thread in the mandrill request
    mReq := strings.Replace(string(TestMandrillInbound), "$INBOUNDMAIL$", createdThreadId.Hex() + "@" + config["INBOUND_EMAIL_DOMAIN"], 1)
    data := url.Values{}
    data.Set("mandrill_events", mReq)
    request, _ := http.NewRequest("POST", "/inbound", bytes.NewBufferString(data.Encode()))
    request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    body, _ := ioutil.ReadAll(response.Body)
    fmt.Println(string(body))
    require := require.New(t)
    require.Equal(response.Code, http.StatusOK)
    //now check it was correctly added
    request, _ = http.NewRequest("GET", "/threads/" + createdThreadId.Hex(), nil)
    AuthRequest(request, TestAuthHeader)
    response = httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require.Equal(response.Code, http.StatusOK)
    var nt Thread
    UnmarshalObject(response.Body, &nt)
    //build the truth msg
    var nm Message
    err := json.NewDecoder(bytes.NewBuffer(TestInsertedInboundMsg)).Decode(&nm)
    require.Nil(err)
    require.Equal(reflect.DeepEqual(nm, nt.Messages[len(nt.Messages) -1]), true)

}

func proofInvalidJSON(url string, t *testing.T){
    request := BuildJSONReq("POST", url, TestBadJSON)
    AuthRequest(request, TestAuthHeader)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require := require.New(t) 
    require.Equal(response.Code, http.StatusBadRequest)
}

func proofNoAuth(url string, method string, body []byte, t *testing.T){
    request := BuildJSONReq(method, url, body)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require := require.New(t)
    require.Equal(response.Code, http.StatusBadRequest)
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


