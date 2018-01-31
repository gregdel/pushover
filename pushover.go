// Package pushover provides a wrapper around the Pushover API
package pushover

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
)

// Regexp validation
var tokenRegexp *regexp.Regexp

func init() {
	tokenRegexp = regexp.MustCompile(`^[A-Za-z0-9]{30}$`)
}

// APIEndpoint is the API base URL for any request
var APIEndpoint = "https://api.pushover.net/1"

// Pushover custom errors
var (
	ErrHTTPPushover               = errors.New("pushover: http error")
	ErrEmptyToken                 = errors.New("pushover: empty API token")
	ErrEmptyURL                   = errors.New("pushover: empty URL, URLTitle needs an URL")
	ErrEmptyRecipientToken        = errors.New("pushover: empty recipient token")
	ErrInvalidRecipientToken      = errors.New("pushover: invalid recipient token")
	ErrInvalidRecipient           = errors.New("pushover: invalid recipient")
	ErrInvalidHeaders             = errors.New("pushover: invalid headers in server response")
	ErrInvalidPriority            = errors.New("pushover: invalid priority")
	ErrInvalidToken               = errors.New("pushover: invalid API token")
	ErrMessageEmpty               = errors.New("pushover: message empty")
	ErrMessageTitleTooLong        = errors.New("pushover: message title too long")
	ErrMessageTooLong             = errors.New("pushover: message too long")
	ErrMessageAttachementTooLarge = errors.New("pushover: message attachement is too large")
	ErrMessageURLTitleTooLong     = errors.New("pushover: message URL title too long")
	ErrMessageURLTooLong          = errors.New("pushover: message URL too long")
	ErrMissingEmergencyParameter  = errors.New("pushover: missing emergency parameter")
	ErrInvalidDeviceName          = errors.New("pushover: invalid device name")
	ErrEmptyReceipt               = errors.New("pushover: empty receipt")
	ErrInvalidAttachementPath     = errors.New("pushover: invalid attachement path")
)

// API limitations
const (
	// MessageMaxLength is the max message number of characters
	MessageMaxLength = 1024
	// MessageTitleMaxLength is the max title number of characters
	MessageTitleMaxLength = 250
	// MessageURLMaxLength is the max URL number of characters
	MessageURLMaxLength = 512
	// MessageURLTitleMaxLength is the max URL title number of characters
	MessageURLTitleMaxLength = 100
	// MessageMaxAttachementByte is the max attachement size in byte
	MessageMaxAttachementByte = 2621440
)

// Message priorities
const (
	PriorityLowest    = -2
	PriorityLow       = -1
	PriorityNormal    = 0
	PriorityHigh      = 1
	PriorityEmergency = 2
)

// Sounds
const (
	SoundPushover     = "pushover"
	SoundBike         = "bike"
	SoundBugle        = "bugle"
	SoundCashRegister = "cashregister"
	SoundClassical    = "classical"
	SoundCosmic       = "cosmic"
	SoundFalling      = "falling"
	SoundGamelan      = "gamelan"
	SoundIncoming     = "incoming"
	SoundIntermission = "intermission"
	SoundMagic        = "magic"
	SoundMechanical   = "mechanical"
	SoundPianobar     = "pianobar"
	SoundSiren        = "siren"
	SoundSpaceAlarm   = "spacealarm"
	SoundTugBoat      = "tugboat"
	SoundAlien        = "alien"
	SoundClimb        = "climb"
	SoundPersistent   = "persistent"
	SoundEcho         = "echo"
	SoundUpDown       = "updown"
	SoundNone         = "none"
)

// Pushover is the representation of an app using the pushover API
type Pushover struct {
	token string
}

// New returns a new app to talk to the pushover API
func New(token string) *Pushover {
	return &Pushover{token}
}

// Request to the API
type Request struct {
	Message   *Message
	Recipient *Recipient
}

// SendMessage is used to send message to a recipient
func (p *Pushover) SendMessage(message *Message, recipient *Recipient) (*Response, error) {
	url := fmt.Sprintf("%s/messages.json", APIEndpoint)

	// Encode params and perform data validation
	params, err := p.encodeRequest(message, recipient)
	if err != nil {
		return nil, err
	}

	// Post the from and check the headers of the response
	return p.postForm(url, params, true)
}

// Validate Pushover token
func (p *Pushover) validate() error {
	// Check empty token
	if p.token == "" {
		return ErrEmptyToken
	}

	// Check invalid token
	if tokenRegexp.MatchString(p.token) == false {
		return ErrInvalidToken
	}
	return nil
}

