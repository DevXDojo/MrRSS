package siyuan

import (
	"html"
	"time"

	"MrRSS/internal/models"
	"MrRSS/internal/utils/textutil"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

func ArticleMarkdown(article *models.Article, content string) (string, error) {
	metadata := "<h1>" + html.EscapeString(article.Title) + "</h1>" +
		"<p><a href=\"" + html.EscapeString(article.URL) + "\">" + html.EscapeString(article.URL) + "</a></p>" +
		"<p>" + html.EscapeString(article.FeedTitle) + " · " + html.EscapeString(article.Author) + " · " + article.PublishedAt.Format(time.RFC3339) + "</p><hr>"
	// Normalize the body separately so Markdown feeds still render as Markdown
	// before adding the HTML metadata. Remote images stay at their original URLs.
	safeHTML := textutil.PrepareArticleContent(metadata, article.URL) + textutil.PrepareArticleContent(content, article.URL)
	return md.NewConverter(article.URL, true, nil).ConvertString(safeHTML)
}
