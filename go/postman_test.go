
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
)

const TestAuthHeader string = "Basic NTE4Y2JiMTM4OWRhNzlkM2EyNTQ1M2Y5Om5vcGFzc3c="
var TestOwnerId string = "518cbb1389da79d3a25453f9"
var TestMsg = []byte(`{"from": "pinco@random.com","to": "pinco@random.com","msg": "hello!"}`)
var TestThread = []byte(`{
    "owner": {
        "id": "518cbb1389da79d3a25453f9"
    },
    "messages": [{
        "from": "pinco@random.com",
        "to": "pinco@random.com",
        "msg": "hello!"
    }]
}`)
var negro = negroni.Classic()
var createdThreadId string

func init() {
	router := mux.NewRouter().StrictSlash(true)
    //setup routing and middleware     
    PrepareRouting(router, negro)
}

func TestCreateThread(t *testing.T) {
	request := BuildJSONReq("POST", "/threads", TestMsg)
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
    nt.Id = ""
    require.Equal(reflect.DeepEqual(tt, nt), true)
}

func TestGetAllThreads(t *testing.T) {
	request, _ := http.NewRequest("GET", "/threads", nil)
	request.Header.Set("Authorization", TestAuthHeader)
    response := httptest.NewRecorder()
    negro.ServeHTTP(response, request)
    require.Equal(response.Code, http.StatusOK)
    
}

/*func TestGetOneThread(t *testing.T) {
	request, _ := http.NewRequest("GET", "/threads", nil)
    response := httptest.NewRecorder()
    if response.Code != http.StatusOK {
        t.Fatalf("Response body did not contain expected %v:\n\tbody: %v", "200", response.Code)
    }
}*/

func BuildJSONReq(method string, url string, mJson []byte) *http.Request{
	contentReader := bytes.NewBuffer(mJson)
	req, _ := http.NewRequest(method, url, contentReader)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("Authorization", TestAuthHeader)
	return req
}



