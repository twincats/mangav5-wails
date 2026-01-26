export const scrapingRuleSnippets = [
  {
    label: 'scraping-rule-manga-template',
    kind: 27, // monaco.languages.CompletionItemKind.Snippet
    sortText: 'z_20_scraping',
    insertText: `{
  "site": "\${1:My Manga Site}",
  "domains": ["\${2:example.com}"],
  "strategy": "\${3|static,api,browser,auto|}",
  \${4:"entry": {
    "url": "\${5:https://example.com/manga/123}",
    "method": "GET",
    "headers": {\\}
  \\},}
  "extract": [
    {
      "name": "title",
      "type": "\${6|css,json,template,text|}",
      "selector": "h1"
    },
    {
      "name": "cover",
      "type": "\${7|css,json,template,text|}",
      "selector": "h1"
    },
    {
      "name": "chapters",
      "type": "\${8|css,json,template,text|}",
      "selector": ".page img",
      "multiple": true,
      "children": [
        {
          "name": "chapter_id",
          "type": "css",
          "selector": ".page img"
        },
        {
          "name": "chapter",
          "type": "css",
          "selector": ".page img"
        },
        {
          "name": "group_name",
          "type": "text",
          "text": "\${9:Westmanga}"
        },
        {
          "name": "language",
          "type": "text",
          "text": "\${10|en,id|}"
        },
        {
          "name": "time",
          "type": "css",
          "selector": ".page img"
        }
      ]
    }
  ]
}`,
    insertTextRules: 4, // monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet
  },
  {
    label: 'scraping-rule-chapter-template',
    kind: 27,
    sortText: 'z_20_scraping',
    insertText: `{
  "site": "\${1:My Manga Site}",
  "domains": ["\${2:example.com}"],
  "strategy": "\${3|static,api,browser,auto|}",
  \${4:"entry": {
    "url": "\${5:https://example.com/manga/123}",
    "method": "GET",
    "headers": {\\}
  \\},}
  \${6:"api": {
    "steps": [
      {
        "id": "step1",
        "request": {
          "url": "\${7:https://api.mangadex.org/manga/{id\\}}"
        \\}
      \\}
    ]
  \\},}
  "extract": [
    {
      "name": "pages",
      "type": "\${8|css,json,template,text|}",
      "multiple": true,
      \${9:"from": "step1",}
      "selector": "h1"
    }
  ]
}`,
    insertTextRules: 4,
  },
  {
    label: 'scraping-rule-static',
    kind: 27,
    sortText: 'z_20_scraping',
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
                "name": "chapter_id",
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
              "name": "group_name",
              "type": "text",
              "text": "\${6:Asmotoon}"
            },
            {
              "name": "language",
              "type": "text",
              "text": "\${7|en,id|}"
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
    sortText: 'z_20_scraping',
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
    sortText: 'z_20_scraping',
    insertText: `{
  "site": "\${1:Mangadex}",
  "domains": ["\${2:mangadex.org}"],
  "strategy": "api",
  "entry": {
    "url": "\${3:https://mangadex.org/title/829141f2-192a-4422-a9b1-2b63458e6981}",
    "regex": "/title/(?P<id>[^/]+)"
  },
  "api": {
    "steps": [
      {
        "id": "\${4:step1}",
        "request": {"url": "\${5:https://api.mangadex.org/at-home/server/{id}}"}
      }
    ]
  },
  "extract": [
    {
      "name": "pages",
      "type": "json",
      "from": "\${4:step1}",
      "path": "chapter.data",
      "multiple": true,
      "children": [
        {
          "name": "image_url",
          "type": "template",
          "template": "{baseUrl}/data/{hash}/{file}",
          "children": [
             {"name": "baseUrl", "type": "json", "from": "step1", "path": "baseUrl"},
             {"name": "hash", "type": "json", "from": "step1", "path": "chapter.hash"},
             {"name": "file", "type": "template", "template": "{_self}"}
          ]
        }
      ]
    }
  ]
}`,
    insertTextRules: 4,
  },
  // Field Rules Snippets
  {
    label: 'field-text',
    kind: 27,
    sortText: 'z_10_field',
    detail: 'Static Text Field Rule',
    documentation: 'Extracts a static text value',
    insertText: `{
  "name": "\${1:fieldName}",
  "type": "text",
  "text": "\${2:value}"
}`,
    insertTextRules: 4,
  },
  {
    label: 'field-css',
    kind: 27,
    sortText: 'z_10_field',
    detail: 'CSS Selector Field Rule',
    documentation: 'Extracts content using CSS selector',
    insertText: `{
  "name": "\${1:fieldName}",
  "type": "css",
  "selector": "\${2:selector}",
  "trim": true
}`,
    insertTextRules: 4,
  },
  {
    label: 'field-attr',
    kind: 27,
    sortText: 'z_10_field',
    detail: 'Attribute Extraction Field Rule',
    documentation: 'Extracts an attribute from an element',
    insertText: `{
  "name": "\${1:fieldName}",
  "type": "css",
  "selector": "\${2:selector}",
  "attr": ["\${3:src}"],
  "trim": true
}`,
    insertTextRules: 4,
  },
  {
    label: 'field-json',
    kind: 27,
    sortText: 'z_10_field',
    detail: 'JSON Path Field Rule',
    documentation: 'Extracts value from JSON response using GJSON path',
    insertText: `{
  "name": "\${1:fieldName}",
  "type": "json",
  "path": "\${2:data.title}"
}`,
    insertTextRules: 4,
  },
  {
    label: 'field-template',
    kind: 27,
    sortText: 'z_10_field',
    detail: 'Template Field Rule',
    documentation: 'Combines multiple values using a template',
    insertText: `{
  "name": "\${1:fieldName}",
  "type": "template",
  "template": "{{ .\${2:variable} }}",
  "trim": true
}`,
    insertTextRules: 4,
  },
  {
    label: 'field-regex',
    kind: 27,
    sortText: 'z_10_field',
    detail: 'Regex Extraction Field Rule',
    documentation: 'Extracts content using Regex capture group',
    insertText: `{
  "name": "\${1:fieldName}",
  "type": "css",
  "selector": "\${2:selector}",
  "regex": "\${3:pattern (capture)}"
}`,
    insertTextRules: 4,
  },
  {
    label: 'field-children',
    kind: 27,
    sortText: 'z_10_field',
    detail: 'Nested Children Field Rule',
    documentation: 'Extracts a list of items',
    insertText: `{
  "name": "\${1:items}",
  "type": "css",
  "selector": "\${2:.item}",
  "multiple": true,
  "children": [
    {
      "name": "\${3:subField}",
      "type": "css",
      "selector": "\${4:.sub}"
    }
  ]
}`,
    insertTextRules: 4,
  },
]
