package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sendgrid/sendgrid-go"
	"os"
)

const (
	host = "https://api.sendgrid.com"
)

// SendgridResult is the result of queries to the Sengrid API.
type SendgridResult struct {
	Created int    `json:"created"`
	Email   string `json:"email"`
	Reason  string `json:"reason"`
	Status  string `json:"status"`
}

// SpamReportResult returned info on SPAM reports for given email address.
type SpamReportResult struct {
	Created int    `json:"created"`
	Email   string `json:"email"`
	IP      string `json:"ip"`
}

// SendgridHandler handles all things sendgrid...
type SendgridHandler struct {
	ApiKey string
}

// NewSendgridHandler creates a new instance of SendgridHandler
func NewSendgridHandler() *SendgridHandler {
	sg := SendgridHandler{}
	sg.ApiKey = os.Getenv("SENDGRID_KEY")
	return &sg
}

// CheckSpam checks if address has been marked as spammy
func (sg *SendgridHandler) CheckSpam(email string) ([]*SpamReportResult, error) {
	url := fmt.Sprintf("/v3/suppression/spam_reports/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)

	request.Method = "GET"
	response, err := sendgrid.API(request)
	if err != nil {
		return nil, err
	}

	var res []*SpamReportResult
	json.Unmarshal([]byte(response.Body), &res)
	return res, nil
}

// CheckBlock checks if a block has been registered against the email address
func (sg *SendgridHandler) CheckBlock(email string) ([]*SendgridResult, error) {
	url := fmt.Sprintf("/v3/suppression/blocks/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)

	request.Method = "GET"
	response, err := sendgrid.API(request)
	if err != nil {
		return nil, err
	}

	var res []*SendgridResult
	json.Unmarshal([]byte(response.Body), &res)
	return res, nil
}

// CheckBounce checks if a bounce has been registered against the email address
func (sg *SendgridHandler) CheckBounce(email string) ([]*SendgridResult, error) {

	url := fmt.Sprintf("/v3/suppression/bounces/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)

	request.Method = "GET"
	response, err := sendgrid.API(request)
	if err != nil {
		return nil, err
	}

	var res []*SendgridResult
	json.Unmarshal([]byte(response.Body), &res)
	return res, nil
}

// CheckInvalid checks if an invalid email flag has been registered against the email address
func (sg *SendgridHandler) CheckInvalid(email string) ([]*SendgridResult, error) {

	url := fmt.Sprintf("/v3/suppression/invalid_emails/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)

	request.Method = "GET"
	response, err := sendgrid.API(request)
	if err != nil {
		return nil, err
	}

	var res []*SendgridResult
	json.Unmarshal([]byte(response.Body), &res)
	return res, nil
}

// DeleteBounce removes a bounce against an email addr.
func (sg *SendgridHandler) DeleteBounce(email string) error {
	url := fmt.Sprintf("/v3/suppression/bounces/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)
	request.Method = "DELETE"
	queryParams := make(map[string]string)
	queryParams["email_address"] = email
	request.QueryParams = queryParams
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return errors.New("Unable to debounce")
	}
	return nil
}

// DeleteBlock removes a block against an email addr.
func (sg *SendgridHandler) DeleteBlock(email string) error {
	url := fmt.Sprintf("/v3/suppression/blocks/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)
	request.Method = "DELETE"
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return errors.New("Unable to unblock")
	}

	return nil
}

// DeleteSpam removes a spam against an email addr.
func (sg *SendgridHandler) DeleteSpam(email string) error {
	url := fmt.Sprintf("/v3/suppression/spam_reports/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)
	request.Method = "DELETE"
	queryParams := make(map[string]string)
	queryParams["email_address"] = email
	request.QueryParams = queryParams
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return errors.New("Unable to despam-spam-spam-spam-spam.......... wonderful spam...")
	}
	return nil
}

// DeleteInvalid removes an invalid mark against an email addr.
func (sg *SendgridHandler) DeleteInvalid(email string) error {
	url := fmt.Sprintf("/v3/suppression/invalid_emails/%s", email)
	request := sendgrid.GetRequest(sg.ApiKey, url, host)
	request.Method = "DELETE"
	queryParams := make(map[string]string)
	queryParams["email_address"] = email
	request.QueryParams = queryParams
	response, err := sendgrid.API(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return errors.New("Unable to remove invalidation")
	}
	return nil
}
