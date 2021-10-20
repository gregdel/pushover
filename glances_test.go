package pushover

import (
	"testing"
)

func TestGlancesValidation(t *testing.T) {
	tests := []struct {
		name        string
		fields      *Glance
		expectedErr error
	}{
		{
			name: "valid message 1",
			fields: &Glance{
				Title:   StrP("Hello World!"),
				Text:    StrP("Hi!"),
				Subtext: StrP("Hello!"),
				Count:   IntP(10),
				Percent: IntP(15),
			},
			expectedErr: nil,
		},
		{
			name: "valid message 2",
			fields: &Glance{
				Title: StrP("quam nulla porttitor massa id"),
			},
			expectedErr: nil,
		},
		{
			name: "invalid message (long title)",
			fields: &Glance{
				Title: StrP("facilisi etiam dignissim diam quis enim lobortis scelerisque fermentum dui faucibus in ornare quam viverra"),
			},
			expectedErr: ErrGlancesTitleTooLong,
		},
		{
			name: "invalid message (long text)",
			fields: &Glance{
				Text: StrP("facilisi etiam dignissim diam quis enim lobortis scelerisque fermentum dui faucibus in ornare quam viverra"),
			},
			expectedErr: ErrGlancesTextTooLong,
		},
		{
			name: "invalid message (long subtext)",
			fields: &Glance{
				Subtext: StrP("facilisi etiam dignissim diam quis enim lobortis scelerisque fermentum dui faucibus in ornare quam viverra"),
			},
			expectedErr: ErrGlancesSubtextTooLong,
		},
		{
			name: "invalid message (percentage)",
			fields: &Glance{
				Percent: IntP(101),
			},
			expectedErr: ErrGlancesInvalidPercent,
		},
		{
			name: "invalid device",
			fields: &Glance{
				Title:      StrP("hi!"),
				DeviceName: "device!test",
			},
			expectedErr: ErrInvalidDeviceName,
		},
		{
			name: "missing data",
			fields: &Glance{
				DeviceName: "a",
			},
			expectedErr: ErrGlancesMissingData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Glance{
				Title:      tt.fields.Title,
				Text:       tt.fields.Text,
				Subtext:    tt.fields.Subtext,
				Count:      tt.fields.Count,
				Percent:    tt.fields.Percent,
				DeviceName: tt.fields.DeviceName,
			}
			if err := m.validate(); err != tt.expectedErr {
				t.Errorf("validate() error = %v, expected err %v", err, tt.expectedErr)
			}
		})
	}
}
