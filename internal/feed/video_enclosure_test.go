package feed

import (
	"github.com/mmcdole/gofeed"
	"testing"
)

func TestGenericVideoExtraction(t *testing.T) {
	cases := []struct{ name, content, mime, enclosure, want string }{
		{"enclosure", "", "VIDEO/MP4; codecs=avc1", "https://media.example/a?token=abc", "https://media.example/a?token=abc"},
		{"relative enclosure", "", "video/webm", "/a.webm", "https://example.org/a.webm"},
		{"inline video", `<video src="/v.mp4?x=1&amp;y=2"></video>`, "", "", "https://example.org/v.mp4?x=1&y=2"},
		{"source", `<video><source src="v.webm"></video>`, "", "", "https://example.org/posts/v.webm"},
		{"picture source", `<picture><source src="cover.webp"></picture>`, "", "", ""},
		{"unsafe enclosure fallback", `<video src="/good.mp4"></video>`, "video/mp4", "javascript:alert(1)", "https://example.org/good.mp4"},
		{"unsafe HTML", `<video src="file:///secret"></video>`, "", "", ""},
		{"audio", "", "audio/mp3", "https://example.org/a.mp3", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item := &gofeed.Item{Link: "https://example.org/posts/a", Content: tc.content, Enclosures: []*gofeed.Enclosure{nil, {Type: tc.mime, URL: tc.enclosure}}}
			if got := extractVideoURL(item); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
	item := &gofeed.Item{Link: "https://www.youtube.com/watch?v=abc", Enclosures: []*gofeed.Enclosure{{Type: "video/mp4", URL: "https://example.org/v.mp4"}}}
	if got := extractVideoURL(item); got != "https://www.youtube.com/embed/abc" {
		t.Fatal(got)
	}
}
