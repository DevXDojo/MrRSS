# XPath Mode for MrRSS

MrRSS supports XPath mode for extracting RSS-like content from websites that don't provide standard RSS/Atom feeds. This mode allows you to define XPath expressions to scrape article data directly from HTML or XML pages.

## How It Works

1. When adding a new feed, select "XPath" as the feed type
2. Choose between "HTML + XPath" or "XML + XPath" depending on your source
3. Provide the source URL and configure XPath expressions for different article elements
4. MrRSS will parse the page and extract articles using your XPath expressions

## XPath Types

XPath reads downloaded HTML/XML, not the DOM after JavaScript execution. If a
list is populated by an API, use a custom script for that API. For example, the
China government latest-policy page has an empty `//ul[@id="list-1-ajax-id"]`.
Use [`gov_policy_feed.py`](../scripts/examples/gov_policy_feed.py) through
[Custom Script mode](CUSTOM_SCRIPT_MODE.md). It supplies titles, stable links
and dates; enable full-text fetching for article bodies. Scripts use their
interpreter's network/proxy environment, not MrRSS proxy settings.

The date XPath must select a date element (such as `.//span[@class="date"]`),
not a link. A date such as `2026-09-17` uses the layout `2006-01-02`.

### HTML + XPath

Use this for regular web pages. The HTML will be parsed and cleaned before applying XPath expressions.

### XML + XPath

Use this for XML-based sources that aren't standard RSS/Atom feeds.

## Required Configuration

### Visual picker (HTML only)

In the add/edit XPath form, enter an HTTP(S) URL and open **Visual XPath picker**.
Click an article and use **Parent element** until the whole repeating container
is selected. Choose it, then select title/link and optional date, summary and
thumbnail elements inside that same article. Check the matched containers and
samples, then apply. Applying replaces the XPath configuration; nothing is saved
until you submit the feed form. Set the date layout there as needed.

The preview preserves document structure with neutral styling and image
placeholders. It does not execute page scripts, load external resources or use
your browser login. Links cannot navigate. Proxy settings from the form apply to
both the preview and subscription validation. Generated paths describe the
current page structure and may need editing if that structure changes.

### Source URL

The URL of the webpage or XML document to scrape.

### Item XPath (Required)

The XPath expression that selects all article containers on the page. This is the most important expression as it defines what constitutes an "article".

**Example:** `//div[contains(@class, "post")]` - selects all div elements with class containing "post"

## Optional XPath Expressions

### Title XPath

XPath expression to extract the article title, relative to each item.

**Example:** `.//h1[contains(@class, "title")]` - finds h1 elements with "title" in class within each article

### URL XPath

XPath expression to extract the article URL/link, relative to each item.

**Example:** `.//a[contains(@class, "link")]/@href` - extracts the href attribute of anchor tags

### Content XPath

XPath expression to extract the article content/summary, relative to each item.

**Example:** `.//div[contains(@class, "content")]` - selects content divs within each article

### Author XPath

XPath expression to extract the article author, relative to each item.

**Example:** `.//span[contains(@class, "author")]` - selects author spans within each article

### Timestamp XPath

XPath expression to extract the publication date/time, relative to each item.

**Example:** `.//time/@datetime` - extracts datetime attributes from time elements

### Time Format

The format of the timestamp extracted above. Uses Go time format layout.

**Common formats:**

- `2006-01-02 15:04:05` - RFC3339-like format
- `Mon, 02 Jan 2006 15:04:05 -0700` - RFC1123 format
- `2006-01-02T15:04:05Z` - ISO 8601 format

### Thumbnail XPath

XPath expression to extract article thumbnail images, relative to each item.

**Example:** `.//img/@src` - extracts src attributes from img elements

### Categories XPath

XPath expression to extract article categories/tags, relative to each item.

**Example:** `.//span[contains(@class, "tag")]` - selects tag spans within each article

### UID XPath

XPath expression to extract a unique identifier for each article, relative to each item.

**Example:** `.//article/@id` - extracts id attributes from article elements

## XPath Basics

XPath is a language for selecting nodes in XML/HTML documents. Here are some common patterns:

### Basic Selectors

- `//div` - Select all div elements anywhere in the document
- `/html/body/div` - Select div elements that are direct children of body
- `.//p` - Select all p elements within the current context (relative path)

### Attribute Selection

- `//a/@href` - Select href attributes of all links
- `//img/@src` - Select src attributes of all images
- `//div[@class="post"]` - Select divs with exact class "post"

### Class-Based Selection

- `//div[contains(@class, "post")]` - Select divs where class contains "post"
- `//div[@class="post" or @class="article"]` - Select divs with either class

### Text Content

- `//h1/text()` - Get text content of h1 elements
- `//div[@class="content"]//text()` - Get all text within content divs

### Position-Based Selection

- `//div[@class="post"][1]` - Select first post div
- `//div[@class="post"][position() <= 5]` - Select first 5 post divs

## Examples

### Blog with Article Cards

For a blog where articles are in divs with class "article-card":

- **Item XPath:** `//div[contains(@class, "article-card")]`
- **Title XPath:** `.//h2/a/text()`
- **URL XPath:** `.//h2/a/@href`
- **Content XPath:** `.//div[contains(@class, "excerpt")]`
- **Author XPath:** `.//span[contains(@class, "author")]/text()`
- **Timestamp XPath:** `.//time/@datetime`
- **Time Format:** `2006-01-02`

### News Site with Article List

For a news site with articles in li elements:

- **Item XPath:** `//ul[@class="news-list"]/li`
- **Title XPath:** `.//h3/a/text()`
- **URL XPath:** `.//h3/a/@href`
- **Content XPath:** `.//p[@class="summary"]/text()`
- **Timestamp XPath:** `.//span[@class="date"]/text()`
- **Time Format:** `Jan 2, 2006`

## Testing XPath Expressions

To test your XPath expressions:

1. Open the target webpage in a browser
2. Use browser developer tools (F12)
3. In the Console, you can test XPath with: `$x("//your/xpath/here")`
4. Adjust expressions until they select the desired elements

## Troubleshooting

### No Articles Found

- Check that your Item XPath is correct and matches article containers
- Verify the source URL is accessible
- Ensure the page structure hasn't changed

### Wrong Content Extracted

- Double-check your XPath expressions are relative (start with `.//`)
- Test expressions individually in browser dev tools
- Make sure class names or element structures haven't changed

### Date Parsing Issues

- Verify the Time Format matches your timestamp format
- Check that Timestamp XPath extracts the date string correctly
- Common formats: RFC3339, ISO 8601, or custom formats

## Advanced Usage

### Complex Content Extraction

For sites with complex content structures, you can use more advanced XPath:

```xpath
.//div[contains(@class, "content")]//p[not(contains(@class, "ads"))]/text()
```

This selects text from paragraphs within content divs, excluding ad paragraphs.

### Multiple Categories

To extract multiple categories/tags:

```xpath
.//span[contains(@class, "tag")]/text()
```

This will collect all tag texts within each article.

## Related Documentation

- [Custom Script Mode](CUSTOM_SCRIPT_MODE.md) - Alternative method using JavaScript
- [FreshRSS XPath Documentation](https://freshrss.github.io/FreshRSS/en/developers/OPML.html) - Reference for XPath usage in RSS readers
