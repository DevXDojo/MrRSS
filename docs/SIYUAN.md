# SiYuan article clipping

Enable **Settings → Plugins → SiYuan integration**, then configure:

- **Address:** `http://127.0.0.1:6806` for SiYuan running alongside MrRSS. Server mode connects from the server, so use an address reachable from that machine.
- **API token:** Copy it from SiYuan's **Settings → Authentication**. MrRSS stores it encrypted. An empty token is supported for an instance that allows unauthenticated local API access.
- **Notebook ID:** The ID of an open destination notebook, not its display name.
- **Document folder:** A path starting with `/`, such as `/MrRSS` or `/Clippings/Articles`.

Open an article and click **Export to SiYuan** in its toolbar. The action is also available in the article modal and can be hidden or reordered in toolbar customization.

The exported Markdown includes the article title, source link, feed, author, publication time, and available article content. Images remain remote links; this feature does not download attachments into SiYuan. MrRSS appends the article ID to the document title to distinguish identical titles. Repeating an export does not overwrite an existing document at the same path.

Keep SiYuan running and the notebook open. If export fails, check the address, notebook ID, token, and folder. No connection or export happens just by enabling the setting. Local loopback connections bypass the global proxy; other connections follow the configured proxy. Export requests do not follow redirects.

API reference: [SiYuan's Markdown document creation API](https://github.com/siyuan-note/siyuan/blob/master/docs/API.md#create-a-document-with-markdown).
