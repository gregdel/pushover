package pushover

import "testing"

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
