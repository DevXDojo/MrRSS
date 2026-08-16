# MrRSS Skills

English | [简体中文](SKILLS.zh.md)

MrRSS ships a Codex skill package for users who want an AI agent to inspect and operate their MrRSS data through the local REST API.

## What Is Included

The release asset `MrRSS-<version>-skills.zip` contains:

- `mrrss-assistant/SKILL.md` - agent workflow and safety rules.
- `mrrss-assistant/references/api.md` - generated API reference from `docs/SERVER_MODE/swagger.json`.
- `mrrss-assistant/scripts/generate_api_reference.py` - maintenance script for regenerating the API reference.
- `mrrss-assistant/agents/openai.yaml` - Codex UI metadata.

## Install

1. Download `MrRSS-<version>-skills.zip` from the GitHub release page.
2. Extract it.
3. Copy the extracted `mrrss-assistant` folder into your Codex skills directory:

Windows:

```powershell
Copy-Item -Recurse .\mrrss-assistant "$env:USERPROFILE\.codex\skills\"
```

macOS or Linux:

```bash
cp -R ./mrrss-assistant ~/.codex/skills/
```

4. Restart Codex so it can discover the new skill.

## Use

Start MrRSS in server mode or run the desktop app with its local API available.

Server mode:

```bash
docker run -p 1234:1234 ghcr.io/devxdojo/mrrss:latest-amd64
```

Then ask Codex:

```text
Use $mrrss-assistant to inspect my unread MrRSS articles and summarize the most important items.
```

The skill defaults to `http://localhost:1234/api`. If your instance uses another host or port, include it in the prompt.

## Safety Model

The skill tells agents to:

- prefer read-only API calls first;
- ask before destructive operations, bulk status changes, cache clearing, or settings updates;
- redact credentials and API keys from responses;
- use MrRSS API endpoints instead of direct SQLite access unless the user explicitly asks for offline inspection.

## Maintainers

When API routes change, regenerate the reference:

```bash
python skills/mrrss-assistant/scripts/generate_api_reference.py docs/SERVER_MODE/swagger.json skills/mrrss-assistant/references/api.md
```

CI checks that the generated reference stays in sync. The release workflow packages the skill as `MrRSS-<version>-skills.zip` and uploads it as a release asset.
