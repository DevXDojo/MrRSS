#!/usr/bin/env python3
"""
i18n Usage Analyzer for MrRSS

This script analyzes i18n key usage across the codebase and generates a markdown table
with line numbers and clickable links to source files.

IMPORTANT NOTES:
- i18n keys MUST be called with t() function (e.g., t('key.name'))
- Searches .vue, .ts, .tsx, .js, .jsx files for i18n usage
- Skips comments and only searches actual code
- Identifies missing keys (used but not defined) and unused keys (defined but never used)
"""

import os
import re
import sys
from pathlib import Path
from typing import Dict, List, Tuple, Set
from urllib.parse import quote


def safe_print(text: str) -> None:
    """Print text safely, handling encoding issues on Windows console."""
    try:
        print(text)
    except UnicodeEncodeError:
        # Fallback for Windows console that doesn't support UTF-8
        # Replace emoji and other non-ASCII characters with ASCII equivalents
        text = text.replace('⚠️', '(!)').replace('⚠', '(!)')
        try:
            print(text)
        except UnicodeEncodeError:
            # Last resort: write to stdout buffer directly
            sys.stdout.buffer.write(text.encode('utf-8', errors='replace'))
            sys.stdout.buffer.write(b'\n')
            sys.stdout.flush()

# File paths (relative to project root)
EN_FILE = Path("frontend/src/i18n/locales/en.ts")
ZH_FILE = Path("frontend/src/i18n/locales/zh.ts")
TYPES_FILE = Path("frontend/src/i18n/types.ts")
FRONTEND_DIR = Path("frontend/src")


def extract_keys_from_locale(file_path: Path) -> Tuple[Dict[str, int], Set[str]]:
    """
    Extract i18n keys and their line numbers from locale files.
    Returns tuple of (keys_dict, category_keys).
    - keys_dict: maps key name to line number
    - category_keys: set of keys that are categories (objects with children, no direct value)

    Handles formatted code where key names and values may be on different lines.
    """
    keys = {}
    categories = set()

    if not file_path.exists():
        print(f"Warning: {file_path} not found")
        return keys, categories

    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()
        lines = content.split('\n')

    # Remove comments to avoid false matches
    # Single-line comments
    content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
    # Multi-line comments
    content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)

    # Pattern to match object keys (categories): keyName: {
    obj_pattern = re.compile(r'^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*\{', re.MULTILINE)

    # Pattern to match value keys at the start of a line: keyName: 'value'
    # Using ^ to ensure we only match keys at the start of a line (after optional whitespace)
    # This prevents matching strings like 'Progress: ' inside values
    value_pattern = re.compile(r'^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*[\'"`]', re.MULTILINE)

    # Collect all matches with their positions and line numbers
    all_matches = []  # List of (position, line_num, key, is_category)

    # Find all category matches
    for match in obj_pattern.finditer(content):
        key = match.group(1)
        line_num = content[:match.start()].count('\n') + 1
        all_matches.append((match.start(), line_num, key, True))

    # Find all value matches
    for match in value_pattern.finditer(content):
        key = match.group(1)
        line_num = content[:match.start()].count('\n') + 1
        all_matches.append((match.start(), line_num, key, False))

    # Sort all matches by position
    all_matches.sort(key=lambda x: x[0])

    # Track current nesting path and indentation
    path_stack = []

    # Process matches in order
    for position, line_num, key, is_category in all_matches:
        # Get the line to check indentation
        current_line = lines[line_num - 1] if line_num <= len(lines) else ""
        indent_match = re.match(r'^(\s*)', current_line)
        indent_level = len(indent_match.group(1)) if indent_match else 0

        # Update path stack based on indentation
        while path_stack and path_stack[-1][1] >= indent_level:
            path_stack.pop()

        if is_category:
            # This is a nested object (category)
            path_stack.append((key, indent_level))

            # Build full key path
            full_key = ".".join([p[0] for p in path_stack])
            if full_key not in keys:
                keys[full_key] = line_num
            categories.add(full_key)
        else:
            # This is a leaf value (actual translation)
            # Build full key path
            full_key = ".".join([p[0] for p in path_stack] + [key]) if path_stack else key

            if full_key not in keys:
                keys[full_key] = line_num

    return keys, categories


