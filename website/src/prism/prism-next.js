(function (Prism) {
  const identifier = /[a-zA-Z_]\w*/;

  Prism.languages.next = {
    comment: [
      {
        pattern: /\/\/.*$/m,
        greedy: true,
      },
      {
        pattern: /\/\*[\s\S]*?(?:\*\/|$)/,
        greedy: true,
      },
    ],
    string: {
      pattern: /"(?:\\.|[^\\\"\r\n])*"/,
      greedy: true,
    },
    keyword: /\b(?:package|iota|const|enum|struct|interface)\b/,
    boolean: /\b(?:true|false)\b/,
    number: /\b0x[\da-f]+\b|(?:\b\d+(?:\.\d*)?|\B\.\d+)(?:e[+-]?\d+)?/i,
    operator: /[<>]=?|[!=]=?=?|--?|\+\+?|&&?|\|\|?|[?*/~^%]/,
    punctuation: /[{}[\];(),.:]/,
    annotation: {
      pattern: RegExp(`(@)${identifier.source}`),
      alias: "symbol",
    },
    type: {
      pattern:
        /\b(?:int(?:8|16|32|64)?|float(?:32|64)|bool|string|byte|bytes|any|map|vector|array)\b/,
      alias: "class-name",
    },
    function: {
      pattern: RegExp(`(?<!\B@)${identifier.source}(?=\\s*\\()`),
      greedy: true,
    },
    builtin: {
      pattern:
        /\b(?:float|sprint|sprintf|sprintln|print|printf|assert(?:_eq|_ne|_lt|_le|_gt|_ge)?)\b/,
      alias: "keyword",
    },
  };

  Prism.languages.insertBefore("next", "keyword", {
    "package-declaration": {
      pattern: RegExp(`(\\bpackage\\s+)${identifier.source}`),
      lookbehind: true,
      alias: "namespace",
    },
    "const-declaration": {
      pattern: RegExp(`(\\bconst\\s+)${identifier.source}\\s*=`),
      lookbehind: true,
      alias: "constant",
    },
    "enum-declaration": {
      pattern: /\benum\s+\w+\s*\{[\s\S]*?\}/m,
      inside: {
        keyword: /\benum\b/,
        "enum-name": {
          pattern: /\b\w+\b(?=\s*\{)/,
          lookbehind: true,
          alias: "class-name",
        },
        punctuation: /[{};]/,
        "enum-member": {
          pattern: /^\s*\w+\b(?=\s*(?:=\s*(?:"[^"]*"|'[^']*'|[^,;]+)|,|;))/m,
          alias: "constant",
          greedy: true,
        },
        rest: Prism.languages.next,
      },
    },
    "struct-declaration": {
      pattern: RegExp(`(\\bstruct\\s+)${identifier.source}(?=\\s*\\{)`),
      lookbehind: true,
      alias: "class-name",
    },
    "interface-declaration": {
      pattern: RegExp(`(\\binterface\\s+)${identifier.source}(?=\\s*\\{)`),
      lookbehind: true,
      alias: "class-name",
    },
  });
})(Prism);
