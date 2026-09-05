# Native AI provider configuration / 原生 AI 配置

In Settings → AI, create an AI profile, enter the provider API key and a model ID available to that key, then use Test Connection. Select the profile for summary/translation; chat uses the configured chat AI settings/profile. A successful short test does not guarantee that the account accepts long article requests.

在「设置 → AI」新增配置，填写服务商 API Key 和该账号可用的模型 ID，测试连接后，为摘要/翻译选择该配置；聊天使用其配置的 AI 设置或配置档。短连接测试成功不代表账号一定接受长文章请求。

| Protocol / 协议 | Endpoint / 地址 | Authentication / 认证 |
| --- | --- | --- |
| Claude native Messages | `https://api.anthropic.com/v1/messages` (also accepts official root or `/v1`) | Automatic `x-api-key`, `anthropic-version`; no relay required / 自动设置请求头，无需中转 |
| Gemini native | `https://generativelanguage.googleapis.com/v1beta` | Model field builds the `models/<model>:generateContent` path; key supplied automatically / 模型字段生成路径，自动携带密钥 |
| Gemini compatible Chat Completions | `https://generativelanguage.googleapis.com/v1beta/openai/chat/completions` | Bearer key; retain the `/openai/` path / 自动使用 Bearer，保留路径中的 `/openai/` |
| Compatible gateways / 兼容网关 | Provider's full `/chat/completions`, `/messages`, or `:generateContent` URL | Protocol selected from the explicit route / 根据完整接口路径选择协议 |

The model field takes precedence over an old model name in a native Gemini `:generateContent` URL. Custom route prefixes and query options remain intact. Claude and Gemini chat system instructions are sent in their native top-level fields.

Gemini 原生 URL 中旧的模型名会使用「模型」字段更新，保留自定义网关前缀和查询参数。Claude 与 Gemini 的聊天系统提示使用各自原生顶层字段发送。

Known protocol failures retain the provider error instead of trying unrelated request formats. Report the app version, protocol, redacted endpoint (no key/query credentials), model, operation and error code when reporting problems. Do not paste API keys or sensitive article content.

已识别协议失败时保留原始错误用于分类，不再尝试无关协议。反馈问题请附版本、协议、去掉密钥和敏感查询参数的地址、模型、操作与错误码；不要提交 API Key 或敏感文章。

References: [Claude Messages](https://platform.claude.com/docs/en/api/messages/create), [Gemini generateContent](https://ai.google.dev/api/generate-content), [Gemini compatibility](https://ai.google.dev/gemini-api/docs/openai).