def extract_keys_from_types(file_path: Path) -> Dict[str, int]:
    """
    Extract i18n keys and their line numbers from types.ts

    Handles formatted code where key names and types may be on different lines.
    """
    keys = {}

    if not file_path.exists():
        print(f"Warning: {file_path} not found")
        return keys

    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()
        lines = content.split('\n')

    # Remove comments to avoid false matches
    # Single-line comments
    content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
    # Multi-line comments
    content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)

    # Pattern to match nested object types: keyName: {
    obj_pattern = re.compile(r'^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*\{', re.MULTILINE)

    # Pattern to match leaf value types at the start of a line: keyName: string;
    # Using ^ to ensure we only match keys at the start of a line
    value_pattern = re.compile(r'^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:\s*string;', re.MULTILINE)

    # Collect all matches with their positions and line numbers
    all_matches = []  # List of (position, line_num, key, is_category)

    # Find all category matches
    for match in obj_pattern.finditer(content):
        key = match.group(1)
        line_num = content[:match.start()].count('\n') + 1
        all_matches.append((match.start(), line_num, key, True))

    # Find all value matches
    for match in value_pattern.finditer(content):
        key = match.group(1)
        line_num = content[:match.start()].count('\n') + 1
        all_matches.append((match.start(), line_num, key, False))

    # Sort all matches by position
    all_matches.sort(key=lambda x: x[0])

    # Track current nesting path and indentation
    path_stack = []

    # Process matches in order
    for position, line_num, key, is_category in all_matches:
        # Get the line to check indentation
        current_line = lines[line_num - 1] if line_num <= len(lines) else ""
        indent_match = re.match(r'^(\s*)', current_line)
        indent_level = len(indent_match.group(1)) if indent_match else 0

        # Update path stack based on indentation
        while path_stack and path_stack[-1][1] >= indent_level:
            path_stack.pop()

        if is_category:
            # This is a nested object (category)
            path_stack.append((key, indent_level))

            # Build full key path
            full_key = ".".join([p[0] for p in path_stack])
            if full_key not in keys:
                keys[full_key] = line_num
        else:
            # This is a leaf value (actual translation)
            # Build full key path
            full_key = ".".join([p[0] for p in path_stack] + [key]) if path_stack else key

            if full_key not in keys:
                keys[full_key] = line_num

    return keys


def find_i18n_usage(key: str, search_dir: Path) -> List[Tuple[str, int]]:
    """
    Find all usages of an i18n key in the codebase.
    Returns list of (file_path, line_number) tuples.
    Only matches t() calls - i18n keys MUST start with t(.

    Valid i18n key characters: letters, numbers, dots, underscores, and hyphens only.
    This prevents false matches from function call patterns like 'emit(', 'closest(', etc.
    Handles multi-line t() calls (e.g., after formatting).
    """
    usages = []
    # Pattern ensures the key is inside a t() call with proper quoting
    # Handles: t('key'), t("key"), t('key', { param }), t("key", { count: 5 })
    # Uses (?<!\w) to ensure 't' is not part of another identifier (e.g., createElement)
    # Uses re.DOTALL to match across newlines (for formatted code)
    pattern = re.compile(
        rf"""(?<!\w)t\(\s*     # t( with optional whitespace, t must be a standalone identifier
        ['"`]({re.escape(key)})['"`]  # the key in quotes
        \s*(?:,.+?)?          # optionally followed by comma and parameters
        \s*\)                  # closing ) with optional whitespace
        """,
        re.VERBOSE | re.DOTALL,
    )

    # Walk through all Vue, TypeScript, and JavaScript files
    for ext in ["*.vue", "*.ts", "*.tsx", "*.js", "*.jsx"]:
        for file_path in search_dir.rglob(ext):
            # Skip the i18n locale files themselves and i18n type definitions
            if "i18n/" in str(file_path):
                continue

            try:
                with open(file_path, "r", encoding="utf-8") as f:
                    content = f.read()

                # Remove both single-line and multi-line comments
                # Single-line comments
                content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
                # Multi-line comments
                content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)

                # Search for matches with line numbers
                for match in pattern.finditer(content):
                    # Calculate line number from position
                    line_num = content[:match.start()].count('\n') + 1

                    # Convert to relative path from search_dir
                    try:
                        rel_path = file_path.relative_to(search_dir)
                        usages.append((str(rel_path), line_num))
                    except ValueError:
                        # On Windows with different drives or path issues
                        usages.append((str(file_path), line_num))
            except Exception as e:
                print(f"Warning: Could not read {file_path}: {e}")

    return usages


