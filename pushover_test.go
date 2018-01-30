package pushover

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// Fake values to be used in the tests
var fakePushover = New("uQiRzpo4DXghDmr9QzzfQu27cmVRsG")
var fakeRecipient = NewRecipient("gznej3rKEVAvPUxu9vvNnqpmZpokzF")

func TestMessageValidation(t *testing.T) {
	// Create random strings to be used in messages
	randomStringsWithSize := make(map[int]string, 8)
	for _, size := range []int{
		MessageMaxLength,
		MessageMaxLength + 1,
		MessageTitleMaxLength,
		MessageTitleMaxLength + 1,
		MessageURLMaxLength,
		MessageURLMaxLength + 1,
		MessageURLTitleMaxLength,
		MessageURLTitleMaxLength + 1,
	} {
		rands, err := getRandomString(size)
		if err != nil {
			log.Fatalf("failed to create a random string of size %d", size)
		}
		randomStringsWithSize[size] = rands
	}

	tt := []struct {
		name        string
		message     Message
		expectedErr error
	}{
		{
			name: "valid message",
			message: Message{
				Message:    "Hello world !",
				Title:      "Example",
				DeviceName: "My_Device",
				URL:        "http://google.com",
				URLTitle:   "Go check this URL",
				Priority:   PriorityNormal,
			},
			expectedErr: nil,
		},
		{
			name:        "empty message",
			message:     Message{},
			expectedErr: ErrMessageEmpty,
		},
		{
			name: "message with valid size",
			message: Message{
				Message: randomStringsWithSize[MessageMaxLength],
			},
			expectedErr: nil,
		},
		{
			name: "message too long",
			message: Message{
				Message: randomStringsWithSize[MessageMaxLength+1],
			},
			expectedErr: ErrMessageTooLong,
		},
		{
			name: "message with valid title length",
			message: Message{
				Message: "fake message",
				Title:   randomStringsWithSize[MessageTitleMaxLength],
			},
			expectedErr: nil,
		},
		{
			name: "message with too long title",
			message: Message{
				Message: "fake message",
				Title:   randomStringsWithSize[MessageTitleMaxLength+1],
			},
			expectedErr: ErrMessageTitleTooLong,
		},
		{
			name: "message with valid URL",
			message: Message{
				Message: "fake message",
				URL:     randomStringsWithSize[MessageURLMaxLength],
			},
			expectedErr: nil,
		},
		{
			name: "message with too long URL",
			message: Message{
				Message: "fake message",
				URL:     randomStringsWithSize[MessageURLMaxLength+1],
			},
			expectedErr: ErrMessageURLTooLong,
		},
		{
			name: "message with valid URL title",
			message: Message{
				Message:  "Test message",
				URL:      "http://google.com",
				URLTitle: randomStringsWithSize[MessageURLTitleMaxLength],
			},
			expectedErr: nil,
		},
		{
			name: "message with too long URL title",
			message: Message{
				Message:  "Test message",
				URL:      "http://google.com",
				URLTitle: randomStringsWithSize[MessageURLTitleMaxLength+1],
			},
			expectedErr: ErrMessageURLTitleTooLong,
		},
		{
			name: "message with URL without URL title",
			message: Message{
				Message:  "Test message",
				URLTitle: "URL Title",
			},
			expectedErr: ErrEmptyURL,
		},
		{
			name: "message with emergency priority without emergency parameters",
			message: Message{
				Message:  "Test message",
				Priority: PriorityEmergency,
			},
			expectedErr: ErrMissingEmergencyParameter,
		},
		{
			name: "message with emergency priority",
			message: Message{
				Message:  "Test message",
				Priority: PriorityEmergency,
				Expire:   time.Hour,
				Retry:    60 * time.Second,
			},
			expectedErr: nil,
		},
		{
			name: "message with invalid priority",
			message: Message{
				Message:  "Test message",
				Priority: 6,
			},
			expectedErr: ErrInvalidPriority,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.message.validate(); err != tc.expectedErr {
				t.Errorf("expected %v; got %v", tc.expectedErr, err)
			}
		})
	}
}

