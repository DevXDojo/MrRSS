# Custom CSS Guide


## Application themes

Custom CSS loads once at application startup, also when no article is open, and remains active when switching feeds or closing the reader. Uploading a replacement or deleting it takes effect immediately. No external theme service is required.

Use the application color variables for both light and dark themes; keep article-only rules under `.prose-content`. Existing article styles remain compatible.

```css
:root { --accent-color: #7c3aed; --accent-hover: #6d28d9; }
.dark-mode { --bg-primary: #202028; --bg-secondary: #282832; }
.prose-content h2 { color: var(--accent-color); }
```


## Overview

MrRSS allows you to customize the appearance of the application and article content by uploading your own CSS file. This feature enables you to:

- Adjust typography (fonts, sizes, line heights)
- Change colors and spacing
- Style headings, lists, quotes, and code blocks
- Customize table and image display
- Add hover effects and transitions

## How to Use

1. Prepare a CSS file with your custom styles
2. Go to **Settings** → **General** → **Reading & Display**
3. Click **Upload CSS** button
4. Select your CSS file
5. Your custom styles apply immediately across the application

To remove custom CSS, click **Delete CSS** button in the same settings section.

## CSS Targeting

Your custom CSS will be applied to article content rendered in the view mode. The article content is wrapped in a container with these classes:

```html
<div class="prose prose-content custom-css-active">
  <!-- Article content here -->
</div>
```

### Available Classes

- `.prose` - Base prose styling from Tailwind Typography
- `.prose-content` - Article content wrapper
- `.custom-css-active` - Applied when custom CSS is loaded

## Common Customizations

### Typography

```css
/* Change font family for article content */
.prose-content {
  font-family: 'Georgia', serif;
}

/* Adjust font size */
.prose-content {
  font-size: 18px;
  line-height: 1.8;
}

/* Style headings */
.prose-content h1,
.prose-content h2,
.prose-content h3 {
  color: #2c3e50;
  font-weight: 700;
  margin-top: 1.5em;
  margin-bottom: 0.5em;
}
```

### Links

```css
/* Style links */
.prose-content a {
  color: #3498db;
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
}

.prose-content a:hover {
  border-bottom-color: #3498db;
}
```

### Blockquotes

```css
/* Style blockquotes */
.prose-content blockquote {
  border-left: 4px solid #3498db;
  padding-left: 1em;
  margin: 1.5em 0;
  font-style: italic;
  color: #7f8c8d;
}
```

### Code Blocks

```css
/* Style inline code */
.prose-content code {
  background-color: #f4f4f4;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
}

/* Style code blocks */
.prose-content pre {
  background-color: #2c3e50;
  color: #ecf0f1;
  padding: 1em;
  border-radius: 5px;
  overflow-x: auto;
}

.prose-content pre code {
  background-color: transparent;
  padding: 0;
}
```

### Tables

```css
/* Style tables */
.prose-content table {
  width: 100%;
  border-collapse: collapse;
  margin: 1.5em 0;
}

.prose-content th,
.prose-content td {
  border: 1px solid #ddd;
  padding: 0.75em;
  text-align: left;
}

.prose-content th {
  background-color: #f4f4f4;
  font-weight: 600;
}

.prose-content tr:hover {
  background-color: #f9f9f9;
}
```

### Images

```css
/* Style images */
.prose-content img {
  max-width: 100%;
  height: auto;
  border-radius: 5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin: 1.5em 0;
}

/* Center images */
.prose-content img {
  display: block;
  margin-left: auto;
  margin-right: auto;
}
```

### Lists

```css
/* Style lists */
.prose-content ul,
.prose-content ol {
  padding-left: 1.5em;
  margin: 1em 0;
}

.prose-content li {
  margin: 0.5em 0;
}

/* Custom bullet points */
.prose-content ul li {
  list-style-type: square;
}

.prose-content ul li::marker {
  color: #3498db;
}
```

## Advanced Examples

### Dark Mode Theme

```css
.prose-content {
  background-color: #1a1a1a;
  color: #e0e0e0;
  padding: 2em;
  border-radius: 8px;
}

.prose-content h1,
.prose-content h2,
.prose-content h3 {
  color: #ffffff;
}

.prose-content a {
  color: #64b5f6;
}

.prose-content code {
  background-color: #2d2d2d;
  color: #e0e0e0;
}
```

### Print-Friendly Styles

```css
.prose-content {
  font-size: 12pt;
  line-height: 1.6;
  color: #000;
}

.prose-content a {
  color: #000;
  text-decoration: underline;
}

.prose-content img {
  max-width: 100%;
  page-break-inside: avoid;
}
```

### Minimalist Reading Experience

```css
.prose-content {
  font-family: 'Merriweather', Georgia, serif;
  font-size: 16px;
  line-height: 1.8;
  max-width: 680px;
  margin: 0 auto;
  color: #333;
}

.prose-content h1,
.prose-content h2 {
  font-family: 'Helvetica Neue', Arial, sans-serif;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.prose-content p {
  margin-bottom: 1.5em;
}
```

## Tips and Best Practices

1. **Use Specific Selectors**: Target `.prose-content` to avoid affecting other parts of the UI
2. **Test Thoroughly**: Check how your CSS looks with different types of articles
3. **Keep It Simple**: Start with basic changes and gradually add more
4. **Use Browser DevTools**: Inspect article content to find the right elements to style
5. **File Size Limit**: CSS files must be under 1MB
6. **Valid CSS Required**: Invalid CSS may fail to apply

## Troubleshooting

### Styles Not Applying

- Check browser console for CSS errors
- Verify your CSS syntax is valid
- Ensure the file has `.css` extension
- Try reloading the article after uploading

### Styles Affecting Wrong Elements

- Use more specific selectors (e.g., `.prose-content p` instead of just `p`)
- Avoid using generic selectors like `div` or `span`
- Test your CSS with different article types

### Performance Issues

- Keep CSS file size reasonable (under 100KB recommended)
- Avoid complex selectors and expensive properties
- Limit use of `!important`

## Example CSS File

A complete example is available in the repository as `test-custom-style.css`, which demonstrates:

- Gradient backgrounds for headings
- Custom typography
- Styled links with hover effects
- Enhanced code blocks with syntax colors
- Responsive tables
- Rounded corners on images
- Smooth transitions

## Resources

- [MDN CSS Reference](https://developer.mozilla.org/en-US/docs/Web/CSS)
- [Tailwind Typography Plugin](https://tailwindcss.com/docs/typography-plugin)
- [CSS Tricks](https://css-tricks.com/)
