export const scrapingRuleSnippets = [
  {
    label: 'scraping-rule-template',
    kind: 27, // monaco.languages.CompletionItemKind.Snippet
    insertText: `{
  "site": "\${1:My Manga Site}",
  "domains": ["\${2:example.com}"],
  "strategy": "\${3:static}",
  "wait_config": {
    "container_selectors": [],
    "content_selectors": [],
    "min_text_length": 0,
    "require_image_loaded": false,
    "timeout_ms": 15000,
    "poll_ms": 500
  },
  "entry": {
    "url": "\${4:https://example.com/manga/123}",
    "method": "GET",
    "headers": {}
  },
  "extract": [
    {
      "name": "title",
      "type": "css",
      "selector": "h1",
      "trim": true
    },
    {
      "name": "pages",
      "type": "css",
      "selector": ".page img",
      "attr": ["src"],
      "multiple": true
    }
  ]
}`,
    insertTextRules: 4, // monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
  },
  {
    label: 'scraping-rule-static',
    kind: 27,
    insertText: `{
  "site": "\${1:Asmotoon}",
  "domains": ["\${2:asmotoon.com}"],
  "strategy": "static",
  "entry": {
    "url": "\${3:https://asmotoon.com/chapter/6478dd70f1a-6478e2d062d/}"
  },
  "extract": [
    {
      "name": "title",
      "type": "css",
      "selector": "\${4:h1}",
      "trim": true
    },
     {
        "name": "chapters",
        "multiple": true,
        "type": "css",
        "selector": "\${5:#chapters a}",
        "children": [
            {
                "name": "url",
                "type": "css",
                "attr": ["href"],
                "regex": "/chapter/([^/]+)"
            },
            {
                "name": "chapter",
                "type": "css",
                "selector": ".text-sm.truncate",
                "trim": true
            },
            {
                "name": "time",
                "type": "css",
                "selector": ".text-xs.text-white\\/50",
                "trim": true
            }
        ]
    }
  ]
}`,
    insertTextRules: 4,
  },
  {
    label: 'scraping-rule-browser',
    kind: 27,
    insertText: `{
  "site": "\${1:Westmanga}",
  "domains": ["\${2:westmanga.me}"],
  "strategy": "browser",
  "entry": {
    "url": "\${3:https://westmanga.me/comic/{id}}"
  },
  "wait_config": {
    "content_selectors": [
      "div.grid > div[data-slot='card']"
    ]
  },
  "extract": [
    {
      "name": "title",
      "type": "css",
      "selector": "\${4:div[data-slot='card-title']}"
    },
    {
      "name": "chapters",
      "type": "css",
      "selector": "\${5:div.grid > div[data-slot='card']}",
      "multiple": true,
      "children": [
        {
          "name": "chapter",
          "type": "css",
          "selector": "\${5:a > p:first-child}"
        },
        {
          "name": "time",
          "type": "css",
          "selector": "\${6:a > p:last-child}"
        }
      ]
    }
  ]
}`,
    insertTextRules: 4,
  },
  {
    label: 'scraping-rule-api',
    kind: 27,
    insertText: `{
  "site": "\${1:Mangadex}",
  "domains": [\${2:"api.mangadex.org","mangadex.org"}],
  "strategy": "api",
  "api": {
    "steps": [
      {
        "id": "\${3:step1}",
        "request": {"url": "\${4:http://api.mangadex.org/manga/{id}}"}
      }
    ]
  },
  "extract": [
    {
      "name": "title",
      "type": "json",
      "from": "\${3:step1}",
      "path": "\$5{$.data.attributes.title['ja-ro']}"
    }
  ]
}`,
    insertTextRules: 4,
  }
]