// TestTokenFormat tests the token format
func TestTokenFormat(t *testing.T) {
	tt := []struct {
		name  string
		token string
		err   error
	}{
		{"empty token", "", ErrEmptyToken},
		{"invalid token 1", "uQiR-po4DXghDmr9QzzfQu27cmVRsG", ErrInvalidToken},
		{"invalid token 2", "agznej3rKEVAvPUxu9vvNnqpmZpokzF", ErrInvalidToken},
		{"valid token 1", "uQiRzpo4DXghDmr9QzzfQu27cmVRsG", nil},
		{"valid token 2", "gznej3rKEVAvPUxu9vvNnqpmZpokzF", nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := New(tc.token)
			if err := p.validate(); err != tc.err {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
		})
	}
}

// TestRecipientTokenFormat tests the token format
func TestRecipientTokenFormat(t *testing.T) {
	tt := []struct {
		name      string
		recipient string
		err       error
	}{
		{"empty recipient", "", ErrEmptyRecipientToken},
		{"invalid recipient 1", "uQiR-po4DXghDmr9QzzfQu27cmVRsG", ErrInvalidRecipientToken},
		{"invalid recipient 2", "agznej3rKEVAvPUxu9vvNnqpmZpokzF", ErrInvalidRecipientToken},
		{"valid recipient 1", "uQiRzpo4DXghDmr9QzzfQu27cmVRsG", nil},
		{"valid recipient 2", "gznej3rKEVAvPUxu9vvNnqpmZpokzF", nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := NewRecipient(tc.recipient)
			if err := p.validate(); err != tc.err {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
		})
	}
}

// TestMessageDeviceName tests the message device name format
func TestMessageDeviceName(t *testing.T) {
	tt := []struct {
		name   string
		device string
		err    error
	}{
		{"good device name 1", "yo_mama", nil},
		{"good device name 2", "droid-2", nil},
		{"good device name 2", "fasdfafdadfasdfa", nil},
		{"invalid device name 1", "yo&mama", ErrInvalidDeviceName},
		{"invalid device name 2", "my^device", ErrInvalidDeviceName},
		{"invalid device name 3", "d34342fasdfasdfasdfasdfasdfasd", ErrInvalidDeviceName},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			message := Message{
				Message:    "Test message",
				DeviceName: tc.device,
			}
			if err := message.validate(); err != tc.err {
				t.Fatalf("expected %v, got %v", tc.err, err)
			}
		})
	}
}

// TestEmptyReceipt tests if the receipt is empty trying to get details
func TestEmptyReceipt(t *testing.T) {
	app := New("uQiRzpo4DXghDmr9QzzfQu27cmVRsG")
	_, err := app.GetReceiptDetails("")
	if err == nil {
		t.Errorf("GetReceiptDetails should return an error")
	}

	if err != ErrEmptyReceipt {
		t.Errorf("Should get an ErrEmptyReceipt")
	}
}

// TestNewMessageWithTitle
func TestNewMessageWithTitle(t *testing.T) {
	message := NewMessageWithTitle("World", "Hello")

	expected := &Message{
		Title:   "Hello",
		Message: "World",
	}

	if !reflect.DeepEqual(message, expected) {
		t.Errorf("Invalid message from NewMessage")
	}
}

