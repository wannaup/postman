package main

import (
    "log"
    "github.com/parnurzeal/gorequest"
)

type MandrillRecipient struct{
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


func MandrillSendMail(config map[string]string, from string, to []string, msg string) {
	//add recipients
    rcpts := []MandrillRecipient{}
    for i, val := range to {
    	rcpts[i] = MandrillRecipient{Email: val}
    }
    hdr := map[string]string{"Reply-To": config["INBOUND_EMAIL"]}
	mmsg := MandrillMsg{
		msg,
		msg,
		"Test subject",
		from,
		from,
		rcpts,
		hdr,
	}
	postData := map[string]interface{}{"key": config["MANDRILL_API_KEY"], "message": mmsg}
	//send the mail using the HTTP JSON API
	_, body, errs := gorequest.New().Post(config["MANDRILL_API_URL"]).
		Send(postData).
  		End()
  	if errs != nil {
  		log.Println("Error sending mail")
  		return
  	}
    log.Println("Sent mail:", body)
    
}