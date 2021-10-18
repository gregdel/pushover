package pushover

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	GlancesAllDevices              = ""
	GlancesMessageMaxTitleLength   = 100
	GlancesMessageMaxTextLength    = 100
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
	Percent int
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

func (p *Pushover) SendGlances(deviceName string, message *GlancesMessage, recipient *Recipient) (*GlancesResponse, error) {
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

	return message.send(deviceName, p.token, recipient.token)
}

func (m *GlancesMessage) validate() error {
	if len(m.Title) > GlancesMessageMaxTitleLength {
		return ErrGlancesTitleTooLong
	}
	if len(m.Text) > GlancesMessageMaxTextLength {
		return ErrGlancesTextTooLong
	}
	if len(m.Subtext) > GlancesMessageMaxSubtextLength {
		return ErrGlancesSubtextTooLong
	}
	if m.Percent < 0 || m.Percent > 100 {
		return ErrGlancesInvalidPercent
	}
	return nil
}

func (m *GlancesMessage) send(deviceName string, pToken, rToken string) (*GlancesResponse, error) {
	url := fmt.Sprintf("%s/glances.json", APIEndpoint)

	params := map[string]string{
		"token":   pToken,
		"user":    rToken,
		"device":  deviceName,
		"count":   strconv.Itoa(m.Count),
		"percent": strconv.Itoa(m.Percent),
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
