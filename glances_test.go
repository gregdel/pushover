package pushover

import (
	"testing"
)

func TestGlancesMessage_validate(t *testing.T) {
	tests := []struct {
		name    string
		fields  *GlancesMessage
		wantErr bool
	}{
		{
			name: "valid message 1",
			fields: &GlancesMessage{
				Title:   "Hello World!",
				Text:    "Hi!",
				Subtext: "Hello!",
				Count:   10,
				Percent: 15,
			},
			wantErr: false,
		},
		{
			name: "valid message 2",
			fields: &GlancesMessage{
				Title: "quam nulla porttitor massa id",
			},
			wantErr: false,
		},
		{
			name: "invalid message (long name)",
			fields: &GlancesMessage{
				Title: "facilisi etiam dignissim diam quis enim lobortis scelerisque fermentum dui faucibus in ornare quam viverra",
			},
			wantErr: true,
		},
		{
			name: "invalid message (percentage)",
			fields: &GlancesMessage{
				Percent: 101,
			},
			wantErr: true,
		},
		{
			name: "invalid device",
			fields: &GlancesMessage{
				DeviceName: "device!test",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &GlancesMessage{
				Title:      tt.fields.Title,
				Text:       tt.fields.Text,
				Subtext:    tt.fields.Subtext,
				Count:      tt.fields.Count,
				Percent:    tt.fields.Percent,
				DeviceName: tt.fields.DeviceName,
			}
			if err := m.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
