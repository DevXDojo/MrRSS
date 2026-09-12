package feed

import (
	"testing"

	"MrRSS/internal/models"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

func TestFeedThumbnailFallbacks(t *testing.T) {
	for _, tc := range []struct {
		name string
		item gofeed.Item
		want string
	}{
		{
			name: "empty image metadata does not mask a video poster",
			item: gofeed.Item{
				Image:   &gofeed.Image{URL: " "},
				Link:    "https://site.example/posts/entry/",
				Content: `<VIDEO controls title="a > b" poster='cover.jpg?x=1&amp;y=2'><source src="video.mp4"></VIDEO>`,
			},
			want: "https://site.example/posts/entry/cover.jpg?x=1&y=2",
		},
		{
			name: "description cover is available even with text-only full content",
			item: gofeed.Item{
				Content:     "A video entry",
				Description: `<video poster=/cover.jpg></video>`,
			},
			want: "https://feed.example/cover.jpg",
		},
		{
			name: "lazy image attributes and single quotes",
			item: gofeed.Item{Content: `<IMG src='/placeholder.gif' data-src='//cdn.example/cover.jpg'>`},
			want: "https://cdn.example/cover.jpg",
		},
		{
			name: "invalid poster does not hide a usable HTML image",
			item: gofeed.Item{Content: `<video poster="javascript:alert(1)"></video><img src='/image.jpg'>`},
			want: "https://feed.example/image.jpg",
		},
		{
			name: "skip empty enclosure and accept image MIME case",
			item: gofeed.Item{Enclosures: []*gofeed.Enclosure{
				nil, {Type: "image/jpeg"}, {Type: "IMAGE/JPEG", URL: " /enclosure.jpg "},
			}},
			want: "https://feed.example/enclosure.jpg",
		},
		{
			name: "explicit image metadata keeps priority over poster",
			item: gofeed.Item{
				Image:   &gofeed.Image{URL: "/chosen.jpg"},
				Content: `<video poster="/poster.jpg"></video>`,
			},
			want: "https://feed.example/chosen.jpg",
		},
		{
			name: "ignore video markup in script text and comments",
			item: gofeed.Item{Content: `<script>const x = '<video poster="/fake.jpg">';</script><!-- <video poster="/comment.jpg"> --><video poster="/real.jpg"></video>`},
			want: "https://feed.example/real.jpg",
		},
		{
			name: "video source is not used as an image without a poster",
			item: gofeed.Item{Content: `<video src="https://cdn.example/video.mp4"></video>`},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractImageURL(&tc.item, "https://feed.example/rss.xml"); got != tc.want {
				t.Errorf("thumbnail = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestMediaThumbnailsSkipEmptyEntries(t *testing.T) {
	item := &gofeed.Item{Extensions: ext.Extensions{"media": {
		"group": {
			{Children: map[string][]ext.Extension{"thumbnail": {{Attrs: map[string]string{"url": " "}}}}},
			{Children: map[string][]ext.Extension{"thumbnail": {
				{}, {Attrs: map[string]string{"url": "https://cdn.example/group.jpg"}},
			}}},
		},
		"thumbnail": {{}, {Attrs: map[string]string{"url": "https://cdn.example/direct.jpg"}}},
	}}}
	if got := extractMediaThumbnail(item); got != "https://cdn.example/group.jpg" {
		t.Errorf("group thumbnail = %q", got)
	}
	delete(item.Extensions["media"], "group")
	if got := extractMediaThumbnail(item); got != "https://cdn.example/direct.jpg" {
		t.Errorf("direct thumbnail = %q", got)
	}
}

func TestProcessVideoEntryRetainsPosterForArticleList(t *testing.T) {
	fetcher := &Fetcher{}
	entries := fetcher.processArticles(models.Feed{ID: 1, URL: "https://feed.example/rss"}, []*gofeed.Item{{
		Title:   "Video entry",
		Link:    "https://site.example/post/",
		Content: `<video poster="cover.jpg" controls><source src="video.mp4"></video>`,
	}})
	if len(entries) != 1 || entries[0].Article.ImageURL != "https://site.example/post/cover.jpg" {
		t.Fatalf("video cover did not reach the article model: %#v", entries)
	}
}
