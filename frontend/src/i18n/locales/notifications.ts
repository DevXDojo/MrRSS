const messages = {
  unsavedTitle: ['Discard unsaved changes?', '放弃未保存的修改？'],
  unsavedMessage: [
    'The destination or rule you are editing has not been saved. Leave this editor and discard its changes?',
    '当前编辑的接收目标或规则尚未保存。离开后将放弃这些修改，确定离开吗？',
  ],
  waitForSave: ['Please wait for settings to finish saving.', '设置正在保存，请稍候。'],
  title: ['Notifications', '消息推送'],
  description: [
    'Keep up with what matters. Send article alerts and scheduled briefings to your favorite channels.',
    '把值得关注的内容送到身边。通过提醒规则和定时摘要，让订阅消息按你的节奏到达。',
  ],
  enable: ['Enable notifications', '启用消息推送'],
  running: ['Enabled', '已启用'],
  off: ['Not enabled', '尚未启用'],
  overview: [
    '{channels} active destinations · {rules} active rules',
    '{channels} 个启用的接收目标 · {rules} 条启用的规则',
  ],
  runtimeHint: [
    'MrRSS must be running. New or edited rules start with future articles; enabling notifications never sends your entire archive.',
    '需要保持 MrRSS 运行。新建或修改规则从之后收到的文章开始，不会一次性推送历史文章。',
  ],
  commonSettings: ['Shared preferences', '公共设置'],
  commonHint: [
    'Apply to all destinations. Changes here are saved automatically.',
    '对所有渠道生效，修改后自动保存。',
  ],
  timezone: ['Schedule timezone', '计划时区'],
  systemTimezone: ['System timezone (where MrRSS runs)', '系统时区（MrRSS 运行设备）'],
  quietHours: ['Quiet hours', '静默时段'],
  quietStart: ['From', '开始时间'],
  quietEnd: ['Until', '结束时间'],
  quietHint: [
    'Messages wait until quiet hours end. Overnight periods are supported. Connection tests can still be sent.',
    '静默期间暂缓自动推送，结束后继续处理；支持跨午夜。测试消息仍可主动发送。',
  ],
  channels: ['Delivery channels', '推送渠道'],
  channelsHint: [
    'Connect one or more destinations per service. Credentials are encrypted on this device.',
    '每种服务可添加多个接收目标，凭据在本机加密保存。',
  ],
  feishu: ['Feishu / Lark', '飞书 / Lark'],
  configured: ['{count} configured', '已配置 {count} 个'],
  notConfigured: ['Not configured', '未配置'],
  addChannel: ['Add destination', '添加接收目标'],
  channelName: ['Destination name', '接收目标名称'],
  channelEnabled: ['Enable this destination', '启用此接收目标'],
  credentialsHint: [
    'Saved secrets are hidden. Replace a field to update it; leave the dots to keep it.',
    '已保存的密钥显示为圆点；保留圆点即保持原值，输入新内容即可替换。',
  ],
  showSecrets: ['Show entered credentials', '显示输入的凭据'],
  hideSecrets: ['Hide credentials', '隐藏凭据'],
  token: ['Bot token', '机器人 Token'],
  chatId: ['Chat ID or channel username', '聊天 ID 或频道用户名'],
  threadId: ['Topic ID (0 = general)', '话题 ID（0 表示普通聊天）'],
  webhook: ['Webhook URL', 'Webhook 地址'],
  signingSecret: ['Signing secret (optional)', '签名密钥（选填）'],
  optional: ['Optional', '选填'],
  testHint: [
    'Send a fixed test using these fields, even before saving. No articles are included. Global proxy settings apply.',
    '使用当前填写的信息发送固定测试消息，无需先保存，不包含文章内容。遵循应用的全局代理设置。',
  ],
  test: ['Send test', '发送测试消息'],
  testing: ['Sending…', '正在发送…'],
  testSuccess: ['Test delivered. Check the receiving chat.', '测试已送达，请查看对应聊天。'],
  saveChannel: ['Save destination', '保存接收目标'],
  saving: ['Saving…', '正在保存…'],
  saved: ['Notification settings saved', '推送设置已保存'],
  ready: ['Enabled', '已启用'],
  paused: ['Paused', '已暂停'],
  editChannel: ['Edit {name}', '编辑 {name}'],
  removeChannel: ['Remove destination', '删除接收目标'],
  removeChannelConfirm: [
    'Remove the notification configuration for “{name}”?',
    '删除“{name}”的推送配置？',
  ],
  channelInUse: [
    'This destination is used by a rule. Remove it from those rules first.',
    '此目标仍被规则使用，请先在规则中取消选择此目标。',
  ],
  rules: ['Alerts & briefings', '提醒与摘要'],
  rulesHint: [
    'Choose what to watch, where it goes, and when it arrives. Every matching rule runs independently.',
    '选择关注的内容、接收渠道和发送时间。每条匹配的规则都会独立执行。',
  ],
  emptyRules: ['Make your first notification yours', '从一条适合你的推送开始'],
  emptyRulesHint: [
    'Follow a keyword with an alert, or collect new articles into a morning briefing.',
    '用关键词提醒关注动态，或用早间摘要集中了解新文章。',
  ],
  connectFirst: [
    'First connect a destination above, then create an alert or a daily briefing.',
    '先在上方连接一个接收目标，再添加提醒规则或每日摘要。',
  ],
  addAlert: ['Add alert', '添加提醒规则'],
  addDigest: ['Add briefing', '添加定时摘要'],
  newAlert: ['Keyword alert', '关键词提醒'],
  newDigest: ['Morning briefing', '早间摘要'],
  ruleName: ['Rule name', '规则名称'],
  ruleEnabled: ['Enable this rule', '启用此规则'],
  sendTo: ['Send to', '发送到'],
  deliveryMode: ['Delivery mode', '推送方式'],
  alert: ['Article alert', '文章提醒'],
  digest: ['Scheduled briefing', '定时摘要'],
  interval: ['Check interval (minutes)', '检查间隔（分钟）'],
  deliveryTime: ['Send at (schedule timezone)', '发送时间（计划时区）'],
  weekdays: ['Days of the week', '发送星期'],
  everyDayHint: ['No days selected means every day.', '不选择任何星期表示每天发送。'],
  conditions: ['Match conditions', '匹配条件'],
  matchLogic: ['Condition logic', '条件逻辑'],
  matchAll: ['Match all (AND)', '满足全部（AND）'],
  matchAny: ['Match any (OR)', '满足任一（OR）'],
  noConditions: [
    'No conditions: include all new, visible articles.',
    '未添加条件：包含全部新收到的可见文章。',
  ],
  field: ['Field', '字段'],
  operator: ['Operator', '运算符'],
  value: ['Value', '条件值'],
  yes: ['Yes', '是'],
  no: ['No', '否'],
  keywordPlaceholder: ['Enter a keyword or phrase', '输入关键词或短语'],
  removeCondition: ['Remove condition', '删除条件'],
  addCondition: ['Add condition', '添加条件'],
  maxItems: ['Articles per run (1–20)', '每次最多文章数（1–20）'],
  unreadOnly: ['Only unread articles', '仅包含未读文章'],
  includeSummary: ['Include article summaries', '包含文章摘要'],
  includeLink: ['Include original links', '包含原文链接'],
  ruleHint: [
    'Checks newly received articles. Summaries use cached AI summaries or source excerpts without extra AI requests. Excess matches wait for the next run. Editing or pausing resets the rule baseline and cancels its queued messages.',
    '仅检查新收到的文章。优先使用已有 AI 摘要，其次使用来源简介，不额外调用 AI。超出数量上限的匹配文章留到下次。修改或暂停会重置对应规则的起点，并取消其待发送消息。',
  ],
  preview: ['Preview matches', '预览匹配内容'],
  previewing: ['Preparing…', '正在预览…'],
  saveRule: ['Save rule', '保存规则'],
  previewResult: [
    '{matched} matches in the latest {scanned} articles',
    '最近 {scanned} 篇文章中匹配 {matched} 篇',
  ],
  previewHint: [
    'Preview only — nothing is sent. Uses the latest 500 articles to illustrate your rule. Automatic delivery uses articles received after activation.',
    '仅预览，不会发送。使用最近最多 500 篇文章展示规则效果；自动推送只处理启用后收到的文章。',
  ],
  noMatches: [
    'No articles match yet. Try a broader condition or wait for new articles.',
    '暂时没有匹配文章。可以放宽条件，或等待新文章到达。',
  ],
  dailyAt: ['Briefing at {time}', '{time} 定时摘要'],
  everyMinutes: ['Every {minutes} min', '每 {minutes} 分钟检查'],
  editRule: ['Edit {name}', '编辑 {name}'],
  removeRule: ['Remove rule', '删除规则'],
  removeRuleConfirm: [
    'Remove “{name}” and cancel its pending messages?',
    '删除“{name}”并取消其待发送消息？',
  ],
  history: ['Delivery history', '发送记录'],
  refreshHistory: ['Refresh delivery history', '刷新发送记录'],
  historyHint: [
    'Latest 100 messages; records are kept for up to 30 days. Temporary failures retry up to 3 attempts. Long messages are split.',
    '显示最近 100 条消息，记录最多保留 30 天。临时故障最多尝试 3 次，长消息会分段发送。',
  ],
  historyEmpty: [
    'No automatic deliveries yet. Tests are shown in the destination editor.',
    '暂无自动推送记录。测试结果显示在接收目标编辑区。',
  ],
  historyError: [
    'Could not load delivery history. Use refresh to try again.',
    '无法读取发送记录，请点击刷新重试。',
  ],
  deletedChannel: ['Removed destination', '已删除的接收目标'],
  attempts: ['{count} attempts', '已尝试 {count} 次'],
  retryAt: ['Retry after {time}', '下次重试：{time}'],
  loading: ['Loading notification settings…', '正在加载推送设置…'],
  reload: ['Reload', '重新加载'],
};

