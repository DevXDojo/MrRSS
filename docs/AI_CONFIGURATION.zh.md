# AI 配置指南

MrRSS 支持 AI 功能，包括翻译、摘要和聊天。本指南介绍如何配置不同的 AI 服务。

## 支持的 AI 服务

MrRSS 可与任何兼容 OpenAI 的 API 服务以及用于本地模型的 Ollama 配合使用。

## 配置步骤

### 1. OpenAI 配置

#### 前置条件

1. 访问 [OpenAI Platform](https://platform.openai.com/api-keys)
2. 创建一个 API 密钥

#### 配置

- **API 密钥**：输入您的 OpenAI API 密钥
- **端点**：`https://api.openai.com/v1/chat/completions`（对于 Azure OpenAI，使用您的 Azure 端点 URL）
- **模型**：使用任何支持的模型（例如 `gpt-4o`、`gpt-4o-mini`、`gpt-4.1`）

### 2. Ollama 配置

#### 前置条件

1. 安装 [Ollama](https://ollama.com/)
2. 拉取模型：`ollama pull llama3.2:1b`（替换为所需的模型）

#### 配置

- **API 密钥**：留空（本地 Ollama 不需要）
- **端点**：`http://localhost:11434/api/generate`
- **模型**：使用您拉取的模型名称（例如 `llama3.2:1b`）

### 3. 其他兼容 OpenAI 的服务

#### DeepSeek

- **端点**：`https://api.deepseek.com/v1/chat/completions`
- **模型**：`deepseek-chat` 或 `deepseek-coder`

#### Moonshot（月之暗面）

- **端点**：`https://api.moonshot.cn/v1/chat/completions`
- **模型**：`moonshot-v1-8k`、`moonshot-v1-32k`、`moonshot-v1-128k`

#### 智谱 AI (GLM)

- **端点**：`https://open.bigmodel.cn/api/paas/v4/chat/completions`
- **模型**：`glm-4-plus`、`glm-4-air`、`glm-4-flash`

#### 阿里云百炼

- **端点**：`https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions`
- **模型**：`qwen-plus`、`qwen-turbo`、`qwen-max`

## 重要注意事项

### 成本管理

1. **设置使用限制**：在设置中配置最大 token 限制
2. **监控使用情况**：定期检查使用统计信息
3. **选择合适的模型**：
   - 如果使用兼容 OpenAI 的 API 服务，小型模型如 `gpt-4o-mini` 可以降低成本并满足大多数使用场景。
   - 对于 Ollama，使用更小或量化的模型如 `llama3.2:1b` 可以节省资源并加快响应时间。

## 故障排除

### "身份验证失败"

- 检查您的 API 密钥是否正确
- 确保密钥未过期
- 验证密钥具有所需的权限

### "未找到模型"

- 检查模型名称拼写
- 确保该模型在您的账户中可用
- 对于 Ollama：运行 `ollama list` 查看可用模型

### "连接失败"

- 检查您的网络连接
- 验证端点 URL 是否正确
- 对于本地模型（Ollama）：确保 Ollama 正在运行
- 检查是否需要代理设置

### 响应缓慢

- 尝试使用更小的模型
- 检查您的网络连接速度
- 对于 Ollama：考虑使用量化模型

## 隐私注意事项

- **使用 AI 功能时，文章内容将发送到 AI 提供商**
- 对于敏感内容，请使用本地模型
- 查看 AI 提供商的隐私政策

## 其他资源

- [OpenAI API 文档](https://platform.openai.com/docs)
- [Ollama 文档](https://github.com/ollama/ollama)
- [Azure OpenAI 文档](https://learn.microsoft.com/zh-cn/azure/ai-services/openai/)
