package pushover

import (
	"os"
	"regexp"
	"time"
)

var deviceNameRegexp *regexp.Regexp

func init() {
	deviceNameRegexp = regexp.MustCompile(`^[A-Za-z0-9_-]{1,25}$`)
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

	// Attachment file path
	AttachmentPath string
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

	// Test file attachement
	if m.AttachmentPath != "" {
		stat, err := os.Stat(m.AttachmentPath)
		if err != nil {
			return ErrInvalidAttachementPath
		}

		if stat.Size() > MessageMaxAttachementByte {
			return ErrMessageAttachementTooLarge
		}
	}

	return nil
}
