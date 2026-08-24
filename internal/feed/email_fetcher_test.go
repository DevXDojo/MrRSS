package feed

import (
	"testing"

	"github.com/emersion/go-imap"
)

func TestEmailMatchesSenderFilter(t *testing.T) {
	msg := &imap.Message{
		Envelope: &imap.Envelope{
			From: []*imap.Address{
				{
					PersonalName: "Example Newsletter",
					MailboxName:  "news",
					HostName:     "example.com",
				},
			},
		},
	}

	tests := []struct {
		name   string
		filter string
		want   bool
	}{
		{name: "empty filter", filter: "", want: true},
		{name: "email match", filter: "news@example.com", want: true},
		{name: "domain match", filter: "example.com", want: true},
		{name: "name match", filter: "newsletter", want: true},
		{name: "different sender", filter: "alerts@example.org", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := emailMatchesSenderFilter(msg, tt.filter); got != tt.want {
				t.Fatalf("emailMatchesSenderFilter() = %v, want %v", got, tt.want)
			}
		})
	}
}
