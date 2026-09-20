const messages = {
  title: ['Ad filtering', '广告过滤'],
  description: [
    'Keep your reading focused, with reversible filtering and AI review.',
    '通过可恢复的过滤和 AI 审核，减少阅读中的广告干扰。',
  ],
  basic: ['Enable ad filtering', '启用广告过滤'],
  basicHint: [
    'Removes dedicated ad-network resources and verified tracking pixels. Ambiguous promotional text stays until AI review or a custom rule handles it.',
    '处理明确的广告网络资源与跟踪像素；模糊的推广文字交由 AI 审核或自定义规则处理。',
  ],
  ai: ['AI advertising analysis', 'AI 广告识别'],
  aiHint: [
    'Uses your existing AI profiles. Article text is sent only when you request analysis, or enable automatic mode. Local Ollama profiles are supported.',
    '复用已有 AI 配置。点击分析或开启自动模式后才会向所选模型发送文章文本，也支持本地 Ollama。',
  ],
  profile: ['AI profile for ad filtering', '广告过滤使用的 AI 配置'],
  defaultProfile: ['Use the default AI profile', '使用默认 AI 配置'],
  profileHint: [
    'Manage providers, models and credentials in Settings → AI. Profile selection saves automatically; filtering options use the Save button below.',
    '在「设置 → AI」管理模型、服务和凭据。AI 配置选择自动保存，下方过滤选项需点击保存。',
  ],
  mode: ['When to apply AI results', 'AI 结果应用方式'],
  review: ['Review first — analyze on demand', '先审核：手动分析并确认'],
  automatic: ['Automatically analyze and filter when reading', '阅读时自动分析并过滤'],
  automaticHint: [
    'The article remains readable while AI works. Up to two model requests per uncached document, subject to the shared AI usage limit. Repeated content reuses local decisions.',
    '分析期间可继续阅读。每份未缓存的内容最多进行两次模型请求，遵循公共 AI 用量上限；相同内容复用本地判断缓存。',
  ],
  selfPromotion: ['Also identify publisher self-promotion', '同时识别发布者自身推广'],
  selfPromotionHint: [
    'Includes newsletter, subscription and product pitches; neutral disclosures and editorial discussion are preserved.',
    '包括新闻邮件、订阅服务和产品推销；保留中性声明与正文讨论。',
  ],
  scope: [
    'AI applies to RSS and extracted full-text reading. Webpage view only filters known resources and static custom rules. It does not install extensions or affect external browsers.',
    'AI 作用于 RSS 正文和提取的全文；网页视图只过滤明确资源和静态自定义规则，不安装扩展，也不影响外部浏览器。',
  ],
  exceptions: ['Site exceptions', '站点例外'],
  exceptionsHint: [
    'One hostname per line; includes its subdomains. These sites skip both resource filtering and AI analysis.',
    '每行一个主机名，包含其子域名。这些站点跳过基础过滤和 AI 分析。',
  ],
  rules: ['Advanced: site-specific cosmetic rules', '高级：站点专用元素规则'],
  rulesHint: [
    "One domain##CSS selector per line. domain#{'@'}#selector protects an element from basic filtering. No remote lists, scripts or regex filters. Rules do not instruct the AI.",
    "每行 domain##CSS选择器；domain#{'@'}#选择器保护元素不被基础规则过滤。不支持远程列表、脚本或正则过滤。这些规则不会指示 AI。",
  ],
  save: ['Save filtering options', '保存过滤选项'],
  saved: ['Ad filtering settings saved', '广告过滤设置已保存'],
  reset: ['Discard changes', '放弃修改'],
  loading: ['Loading filtering settings…', '正在加载过滤设置…'],
  reload: ['Reload saved settings', '重新加载已保存设置'],
  unsavedTitle: ['Discard filtering changes?', '放弃过滤设置修改？'],
  unsavedMessage: [
    'Your filtering options have not been saved. Leave and discard these changes?',
    '广告过滤选项尚未保存，离开并放弃这些修改吗？',
  ],
  preview: ['Try a sample before saving', '保存前试用样例'],
  sampleURL: ['Sample article URL', '样例文章网址'],
  sampleHTML: ['Sample HTML', '样例 HTML'],
  previewBasic: ['Preview basic filtering', '预览基础过滤'],
  previewAI: ['Analyze sample with AI', '使用 AI 分析样例'],
  previewHint: [
    'Basic preview stays local. AI preview sends this sample to the selected model and uses the shared AI quota. Preview never changes saved articles.',
    '基础预览完全在本机处理；AI 预览会把此样例发送给所选模型，并计入公共 AI 用量。预览不会更改已保存文章。',
  ],
  previewResult: ['Result (safe HTML source)', '结果（安全 HTML 源码）'],
  removed: ['{count} resource/block removed', '已过滤 {count} 个资源或区块'],
  showOriginal: ['Show unfiltered content', '显示未过滤内容'],
  showFiltered: ['Show filtered content', '显示过滤后内容'],
  originalShown: ['Showing unfiltered content', '正在显示未过滤内容'],
  analyze: ['Analyze ads with AI', '使用 AI 分析广告'],
  analyzeAgain: ['Analyze again', '重新分析'],
  analyzing: ['AI is checking the article…', 'AI 正在检查文章…'],
  applySuggestions: ['Apply {count} suggestion(s)', '应用 {count} 项过滤建议'],
  aiApplied: ['AI filtered {count} segment(s)', 'AI 已过滤 {count} 个片段'],
  contentPreserved: ['The original content is preserved.', '原始内容已保留。'],
  guarded: [
    'The proposed removals were too broad. The article was preserved.',
    '建议移除的内容范围过大，已保留文章。',
  ],
  noSuggestions: [
    'No sufficiently supported advertising segments were found.',
    '未发现证据充分的广告片段。',
  ],
  partial: [
    'Checked {count} of {total} eligible blocks; remaining content was preserved.',
    '已检查 {total} 个可分析区块中的 {count} 个，其余内容保持不变。',
  ],
  reviewEvidence: ['Review the evidence', '查看判断依据'],
  aiCaution: [
    'AI can make mistakes. Two-pass agreement is a safeguard, not a guarantee of accuracy. Restore the unfiltered content whenever needed.',
    'AI 仍可能误判，两次判断一致也不代表准确率保证。需要时可随时恢复未过滤内容。',
  ],
} as const;
const errors = {
  invalid_request: ['Invalid request.', '请求格式无效。'],
  invalid_url: ['Enter a valid HTTP(S) article URL.', '请输入有效的 HTTP(S) 文章网址。'],
  invalid_rule: [
    'Invalid or unsupported CSS rule on line {line}.',
    '第 {line} 行 CSS 规则无效或不受支持。',
  ],
  invalid_domain: [
    'Use hostnames only, without a protocol, path or wildcard.',
    '请只填写主机名，不包含协议、路径或通配符。',
  ],
  rules_too_large: ['Rules exceed 32 KB.', '规则不能超过 32 KB。'],
  too_many_rules: ['Use no more than 100 custom rules.', '自定义规则不能超过 100 条。'],
  allowlist_too_large: ['Site exceptions exceed 8 KB.', '站点例外不能超过 8 KB。'],
  invalid_ai_options: ['Choose a supported AI mode.', '请选择有效的 AI 模式。'],
  conflict: [
    'Settings changed in another window. Reload before editing again.',
    '其他窗口已修改设置，请重新加载后编辑。',
  ],
  config_unavailable: ['Could not load filtering settings.', '无法加载过滤设置。'],
  save_failed: ['Could not save filtering settings.', '无法保存过滤设置。'],
  ai_not_enabled: [
    'Enable AI analysis and check the site exceptions first.',
    '请先启用 AI 分析，并检查站点例外。',
  ],
  ai_not_configured: [
    'Add an AI profile in Settings → AI, then select it for ad filtering.',
    '请在「设置 → AI」添加 AI 配置，再为广告过滤选择该配置。',
  ],
  ai_limit: ['The shared AI usage limit has been reached.', '已达到公共 AI 用量上限。'],
  ai_unavailable: [
    'AI analysis could not complete. Check your model and connection.',
    'AI 分析未能完成，请检查模型与网络连接。',
  ],
  ai_response: [
    'The model returned an invalid decision. Nothing was removed.',
    '模型返回的判断无效，未移除内容。',
  ],
  ai_timeout: [
    'AI analysis timed out. Try again or use a faster model.',
    'AI 分析超时，请重试或使用响应更快的模型。',
  ],
  content_too_large: [
    'This document exceeds the safe analysis limit.',
    '这份文档超过了安全分析范围。',
  ],
} as const;
const categories = {
  advertisement: ['Commercial advertising', '商业广告'],
  affiliate_promotion: ['Affiliate promotion', '返佣推广'],
  self_promotion: ['Publisher self-promotion', '发布者自身推广'],
} as const;
const locale = (index: 0 | 1) => ({
  ...Object.fromEntries(Object.entries(messages).map(([key, value]) => [key, value[index]])),
  errors: Object.fromEntries(Object.entries(errors).map(([key, value]) => [key, value[index]])),
  categories: Object.fromEntries(
    Object.entries(categories).map(([key, value]) => [key, value[index]])
  ),
});
export const enAdFilter = locale(0);
export const zhAdFilter = locale(1);