const providerHint = {
  telegram: ['Private chats, groups, channels & topics', '私聊、群组、频道与话题'],
  discord: ['Channel webhooks, with mentions suppressed', '频道 Webhook，自动屏蔽群体提及'],
  feishu: ['Custom group bots, with optional signing', '群自定义机器人，支持签名校验'],
};
const setup = {
  telegram: [
    'Create a bot with BotFather. Start a chat with it or add it to your group, then enter the chat ID. The bot needs permission to post to channels.',
    '通过 BotFather 创建机器人。先向它发起聊天或将其加入群组，再填写聊天 ID。推送到频道时需赋予机器人发消息权限。',
  ],
  discord: [
    'In Discord, open channel settings → Integrations → Webhooks. Create a webhook and paste its URL here.',
    '在 Discord 频道设置 → 整合 → Webhook 中创建 Webhook，将地址粘贴到下方。',
  ],
  feishu: [
    'Add a custom bot to your group and copy its webhook. If signature verification is enabled, enter the secret. Keyword security should allow “MrRSS”. Lark webhooks are supported too.',
    '在飞书群中添加自定义机器人并复制 Webhook。若启用了签名校验，请同时填写密钥；若使用关键词校验，请允许关键词“MrRSS”。也支持 Lark 国际版地址。',
  ],
};
const fields = {
  title: ['Article title', '文章标题'],
  summary: ['Summary / excerpt', '摘要 / 简介'],
  feed: ['Feed name', '订阅源名称'],
  category: ['Feed category', '订阅源分类'],
  author: ['Author', '作者'],
  url: ['Article URL', '文章链接'],
  favorite: ['Favorite', '已收藏'],
  read_later: ['Read later', '稍后阅读'],
};
const operators = {
  contains: ['Contains', '包含'],
  not_contains: ['Does not contain', '不包含'],
  equals: ['Equals', '等于'],
  not_equals: ['Does not equal', '不等于'],
};
const status = {
  sent: ['Delivered', '已送达'],
  pending: ['Queued / retrying', '等待发送 / 重试'],
  failed: ['Failed', '发送失败'],
  cancelled: ['Cancelled', '已取消'],
};
const errors = {
  internal: ['The operation could not be completed. Please try again.', '操作未完成，请稍后重试。'],
  conflict: [
    'Settings changed in another window. Cancel this draft and reload before saving again.',
    '设置已在其他窗口修改，请取消当前编辑并重新加载后再保存。',
  ],
  channel_name: [
    'Enter a destination name (up to 120 characters).',
    '请填写接收目标名称（最多 120 字符）。',
  ],
  telegram_credentials: [
    'Check the bot token, chat ID and topic ID.',
    '请检查机器人 Token、聊天 ID 和话题 ID。',
  ],
  webhook_url: [
    'Enter a valid HTTPS webhook from Discord, Feishu or Lark, without query parameters.',
    '请填写 Discord、飞书或 Lark 的有效 HTTPS Webhook 地址，不要附加查询参数。',
  ],
  timezone: ['Choose a valid timezone.', '请选择有效的时区。'],
  quiet_time: [
    'Set different start and end times for quiet hours.',
    '请为静默时段设置不同的开始与结束时间。',
  ],
  limit: [
    'Up to 20 destinations and 50 rules are supported.',
    '最多支持 20 个接收目标和 50 条规则。',
  ],
  duplicate_id: [
    'Duplicate configuration detected. Reload and try again.',
    '检测到重复配置，请重新加载后重试。',
  ],
  rule_name: ['Enter a rule name (up to 120 characters).', '请填写规则名称（最多 120 字符）。'],
  schedule: [
    'Check the time, interval (1–1440 minutes) and article limit (1–20).',
    '请检查时间、间隔（1–1440 分钟）和文章上限（1–20）。',
  ],
  rule_channels: [
    'Select at least one destination. A rule supports up to 10 conditions.',
    '请至少选择一个接收目标；每条规则最多支持 10 个条件。',
  ],
  condition: [
    'Complete every condition with a valid field, operator and value.',
    '请为每个条件填写有效的字段、运算符与条件值。',
  ],
  credentials_missing: [
    'Saved credentials are unavailable. Enter them again.',
    '已保存的凭据不可用，请重新输入。',
  ],
  credentials_unavailable: [
    'Could not read or encrypt credentials on this device.',
    '无法在此设备读取或加密凭据。',
  ],
  config_unavailable: [
    'Could not load notification settings. Reload to try again.',
    '无法读取推送设置，请重新加载。',
  ],
  config_invalid: ['The stored notification configuration is invalid.', '已保存的推送配置无效。'],
  invalid_request: [
    'Some settings are invalid. Check your entries and try again.',
    '部分设置无效，请检查填写内容后重试。',
  ],
  proxy: [
    'Could not initialize the connection. Check global proxy settings.',
    '无法建立连接，请检查全局代理设置。',
  ],
  network: [
    'Connection failed or timed out. Check the network and proxy; the provider may already have received the message.',
    '连接失败或超时，请检查网络和代理；接收平台可能已经收到消息。',
  ],
  rate_limit: [
    'The service is rate limiting requests. Automatic deliveries will retry later.',
    '服务触发频率限制，自动推送会稍后重试。',
  ],
  unauthorized: [
    'Access denied. Check credentials and bot permissions.',
    '访问被拒绝，请检查凭据与机器人的发送权限。',
  ],
  rejected: [
    'The service rejected the message. Check credentials, destination and bot permissions.',
    '平台拒绝了消息，请检查凭据、接收目标和机器人权限。',
  ],
  feishu_rejected: [
    'Feishu rejected the message. Check the signing secret, clock, keyword “MrRSS” and bot security settings.',
    '飞书拒绝了消息，请检查签名密钥、设备时间、关键词“MrRSS”和机器人的安全设置。',
  ],
  response: [
    'The service did not confirm delivery. Check the destination before trying again.',
    '平台未确认送达，请检查接收目标后再重试。',
  ],
};

function language(index: number) {
  const select = (items: Record<string, string[]>) =>
    Object.fromEntries(Object.entries(items).map(([key, values]) => [key, values[index]]));
  return {
    ...select(messages),
    providerHint: select(providerHint),
    setup: select(setup),
    fields: select(fields),
    operators: select(operators),
    status: select(status),
    errors: select(errors),
    days: Object.fromEntries(
      (index === 0
        ? ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
        : ['日', '一', '二', '三', '四', '五', '六']
      ).map((day, i) => [i, day])
    ),
  };
}
export const enNotifications = language(0);
export const zhNotifications = language(1);
