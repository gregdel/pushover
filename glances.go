package pushover

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	// GlancesAllDevices can be passed as a device name to send a glances-message to all devices
	GlancesAllDevices = ""
	// GlancesMessageMaxTitleLength is the max title length in a pushover glance update
	GlancesMessageMaxTitleLength = 100
	// GlancesMessageMaxTextLength is the max text length in a pushover glance update
	GlancesMessageMaxTextLength = 100
	// GlancesMessageMaxSubtextLength is the max subtext length in a pushover glance update
	GlancesMessageMaxSubtextLength = 100
)

// Glance represents a pushover glances update request.
type Glance struct {
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

func (m *Glance) validate() error {
	// check if data is present
	if m.Title == "" && m.Text == "" && m.Subtext == "" && m.Count == 0 && m.Percent == 0 {
		return ErrGlancesMissingData
	}
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
func (m *Glance) send(pToken, rToken string) (*Response, error) {
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

	resp := new(Response)
	if err = do(req, resp, true); err != nil {
		return nil, err
	}

	return resp, nil
}