// TestEncodeRequest tests if the url params are encoded properly
func TestEncodeRequest(t *testing.T) {
	fakeTime := time.Now()
	fakeMessage := &Message{
		Message:     "My awesome message",
		Title:       "My title",
		Priority:    PriorityEmergency,
		URL:         "http://google.com",
		URLTitle:    "Google",
		Timestamp:   fakeTime.Unix(),
		Retry:       60 * time.Second,
		Expire:      time.Hour,
		DeviceName:  "SuperDevice",
		CallbackURL: "http://yourapp.com/callback",
		Sound:       SoundCosmic,
		HTML:        true,
	}

	// Expected arguments
	expected := map[string]string{
		"token":     fakePushover.token,
		"user":      fakeRecipient.token,
		"message":   fakeMessage.Message,
		"title":     fakeMessage.Title,
		"priority":  "2",
		"url":       "http://google.com",
		"url_title": "Google",
		"timestamp": fmt.Sprintf("%d", fakeTime.Unix()),
		"retry":     "60",
		"expire":    "3600",
		"device":    "SuperDevice",
		"callback":  "http://yourapp.com/callback",
		"sound":     "cosmic",
		"html":      "1",
	}

	// Encode request
	result, err := fakePushover.encodeRequest(fakeMessage, fakeRecipient)
	if err != nil {
		t.Errorf("Failed to encode request")
	}

	if reflect.DeepEqual(result, expected) == false {
		t.Errorf("Invalid message from NewMessage")
	}
}

// TestPostForm
func TestValidPostForm(t *testing.T) {
	// Fake server with the right headers
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Limit-App-Limit", "7500")
		w.Header().Set("X-Limit-App-Remaining", "6000")
		w.Header().Set("X-Limit-App-Reset", "1393653600")
		fmt.Fprintln(w, `{"status":1,"request":"e460545a8b333d0da2f3602aff3133d6"}`)
	}))
	defer ts.Close()

	p := &Pushover{}
	got, err := p.postForm(ts.URL, map[string]string{}, true)
	if err != nil {
		t.Fatalf("expected no error, got %q", err)
	}

	expected := &Response{
		Status:  1,
		ID:      "e460545a8b333d0da2f3602aff3133d6",
		Errors:  nil,
		Receipt: "",
		Limit: &Limit{
			Total:     7500,
			Remaining: 6000,
			NextReset: time.Unix(int64(1393653600), 0),
		},
	}

	if reflect.DeepEqual(got, expected) == false {
		t.Errorf("unexpected response from postFrom")
	}
}

// TestPostFormErrors
func TestPostFormErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":0,"request":"e460545a8b333d0da2f3602aff3133d6", "errors":["error1", "error2"]}`)
	}))
	defer ts.Close()

	p := &Pushover{}
	got, err := p.postForm(ts.URL, map[string]string{}, false)
	if got != nil {
		t.Fatalf("expected no result, got %q", got)
	}

	expected := Errors{"error1", "error2"}
	if reflect.DeepEqual(err, expected) == false {
		t.Errorf("failed to get postFormErrors")
	}
}

// TestErrorsString tests the custom error string
func TestErrorsString(t *testing.T) {
	e := &Errors{"error1", "error2"}
	got := e.Error()
	expected := fmt.Sprintf("Errors:\nerror1\nerror2")

	if got != expected {
		t.Errorf("invalid error string\ngot:\n%s\nexpected:\n%s\n", got, expected)
	}
}

// TestGetRecipienDetails
func TestGetRecipienDetails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":1,"request":"e460545a8b333d0da2f3602aff3133d6"}`)
	}))
	defer ts.Close()

	APIEndpoint = ts.URL
	got, err := fakePushover.GetRecipientDetails(fakeRecipient)
	if err != nil {
		t.Fatalf("expected no error, got %q", err)
	}

	expected := &RecipientDetails{
		Status:    1,
		Group:     0,
		Devices:   nil,
		RequestID: "e460545a8b333d0da2f3602aff3133d6",
		Errors:    nil,
	}

	if reflect.DeepEqual(got, expected) == false {
		t.Errorf("unexpected response from postFrom")
	}
}

// TestGetRecipienDetailsError
func TestGetRecipienDetailsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":0,"request":"e460545a8b333d0da2f3602aff3133d6"}`)
	}))
	defer ts.Close()

	APIEndpoint = ts.URL
	got, err := fakePushover.GetRecipientDetails(fakeRecipient)
	if got != nil {
		t.Fatalf("expected no recipient details, got %q", got)
	}

	if err == nil {
		t.Fatalf("expected an error")
	}

	if err != ErrInvalidRecipient {
		t.Fatalf("expected %q, got %q", ErrInvalidRecipient, got)
	}
}

