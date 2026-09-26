---
name: searching-query
description: Use when you need to search something on the web
---
# Effective web searching

Effective web searching requires clear keywords, specific operators, and alternative tools.

Request only one search query per tool call!

## Search Tips

Use 2 to 4 keywords instead of long sentences. Add operators like `site:`
to look inside a specific web address. Remove extra words if you get too many results.

## Reliable web sources

Prefer to use information from these sites:

- github.com
- gitlab.com
- medium.com
- habr.com
- redis.io
- go.dev
- postgresql.org
- apache.org
- citusdata.com
- clickhouse.com
- gohugo.io
- kubernetes.io
- helm.sh
- python.org

## Formatting query

Wrong:

```txt
site:github.com/chromedp/chromedp query text
```

Site must contain only domain, that's why correct version:

```txt
site:github.com chromedp/chromedp query text
```

## Get url page content

You MUST use `browser markdown` tool for fetching web pages content in markdown format. Set maxBytes 128000.
