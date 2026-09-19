#!/usr/bin/env python3
"""China government latest policies: public JSON -> RSS; standard library only."""

import json
import sys
from datetime import datetime, timedelta, timezone
from email.utils import format_datetime
from urllib.parse import urljoin, urlsplit
from urllib.request import Request, urlopen
from xml.etree import ElementTree as ET

PAGE_URL = "https://www.gov.cn/zhengce/zuixin/"
DATA_URL = urljoin(PAGE_URL, "ZUIXINZHENGCE.json")
MAX_BYTES = 4 * 1024 * 1024


def generate_rss(records):
    if not isinstance(records, list):
        raise ValueError("Expected an array of policy articles")
    root = ET.Element("rss", version="2.0")
    channel = ET.SubElement(root, "channel")
    ET.SubElement(channel, "title").text = "中国政府网 · 最新政策"
    ET.SubElement(channel, "link").text = PAGE_URL
    ET.SubElement(channel, "description").text = "中国政府网最新政策，来自公开列表数据。"
    seen = set()
    for record in records[:100]:
        if not isinstance(record, dict):
            continue
        title, link = record.get("TITLE"), record.get("URL")
        if not isinstance(title, str) or not title.strip() or not isinstance(link, str):
            continue
        link = urljoin(PAGE_URL, link)
        if urlsplit(link).scheme not in ("https", "http") or link in seen:
            continue
        seen.add(link)
        item = ET.SubElement(channel, "item")
        ET.SubElement(item, "title").text = title.strip()
        ET.SubElement(item, "link").text = link
        ET.SubElement(item, "guid", isPermaLink="true").text = link
        # Leave body empty: MrRSS can fetch full text from the article URL.
        try:
            date = datetime.strptime(record.get("DOCRELPUBTIME", ""), "%Y-%m-%d")
            date = date.replace(tzinfo=timezone(timedelta(hours=8)))
            ET.SubElement(item, "pubDate").text = format_datetime(date)
        except (ValueError, TypeError):
            pass
    if not seen:
        raise ValueError("No policy articles found; the source format may have changed")
    return ET.tostring(root, encoding="utf-8", xml_declaration=True)


def main():
    request = Request(DATA_URL, headers={"User-Agent": "MrRSS policy feed", "Accept": "application/json"})
    with urlopen(request, timeout=20) as response:
        data = response.read(MAX_BYTES + 1)
    if len(data) > MAX_BYTES:
        raise ValueError("Policy response exceeds size limit")
    # Binary stdout preserves UTF-8 on Windows regardless of console code page.
    sys.stdout.buffer.write(generate_rss(json.loads(data.decode("utf-8-sig"))))


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"Policy feed failed: {error}", file=sys.stderr)
        sys.exit(1)
