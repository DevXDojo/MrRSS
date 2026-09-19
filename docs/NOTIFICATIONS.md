# Message notifications / 消息推送

Open **Settings → Notifications / 设置 → 消息推送**. Notifications are off by default and work in both desktop and server builds while MrRSS is running.

## Quick start / 快速开始

1. Expand a service card and add a destination. Each service supports multiple destinations, such as a personal chat and a team channel.
2. Enter its credentials, send a test if desired, and save. Tests use the current form without requiring a save and contain no article content.
3. Add an article alert or scheduled briefing. Select the receiving destinations, optional conditions and schedule, then preview matches before saving.
4. Enable the master switch. Automatic delivery starts with articles received after activation; it does not send the existing archive.

展开渠道卡片并添加接收目标，填写凭据后可先测试再保存。然后创建提醒或摘要规则，选择接收目标、匹配条件和计划，通过预览确认效果。最后打开总开关；只处理启用后新收到的文章，不会批量发送历史文章。

## Channels / 渠道

| Service | Configuration |
| --- | --- |
| Telegram | Create a bot with BotFather. Start a conversation or add it to a group. Supply its bot token and chat ID or channel username. An optional topic ID targets a forum topic. The bot needs posting permission in channels. |
| Discord | Create a webhook under the receiving channel's **Settings → Integrations → Webhooks**. Paste its HTTPS URL. Automatic mentions and link previews are disabled. Forum webhooks requiring a thread are not supported in this release. |
| Feishu / Lark | Add a custom group bot and paste its webhook. If signature verification is enabled, enter the signing secret. Keyword restrictions must permit `MrRSS`; IP restrictions must permit the sending device. Keep its system clock accurate. |

Telegram 支持聊天、群组、频道和话题；Discord 使用频道 Webhook；飞书及 Lark 使用自定义群机器人，支持签名校验。飞书关键词校验请允许 `MrRSS`，开启 IP 白名单时请允许当前设备的出口 IP。

Endpoints are restricted to the official Telegram, Discord, Feishu and Lark services over HTTPS. Redirects are rejected. Connections follow the application's global proxy settings. Custom relay endpoints are not supported.

Credentials use MrRSS's existing machine-bound encryption. Configuration responses contain placeholders instead of tokens, webhook URLs or signing secrets. Leaving a placeholder unchanged preserves its saved value; replacing it updates the secret, and clearing an optional signing secret removes it. Do not expect copied encrypted settings to work on a different machine.

## Rules and content / 规则与内容

- **Article alerts** check at an interval of 1–1440 minutes. The scheduler wakes approximately every 15 seconds; network work can delay a cycle.
- **Scheduled briefings** run at a selected local time and on optional weekdays. No selected weekdays means every day. Choose the timezone where MrRSS runs, or an explicit timezone. A missed time can run later on the same scheduled day when the app is available.
- Combine up to 10 conditions using **all (AND)** or **any (OR)**. Conditions support title, summary/excerpt, feed name, feed category, author, URL, favorite and read-later state. Text comparisons ignore case; `does not contain` and `does not equal` provide exclusions. Favorite/read-later conditions examine state when a new article is scanned; changing an old article's state does not create a new event.
- Hidden articles are excluded. Unread-only is optional. Matching rules run independently, so overlapping rules can intentionally send the same article more than once.
- Each run includes 1–20 articles, in arrival order. Additional matches remain for the next scheduled run. Selective rules page through new articles, so a matching article is not restricted to the first page of incoming articles.
- Include or omit source links and article summaries. Summaries prefer cached AI summaries and otherwise use feed excerpts. This feature does not make additional AI calls or synthesize a cross-article AI report. It also works without AI configured.
- Previews inspect the latest 500 local articles, show the match count and sample messages, and never send or change reading state. This differs from the future-article baseline used by automatic delivery.

支持 AND / OR 条件组合与反向匹配，可按标题、摘要、订阅源、分类、作者、链接、收藏或稍后阅读状态筛选。每次最多 20 篇，剩余匹配文章留待下次计划。摘要优先使用缓存的 AI 摘要，其次使用 RSS 简介，不产生额外 AI 调用。预览使用最近 500 篇文章，不会实际发送。

## Quiet hours and changes / 静默与配置变更

Shared settings save automatically. Destination and rule editors have explicit Save and Cancel actions. Leaving an unsaved editor requires a discard confirmation. Concurrent configuration edits use a revision check to prevent a stale window overwriting newer settings.

Quiet hours support overnight ranges. Automatic scanning and delivery pause during that range and resume afterward; tests remain available. A daily briefing whose scheduled time falls during quiet hours is deferred while it is still the scheduled day. If quiet hours extend beyond midnight, it waits for the next eligible scheduled time.

Editing a rule or destination, disabling a destination/rule, or switching off notifications cancels the affected pending messages and resets their baseline. Re-enabling begins with newly received articles. Changing only a destination's display name preserves its baseline. To delay delivery without resetting it, use quiet hours.

公共选项自动保存，渠道与规则独立保存。静默时段内暂缓自动扫描和发送；跨日后按照下一个符合条件的定时时间执行。修改规则或接收目标、暂停规则/目标或关闭总开关，会取消对应待发消息；重新启用从新的文章开始。只想临时延后发送时，请使用静默时段。

## Delivery behavior / 发送行为

Article cursors and queued messages commit in the same SQLite transaction. Pending messages survive a restart. Long content is split into bounded messages, and each completed part is recorded separately. Each channel sends at most one queued part per scheduler cycle, preserving order during backoff.

Temporary network errors, HTTP 429 and server errors are retried up to three total attempts, with backoff and supported provider retry delays. Authentication and application-level rejections fail without repeated sends. As with most webhook APIs, an ambiguous network failure or crash after provider acceptance can result in a duplicate; exactly-once external delivery is not guaranteed.

Delivery history shows the latest 100 records. Completed records are retained for at most 30 days and capped at 1,000. Message bodies are cleared after success, final failure or cancellation; the history API never returns bodies or credentials. Configuration and queue storage remain local, and only explicitly enabled destinations receive article data.

发送队列与扫描进度原子保存，重启可恢复。长消息分段处理，每段独立记录。临时错误最多尝试三次，鉴权或平台业务错误直接显示失败。网络结果不确定时仍可能重复送达。记录显示最近 100 条，完成记录最多保留 30 天、1,000 条；成功、最终失败或取消后清除消息正文。

## HTTP API

| Method | Endpoint | Purpose |
| --- | --- | --- |
| GET | `/api/notifications/config` | Configuration with redacted secrets and revision. |
| PUT | `/api/notifications/config` | Validate and atomically replace configuration; include the loaded revision. `409` means reload before editing again. |
| POST | `/api/notifications/test` | Send a fixed test using a `Channel` object; masked fields are restored only for the matching stored channel ID and provider. |
| POST | `/api/notifications/preview` | Preview a `Rule` against recent articles without sending. |
| GET | `/api/notifications/history` | Recent delivery metadata; excludes message bodies and credentials. |

Payloads are defined in [`notification/config.go`](../internal/notification/config.go) and mirrored by [`notification.ts`](../frontend/src/types/notification.ts). The schema-driven `notification_config` setting is encrypted and owned by this API; generic settings autosaves cannot overwrite it. No new listener or access boundary is introduced.

Provider references: [Telegram Bot API](https://core.telegram.org/bots/api#sendmessage), [Discord webhooks](https://docs.discord.com/developers/resources/webhook#execute-webhook), [Feishu custom bots](https://open.feishu.cn/document/client-docs/bot-v3/add-custom-bot).
