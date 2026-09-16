# AI Configuration Guide

MrRSS supports AI-powered features including translation, summarization, and chat. This guide explains how to configure different AI services.

## Reading reports and article assistance

Open **More → AI reading report** above the article list. Select 1–20 articles
from the current list order, preview their sources, optionally enter a focus,
and generate a report. The report groups topics, links supporting source numbers
to the original articles, and suggests what to read next. Choose a profile or
reuse the summary profile/global configuration. A failed retry keeps the previous
report; Stop cancels the pending request. Use **Copy report** before closing if
you want to keep the result.

Reports use local article content, falling back to RSS excerpts. Missing content
is skipped and shown in the preview; websites are not fetched automatically.
Input is limited to 32,000 Unicode characters of source text in total, with at
most 6,000 per article, shared evenly across the selected articles. Partial
sources are labelled. The model's reference IDs must resolve to a supplied source,
but this checks citation identity, not factual entailment. Important conclusions
still need checking against the original articles. No report is saved automatically
and no article is marked read. Content is sent only after clicking Generate.

Article-chat follow-ups retain the supplied article, use the selected profile's
headers, and report actionable errors. The evidence quick prompt asks for brief
supporting excerpts and unsupported claims. Built-in summaries now emphasize the
main takeaway, key facts, and uncertainty while preserving custom prompts.

For local integrations, `POST /api/ai/reading-report/preview` accepts
`{"article_ids":[1,2]}`; `POST /api/ai/reading-report` also accepts optional
`profile_id` and `focus` (up to 500 characters). Both work in desktop and server
modes. See the [API reference](SERVER_MODE/swagger.json).

## Supported AI Services

MrRSS works with any OpenAI-compatible API service and Ollama for local models.

## Configuration Steps

### 1. OpenAI Configuration

#### Prerequisites

1. Go to [OpenAI Platform](https://platform.openai.com/api-keys)
2. Create an API key

#### Configuration

- **API Key**: Enter your OpenAI API key
- **Endpoint**: `https://api.openai.com/v1/chat/completions` (For Azure OpenAI, use your Azure endpoint URL)
- **Model**: Use any supported model (e.g., `gpt-4o`, `gpt-4o-mini`, `gpt-5.2`)

### 2. Ollama Configuration

#### Prerequisites

1. Install [Ollama](https://ollama.com/)
2. Pull a model: `ollama pull llama3.2:1b` (replace with desired model)

#### Configuration

- **API Key**: Leave empty (not required for local Ollama)
- **Endpoint**: `http://localhost:11434/api/generate`
- **Model**: Use the model name you pulled (e.g., `llama3.2:1b`)

### 3. Other OpenAI-Compatible Services

#### DeepSeek

- **Endpoint**: `https://api.deepseek.com/v1/chat/completions`
- **Model**: `deepseek-chat` or `deepseek-coder`

#### Moonshot (Kimi)

- **Endpoint**: `https://api.moonshot.cn/v1/chat/completions`
- **Model**: `moonshot-v1-8k`, `moonshot-v1-32k`, `moonshot-v1-128k`

## Important Considerations

### Cost Management

1. **Set Usage Limits**: Configure a maximum token limit in settings
2. **Monitor Usage**: Check the usage statistics regularly
3. **Choose Appropriate Models**:
   - If you use OpenAI-compatible API services, small models like `gpt-4o-mini` can reduce costs and satisfy most use cases.
   - For Ollama, use smaller or quantized models like `llama3.2:1b` to save resources and accelerate response times.

## Troubleshooting

### "Authentication Failed"

- Check your API key is correct
- Ensure the key has not expired
- Verify the key has the required permissions

### "Model Not Found"

- Check the model name spelling
- Ensure the model is available in your account
- For Ollama: Run `ollama list` to see available models

### "Connection Failed"

- Check your internet connection
- Verify the endpoint URL is correct
- For local models (Ollama): Ensure Ollama is running
- Check if proxy settings are required

### Slow Response

- Try a smaller model
- Check your internet connection speed
- For Ollama: Consider using a quantized model

## Privacy Considerations

- **Article content is sent to the AI provider** when using AI features
- For sensitive content, use local models
- Review the AI provider's privacy policy

## Additional Resources

- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Ollama Documentation](https://github.com/ollama/ollama)
- [Azure OpenAI Documentation](https://learn.microsoft.com/en-us/azure/ai-services/openai/)
