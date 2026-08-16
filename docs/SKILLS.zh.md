# MrRSS Skills

[English](SKILLS.md) | 简体中文

MrRSS 提供 Codex skill 包，方便用户让 AI agent 通过本地 REST API 检查和操作自己的 MrRSS 数据。

## 包含内容

Release 资产 `MrRSS-<version>-skills.zip` 包含：

- `mrrss-assistant/SKILL.md` - agent 工作流和安全规则。
- `mrrss-assistant/references/api.md` - 从 `docs/SERVER_MODE/swagger.json` 生成的 API 参考。
- `mrrss-assistant/scripts/generate_api_reference.py` - 用于维护 API 参考的生成脚本。
- `mrrss-assistant/agents/openai.yaml` - Codex UI 元数据。

## 安装

1. 从 GitHub release 页面下载 `MrRSS-<version>-skills.zip`。
2. 解压。
3. 将解压后的 `mrrss-assistant` 文件夹复制到 Codex skills 目录：

Windows:

```powershell
Copy-Item -Recurse .\mrrss-assistant "$env:USERPROFILE\.codex\skills\"
```

macOS 或 Linux:

```bash
cp -R ./mrrss-assistant ~/.codex/skills/
```

4. 重启 Codex，让它重新发现新的 skill。

## 使用

启动 MrRSS server 模式，或运行带本地 API 的桌面版。

Server 模式示例：

```bash
docker run -p 1234:1234 ghcr.io/devxdojo/mrrss:latest-amd64
```

然后在 Codex 中输入：

```text
Use $mrrss-assistant to inspect my unread MrRSS articles and summarize the most important items.
```

该 skill 默认访问 `http://localhost:1234/api`。如果你的实例使用其他主机或端口，请在提示词中说明。

## 安全模型

该 skill 会要求 agent：

- 优先使用只读 API；
- 在删除、批量状态修改、缓存清理、设置修改前征求确认；
- 在回复中隐藏凭据和 API key；
- 除非用户明确要求离线检查，否则使用 MrRSS API，而不是直接读取 SQLite 数据库。

## 维护

当 API 路由变更时，重新生成参考文档：

```bash
python skills/mrrss-assistant/scripts/generate_api_reference.py docs/SERVER_MODE/swagger.json skills/mrrss-assistant/references/api.md
```

CI 会检查生成结果是否同步。release workflow 会将 skill 打包为 `MrRSS-<version>-skills.zip` 并上传为 release asset。