def extract_label_keys_from_file(file_path: Path, search_dir: Path) -> Dict[str, List[Tuple[str, int]]]:
    """
    Extract i18n keys from labelKey properties in configuration files.
    This handles special cases like useRuleOptions.ts where keys are stored
    in labelKey properties instead of direct t() calls.
    Returns dict mapping key to list of (file_path, line_number) tuples.
    """
    label_keys = {}

    if not file_path.exists():
        return label_keys

    try:
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        # Remove comments to avoid false matches
        content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
        content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)

        # Pattern to match labelKey: 'key' or labelKey: "key"
        # Matches keys with valid i18n characters: a-zA-Z0-9._-
        pattern = re.compile(
            r"""labelKey\s*:\s*           # labelKey: with optional whitespace
            ['"`]([a-zA-Z0-9._\-]+)['"`]  # the key in quotes (valid i18n chars only)
            """,
            re.VERBOSE,
        )

        for match in pattern.finditer(content):
            key = match.group(1)
            line_num = content[:match.start()].count('\n') + 1

            if key not in label_keys:
                label_keys[key] = []

            # Convert to relative path from search_dir
            try:
                rel_path = file_path.relative_to(search_dir)
                label_keys[key].append((str(rel_path), line_num))
            except ValueError:
                label_keys[key].append((str(file_path), line_num))

    except Exception as e:
        print(f"Warning: Could not read {file_path}: {e}")

    return label_keys


def find_dynamic_i18n_patterns(search_dir: Path) -> List[Tuple[str, int, str]]:
    """
    Find dynamic i18n key patterns that cannot be statically analyzed.
    Returns list of (file_path, line_number, pattern) tuples.
    Detects:
    - Template literals: t(`prefix.${variable}`)
    - String concatenation: t('prefix.' + variable)
    """
    dynamic_patterns = []

    # Pattern to match t() calls with template literals (backticks)
    # Matches: t(`prefix.${variable}`) or t(`prefix.${variable}.suffix`)
    template_pattern = re.compile(
        r"""(?<!\w)t\(\s*                    # t( with optional whitespace
        `                                  # opening backtick
        ([^`]*?\$\{[^}]+\}[^`]*?)          # template content with ${variable}
        `                                  # closing backtick
        \s*\)                              # closing ) with optional whitespace
        """,
        re.VERBOSE | re.DOTALL,
    )

    # Pattern to match t() calls with string concatenation
    # Matches: t('prefix.' + variable) or t("prefix." + variable)
    concat_pattern = re.compile(
        r"""(?<!\w)t\(\s*                    # t( with optional whitespace
        ['"`]([a-zA-Z0-9._\-]+)['"`]\s*\+  # string part followed by +
        \s*(?:[a-zA-Z0-9._\-]+\s*\+)*      # optional additional concatenations
        \s*[a-zA-Z0-9._\-]+                # variable part
        \s*\)                              # closing ) with optional whitespace
        """,
        re.VERBOSE,
    )

    # Search in all TypeScript, JavaScript, and Vue files
    for ext in ["*.vue", "*.ts", "*.tsx", "*.js", "*.jsx"]:
        for file_path in search_dir.rglob(ext):
            # Skip the i18n locale files and types files
            if "i18n/" in str(file_path):
                continue

            try:
                with open(file_path, "r", encoding="utf-8") as f:
                    content = f.read()

                # Remove both single-line and multi-line comments
                content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
                content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)

                # Search for template literal patterns
                for match in template_pattern.finditer(content):
                    line_num = content[:match.start()].count('\n') + 1
                    pattern = match.group(1)

                    try:
                        rel_path = file_path.relative_to(search_dir)
                        dynamic_patterns.append((str(rel_path), line_num, pattern))
                    except ValueError:
                        dynamic_patterns.append((str(file_path), line_num, pattern))

                # Search for string concatenation patterns
                for match in concat_pattern.finditer(content):
                    line_num = content[:match.start()].count('\n') + 1
                    pattern = match.group(1) + " + ..."

                    try:
                        rel_path = file_path.relative_to(search_dir)
                        dynamic_patterns.append((str(rel_path), line_num, pattern))
                    except ValueError:
                        dynamic_patterns.append((str(file_path), line_num, pattern))

            except Exception as e:
                print(f"Warning: Could not read {file_path}: {e}")

    return dynamic_patterns


