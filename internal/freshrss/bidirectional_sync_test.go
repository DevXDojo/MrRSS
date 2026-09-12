package freshrss

import "testing"

func TestIsGoogleReaderStateSupportsFreshRSSAndMinifluxIDs(t *testing.T) {
	for _, category := range []string{
		"user/-/state/com.google/read",
		"user/1/state/com.google/read",
	} {
		if !isGoogleReaderState(category, "read") {
			t.Fatalf("%q was not recognized as read", category)
		}
	}
	if isGoogleReaderState("user/1/state/com.google/starred", "read") {
		t.Fatal("starred category was recognized as read")
	}
}
