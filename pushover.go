// Package pushover provides a wrapper around the Pushover API
package pushover

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

// APIEndpoint is the API base URL for any request
const APIEndpoint = "https://api.pushover.net/1"

// Pushover custom errors
var (
	ErrHTTPPushover              = errors.New("pushover http error")
	ErrEmptyToken                = errors.New("empty API token")
	ErrEmptyURL                  = errors.New("empty URL, URLTitle needs an URL")
	ErrEmptyRecipientToken       = errors.New("empty recipient token")
	ErrInvalidRecipientToken     = errors.New("invalid recipient token")
	ErrInvalidRecipient          = errors.New("invalid recipient")
	ErrInvalidHeaders            = errors.New("invalid headers in server response")
	ErrInvalidPriority           = errors.New("invalid priority")
	ErrInvalidToken              = errors.New("invalid API token")
	ErrMessageEmpty              = errors.New("message empty")
	ErrMessageTitleTooLong       = errors.New("message title too long")
	ErrMessageTooLong            = errors.New("message too long")
	ErrMessageURLTitleTooLong    = errors.New("message URL title too long")
	ErrMessageURLTooLong         = errors.New("message URL too long")
	ErrMissingEmergencyParameter = errors.New("missing emergency parameter")
	ErrInvalidDeviceName         = errors.New("invalid device name")
	ErrEmptyReceipt              = errors.New("empty receipt")
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

// Regexp validation
var tokenRegexp, recipientRegexp, deviceNameRegexp *regexp.Regexp

func init() {
	tokenRegexp = regexp.MustCompile(`^[A-Za-z0-9]{30}$`)
	recipientRegexp = regexp.MustCompile(`^[A-Za-z0-9]{30}$`)
	deviceNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]{1,25}$`)
}

// Pushover is the representation of an app using the pushover API
type Pushover struct {
	token string
}

// New returns a new app to talk to the pushover API
func New(token string) *Pushover {
	return &Pushover{token}
}

// Errors represents the errors returned by pushover
type Errors []string

// Request to the API
type Request struct {
	Message   *Message
	Recipient *Recipient
}

// Response represents a response from the API
type Response struct {
	Status         int    `json:"status"`
	ID             string `json:"request"`
	Errors         Errors `json:"errors,omitempty"`
	Receipt        string `json:"receipt,omitempty"`
	ReceiptDetails *ReceiptDetails
	Limit          *Limit
}

// Limit represents the limitation of the application. This information is
// fetched when posting a new message.
//	Headers exemple:
//		X-Limit-App-Limit: 7500
// 		X-Limit-App-Remaining: 7496
// 		X-Limit-App-Reset: 1393653600
type Limit struct {
	Total     int
	Remaining int
	NextReset time.Time
}

func newLimit(headers http.Header) (*Limit, error) {
	headersStrings := []string{
		"X-Limit-App-Limit",
		"X-Limit-App-Remaining",
		"X-Limit-App-Reset",
	}
	headersValues := map[string]int{}

	for _, header := range headersStrings {
		// Check if the header is present
		h, ok := headers[header]
		if !ok {
			return nil, ErrInvalidHeaders
		}

		// The header must have only one element
		if len(h) != 1 {
			return nil, ErrInvalidHeaders
		}

		i, err := strconv.Atoi(h[0])
		if err != nil {
			return nil, err
		}

		headersValues[header] = i
	}

	return &Limit{
		Total:     headersValues["X-Limit-App-Limit"],
		Remaining: headersValues["X-Limit-App-Remaining"],
		NextReset: time.Unix(int64(headersValues["X-Limit-App-Reset"]), 0),
	}, nil
}

// Error represents the error as a string
func (e Errors) Error() string {
	ret := ""
	for _, err := range e {
		ret = fmt.Sprintf("%s%s\n", ret, err)
	}
	return ret
}

// String represents a printable form of the response
func (r Response) String() string {
	ret := fmt.Sprintf("Request id: %s", r.ID)
	if r.Receipt != "" {
		ret += fmt.Sprintf("\nReceipt: %s", r.Receipt)
	}
	if r.Limit != nil {
		ret += fmt.Sprintf("\nUsage %d/%d messages - Next reset : %s", r.Limit.Remaining, r.Limit.Total, r.Limit.NextReset)
	}
	return ret
}

// Recipient represents the a recipient to notify
type Recipient struct {
	token string
}

// Validates recipient token
func (u *Recipient) validate() error {
	// Check empty token
	if u.token == "" {
		return ErrEmptyRecipientToken
	}

	// Check invalid token
	if recipientRegexp.MatchString(u.token) == false {
		return ErrInvalidRecipientToken
	}
	return nil
}

// NewRecipient is the representation of the recipient to notify
func NewRecipient(token string) *Recipient {
	return &Recipient{token}
}

// SendMessage is used to send message to a recipient
func (p *Pushover) SendMessage(message *Message, recipient *Recipient) (*Response, error) {
	url := fmt.Sprintf("%s/messages.json", APIEndpoint)

	// Encode params and perform data validation
	urlValues, err := p.encodeRequest(message, recipient)
	if err != nil {
		return nil, err
	}

	// Send request
	resp, err := http.PostForm(url, *urlValues)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Only 500 errors will not repsond a readable result
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, ErrHTTPPushover
	}

	// Get response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var response *Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Check response status
	if response.Status != 1 {
		return nil, response.Errors
	}

	// Get app limits from headers
	appLimits, err := newLimit(resp.Header)
	if err != nil {
		return nil, err
	}
	response.Limit = appLimits

	return response, nil
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

// Message represents the infos a
type Message struct {
	// Required
	Message string

	// Optional
	Title       string
	Priority    int
	URL         string
	URLTitle    string
	Timestamp   int64
	Retry       time.Duration
	Expire      time.Duration
	CallbackURL string
	DeviceName  string
	Sound       string
}

// NewMessage returns a simple new message
func NewMessage(message string) *Message {
	return &Message{Message: message}
}

// NewMessageWithTitle returns a simple new message with a title
func NewMessageWithTitle(message, title string) *Message {
	return &Message{Message: message, Title: title}
}

// Validate the message values
func (m *Message) validate() error {
	// Message should no be empty
	if m.Message == "" {
		return ErrMessageEmpty
	}

	// Validate message length
	if len(m.Message) > MessageMaxLength {
		return ErrMessageTooLong
	}

	// Validate Title field length
	if len(m.Title) > MessageTitleMaxLength {
		return ErrMessageTitleTooLong
	}

	// Validate URL field
	if len(m.URL) > MessageURLMaxLength {
		return ErrMessageURLTooLong
	}

	// Validate URL title field
	if len(m.URLTitle) > MessageURLTitleMaxLength {
		return ErrMessageURLTitleTooLong
	}

	// URLTitle should not be set with an empty URL
	if m.URL == "" && m.URLTitle != "" {
		return ErrEmptyURL
	}

	// Validate priorities
	if m.Priority > PriorityEmergency || m.Priority < PriorityLowest {
		return ErrInvalidPriority
	}

	// Validate emergency priority
	if m.Priority == PriorityEmergency {
		if m.Retry == 0 || m.Expire == 0 {
			return ErrMissingEmergencyParameter
		}
	}

	// Test device name
	if m.DeviceName != "" {
		if deviceNameRegexp.MatchString(m.DeviceName) == false {
			return ErrInvalidDeviceName
		}
	}

	return nil
}

// Encode pushover request and validate each data before sending
func (p *Pushover) encodeRequest(message *Message, recipient *Recipient) (*url.Values, error) {
	// Validate pushover
	err := p.validate()
	if err != nil {
		return nil, err
	}

	// Validate recipient
	err = recipient.validate()
	if err != nil {
		return nil, err
	}

	// Validate message
	err = message.validate()
	if err != nil {
		return nil, err
	}

	// Create the url values
	urlValues := &url.Values{}
	urlValues.Set("token", p.token)
	urlValues.Add("user", recipient.token)
	urlValues.Add("message", message.Message)
	urlValues.Add("priority", fmt.Sprintf("%d", message.Priority))

	if message.Message != "" {
		urlValues.Add("title", message.Title)
	}

	if message.URL != "" {
		urlValues.Add("url", message.URL)
	}

	if message.URLTitle != "" {
		urlValues.Add("url_title", message.URLTitle)
	}

	if message.Sound != "" {
		urlValues.Add("sound", message.Sound)
	}

	if message.Timestamp != 0 {
		urlValues.Add("timestamp", strconv.FormatInt(message.Timestamp, 10))
	}

	if message.Priority == PriorityEmergency {
		urlValues.Add("retry", strconv.FormatFloat(message.Retry.Seconds(), 'f', -1, 64))
		urlValues.Add("expire", strconv.FormatFloat(message.Expire.Seconds(), 'f', -1, 64))
		if message.CallbackURL != "" {
			urlValues.Add("callback", message.CallbackURL)
		}
	}

	return urlValues, nil
}

// ReceiptDetails represents the receipt informations in case of emergency priority
type ReceiptDetails struct {
	Status          int
	Acknowledged    bool
	AcknowledgedBy  string
	Expired         bool
	CalledBack      bool
	ID              string
	AcknowledgedAt  *time.Time
	LastDeliveredAt *time.Time
	ExpiresAt       *time.Time
	CalledBackAt    *time.Time
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
	defer resp.Body.Close()

	// Get response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var details *ReceiptDetails
	err = json.Unmarshal(body, &details)
	if err != nil {
		return nil, err
	}

	return details, nil
}

// UnmarshalJSON is a custom unmarshal function to handle timestamps and
// boolean as int and convert them to the right type
func (r *ReceiptDetails) UnmarshalJSON(data []byte) error {
	dataBytes := bytes.NewReader(data)
	var aux struct {
		ID              string     `json:"request"`
		Status          int        `json:"status"`
		Acknowledged    intBool    `json:"acknowledged"`
		AcknowledgedBy  string     `json:"acknowledged_by"`
		Expired         intBool    `json:"expired"`
		CalledBack      intBool    `json:"called_back"`
		AcknowledgedAt  *timestamp `json:"acknowledged_at"`
		LastDeliveredAt *timestamp `json:"last_delivered_at"`
		ExpiresAt       *timestamp `json:"expires_at"`
		CalledBackAt    *timestamp `json:"called_back_at"`
	}

	// Decode json into the aux struct
	if err := json.NewDecoder(dataBytes).Decode(&aux); err != nil {
		return err
	}

	// Set the RecipientDetails with the right types
	r.Status = aux.Status
	r.Acknowledged = bool(aux.Acknowledged)
	r.AcknowledgedBy = aux.AcknowledgedBy
	r.Expired = bool(aux.Expired)
	r.CalledBack = bool(aux.CalledBack)
	r.ID = aux.ID
	r.AcknowledgedAt = aux.AcknowledgedAt.Time
	r.LastDeliveredAt = aux.LastDeliveredAt.Time
	r.ExpiresAt = aux.ExpiresAt.Time
	r.CalledBackAt = aux.CalledBackAt.Time

	return nil
}

// Helper to unmarshal a timestamp as string to a time.Time
type timestamp struct{ *time.Time }

func (t *timestamp) UnmarshalJSON(data []byte) error {
	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	if i > 0 {
		unixTime := time.Unix(i, 0)
		*t = timestamp{&unixTime}
	}

	return nil
}

// Helper to unmarshal a int as a boolean
type intBool bool

func (i *intBool) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	switch v {
	case 0:
		*i = false
	case 1:
		*i = true
	default:
		return fmt.Errorf("Failed to unmarshal int to bool")
	}

	return nil
}

// RecipientDetails represents the receipt informations in case of emergency priority
type RecipientDetails struct {
	Status    int      `json:"status"`
	Group     int      `json:"group"`
	Devices   []string `json:"devices"`
	RequestID string   `json:"request"`
	Errors    Errors   `json:"errors"`
}

// GetRecipientDetails allows to check if a recipient exists, if it's a group
// and the devices associated to this recipient. It returns an
// ErrInvalidRecipient if the recipient is not valid in the Pushover API.
func (p *Pushover) GetRecipientDetails(recipient *Recipient) (*RecipientDetails, error) {
	endpoint := fmt.Sprintf("%s/users/validate.json", APIEndpoint)

	// Validate pushover
	err := p.validate()
	if err != nil {
		return nil, err
	}

	// Validate recipient
	err = recipient.validate()
	if err != nil {
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
	defer resp.Body.Close()

	// Only 500 errors will not repsond a readable result
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, ErrHTTPPushover
	}

	// Get response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal
	var response *RecipientDetails
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	// Check response status
	if response.Status != 1 {
		return nil, ErrInvalidRecipient
	}

	return response, nil
}
