package pushover

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

// Fake values to be used in the tests
var fakeTime = time.Now()
var fakePushover = New("uQiRzpo4DXghDmr9QzzfQu27cmVRsG")
var fakeRecipient = NewRecipient("gznej3rKEVAvPUxu9vvNnqpmZpokzF")
var fakeMessage = &Message{
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

func TestValidMessage(t *testing.T) {
	message := Message{
		Message:    "Hello world !",
		Title:      "Exemple",
		DeviceName: "My_Device",
		URL:        "http://google.com",
		URLTitle:   "Go check this URL",
		Priority:   PriorityNormal,
	}

	err := message.validate()
	if err != nil {
		t.Errorf("Message should be valid instead there is an error : %s", err)
	}
}

// TestEmptyMessage tests if the message empty
func TestEmptyMessage(t *testing.T) {
	message := Message{}

	err := message.validate()
	if err == nil {
		t.Errorf("Message should not be valid")
	}

	if err != ErrMessageEmpty {
		t.Errorf("Should get an ErrMessageEmpty")
	}
}

// TestMessageSize tests if the message size is valid
func TestMessageSize(t *testing.T) {
	// Create a random 1024 char string
	randMessage, err := getRandomString(MessageMaxLength)
	if err != nil {
		log.Panic(err)
	}

	message := Message{
		Message: randMessage,
	}

	err = message.validate()
	if err != nil {
		t.Errorf("Message length should be valid")
	}

	// Create a random 1025 char string, it should be too long
	randMessage, err = getRandomString(MessageMaxLength + 1)
	if err != nil {
		log.Panic(err)
	}

	message.Message = randMessage
	err = message.validate()
	if err == nil {
		t.Errorf("Message should not be valid")
	}

	if err != ErrMessageTooLong {
		t.Errorf("Should get an ErrMessageTooLong")
	}
}

// TestMessageTitleSize tests if the message size is valid
func TestMessageTitleSize(t *testing.T) {
	randTitle, err := getRandomString(MessageTitleMaxLength)
	if err != nil {
		log.Panic(err)
	}

	message := Message{
		Message: "Test message",
		Title:   randTitle,
	}

	err = message.validate()
	if err != nil {
		t.Errorf("Message title length should be valid")
	}

	randTitle, err = getRandomString(MessageTitleMaxLength + 1)
	if err != nil {
		log.Panic(err)
	}

	message.Title = randTitle
	err = message.validate()
	if err == nil {
		t.Errorf("Message title should not be valid")
	}

	if err != ErrMessageTitleTooLong {
		t.Errorf("Should get an ErrMessageTitleTooLong")
	}
}

// TestMessageURLSize tests if the message size is valid
func TestMessageURLSize(t *testing.T) {
	randTitle, err := getRandomString(MessageURLMaxLength)
	if err != nil {
		log.Panic(err)
	}

	message := Message{
		Message: "Test message",
		URL:     randTitle,
	}

	err = message.validate()
	if err != nil {
		t.Errorf("Message URL length should be valid")
	}

	randTitle, err = getRandomString(MessageURLMaxLength + 1)
	if err != nil {
		log.Panic(err)
	}

	message.URL = randTitle
	err = message.validate()
	if err == nil {
		t.Errorf("Message URL should not be valid")
	}

	if err != ErrMessageURLTooLong {
		t.Errorf("Should get an ErrMessageURLTooLong")
	}
}

// TestMessageURLTitleSize tests if the url title size is valid
func TestMessageURLTitleSize(t *testing.T) {
	randTitle, err := getRandomString(MessageURLTitleMaxLength)
	if err != nil {
		log.Panic(err)
	}

	message := Message{
		Message:  "Test message",
		URL:      "http://google.com",
		URLTitle: randTitle,
	}

	err = message.validate()
	if err != nil {
		t.Errorf("Message URL title length should be valid")
	}

	randTitle, err = getRandomString(MessageURLTitleMaxLength + 1)
	if err != nil {
		log.Panic(err)
	}

	message.URLTitle = randTitle
	err = message.validate()
	if err == nil {
		t.Errorf("Message URL title should not be valid")
	}

	if err != ErrMessageURLTitleTooLong {
		t.Errorf("Should get an ErrMessageURLTitleTooLong")
	}
}

// TestMessageURLTitleWithNoURL tests if URL is set if the title is not
func TestMessageURLTitleWithNoURL(t *testing.T) {
	message := Message{
		Message:  "Test message",
		URLTitle: "URL Title",
	}

	err := message.validate()
	if err == nil {
		t.Errorf("Message should not be valid if URLTitle is set with no URL")
	}

	if err != ErrEmptyURL {
		t.Errorf("Should get an ErrEmptyURL")
	}
}

// TestMessagePriority tests message priorities
func TestMessageEmergencyPriority(t *testing.T) {
	message := Message{
		Message:  "Test message",
		Priority: 2,
	}

	err := message.validate()
	if err == nil {
		t.Errorf("Message should not be valid with no Emergency infos")
	}

	if err != ErrMissingEmergencyParameter {
		t.Errorf("Should get an ErrInvalidPriority")
	}
}

// TestMessageEmergencyPriority tests message with emergency priority
func TestMessagePriority(t *testing.T) {
	for _, p := range []int{-6, -3, 3, 6} {
		message := Message{
			Message:  "Test message",
			Priority: p,
		}

		err := message.validate()
		if err == nil {
			t.Errorf("Message should not be valid with a %d priority", p)
		}

		if err != ErrInvalidPriority {
			t.Errorf("Should get an ErrInvalidPriority")
		}
	}

	for _, p := range []int{PriorityLowest, PriorityLow, PriorityNormal, PriorityHigh, PriorityEmergency} {
		message := Message{
			Message:  "Test message",
			Priority: p,
			Expire:   time.Hour,
			Retry:    60 * time.Second,
		}

		err := message.validate()
		if err != nil {
			t.Errorf("Message should be valid with a %d good priority", p)
		}
	}
}

// TestTokenFormat tests the token format
func TestTokenFormat(t *testing.T) {
	// Test empty token
	p := New("")
	err := p.validate()
	if err == nil {
		t.Errorf("An empty token should not be valid")
	}

	if err != ErrEmptyToken {
		t.Errorf("Should get an ErrEmptyToken")
	}

	// Test good token
	goodTokens := []string{"uQiRzpo4DXghDmr9QzzfQu27cmVRsG", "gznej3rKEVAvPUxu9vvNnqpmZpokzF"}
	for _, goodToken := range goodTokens {
		p := New(goodToken)
		err := p.validate()
		if err != nil {
			t.Errorf("Token '%s' should be valid", goodToken)
		}
	}

	// Test bad token
	badTokens := []string{"uQiR-po4DXghDmr9QzzfQu27cmVRsG", "agznej3rKEVAvPUxu9vvNnqpmZpokzF", "test"}
	for _, badToken := range badTokens {
		p := New(badToken)
		err := p.validate()
		if err == nil {
			t.Errorf("Token '%s' should not be valid", badToken)
		}
	}
}

// TestRecipientTokenFormat tests the token format
func TestRecipientTokenFormat(t *testing.T) {
	// Test empty token
	u := NewRecipient("")
	err := u.validate()
	if err == nil {
		t.Errorf("An empty recipient token should not be valid")
	}

	if err != ErrEmptyRecipientToken {
		t.Errorf("Should get an ErrEmptyRecipientToken")
	}

	// Test good token
	goodTokens := []string{"uQiRzpo4DXghDmr9QzzfQu27cmVRsG", "gznej3rKEVAvPUxu9vvNnqpmZpokzF"}
	for _, goodToken := range goodTokens {
		u := NewRecipient(goodToken)
		err := u.validate()
		if err != nil {
			t.Errorf("Token '%s' should be valid", goodToken)
		}
	}

	// Test bad token
	badTokens := []string{"uQiR-po4DXghDmr9QzzfQu27cmVRsG", "agznej3rKEVAvPUxu9vvNnqpmZpokzF", "test"}
	for _, badToken := range badTokens {
		u := NewRecipient(badToken)
		err := u.validate()
		if err == nil {
			t.Errorf("Token '%s' should not be valid", badToken)
		}
	}
}

// TestMessageDeviceName tests the message device name format
func TestMessageDeviceName(t *testing.T) {
	message := NewMessage("Hello world")

	// Test good device names
	goodDeviceNames := []string{"yo_mama", "droid-2", "dfasdfasdfasdfasdfasdfasd"}
	for _, goodDeviceName := range goodDeviceNames {
		message.DeviceName = goodDeviceName
		err := message.validate()
		if err != nil {
			t.Errorf("Device name '%s' should be valid", goodDeviceName)
		}
	}

	// Test bad device names
	badDeviceNames := []string{"yo&mama", "super^device", "d34342fasdfasdfasdfasdfasdfasd"}
	for _, badDeviceName := range badDeviceNames {
		message.DeviceName = badDeviceName
		err := message.validate()
		if err == nil {
			t.Errorf("Device name '%s' should not be valid", badDeviceName)
		}
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

	if reflect.DeepEqual(message, expected) == false {
		t.Errorf("Invalid message from NewMessage")
	}
}

// TestEncodeRequest tests if the url params are encoded properly
func TestEncodeRequest(t *testing.T) {
	// Expected arguments
	expected := &url.Values{}
	expected.Set("token", fakePushover.token)
	expected.Add("user", fakeRecipient.token)
	expected.Add("message", fakeMessage.Message)
	expected.Add("title", fakeMessage.Title)
	expected.Add("priority", "2")
	expected.Add("url", "http://google.com")
	expected.Add("url_title", "Google")
	expected.Add("timestamp", fmt.Sprintf("%d", fakeTime.Unix()))
	expected.Add("retry", "60")
	expected.Add("expire", "3600")
	expected.Add("device", "SuperDevice")
	expected.Add("callback", "http://yourapp.com/callback")
	expected.Add("sound", "cosmic")
	expected.Add("html", "1")

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
	got, err := p.postForm(ts.URL, &url.Values{}, true)
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
	got, err := p.postForm(ts.URL, &url.Values{}, false)
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
	got, err := fakePushover.SendMessage(fakeMessage, fakeRecipient)
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