def find_all_i18n_calls(search_dir: Path) -> Dict[str, List[Tuple[str, int]]]:
    """
    Find ALL t() calls in the codebase and extract the keys.
    Returns dict mapping key to list of (file_path, line_number) tuples.
    IMPORTANT: i18n keys MUST be called with t() function - this is the only valid pattern.
    Searches .vue, .ts, .tsx, .js, .jsx files.

    Valid i18n key characters: letters, numbers, dots, underscores, and hyphens only.
    This prevents false matches from function call patterns like 'emit(', 'closest(', etc.
    Handles multi-line t() calls (e.g., after formatting).
    """
    all_calls = {}

    # Pattern to match t('key') or t("key") with optional whitespace and parameters
    # Only matches valid i18n key characters: a-zA-Z0-9._-
    # Uses (?<!\w) to ensure 't' is not part of another identifier (e.g., createElement)
    # This prevents matching invalid patterns like 'createElement(', 'closest(', etc.
    pattern = re.compile(r"""(?<!\w)t\(\s*              # t( with optional whitespace, t must be a standalone identifier
        ['"`]([a-zA-Z0-9._\-]+)['"`]                # the key in quotes (valid i18n chars only)
        \s*(?:,.+?)?                                # optionally followed by comma and parameters
        \s*\)                                      # closing ) with optional whitespace
        """, re.VERBOSE | re.DOTALL)

    # Search in all TypeScript, JavaScript, and Vue files
    for ext in ["*.vue", "*.ts", "*.tsx", "*.js", "*.jsx"]:
        for file_path in search_dir.rglob(ext):
            # Skip the i18n locale files and types files
            if "i18n/" in str(file_path):
                continue

            try:
                with open(file_path, "r", encoding="utf-8") as f:
                    content = f.read()

                # Remove both single-line and multi-line comments
                # Single-line comments
                content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)
                # Multi-line comments
                content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)

                # Search for matches
                for match in pattern.finditer(content):
                    key = match.group(1)
                    # Calculate line number from position
                    line_num = content[:match.start()].count('\n') + 1

                    if key not in all_calls:
                        all_calls[key] = []

                    # Convert to relative path from search_dir
                    try:
                        rel_path = file_path.relative_to(search_dir)
                        all_calls[key].append((str(rel_path), line_num))
                    except ValueError:
                        all_calls[key].append((str(file_path), line_num))
            except Exception as e:
                print(f"Warning: Could not read {file_path}: {e}")

    # Special handling: Extract labelKey properties from known files
    # These files define i18n keys in labelKey properties that are used dynamically
    special_files = [
        "composables/rules/useRuleOptions.ts",
        "composables/filter/useFilterFields.ts",
    ]

    for special_file in special_files:
        file_path = search_dir / special_file
        if file_path.exists():
            label_keys = extract_label_keys_from_file(file_path, search_dir)
            # Merge label_keys into all_calls
            for key, locations in label_keys.items():
                if key not in all_calls:
                    all_calls[key] = []
                all_calls[key].extend(locations)

    return all_calls


