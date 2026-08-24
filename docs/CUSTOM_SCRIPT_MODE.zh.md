# MrRSS 自定义脚本

MrRSS 支持自定义脚本，可用于从不提供标准 RSS/Atom 订阅源的源获取 RSS 订阅源。

## 使用 AI 编写爬虫脚本

如果您不熟悉编程，可以使用 AI 工具（如 ChatGPT、Claude 等）来帮助您编写爬虫脚本。以下是一个简单的指南和提示模板。

### 使用方法

1. 将下面的提示复制到 AI 工具中
2. 将 `{网页 URL}` 替换为您要抓取的实际网页 URL
3. 将 `{网站名称}` 替换为目标网站的名称
4. 让 AI 生成脚本代码
5. 将生成的脚本保存为 `.py` 文件到 `scripts` 文件夹中

### AI 提示模板

```plaintext
请为我编写一个 Python 脚本来爬取网页内容并将其转换为 RSS XML 格式。

目标网站：{网站名称}
网页 URL：{网页 URL}

脚本要求：

1. 使用 Python 编写，导入必要的库（如 requests、beautifulsoup4 等）
2. 接受 URL 作为命令行参数（如果需要）
3. 获取网页内容并解析 HTML
4. 提取文章列表，包括标题、链接、发布日期、摘要等
5. 输出标准的 RSS 2.0 XML 格式到 stdout
6. 处理可能的编码问题，确保输出为 UTF-8 编码
7. 添加适当的错误处理

请提供完整的、可运行的代码示例。
```

### 注意事项

- 生成的脚本需要满足 MrRSS 脚本要求（见下文）
- 确保脚本输出有效的 RSS XML 格式
- 在您的环境中测试脚本是否正常运行
- 遵守网站的 robots.txt 和使用条款

---

## 工作原理

1. 将脚本放置在 MrRSS 数据目录内的 `scripts` 文件夹中
2. 添加新订阅源时，选择"自定义脚本"而不是"RSS URL"
3. 从下拉列表中选择您的脚本
4. 刷新订阅源时会执行该脚本

## 脚本要求

您的脚本必须向 stdout 输出有效的 RSS 或 Atom XML。输出应遵循以下结构：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>订阅源标题</title>
    <link>https://example.com</link>
    <description>订阅源描述</description>
    <item>
      <title>文章标题</title>
      <link>https://example.com/article1</link>
      <description>文章内容或摘要</description>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
    </item>
    <!-- 更多项目... -->
  </channel>
</rss>
```

## 支持的脚本类型

| 扩展名 | 语言 | 使用的命令 |
| ------ | ------ | ---------- |
| `.py` | Python | `python3`（Windows 上为 `python`） |
| `.sh` | Shell | `bash` |
| `.ps1` | PowerShell | `powershell.exe`（Windows）或 `pwsh` |
| `.js` | Node.js | `node` |
| `.rb` | Ruby | `ruby` |

## 示例脚本

### Python 示例

```python
#!/usr/bin/env python3
import json
from datetime import datetime

# 从您的数据源获取数据
# items = fetch_from_api()

# 生成 RSS 输出
print('<?xml version="1.0" encoding="UTF-8"?>')
print('<rss version="2.0">')
print('  <channel>')
print('    <title>我的自定义订阅源</title>')
print('    <link>https://example.com</link>')
print('    <description>由脚本生成的自定义订阅源</description>')

# 示例项目
items = [
    {"title": "文章 1", "link": "https://example.com/1", "desc": "描述 1"},
    {"title": "文章 2", "link": "https://example.com/2", "desc": "描述 2"},
]

for item in items:
    print(f'    <item>')
    print(f'      <title>{item["title"]}</title>')
    print(f'      <link>{item["link"]}</link>')
    print(f'      <description>{item["desc"]}</description>')
    print(f'    </item>')

print('  </channel>')
print('</rss>')
```

### Shell 示例

```bash
#!/bin/bash

cat << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Shell 脚本订阅源</title>
    <link>https://example.com</link>
    <description>由 shell 脚本生成的订阅源</description>
    <item>
      <title>示例文章</title>
      <link>https://example.com/sample</link>
      <description>这是一个示例文章</description>
    </item>
  </channel>
</rss>
EOF
```

### Node.js 示例

```javascript
#!/usr/bin/env node

const items = [
  { title: '文章 1', link: 'https://example.com/1', desc: '描述 1' },
  { title: '文章 2', link: 'https://example.com/2', desc: '描述 2' },
];

console.log('<?xml version="1.0" encoding="UTF-8"?>');
console.log('<rss version="2.0">');
console.log('  <channel>');
console.log('    <title>Node.js 订阅源</title>');
console.log('    <link>https://example.com</link>');
console.log('    <description>由 Node.js 生成的订阅源</description>');

items.forEach(item => {
  console.log('    <item>');
  console.log(`      <title>${item.title}</title>`);
  console.log(`      <link>${item.link}</link>`);
  console.log(`      <description>${item.desc}</description>`);
  console.log('    </item>');
});

console.log('  </channel>');
console.log('</rss>');
```

## 提示

1. **错误处理**：如果脚本遇到错误，将错误消息写入 stderr。MrRSS 将在订阅源的错误指示器中显示此消息。

2. **超时**：脚本有 30 秒的超时时间。如果脚本运行时间更长，将被终止。

3. **工作目录**：脚本执行时以 scripts 文件夹作为工作目录。

4. **依赖项**：确保系统上安装了任何必需的依赖项（Python 包、Node 模块等）。

5. **测试**：首先从命令行测试您的脚本，以确保它生成有效的 RSS 输出：

   ```bash
   python3 your_script.py | xmllint --noout -
   ```

## 故障排除

- **未找到脚本**：确保脚本文件在 scripts 文件夹中并具有正确的扩展名
- **权限被拒绝**：在类 Unix 系统上，确保脚本是可执行的（`chmod +x script.py`）
- **输出无效**：通过手动测试来验证您的脚本输出有效的 XML
- **编码问题**：在脚本中始终使用 UTF-8 编码

## 数据目录位置

scripts 文件夹位于 MrRSS 数据目录中：

- **Windows**：`%APPDATA%\MrRSS\scripts\`
- **macOS**：`~/Library/Application Support/MrRSS/scripts/`
- **Linux**：`~/.local/share/MrRSS/scripts/`

您也可以在添加新订阅源时直接从 MrRSS 界面打开 scripts 文件夹。

## 其他资源

- [XPath 模式文档](XPATH_MODE.zh.md)
- [RSS 2.0 规范](https://validator.w3.org/feed/docs/rss2.html)
- [Atom 联合格式](https://tools.ietf.org/html/rfc4287)
