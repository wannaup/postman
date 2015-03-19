package main

import (
	"os"
    "log"
    "github.com/parnurzeal/gorequest"
)

type MailProvider interface {
	SendMail (threadId string, from string, to []string, msg string)
}

func NewMailProvider(config map[string]string) MailProvider {
	if config["MAIL_PROVIDER"] == "mandrill" {
    	// load environment vars
		return &MandrillMailProvider{os.Getenv("INBOUND_EMAIL_DOMAIN"), os.Getenv("MANDRILL_API_URL"), os.Getenv("MANDRILL_API_KEY")}
	}
	return nil
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

func (m *MandrillMailProvider) SendMail(threadId string, from string, to []string, msg string) {
	//add recipients
    rcpts := []MandrillRecipient{}
    for i, val := range to {
    	rcpts[i] = MandrillRecipient{Email: val}
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
	postData := map[string]interface{}{"key": m.ApiKey, "message": mmsg}
	//send the mail using the HTTP JSON API
	_, body, errs := gorequest.New().Post(m.ApiUrl).
		Send(postData).
  		End()
  	if errs != nil {
  		log.Println("Error sending mail")
  		return
  	}
    log.Println("Sent mail:", body)
    
}

