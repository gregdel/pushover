package pushover

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func do(req *http.Request, resType interface{}, returnHeaders bool) error {
	client := http.DefaultClient

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Only 500 errors will not respond a readable result
	if resp.StatusCode >= http.StatusInternalServerError {
		return ErrHTTPPushover
	}

	// Decode the JSON response
	if err := json.NewDecoder(resp.Body).Decode(&resType); err != nil {
		return err
	}

	// Check if the unmarshaled data is a response
	r, ok := resType.(*Response)
	if !ok {
		return nil
	}

	// Check response status
	if r.Status != 1 {
		return r.Errors
	}

	// The headers are only returned when posting a new notification
	if returnHeaders {
		// Get app limits from headers
		appLimits, err := newLimit(resp.Header)
		if err != nil {
			return err
		}
		r.Limit = appLimits
	}

	return nil
}

// multipartRequest returns a new multipart request
func multipartRequest(method, url string, params map[string]string) (*http.Request, error) {
	body := &bytes.Buffer{}

	// Write the body as multipart form data
	w := multipart.NewWriter(body)

	// Handle file upload
	filePath, ok := params["attachment_path"]
	if ok {
		// Remove the path from the params
		delete(params, "attachment_path")

		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Write the file in the body
		fw, err := w.CreateFormFile("attachment", "poster")
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fw, file)
		if err != nil {
			return nil, err
		}
	}

	// Handle params
	for k, v := range params {
		if err := w.WriteField(k, v); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, nil
}
