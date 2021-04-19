package pushover

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
	"unicode/utf8"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789абвгдеёжзийклмнопрстуфхцчшщъыьэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ😊😂😉😎🎶🐱\u200d👓")

// Returns a random string with a fixed size
func getRandomString(size int) string {
	b := make([]rune, size)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

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
		rands := getRandomString(size)
		t := utf8.RuneCountInString(rands)
		fmt.Println(t)
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
		{"multiple devices", "device1,device2", nil},
		{"invalid multiple devices", "device1,device^2", ErrInvalidDeviceName},
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

// MultipartRequest
func TestMultipartRequest(t *testing.T) {
	tt := []struct {
		name           string
		attachmentSize int64
		expectedErr    error
	}{
		{
			name:           "valid multipart form",
			attachmentSize: 16,
		},
		{
			name:           "no attachment",
			expectedErr:    ErrMissingAttachment,
			attachmentSize: 0,
		},
		{
			name:           "no attachment",
			expectedErr:    ErrMessageAttachmentTooLarge,
			attachmentSize: MessageMaxAttachmentByte + 1,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			message := NewMessageWithTitle("World", "Hello")

			if tc.attachmentSize > 0 {
				buf := make([]byte, tc.attachmentSize)
				attachment := bytes.NewBuffer(buf)
				message.AddAttachment(attachment)
			}

			req, err := message.multipartRequest("pToken", "rToken", "url")
			if err != tc.expectedErr {
				t.Fatalf("expected %q, got %q", tc.expectedErr, err)
			}

			if err != nil {
				return
			}

			if err := req.ParseMultipartForm(tc.attachmentSize * 2); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			expectedValues := map[string][]string{
				"token":    {"pToken"},
				"user":     {"rToken"},
				"message":  {"World"},
				"priority": {"0"},
				"title":    {"Hello"},
			}

			if !reflect.DeepEqual(req.MultipartForm.Value, expectedValues) {
				t.Fatalf("Invalid form values, expected %v, got %v", expectedValues, req.MultipartForm.Value)
			}

			fileHeaders, ok := req.MultipartForm.File["attachment"]
			if !ok {
				t.Fatalf("expected an attachment, got nothing")
			}

			if len(fileHeaders) != 1 {
				t.Fatalf("expected one attachment, got %d", len(fileHeaders))
			}

			fileHeader := fileHeaders[0]
			if fileHeader.Filename != "attachment" {
				t.Fatalf("invalid attachment name: %s", fileHeader.Filename)
			}

			if fileHeader.Size != tc.attachmentSize {
				t.Fatalf("invalid attachment size, expected %d, got %d", tc.attachmentSize, fileHeader.Size)
			}
		})
	}
}
