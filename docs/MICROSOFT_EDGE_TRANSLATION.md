# Microsoft Edge translation without an API key

In **Settings → Content → Translation**, enable translation and select
**Microsoft Edge (no API key)**. No Azure subscription key or region is required.
Existing **Microsoft Translator** settings continue to select the Azure API.

This option uses the consumer Edge translation protocol: it obtains a temporary
token from `edge.microsoft.com` and sends the selected text to
`api-edge.cognitive.microsofttranslator.com`. Both requests use MrRSS's configured
application proxy. Tokens remain in memory; they are not saved in settings or logs.
Translated results use the existing local translation cache under a separate
provider name.

Long text is split into requests of at most 5,000 UTF-16 units, preferring whitespace
boundaries. A provider instance sends one request at a time. Authentication failures
may refresh the token once; rate-limit responses are returned without repeated
retries. Cancelling an HTTP translation request cancels its pending Edge requests.

The consumer endpoint is not an Azure API subscription and can change or become
unavailable. Network reachability depends on the user's environment; no domestic
connectivity guarantee is implied. If Microsoft rejects a request, use the existing
translation error feedback and try again later or select another configured provider.

Protocol reference: the [Pot Bing translation implementation](https://github.com/pot-app/pot-desktop/blob/master/src/services/translate/bing/index.jsx)
uses the same authentication and translation endpoints. MrRSS implements its own
request handling, proxy integration, token lifetime, cancellation, and response limits.

## 中文说明

在 **设置 → 内容 → 翻译** 中选择 **Microsoft Edge（免 API Key）** 即可使用，
无需填写 Azure 密钥或区域。原有 **Microsoft 翻译** 仍使用 Azure API。

此选项会将待翻译文本发送给微软，使用当前应用代理配置。临时令牌仅保存在内存中；
长文本自动分段，翻译失败不会被当作完整结果缓存。服务是否可达、请求限制及后续接口变化
取决于微软的消费者服务和网络环境，不保证在所有国内网络中均可直连。