// TestGetReceiptDetails
func TestGetReceiptDetails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			{
				"status": 1,
				"acknowledged": 1,
				"acknowledged_at": 1424305421,
				"acknowledged_by": "uYWtrQ4scpDU38cz5X5pvxNvu7b15",
				"last_delivered_at": 1424305379,
				"expired": 1,
				"expires_at": 1424308979,
				"called_back": 0,
				"called_back_at": 0,
				"request": "e95f35c2d75a100a3719b3764f0c8e47"
			}
		`)
	}))
	defer ts.Close()

	APIEndpoint = ts.URL
	got, err := fakePushover.GetReceiptDetails("fasdfadfasdfadfaf")
	if err != nil {
		t.Fatalf("expected no error, got %q", err)
	}

	// Expected times from timestamp
	acknowledgedAt := time.Unix(int64(1424305421), 0)
	lastDeliveredAt := time.Unix(int64(1424305379), 0)
	expiresAt := time.Unix(int64(1424308979), 0)

	// Expected result
	expected := &ReceiptDetails{
		Status:          1,
		Acknowledged:    true,
		AcknowledgedAt:  &acknowledgedAt,
		AcknowledgedBy:  "uYWtrQ4scpDU38cz5X5pvxNvu7b15",
		LastDeliveredAt: &lastDeliveredAt,
		Expired:         true,
		ExpiresAt:       &expiresAt,
		CalledBack:      false,
		CalledBackAt:    nil,
		ID:              "e95f35c2d75a100a3719b3764f0c8e47",
	}

	if reflect.DeepEqual(got, expected) == false {
		t.Errorf("unexpected receipt details")
	}
}

// TestCancelEmergencyNotification
func TestCancelEmergencyNotification(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":1,"request":"e460545a8b333d0da2f3602aff3133d6"}`)
	}))
	defer ts.Close()

	APIEndpoint = ts.URL
	p := &Pushover{}
	got, err := p.CancelEmergencyNotification("fasfaasdfa")
	if err != nil {
		t.Fatalf("expected no error, got %q", err)
	}

	expected := &Response{
		Status:  1,
		ID:      "e460545a8b333d0da2f3602aff3133d6",
		Errors:  nil,
		Receipt: "",
		Limit:   nil,
	}

	if reflect.DeepEqual(got, expected) == false {
		t.Errorf("unexpected response from postFrom")
	}
}

// TestSendMessage
func TestSendMessage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Limit-App-Limit", "7500")
		w.Header().Set("X-Limit-App-Remaining", "6000")
		w.Header().Set("X-Limit-App-Reset", "1393653600")
		fmt.Fprintln(w, `{"status":1,"request":"e460545a8b333d0da2f3602aff3133d6"}`)
	}))
	defer ts.Close()

	APIEndpoint = ts.URL
	got, err := fakePushover.SendMessage(NewMessage("TestMessage"), fakeRecipient)
	if err != nil {
		t.Fatalf("expected no error, got %q", err)
	}

	expected := &Response{
		Status:  1,
		ID:      "e460545a8b333d0da2f3602aff3133d6",
		Errors:  nil,
		Receipt: "",
		Limit: &Limit{
			Total:     7500,
			Remaining: 6000,
			NextReset: time.Unix(int64(1393653600), 0),
		},
	}

	if reflect.DeepEqual(got, expected) == false {
		t.Errorf("unexpected response from postFrom")
	}
}

// Returns a random string with a fixed size
func getRandomString(size int) (string, error) {
	bytesSize := size
	if size%2 == 1 {
		// If the number of bytes is not pair add 1 so it's pair again, the
		// extra char will be removed at the end
		bytesSize++
	}
	bytesSize = (bytesSize / 2)

	// Create a random byte array reading from /dev/urandom
	b := make([]byte, bytesSize)

	// Read
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(b)[:size], nil
}