def create_file_link(file_path: str, line_num: int, is_windows: bool = False) -> str:
    """
    Create a clickable markdown link for a file location.
    The report is in tools/ folder, so we need relative path from tools/ to frontend/src/
    """
    # Convert backslashes to forward slashes for URLs on Windows
    if is_windows:
        file_path = file_path.replace("\\", "/")

    # The report is in tools/ folder, files are in frontend/src/
    # So we need to go up one level (../) then into frontend/src/
    relative_path = f"../../frontend/src/{file_path}"
    display_text = f"{file_path}:{line_num}"

    # Create a clickable markdown link
    # Using relative path from tools/ to the file
    return f"[{display_text}]({relative_path}#L{line_num})"


def create_usage_links(usages: List[Tuple[str, int]], is_windows: bool = False) -> str:
    """Create comma-separated clickable links for usages."""
    if not usages:
        return ""

    links = []
    for file_path, line_num in sorted(usages)[:10]:  # Limit to 10 usages to avoid huge output
        # Shorten path by removing common prefix
        short_path = file_path.replace("frontend/src/", "")
        links.append(create_file_link(short_path, line_num, is_windows))

    result = ", ".join(links)
    if len(usages) > 10:
        result += f" ... ({len(usages) - 10} more)"

    return result


def get_nesting_depth(key: str) -> int:
    """Get the nesting depth of a key (number of dots + 1)."""
    return key.count(".") + 1


def is_top_level_key(key: str) -> bool:
    """Check if a key is a top-level key (not in any category)."""
    # Top-level keys have no dots
    return "." not in key


