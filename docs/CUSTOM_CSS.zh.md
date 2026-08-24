# 自定义 CSS 指南

## 概述

MrRSS 允许您通过上传自己的 CSS 文件来自定义文章内容的外观。此功能使您能够：

- 调整排版（字体、大小、行高）
- 更改颜色和间距
- 设置标题、列表、引用和代码块的样式
- 自定义表格和图像显示
- 添加悬停效果和过渡

## 使用方法

1. 准备一个包含自定义样式的 CSS 文件
2. 转到 **设置** → **通用** → **阅读与显示**
3. 点击 **上传 CSS** 按钮
4. 选择您的 CSS 文件
5. 打开任何文章以查看应用的自定义样式

要删除自定义 CSS，请在同一设置部分点击 **删除 CSS** 按钮。

## CSS 目标定位

您的自定义 CSS 将应用于视图模式中渲染的文章内容。文章内容被包裹在具有以下类别的容器中：

```html
<div class="prose prose-content custom-css-active">
  <!-- 文章内容在这里 -->
</div>
```

### 可用类别

- `.prose` - Tailwind Typography 的基础 prose 样式
- `.prose-content` - 文章内容包装器
- `.custom-css-active` - 加载自定义 CSS 时应用

## 常见自定义

### 排版

```css
/* 更改文章内容的字体系列 */
.prose-content {
  font-family: 'Georgia', serif;
}

/* 调整字体大小 */
.prose-content {
  font-size: 18px;
  line-height: 1.8;
}

/* 设置标题样式 */
.prose-content h1,
.prose-content h2,
.prose-content h3 {
  color: #2c3e50;
  font-weight: 700;
  margin-top: 1.5em;
  margin-bottom: 0.5em;
}
```

### 链接

```css
/* 设置链接样式 */
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

### 引用

```css
/* 设置引用块样式 */
.prose-content blockquote {
  border-left: 4px solid #3498db;
  padding-left: 1em;
  margin: 1.5em 0;
  font-style: italic;
  color: #7f8c8d;
}
```

### 代码块

```css
/* 设置内联代码样式 */
.prose-content code {
  background-color: #f4f4f4;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
  font-size: 0.9em;
}

/* 设置代码块样式 */
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

### 表格

```css
/* 设置表格样式 */
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

### 图片

```css
/* 设置图片样式 */
.prose-content img {
  max-width: 100%;
  height: auto;
  border-radius: 5px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  margin: 1.5em 0;
}

/* 居中图片 */
.prose-content img {
  display: block;
  margin-left: auto;
  margin-right: auto;
}
```

### 列表

```css
/* 设置列表样式 */
.prose-content ul,
.prose-content ol {
  padding-left: 1.5em;
  margin: 1em 0;
}

.prose-content li {
  margin: 0.5em 0;
}

/* 自定义项目符号 */
.prose-content ul li {
  list-style-type: square;
}

.prose-content ul li::marker {
  color: #3498db;
}
```

## 高级示例

### 深色主题

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

### 打印友好样式

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

### 极简阅读体验

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

## 提示和最佳实践

1. **使用特定选择器**：定位 `.prose-content` 以避免影响 UI 的其他部分
2. **充分测试**：检查您的 CSS 在不同类型的文章中的外观
3. **保持简单**：从基本更改开始，然后逐步添加更多
4. **使用浏览器开发工具**：检查文章内容以找到要设置样式的正确元素
5. **文件大小限制**：CSS 文件必须小于 1MB
6. **需要有效的 CSS**：无效的 CSS 可能无法应用

## 故障排除

### 样式未应用

- 检查浏览器控制台是否有 CSS 错误
- 验证您的 CSS 语法是否有效
- 确保文件具有 `.css` 扩展名
- 尝试在上传后重新加载文章

### 样式影响错误的元素

- 使用更具体的选择器（例如 `.prose-content p` 而不是仅仅 `p`）
- 避免使用通用选择器，如 `div` 或 `span`
- 使用不同的文章类型测试您的 CSS

### 性能问题

- 保持合理的 CSS 文件大小（建议小于 100KB）
- 避免复杂的选择器和昂贵的属性
- 限制使用 `!important`

## 示例 CSS 文件

存储库中提供了一个完整的示例 `test-custom-style.css`，它演示了：

- 标题的渐变背景
- 自定义排版
- 带有悬停效果的样式链接
- 增强的代码块，带有语法颜色
- 响应式表格
- 图片的圆角
- 平滑过渡

## 资源

- [MDN CSS 参考](https://developer.mozilla.org/zh-CN/docs/Web/CSS)
- [Tailwind Typography 插件](https://tailwindcss.com/docs/typography-plugin)
- [CSS Tricks](https://css-tricks.com/)
