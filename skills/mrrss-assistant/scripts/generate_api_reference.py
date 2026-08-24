#!/usr/bin/env python3
"""Generate the MrRSS skill API reference from the server-mode Swagger file."""

from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any


IMPORTANT_TAGS = {
    "articles",
    "feeds",
    "ai",
    "chat",
    "statistics",
    "settings",
    "summary",
    "translation",
    "tags",
    "saved-filters",
    "discovery",
    "rules",
}


def endpoint_group(path: str) -> str:
    parts = [part for part in path.split("/") if part]
    if not parts:
        return "other"
    if parts[0] == "api" and len(parts) > 1:
        return parts[1]
    return parts[0]


def format_parameters(operation: dict[str, Any]) -> list[str]:
    rows: list[str] = []
    for param in operation.get("parameters", []):
        name = param.get("name", "")
        location = param.get("in", "")
        required = "required" if param.get("required") else "optional"
        description = param.get("description") or param.get("type") or ""
        rows.append(f"  - `{name}` ({location}, {required}): {description}".rstrip())
    return rows


def generate(swagger: dict[str, Any]) -> str:
    info = swagger.get("info", {})
    base_path = swagger.get("basePath", "/api")
    paths = swagger.get("paths", {})

    lines = [
        "# MrRSS API Reference",
        "",
        "Generated from `docs/SERVER_MODE/swagger.json`. Regenerate with:",
        "",
        "```bash",
        "python skills/mrrss-assistant/scripts/generate_api_reference.py docs/SERVER_MODE/swagger.json skills/mrrss-assistant/references/api.md",
        "```",
        "",
        f"- API version: `{info.get('version', 'unknown')}`",
        f"- API root: `{{base_url}}{base_path}`",
        "- Endpoint paths below are relative to the API root unless they already start with `/api/`.",
        "",
        "## Usage Notes",
        "",
        "- Use `GET` endpoints freely for inspection.",
        "- Ask before `DELETE`, bulk updates, cache clearing, or settings changes.",
        "- Send JSON request bodies with `Content-Type: application/json` unless an endpoint describes file upload.",
        "- Redact credentials and API keys from user-facing output.",
        "",
    ]

    grouped: dict[str, list[tuple[str, str, dict[str, Any]]]] = {}
    for path, path_item in sorted(paths.items()):
        group = endpoint_group(path)
        for method, operation in sorted(path_item.items()):
            if method.lower() not in {"get", "post", "put", "patch", "delete"}:
                continue
            grouped.setdefault(group, []).append((method.upper(), path, operation))

    ordered_groups = sorted(grouped, key=lambda key: (key not in IMPORTANT_TAGS, key))
    for group in ordered_groups:
        lines.extend([f"## {group.replace('-', ' ').title()}", ""])
        for method, path, operation in grouped[group]:
            summary = operation.get("summary") or operation.get("description") or ""
            lines.append(f"### `{method} {path}`")
            if summary:
                lines.append("")
                lines.append(summary.strip())
            params = format_parameters(operation)
            if params:
                lines.append("")
                lines.append("Parameters:")
                lines.extend(params)
            if any(param.get("in") == "body" for param in operation.get("parameters", [])):
                lines.append("")
                lines.append("Request body: see the Swagger schema for full field details.")
            lines.append("")

    return "\n".join(lines).rstrip() + "\n"


def main() -> int:
    if len(sys.argv) != 3:
        print("Usage: generate_api_reference.py <swagger.json> <output.md>", file=sys.stderr)
        return 2

    swagger_path = Path(sys.argv[1])
    output_path = Path(sys.argv[2])
    with swagger_path.open("r", encoding="utf-8") as handle:
        swagger = json.load(handle)

    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(generate(swagger), encoding="utf-8", newline="\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
