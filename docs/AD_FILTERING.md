# Advertising filtering / 广告过滤

Configure **Settings → Content → Ad filtering**. AI profile selection is also
available under **Settings → AI → AI features**. Filtering works in desktop and
server readers; no browser extension is installed.

在「设置 → 内容 → 广告过滤」配置功能，也可以在「设置 → AI → AI 功能」指定模型。
桌面版和服务端阅读器使用相同功能，不安装浏览器扩展。

## Defaults and scope / 默认行为与范围

- Basic filtering is enabled: dedicated advertising network resources, explicit
  network slots and tracking pixels with both tiny dimensions and endpoint
  evidence. Names such as `ad`, “sponsored”, prices or brand names alone do not
  trigger removal. Ordinary links and ambiguous promotional prose remain.
- AI is off until enabled. **Review first** analyzes only when requested, shows
  quoted evidence, and waits for Apply. **Automatic** analyzes while reading and
  applies accepted results; the article stays readable during the request.
- AI uses existing profiles, provider protocols, proxy settings and shared usage
  limits. A local Ollama profile can keep analysis on the user's machine. External
  profiles receive article text, headings and link hostnames; URLs with query
  strings and source HTML are not included in the segment payload.
- AI covers RSS and extracted full text. The in-app webpage proxy applies basic
  resource filtering and static custom CSS rules only. Dynamic page content and
  external browsers are outside AI coverage.
- Original cached content is preserved. **Show unfiltered content** restores the
  sanitized original, including sections hidden by basic or AI filtering.

基础过滤默认启用，仅处理明确广告资源和有额外证据的跟踪像素，不凭关键词删除正文。
AI 默认关闭，启用后默认先审核；只有主动分析或开启自动模式才向模型发送文章文本。
AI 复用已有模型配置、代理和公共用量限制，可选择本地 Ollama。网页视图只支持基础
资源过滤与静态规则，AI 作用于 RSS 正文及提取全文；外部浏览器不受影响。
过滤属于显示操作，不修改原始缓存，可以随时显示经过安全清理的未过滤内容。

## Accuracy safeguards / 准确性保护

The model classifies numbered, non-overlapping blocks instead of generating
HTML or selectors. Editorial context accompanies candidates. Code, quotations,
headings, oversized blocks and structural containers cannot be AI removal
targets. Proposed decisions require known IDs, supported categories, an exact
evidence quote, a high model confidence score and a second confirming audit.
Broad removals (over 35% of readable text or leaving fewer than 80 characters)
are rejected. Publisher self-promotion is separately opt-in.

Each uncached document uses at most two model requests, with cancellation,
timeouts, a shared concurrency bound and a bounded in-memory decision cache.
The cache expires after 24 hours, resets on restart, and includes document,
options and model configuration in its key. Up to 120 eligible blocks / 40,000
characters are scanned; the reader indicates partial scans. Other content stays.

Two model judgments are safeguards, **not a measured accuracy guarantee**.
Model confidence is not a calibrated probability. Default review mode and the
original-content switch are therefore part of the feature, not error recovery
after deletion. Regression tests use deterministic model responses to verify
trust boundaries and preservation; they do not claim production-model accuracy.

模型只能选择编号片段，不能自行生成用于删除的 HTML 或选择器。候选片段附带正文
上下文；代码、引用、标题和结构容器不会成为 AI 删除目标。结果必须有原文证据、
有效编号、高置信度及第二次审核同意，删除范围过大会保留全文。发布者自身推广另有开关。
两次同意不代表准确率保证，模型置信度也不等于统计准确率，因此默认保留人工审核。
自动化测试验证协议与保护机制，不把模拟模型结果当作真实模型准确率评测。

## Exceptions, rules and saving / 例外、规则与保存

Site exceptions contain one hostname per line, including its subdomains. They
disable both basic and AI filtering for that site. Filtering options use an
explicit Save button with stale-write protection and an unsaved-changes prompt.
AI profile selection uses the existing settings autosave.

Advanced cosmetic rules are a restricted CSS subset, not full Adblock Plus or
uBlock list compatibility. Maximum: 100 rules / 32 KB. Example:

```text
example.org##.promotion
example.org#@#.editorial-disclosure
```

The second rule protects the element and its enclosing container from **basic**
filtering. Rules do not instruct the AI. Regex, scripts, remote lists and certain
expensive selectors are unsupported. Preview the supplied or pasted sample before
saving; AI previews consume the shared quota, while basic preview stays local.

站点例外每行一个主机名，包含子域名，同时跳过基础过滤与 AI。高级规则是受限 CSS
子集，不兼容完整的广告过滤列表；`#@#` 只保护基础过滤，不影响 AI 判断。
可在保存前试用样例，基础预览在本地执行，AI 预览计入公共用量。
