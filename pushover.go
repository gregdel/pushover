// Package pushover provides a wrapper around the Pushover API
package pushover

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// APIEndpoint is the API base URL for any request
var APIEndpoint = "https://api.pushover.net/1"

// Pushover custom errors
var (
	ErrHTTPPushover              = errors.New("pushover: http error")
	ErrEmptyToken                = errors.New("pushover: empty API token")
	ErrEmptyURL                  = errors.New("pushover: empty URL, URLTitle needs an URL")
	ErrEmptyRecipientToken       = errors.New("pushover: empty recipient token")
	ErrInvalidRecipientToken     = errors.New("pushover: invalid recipient token")
	ErrInvalidRecipient          = errors.New("pushover: invalid recipient")
	ErrInvalidHeaders            = errors.New("pushover: invalid headers in server response")
	ErrInvalidPriority           = errors.New("pushover: invalid priority")
	ErrInvalidToken              = errors.New("pushover: invalid API token")
	ErrMessageEmpty              = errors.New("pushover: message empty")
	ErrMessageTitleTooLong       = errors.New("pushover: message title too long")
	ErrMessageTooLong            = errors.New("pushover: message too long")
	ErrMessageURLTitleTooLong    = errors.New("pushover: message URL title too long")
	ErrMessageURLTooLong         = errors.New("pushover: message URL too long")
	ErrMissingEmergencyParameter = errors.New("pushover: missing emergency parameter")
	ErrInvalidDeviceName         = errors.New("pushover: invalid device name")
	ErrEmptyReceipt              = errors.New("pushover: empty receipt")
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

// Error represents the error as a string
func (e Errors) Error() string {
	ret := ""
	if len(e) > 0 {
		ret = fmt.Sprintf("Errors:\n")
		ret += strings.Join(e, "\n")
	}
	return ret
}

// Request to the API
type Request struct {
	Message   *Message
	Recipient *Recipient
}

// Response represents a response from the API
type Response struct {
	Status  int    `json:"status"`
	ID      string `json:"request"`
	Errors  Errors `json:"errors"`
	Receipt string `json:"receipt"`
	Limit   *Limit
}

// Limit represents the limitation of the application. This information is
// fetched when posting a new message.
//	Headers example:
//		X-Limit-App-Limit: 7500
// 		X-Limit-App-Remaining: 7496
// 		X-Limit-App-Reset: 1393653600
type Limit struct {
	// Total number of messages you can send during a month
	Total int
	// Remaining number of messages you can send until the next reset
	Remaining int
	// NextReset is the time when all the app counters will be reseted
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

	// Post the from and check the headers of the response
	return p.postForm(url, urlValues, true)
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
	HTML        bool
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

	if message.DeviceName != "" {
		urlValues.Add("device", message.DeviceName)
	}

	if message.Timestamp != 0 {
		urlValues.Add("timestamp", strconv.FormatInt(message.Timestamp, 10))
	}

	if message.HTML {
		urlValues.Add("html", "1")
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

	// Decode the JSON response
	var details *ReceiptDetails
	if err = json.NewDecoder(resp.Body).Decode(&details); err != nil {
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
func (p *Pushover) postForm(url string, urlValues *url.Values, returnHeaders bool) (*Response, error) {
	// Send request
	resp, err := http.PostForm(url, *urlValues)
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

	// URL values
	urlValues := &url.Values{}
	urlValues.Set("token", p.token)

	// Post and do not check headers
	return p.postForm(endpoint, urlValues, false)
}
