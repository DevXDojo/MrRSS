# MrRSS XPath 模式

MrRSS 支持 XPath 模式，可以从不提供标准 RSS/Atom 订阅源的网站中提取类似 RSS 的内容。该模式允许您定义 XPath 表达式，直接从 HTML 或 XML 页面中抓取文章数据。

## 工作原理

1. 添加新订阅源时，选择 "XPath" 作为订阅源类型
2. 根据您的数据源选择 "HTML + XPath" 或 "XML + XPath"
3. 提供源 URL 并为不同的文章元素配置 XPath 表达式
4. MrRSS 将解析页面并使用您的 XPath 表达式提取文章

## XPath 类型

### HTML + XPath

用于常规网页。HTML 将在应用 XPath 表达式之前被解析和清理。

### XML + XPath

用于非标准 RSS/Atom 订阅源的基于 XML 的数据源。

## 必需配置

### 源 URL

要抓取的网页或 XML 文档的 URL。

### 项目 XPath（必需）

选择页面上所有文章容器的 XPath 表达式。这是最重要的表达式，因为它定义了什么构成一个"文章"。

**示例：** `//div[contains(@class, "post")]` - 选择所有 class 包含 "post" 的 div 元素

## 可选的 XPath 表达式

### 标题 XPath

提取文章标题的 XPath 表达式，相对于每个项目。

**示例：** `.//h1[contains(@class, "title")]` - 在每篇文章中查找 class 包含 "title" 的 h1 元素

### URL XPath

提取文章 URL/链接的 XPath 表达式，相对于每个项目。

**示例：** `.//a[contains(@class, "link")]/@href` - 提取锚标签的 href 属性

### 内容 XPath

提取文章内容/摘要的 XPath 表达式，相对于每个项目。

**示例：** `.//div[contains(@class, "content")]` - 选择每篇文章中的内容 div

### 作者 XPath

提取文章作者的 XPath 表达式，相对于每个项目。

**示例：** `.//span[contains(@class, "author")]` - 选择每篇文章中的作者 span

### 时间戳 XPath

提取发布日期/时间的 XPath 表达式，相对于每个项目。

**示例：** `.//time/@datetime` - 从 time 元素中提取 datetime 属性

### 时间格式

上面提取的时间戳的格式。使用 Go 时间格式布局。

**常用格式：**

- `2006-01-02 15:04:05` - 类似 RFC3339 的格式
- `Mon, 02 Jan 2006 15:04:05 -0700` - RFC1123 格式
- `2006-01-02T15:04:05Z` - ISO 8601 格式

### 缩略图 XPath

提取文章缩略图的 XPath 表达式，相对于每个项目。

**示例：** `.//img/@src` - 从 img 元素中提取 src 属性

### 分类 XPath

提取文章分类/标签的 XPath 表达式，相对于每个项目。

**示例：** `.//span[contains(@class, "tag")]` - 选择每篇文章中的标签 span

### UID XPath

为每篇文章提取唯一标识符的 XPath 表达式，相对于每个项目。

**示例：** `.//article/@id` - 从 article 元素中提取 id 属性

## XPath 基础

XPath 是一种用于选择 XML/HTML 文档中节点的语言。以下是一些常用模式：

### 基本选择器

- `//div` - 选择文档中任何位置的 div 元素
- `/html/body/div` - 选择作为 body 直接子元素的 div
- `.//p` - 选择当前上下文中的所有 p 元素（相对路径）

### 属性选择

- `//a/@href` - 选择所有链接的 href 属性
- `//img/@src` - 选择所有图像的 src 属性
- `//div[@class="post"]` - 选择 class 完全等于 "post" 的 div

### 基于类的选择

- `//div[contains(@class, "post")]` - 选择 class 包含 "post" 的 div
- `//div[@class="post" or @class="article"]` - 选择具有任一 class 的 div

### 文本内容

- `//h1/text()` - 获取 h1 元素的文本内容
- `//div[@class="content"]//text()` - 获取内容 div 中的所有文本

### 基于位置的选择

- `//div[@class="post"][1]` - 选择第一个 post div
- `//div[@class="post"][position() <= 5]` - 选择前 5 个 post div

## 示例

### 带有文章卡片的博客

对于文章位于 class 为 "article-card" 的 div 中的博客：

- **项目 XPath：** `//div[contains(@class, "article-card")]`
- **标题 XPath：** `.//h2/a/text()`
- **URL XPath：** `.//h2/a/@href`
- **内容 XPath：** `.//div[contains(@class, "excerpt")]`
- **作者 XPath：** `.//span[contains(@class, "author")]/text()`
- **时间戳 XPath：** `.//time/@datetime`
- **时间格式：** `2006-01-02`

### 带有文章列表的新闻网站

对于文章位于 li 元素中的新闻网站：

- **项目 XPath：** `//ul[@class="news-list"]/li`
- **标题 XPath：** `.//h3/a/text()`
- **URL XPath：** `.//h3/a/@href`
- **内容 XPath：** `.//p[@class="summary"]/text()`
- **时间戳 XPath：** `.//span[@class="date"]/text()`
- **时间格式：** `Jan 2, 2006`

## 测试 XPath 表达式

要测试您的 XPath 表达式：

1. 在浏览器中打开目标网页
2. 使用浏览器开发者工具（F12）
3. 在控制台中，您可以使用以下命令测试 XPath：`$x("//your/xpath/here")`
4. 调整表达式，直到它们选择所需的元素

## 故障排除

### 未找到文章

- 检查您的项目 XPath 是否正确并匹配文章容器
- 验证源 URL 是否可访问
- 确保页面结构未发生变化

### 提取的内容错误

- 仔细检查您的 XPath 表达式是否为相对路径（以 `.//` 开头）
- 在浏览器开发工具中单独测试表达式
- 确保 class 名称或元素结构未发生变化

### 日期解析问题

- 验证时间格式与您的时间戳格式匹配
- 检查时间戳 XPath 是否正确提取日期字符串
- 常用格式：RFC3339、ISO 8601 或自定义格式

## 高级用法

### 复杂内容提取

对于具有复杂内容结构的网站，您可以使用更高级的 XPath：

```xpath
.//div[contains(@class, "content")]//p[not(contains(@class, "ads"))]/text()
```

这将选择内容 div 中段落的文本，不包括广告段落。

### 多个分类

要提取多个分类/标签：

```xpath
.//span[contains(@class, "tag")]/text()
```

这将收集每篇文章中的所有标签文本。

## 相关文档

- [自定义脚本模式](CUSTOM_SCRIPT_MODE.zh.md) - 使用 JavaScript 的替代方法
- [FreshRSS XPath 文档](https://freshrss.github.io/FreshRSS/en/developers/OPML.html) - RSS 阅读器中 XPath 使用的参考
