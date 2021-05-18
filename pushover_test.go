package pushover

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// Fake values to be used in the tests
var fakePushover = New("uQiRzpo4DXghDmr9QzzfQu27cmVRsG")
var fakeRecipient = NewRecipient("gznej3rKEVAvPUxu9vvNnqpmZpokzF")

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
		Monospace:   true,
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
		"monospace": "1",
	}

	// Encode request
	result := fakeMessage.toMap(fakePushover.token, fakeRecipient.token)
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

	req, err := http.NewRequest("GET", ts.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	got := &Response{}
	if err := do(req, got, true); err != nil {
		t.Fatalf("failed to do request: %v", err)
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

	req, err := http.NewRequest("POST", ts.URL, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	got := &Response{}
	err = do(req, got, true)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	expected := Errors{"error1", "error2"}
	if reflect.DeepEqual(err, expected) == false {
		t.Errorf("failed to get postFormErrors")
	}
}

// TestGetRecipientDetails
func TestGetRecipientDetails(t *testing.T) {
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

// TestGetRecipientDetailsError
func TestGetRecipientDetailsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"status":0,"request":"e460545a8b333d0da2f3602aff3133d6", "errors": ["user key is invalid"]}`)
	}))
	defer ts.Close()

	APIEndpoint = ts.URL
	got, err := fakePushover.GetRecipientDetails(fakeRecipient)
	if err != nil {
		t.Fatalf("expected no error, got %q", err)
	}

	expected := &RecipientDetails{
		Status:    0,
		RequestID: "e460545a8b333d0da2f3602aff3133d6",
		Errors:    Errors{"user key is invalid"},
	}

	if reflect.DeepEqual(got, expected) == false {
		t.Errorf("unexpected recipient details\nExpected:\t%v\nGot\t%v", expected, got)
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
