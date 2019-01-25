package utils

import (
	"fmt"
	"net/http"
	"net/url"
)

func SendSMS(smsurl, recipient, message string) bool {
	if recipient == "" || message == "" {
		return false
	}

	sendSMSURL := fmt.Sprintf(smsurl, url.QueryEscape(recipient), url.QueryEscape(message))
	httpReq, _ := http.NewRequest("GET", sendSMSURL, nil)
	client := &http.Client{}
	client.Do(httpReq)
	return true
}
