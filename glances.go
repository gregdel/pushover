// glances.go
// see: https://pushover.net/api/glances

package pushover

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	// GlancesAllDevices can be passed as a device name to send a glances-message to all devices
	GlancesAllDevices = ""
	// GlancesMessageMaxTitleLength: The title can be maximum 100 characters long
	GlancesMessageMaxTitleLength = 100
	// GlancesMessageMaxTextLength: The text can be maximum 100 characters long
	GlancesMessageMaxTextLength = 100
	// GlancesMessageMaxSubtextLength: The subtext can be maximum 100 characters long
	GlancesMessageMaxSubtextLength = 100
)

var (
	ErrGlancesTitleTooLong   = errors.New("pushover: glances title too long")
	ErrGlancesTextTooLong    = errors.New("pushover: glances text too long")
	ErrGlancesSubtextTooLong = errors.New("pushover: glances subtext too long")
	ErrGlancesInvalidPercent = errors.New("pushover: glances percent must be in range of 0-100")
)

type GlancesMessage struct {
	// Title(max 100): a description of the data being shown, such as "Widgets Sold"
	Title string
	// Text(max 100): the main line of data, used on most screens
	Text string
	// Subtext(max 100): a second line of data
	Subtext string
	// Count(can be negative): shown on smaller screens; useful for simple counts
	Count int
	// Percent(0-100): shown on some screens as a progress bar/circle
	Percent    int
	DeviceName string
}

type GlancesResponse struct {
	// Status: when your API call was successful, status will be 1
	Status int `json:"status"`
	// Request: random request identifier that you can use when contacting Pushover for support
	Request string `json:"request"`
	// Errors: If the API call failed for any reason,
	// you will receive an errors array detailing each problem
	Errors Errors `json:"errors"`
}

// SendGlancesMessage is used to send glances-message to a recipient.
// It can be used to display widgets on a smart watch
func (p *Pushover) SendGlancesMessage(msg *GlancesMessage, rec *Recipient) (*GlancesResponse, error) {
	// Validate pushover
	if err := p.validate(); err != nil {
		return nil, err
	}

	// Validate rec
	if err := rec.validate(); err != nil {
		return nil, err
	}

	// Validate msg
	if err := msg.validate(); err != nil {
		return nil, err
	}

	return msg.send(p.token, rec.token)
}

func (m *GlancesMessage) validate() error {
	if utf8.RuneCountInString(m.Title) > GlancesMessageMaxTitleLength {
		return ErrGlancesTitleTooLong
	}
	if utf8.RuneCountInString(m.Text) > GlancesMessageMaxTextLength {
		return ErrGlancesTextTooLong
	}
	if utf8.RuneCountInString(m.Subtext) > GlancesMessageMaxSubtextLength {
		return ErrGlancesSubtextTooLong
	}
	if m.Percent < 0 || m.Percent > 100 {
		return ErrGlancesInvalidPercent
	}
	// Test device name
	if m.DeviceName != "" {
		// Accept comma separated device names
		devices := strings.Split(m.DeviceName, ",")
		for _, d := range devices {
			if !deviceNameRegexp.MatchString(d) {
				return ErrInvalidDeviceName
			}
		}
	}
	return nil
}

// send sends the message using the pushover and the recipient tokens.
func (m *GlancesMessage) send(pToken, rToken string) (*GlancesResponse, error) {
	url := fmt.Sprintf("%s/glances.json", APIEndpoint)

	params := map[string]string{
		"token":   pToken,
		"user":    rToken,
		"count":   strconv.Itoa(m.Count),
		"percent": strconv.Itoa(m.Percent),
	}
	if m.DeviceName != "" {
		params["device"] = m.DeviceName
	}
	if m.Title != "" {
		params["title"] = m.Title
	}
	if m.Text != "" {
		params["text"] = m.Text
	}
	if m.Subtext != "" {
		params["subtext"] = m.Subtext
	}

	req, err := newURLEncodedRequest("POST", url, params)
	if err != nil {
		return nil, err
	}

	resp := &GlancesResponse{}
	if err = do(req, resp, true); err != nil {
		return nil, err
	}

	return resp, nil
}
