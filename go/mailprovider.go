package main

import (
	"os"
    "log"
    "bytes"
	"net/http"
    "github.com/parnurzeal/gorequest"
)

type MailProvider interface {
	SendMail (threadId string, from string, to []string, msg string) bool
}

func NewMailProvider(config map[string]string) MailProvider {
	if config["MAIL_PROVIDER"] == "mandrill" {
    	// load environment vars if not provided in config
    	if config["INBOUND_EMAIL_DOMAIN"] == ""{
    		config["INBOUND_EMAIL_DOMAIN"] = os.Getenv("INBOUND_EMAIL_DOMAIN")
    	}
    	if config["MANDRILL_API_URL"] == ""{
    		config["MANDRILL_API_URL"] = os.Getenv("MANDRILL_API_URL")
    	}
    	if config["MANDRILL_API_HOST"] == ""{
    		config["MANDRILL_API_HOST"] = os.Getenv("MANDRILL_API_HOST")
    	}
    	if config["MANDRILL_API_KEY"] == ""{
    		config["MANDRILL_API_KEY"] = os.Getenv("MANDRILL_API_KEY")
    	}
		return &MandrillMailProvider{config["INBOUND_EMAIL_DOMAIN"], config["MANDRILL_API_HOST"] + config["MANDRILL_API_URL"], config["MANDRILL_API_KEY"]}
	}
	return nil
}

type MandrillReq struct {
    Key string  `json:"key"`
    Message MandrillMsg `json:"message"`
}

type MandrillRecipient struct {
	Email 	string 	`json:"email"`
	Name 	string `json:"name,omitempty"`
}

type MandrillMsg struct {
    Html	string 	`json:"html"`
    Text	string 	`json:"text"`
    Subject	string 	`json:"subject"`
    From_email	string 	`json:"from_email"`
    From_name	string 	`json:"from_name,omitempty"`
    To 		[]MandrillRecipient 	`json:"to"`
   	Headers map[string]string 	`json:"headers"`
}


type MandrillMailProvider struct {
	InboundEmailDomain 	string
	ApiUrl 			string
	ApiKey 			string
}

func (m *MandrillMailProvider) SendMail(threadId string, from string, to []string, msg string) bool{
	//add recipients
    rcpts := []MandrillRecipient{}
    for _, val := range to {
    	rcpts = append(rcpts, MandrillRecipient{Email: val, Name: val})
    }
    hdr := map[string]string{"Reply-To": threadId + "@" + m.InboundEmailDomain}
	mmsg := MandrillMsg{
		msg,
		msg,
		"Test subject",
		from,
		from,
		rcpts,
		hdr,
	}
	postData := MandrillReq{m.ApiKey,mmsg}
	//send the mail using the HTTP JSON API
	r, body, errs := gorequest.New().Post(m.ApiUrl).
		Send(postData).
  		End()
  	if r.StatusCode != http.StatusOK || errs != nil {
  		log.Println("Error sending mail")
  		return false
  	}
    //check response is "sent"
    var resp []map[string]interface{}
    UnmarshalObject(bytes.NewBuffer([]byte(body)), &resp)

    if len(resp) < 1 || resp[0]["status"] != "sent"{
    	log.Println("Error sending mail %v", resp)
  		return false
    }
    return true
}