// Encode pushover request and validate each data before sending
func (p *Pushover) encodeRequest(message *Message, recipient *Recipient) (map[string]string, error) {
	// Validate pushover
	if err := p.validate(); err != nil {
		return nil, err
	}

	// Validate recipient
	if err := recipient.validate(); err != nil {
		return nil, err
	}

	// Validate message
	if err := message.validate(); err != nil {
		return nil, err
	}

	// Create the url values
	params := map[string]string{
		"token":    p.token,
		"user":     recipient.token,
		"message":  message.Message,
		"priority": fmt.Sprintf("%d", message.Priority),
	}

	if message.Title != "" {
		params["title"] = message.Title
	}

	if message.URL != "" {
		params["url"] = message.URL
	}

	if message.URLTitle != "" {
		params["url_title"] = message.URLTitle
	}

	if message.Sound != "" {
		params["sound"] = message.Sound
	}

	if message.DeviceName != "" {
		params["device"] = message.DeviceName
	}

	if message.Timestamp != 0 {
		params["timestamp"] = strconv.FormatInt(message.Timestamp, 10)
	}

	if message.HTML {
		params["html"] = "1"
	}

	if message.AttachmentPath != "" {
		params["attachment_path"] = message.AttachmentPath
	}

	if message.Priority == PriorityEmergency {
		params["retry"] = strconv.FormatFloat(message.Retry.Seconds(), 'f', -1, 64)
		params["expire"] = strconv.FormatFloat(message.Expire.Seconds(), 'f', -1, 64)
		if message.CallbackURL != "" {
			params["callback"] = message.CallbackURL
		}
	}

	return params, nil
}

// GetReceiptDetails return detailed informations about a receipt. This is used
// used to check the acknowledged status of an Emergency notification.
func (p *Pushover) GetReceiptDetails(receipt string) (*ReceiptDetails, error) {
	url := fmt.Sprintf("%s/receipts/%s.json?token=%s", APIEndpoint, receipt, p.token)

	if receipt == "" {
		return nil, ErrEmptyReceipt
	}

	// Send request
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Decode the JSON response
	var details *ReceiptDetails
	if err = json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return nil, err
	}

	return details, nil
}

// GetRecipientDetails allows to check if a recipient exists, if it's a group
// and the devices associated to this recipient. It returns an
// ErrInvalidRecipient if the recipient is not valid in the Pushover API.
func (p *Pushover) GetRecipientDetails(recipient *Recipient) (*RecipientDetails, error) {
	endpoint := fmt.Sprintf("%s/users/validate.json", APIEndpoint)

	// Validate pushover
	if err := p.validate(); err != nil {
		return nil, err
	}

	// Validate recipient
	if err := recipient.validate(); err != nil {
		return nil, err
	}

	// Send request
	urlValues := url.Values{}
	urlValues.Add("token", p.token)
	urlValues.Add("user", recipient.token)
	resp, err := http.PostForm(endpoint, urlValues)
	if err != nil {
		return nil, err
	}

	// Only 500 errors will not respond a readable result
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, ErrHTTPPushover
	}

	// Decode the JSON response
	var response *RecipientDetails
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Check response status
	if response.Status != 1 {
		return nil, ErrInvalidRecipient
	}

	return response, nil
}

// postForm is a generic post function. It checks the response from pushover
// and retrieve headers if the returnHeaders argument is set to "true"
func (p *Pushover) postForm(url string, params map[string]string, returnHeaders bool) (*Response, error) {
	body := &bytes.Buffer{}

	// Write the body as multipart form data
	w := multipart.NewWriter(body)

	// Handle file upload
	filePath, ok := params["attachment_path"]
	if ok {
		// Remove the path from the params
		delete(params, "attachment_path")

		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Write the file in the body
		fw, err := w.CreateFormFile("attachment", "poster")
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fw, file)
		if err != nil {
			return nil, err
		}
	}

	// Handle params
	for k, v := range params {
		if err := w.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	// Send request
	resp, err := http.Post(url, w.FormDataContentType(), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Only 500 errors will not respond a readable result
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, ErrHTTPPushover
	}

	// Decode the JSON response
	var response *Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Check response status
	if response.Status != 1 {
		return nil, response.Errors
	}

	// The headers are only returned when posting a new notification
	if returnHeaders {
		// Get app limits from headers
		appLimits, err := newLimit(resp.Header)
		if err != nil {
			return nil, err
		}
		response.Limit = appLimits
	}

	return response, nil
}

// CancelEmergencyNotification helps stop a notification retry in case of a
// notification with an Emergency priority before reaching the expiration time.
// It requires the response receipt in order to stop the right notification.
func (p *Pushover) CancelEmergencyNotification(receipt string) (*Response, error) {
	endpoint := fmt.Sprintf("%s/receipts/%s/cancel.json", APIEndpoint, receipt)

	// Post and do not check headers
	return p.postForm(endpoint, map[string]string{"token": p.token}, false)
}
