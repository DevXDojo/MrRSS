# Partner advertisements / 合作广告

Edit the repository-root [`ads.json`](../ads.json). The shipped default is `{"version":1,"ads":[]}`: no advertisements. The desktop and server use the same `/api/ads` endpoint. On first access and at most hourly afterwards, the backend requests only the public file at `https://raw.githubusercontent.com/DevXDojo/MrRSS/main/ads.json`, through the configured global proxy. It sends no subscriptions, article contents, identifiers or credentials. There is no analytics, remote media or executable advertising content.

编辑根目录 [`ads.json`](../ads.json) 即可更新广告，默认 `ads` 为空。桌面版和 server 版共用 HTTP 接口；首次访问时优先获取 GitHub `main` 分支上的文件，之后最多每小时检查一次，沿用全局代理。请求不携带订阅、文章、用户标识或凭据，不包含统计、远程图片或可执行广告内容。

Valid remote JSON atomically replaces `<data directory>/ads.json`, preserving **all original fields**, including future fields. On network, HTTP, JSON or size errors, keep the last local configuration; use the embedded root file if no valid cache exists. A successful empty `ads` list removes previous advertisements. An unavailable cache directory does not prevent displaying valid remote data.

有效远程配置会原子替换用户数据目录中的 `ads.json`，**完整保留原始字段**。网络、HTTP、JSON 或大小校验失败时保留本地配置；无有效缓存时使用构建时嵌入的根目录文件。远程 `ads: []` 会撤下旧广告。缓存写入失败时仍可展示有效远程配置。

## Extensible format / 扩展格式

```json
{
  "version": 1,
  "ads": [
    {
      "id": "partner-offer-2026",
      "title": { "en": "Partner AI API offer", "zh": "合作 AI API 优惠" },
      "description": { "en": "Use our partner's AI API relay with a discount code.", "zh": "推荐使用合作中转站的 AI API，填写优惠码可享折扣。" },
      "url": "https://example.com/offer",
      "coupon_code": "EXAMPLE",
      "blocks": [
        { "type": "callout", "text": { "en": "Offer details", "zh": "优惠详情" } },
        { "type": "list", "items": ["Benefit one", { "zh": "优惠说明", "en": "Offer terms" }] }
      ],
      "layout": { "future_field": "retained but ignored by older versions" }
    }
  ]
}
```

The example is for editing/testing; it is not an active partner or a real discount. Replace the URL, code and terms with actual partner details before publishing.

示例仅用于编辑和测试，不代表真实合作方或优惠。发布前需填入实际合作链接、优惠码和条款。

- `title`, `description`, and block text accept either a plain string or a locale map (`zh-CN`, `en-US`, `zh`, `en`, etc.). Plain strings work in any language. The interface tries the exact locale, language code, English, Chinese, then another available locale.
- Keep a stable, nonempty `id` and a valid **HTTPS** `url` without embedded credentials. Invalid entries are skipped independently; duplicate IDs use the first valid entry.
- Supported blocks are `text`, `callout` and `list`. Unknown block types, fields, metadata, layouts and version values are ignored rather than rejecting the document. Rendering uses escaped text; remote HTML, scripts and CSS are never executed.
- For future complex creatives, supply a `fallback` object with plain `title`, `description`, `url` and optional `coupon_code`. Older clients use it when the corresponding primary field is absent or unsupported. Keep the outer `ads` array and compatible `id` as the durable envelope. A new format that removes all supported fields cannot be rendered by an old client; its raw configuration is still preserved for newer clients.
- Bound the document to 1 MiB. Display up to 32 valid ads, 12 supported blocks per ad and 8 list items per block. These limits bound resource use; exceeding display limits does not discard the original cache.

标题、说明和文本块支持字符串或语言映射。旧版本忽略未知字段、布局、版本号及内容块类型，单条无效广告不会影响其他广告。复杂广告应提供 `fallback` 基础展示字段，保留外层 `ads` 数组及广告 `id`，让旧客户端仍可展示标题、说明、链接和优惠码。远程 HTML、脚本和 CSS 均不会执行。配置上限为 1 MiB，展示数量和块数量有资源保护限制，缓存仍保留完整原文。

## Display and preferences / 展示与偏好

The first valid ad appears in the reader's lower-right popup; all valid ads appear at the bottom of **Settings → General**. The popup is hidden while Settings is open. Closing the popup hides it for the current session; the settings cards remain available. **Hide offers for one week** suppresses both locations for seven days and persists in the local database (`ads_snoozed_until`), including across restarts and advertisement changes. In server mode this preference is shared by users of that server database.

首条有效广告显示在主界面右下角；全部有效广告显示在**设置 → 通用**底部。设置打开时隐藏主界面弹窗。关闭弹窗只影响当前会话，设置卡片仍可查看；“一周内不再提醒”会同时隐藏两处广告，并在本地数据库保存七天的期限，重启或更换广告不影响期限。server 模式下使用同一数据库的用户共享该偏好。

For manual UI testing before publishing a real offer, block access to the source URL (or disconnect the network), put the sample in the configured data directory's `ads.json`, and restart. Test light/dark themes, narrow windows, both display locations, external navigation, session close, snooze/restart and unsupported blocks with fallback. Remove the sample afterwards. With network access, a successful remote configuration intentionally supersedes local test data.

人工测试时，可先阻断上述源地址或断网，将示例写入当前数据目录的 `ads.json` 后重启。检查浅色/深色主题、窄窗口、两处展示、访问链接、关闭弹窗、免提醒及重启后的持久化，以及复杂字段和 fallback。测试完移除示例。联网时有效远程配置会按设计覆盖本地测试数据。
