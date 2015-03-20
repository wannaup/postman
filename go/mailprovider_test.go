package main

import (
    "fmt"
	"bytes"
	"testing"
	"github.com/stretchr/testify/require"
	"encoding/json"
	"net/http"
    "net/http/httptest"
    "reflect"
)


var TestMandrillMsg = []byte(`{
        "html": "A test message",
        "text": "A test message",
        "subject": "Test subject",
        "from_email": "pinco@test.com",
        "from_name": "pinco@test.com",
        "to": [
            {
                "email": "pallo@test.com",
                "name": "pallo@test.com"
            }
        ],
        "headers": {
            "Reply-To": ""
        }
    }`)

var TestMandrillOKResp = []byte(`[
    {
        "email": "pallo@test.com",
        "status": "sent",
        "reject_reason": "hard-bounce",
        "_id": "abc123abc123abc123abc123abc123"
    }
]`)
var TestMandrillErrorResp = []byte(`[
    {
        "status": "error",
        "code": 12,
        "name": "Unknown_Subaccount",
        "message": "No subaccount exists with the id 'customer-123'"
    }
]`)

var TestMandrillInbound = []byte(`
    [
    {
        "ts": 198743897,
        "event": "inbound",
        "msg": {
            "html": "A test message",
            "text": "A test message",
            "subject": "Test subject",
            "from_email": "pinco@test.com",
            "from_name": "pinco@test.com",
            "to": [
                {
                    "email": "$INBOUNDMAIL$",
                    "name": "pallo@test.com"
                }
            ],
            "headers": {
                "Reply-To": ""
            }
        }
    }
]`)


func TestSendMail(t *testing.T) {
    LoadConfig("conf_debug.json", &config)
    //fake server
    ts := httptest.NewServer(http.HandlerFunc(MockHandler))
    defer ts.Close()
    config["MANDRILL_API_HOST"] = ts.URL
    config["MANDRILL_API_KEY"] = "test key"
    fmt.Println(ts.URL)
    res := NewMailProvider(config).SendMail("testid", "pinco@test.com", []string{"pallo@test.com"}, "A test message")
    require := require.New(t)
    require.Equal(res, true)
}

func MockHandler(w http.ResponseWriter, r *http.Request) {
    if config["MAIL_PROVIDER"] == "mandrill" {
        var rmm MandrillReq
        UnmarshalObject(r.Body, &rmm)
        //build the truth struct
        var mm MandrillMsg
        err := json.NewDecoder(bytes.NewBuffer(TestMandrillMsg)).Decode(&mm)
        if err != nil{   
            http.Error(w, "Bad req", http.StatusBadRequest)
            return
        }
        mm.Headers["Reply-To"] = "testid" + "@" + config["INBOUND_EMAIL_DOMAIN"]
        var tMReq = MandrillReq{config["MANDRILL_API_KEY"], mm}
        if !reflect.DeepEqual(tMReq, rmm) {
           panic("ciao")
           http.Error(w, "Bad req", http.StatusBadRequest)
           return
        }
        w.Write(TestMandrillOKResp)   
    }
    
}


