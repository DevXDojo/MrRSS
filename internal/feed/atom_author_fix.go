package feed

import (
	"encoding/xml"
	"strings"

	"github.com/mmcdole/gofeed"
)

// fixFeedAuthors extracts simple text authors from feeds (both Atom and RSS)
// and populates the Author.Name field for items that don't already have an author.
// This handles feeds that use <author>Name</author> instead of
// the standard <author><name>Name</name></author> format.
// It only fills in missing authors and never overwrites existing ones.
func fixFeedAuthors(feed *gofeed.Feed, rawXML string) {
	// Build a map of items that need authors filled in
	// Only include items that don't already have an author
	itemsNeedingAuthors := make(map[int]*gofeed.Item)
	for i, item := range feed.Items {
		if item.Author == nil || item.Author.Name == "" {
			itemsNeedingAuthors[i] = item
		}
	}

	// If all items already have authors, no need to process
	if len(itemsNeedingAuthors) == 0 {
		return
	}

	// Try to parse as Atom feed first
	var atomFeed struct {
		XMLName xml.Name `xml:"feed"`
		Entries []struct {
			XMLName xml.Name `xml:"entry"`
			ID      string   `xml:"id"`
			Links   []struct {
				Href string `xml:"href,attr"`
				Rel  string `xml:"rel,attr"`
			} `xml:"link"`
			Title  string `xml:"title"`
			Author string `xml:"author"` // Simple text author
		} `xml:"entry"`
	}

	// Try Atom format
	atomErr := xml.Unmarshal([]byte(rawXML), &atomFeed)
	if atomErr == nil && len(atomFeed.Entries) > 0 {
		// Build author map from Atom entries
		itemAuthorMap := make(map[string]string)
		for _, entry := range atomFeed.Entries {
			if entry.Author != "" {
				// Extract the link href (prefer rel="alternate" or first link)
				link := ""
				for _, l := range entry.Links {
					if l.Rel == "alternate" || l.Rel == "" {
						link = l.Href
						break
					}
				}
				if link == "" && len(entry.Links) > 0 {
					link = entry.Links[0].Href
				}

				// Create a key from link and title
				key := link + "|" + entry.Title
				itemAuthorMap[key] = strings.TrimSpace(entry.Author)
			}
		}

		// Populate missing authors
		for _, item := range itemsNeedingAuthors {
			key := item.Link + "|" + item.Title
			if author, ok := itemAuthorMap[key]; ok {
				if item.Author == nil {
					item.Author = &gofeed.Person{}
				}
				item.Author.Name = author
			}
		}
		return
	}

	// Try RSS format
	var rssFeed struct {
		XMLName xml.Name `xml:"rss"`
		Channel struct {
			Items []struct {
				Title  string `xml:"title"`
				Link   string `xml:"link"`
				Author string `xml:"author"` // Simple text author
			} `xml:"item"`
		} `xml:"channel"`
	}

	rssErr := xml.Unmarshal([]byte(rawXML), &rssFeed)
	if rssErr == nil && len(rssFeed.Channel.Items) > 0 {
		// Build author map from RSS items
		itemAuthorMap := make(map[string]string)
		for _, item := range rssFeed.Channel.Items {
			if item.Author != "" {
				key := item.Link + "|" + item.Title
				itemAuthorMap[key] = strings.TrimSpace(item.Author)
			}
		}

		// Populate missing authors
		for _, item := range itemsNeedingAuthors {
			key := item.Link + "|" + item.Title
			if author, ok := itemAuthorMap[key]; ok {
				if item.Author == nil {
					item.Author = &gofeed.Person{}
				}
				item.Author.Name = author
			}
		}
		return
	}

	// If both parsing attempts failed, silently return
}