def main():
    # Determine if we're on Windows
    is_windows = os.name == "nt"

    # Change to project root directory (tools/ parent)
    script_dir = Path(__file__).parent
    project_root = script_dir.parent.parent
    os.chdir(project_root)

    print(f"Analyzing i18n usage in {project_root}")
    print("=" * 80)

    # Extract keys from locale files with category information
    en_keys, en_categories = extract_keys_from_locale(EN_FILE)
    zh_keys, zh_categories = extract_keys_from_locale(ZH_FILE)
    types_keys = extract_keys_from_types(TYPES_FILE)

    # Combine all categories from both locale files
    all_categories = en_categories | zh_categories

    # Get all unique keys, excluding all categories
    all_keys = sorted(set(en_keys.keys()) | set(zh_keys.keys()) | set(types_keys.keys()))
    filtered_keys = [key for key in all_keys if key not in all_categories]

    # Find ALL i18n calls in the codebase
    all_calls = find_all_i18n_calls(FRONTEND_DIR)

    # Find dynamic i18n patterns that cannot be statically analyzed
    dynamic_patterns = find_dynamic_i18n_patterns(FRONTEND_DIR)

    # Find missing keys (used but not defined)
    missing_keys = set(all_calls.keys()) - set(all_keys)

    # Find unused keys (defined but never used)
    # Build a set of keys used in special files (labelKey properties)
    special_files = [
        "composables/rules/useRuleOptions.ts",
        "composables/filter/useFilterFields.ts",
    ]
    label_key_usage = set()
    for special_file in special_files:
        file_path = FRONTEND_DIR / special_file
        if file_path.exists():
            label_keys = extract_label_keys_from_file(file_path, FRONTEND_DIR)
            label_key_usage.update(label_keys.keys())

    unused_keys = []
    for key in filtered_keys:
        usages = find_i18n_usage(key, FRONTEND_DIR)
        # Also check if key is used in labelKey properties
        used_in_label_key = key in label_key_usage
        if len(usages) == 0 and not used_in_label_key:
            unused_keys.append(key)

    # Check for inconsistencies between en.ts and zh.ts
    # Keys that exist in en.ts but missing in zh.ts
    missing_in_zh = set(en_keys.keys()) - set(zh_keys.keys())
    # Keys that exist in zh.ts but missing in en.ts
    missing_in_en = set(zh_keys.keys()) - set(en_keys.keys())

    # Filter out categories from inconsistency checks
    missing_in_zh_filtered = [k for k in sorted(missing_in_zh) if k not in all_categories]
    missing_in_en_filtered = [k for k in sorted(missing_in_en) if k not in all_categories]

    # Group keys by nesting depth
    keys_by_depth: Dict[int, List[str]] = {}
    for key in filtered_keys:
        depth = get_nesting_depth(key)
        if depth not in keys_by_depth:
            keys_by_depth[depth] = []
        keys_by_depth[depth].append(key)

    safe_print(f"Found {len(all_keys)} unique i18n keys")
    safe_print(f"  - en.ts: {len(en_keys)} keys")
    safe_print(f"  - zh.ts: {len(zh_keys)} keys")
    safe_print(f"  - types.ts: {len(types_keys)} keys")
    safe_print(f"  - Category keys (excluded from report): {len(all_categories)}")
    safe_print(f"  - Leaf keys with values (included in report): {len(filtered_keys)}")
    safe_print(f"  - Used keys found in code: {len(all_calls)}")
    safe_print(f"  - Dynamic i18n patterns (cannot be statically analyzed): {len(dynamic_patterns)} ⚠️")
    safe_print(f"  - Missing keys (used but not defined): {len(missing_keys)} ⚠️")
    safe_print(f"  - Unused keys (defined but never used): {len(unused_keys)} ⚠️")
    safe_print(f"\nLocale file inconsistencies:")
    safe_print(f"  - Missing in zh.ts (exists in en.ts): {len(missing_in_zh_filtered)} ⚠️")
    safe_print(f"  - Missing in en.ts (exists in zh.ts): {len(missing_in_en_filtered)} ⚠️")
    safe_print(f"\nNesting depth distribution:")
    for depth in sorted(keys_by_depth.keys()):
        safe_print(f"  - Level {depth}: {len(keys_by_depth[depth])} keys")
    safe_print("=" * 80)
    safe_print("")

    # Generate markdown report
    output_lines = []
    output_lines.append("# i18n Usage Analysis Report")
    output_lines.append("")

    # Section 0: Locale File Inconsistencies
    if missing_in_zh_filtered or missing_in_en_filtered:
        output_lines.append("## ⚠️ Locale File Inconsistencies (en.ts vs zh.ts)")
        output_lines.append("")
        output_lines.append("This section shows keys that are missing in one of the locale files.")
        output_lines.append("")

        # Keys missing in zh.ts
        if missing_in_zh_filtered:
            output_lines.append("### Keys Missing in zh.ts (Exist in en.ts)")
            output_lines.append("")
            output_lines.append("| Key | en.ts Line |")
            output_lines.append("|-----|------------|")
            for key in missing_in_zh_filtered:
                output_lines.append(f"| {key} | {en_keys[key]} |")
            output_lines.append("")

        # Keys missing in en.ts
        if missing_in_en_filtered:
            output_lines.append("### Keys Missing in en.ts (Exist in zh.ts)")
            output_lines.append("")
            output_lines.append("| Key | zh.ts Line |")
            output_lines.append("|-----|------------|")
            for key in missing_in_en_filtered:
                output_lines.append(f"| {key} | {zh_keys[key]} |")
            output_lines.append("")

        output_lines.append("---")
        output_lines.append("")

    # Section 1: Dynamic i18n Patterns (cannot be statically analyzed)
    if dynamic_patterns:
        output_lines.append("## ⚠️ Dynamic i18n Patterns (Cannot be Statically Analyzed)")
        output_lines.append("")
        output_lines.append("This section shows i18n keys that are built dynamically at runtime.")
        output_lines.append("These patterns use template literals or string concatenation and cannot be")
        output_lines.append("automatically verified against the locale files. Manual review is required.")
        output_lines.append("")
        output_lines.append("| File | Line | Pattern |")
        output_lines.append("|------|------|----------|")

        for file_path, line_num, pattern in sorted(dynamic_patterns):
            short_path = file_path.replace("frontend/src/", "")
            link = create_file_link(short_path, line_num, is_windows)
            # Escape special markdown characters in pattern
            safe_pattern = pattern.replace("|", "\\|").replace("`", "\\`")
            output_lines.append(f"| {link} | {line_num} | `{safe_pattern}` |")

        output_lines.append("")
        output_lines.append("---")
        output_lines.append("")

    # Section 2: Missing Keys (used but not defined)
    if missing_keys:
        output_lines.append("## ⚠️ Missing Keys (Used in Code but Not Defined)")
        output_lines.append("")
        output_lines.append("| Key | Usage Count | Locations |")
        output_lines.append("|-----|-------------|-----------|")

        for key in sorted(missing_keys):
            usages = all_calls[key]
            usage_count = len(usages)
            usage_links = create_usage_links(usages, is_windows)
            output_lines.append(f"| {key} | {usage_count} | {usage_links} |")

        output_lines.append("")
        output_lines.append("---")
        output_lines.append("")

    # Section 2: Unused Keys (defined but never used)
    if unused_keys:
        output_lines.append("## ⚠️ Unused Keys (Defined but Never Used)")
        output_lines.append("")
        output_lines.append("| Key | en.ts | zh.ts | types.ts |")
        output_lines.append("|-----|-------|-------|----------|")

        for key in sorted(unused_keys):
            en_line = en_keys.get(key, "")
            zh_line = zh_keys.get(key, "")
            types_line = types_keys.get(key, "")
            output_lines.append(f"| {key} | {en_line} | {zh_line} | {types_line} |")

        output_lines.append("")
        output_lines.append("---")
        output_lines.append("")

    # Section 3: All Keys Analysis
    output_lines.append("## All Keys Analysis")
    output_lines.append("")
    output_lines.append("| Key | Depth | en.ts | zh.ts | types.ts | Usage Count | Locations |")
    output_lines.append("|-----|-------|-------|-------|----------|-------------|-----------|")

    for key in filtered_keys:
        en_line = en_keys.get(key, "")
        zh_line = zh_keys.get(key, "")
        types_line = types_keys.get(key, "")
        depth = get_nesting_depth(key)

        # Bold format for top-level keys (no dots)
        display_key = f"**{key}**" if is_top_level_key(key) else key

        # Find usages in the codebase
        usages = find_i18n_usage(key, FRONTEND_DIR)
        usage_count = len(usages)

        # Create clickable links
        usage_links = create_usage_links(usages, is_windows)

        output_lines.append(
            f"| {display_key} | {depth} | {en_line} | {zh_line} | {types_line} | {usage_count} | {usage_links} |"
        )

    # Write to output file
    output_file = script_dir / "i18n_usage_report.md"
    with open(output_file, "w", encoding="utf-8") as f:
        f.write("\n".join(output_lines))

    print(f"Report generated: {output_file}")
    print(f"Total keys analyzed: {len(filtered_keys)}")

    # Print warnings
    if missing_keys:
        safe_print(f"\n⚠️  WARNING: {len(missing_keys)} keys are used but not defined!")
        safe_print("   Check the 'Missing Keys' section in the report for details.")
    if unused_keys:
        safe_print(f"\n⚠️  WARNING: {len(unused_keys)} keys are defined but never used!")
        safe_print("   Check the 'Unused Keys' section in the report for details.")


if __name__ == "__main__":
    main()
