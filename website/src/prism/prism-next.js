(function (Prism) {
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
    keyword: /\b(?:package|import|const|enum|struct|interface)\b/,
    boolean: /\b(?:true|false)\b/,
    number: /\b0x[\da-f]+\b|(?:\b\d+(?:\.\d*)?|\B\.\d+)(?:e[+-]?\d+)?/i,
    operator: /[<>]=?|[!=]=?=?|--?|\+\+?|&&?|\|\|?|[?*/~^%]/,
    punctuation: /[{}[\];(),.:]/,
    annotation: {
      pattern: /\B@\w+(?:\([^)]*\))?/,
      alias: "symbol",
    },
    type: {
      pattern:
        /\b(?:int(?:8|16|32|64)?|float(?:32|64)|bool|string|byte|bytes|any|map|vector|array)\b/,
      alias: "class-name",
    },
    function: {
      pattern: /(?<!\B@)\b[a-zA-Z_]\w*(?=\s*\()/,
      greedy: true,
    },
    builtin: {
      pattern:
        /\b(?:float|sprint|sprintf|sprintln|print|printf|assert(?:_eq|_ne|_lt|_le|_gt|_ge)?)\b/,
      alias: "keyword",
    },
  };

  Prism.languages.insertBefore("next", "keyword", {
    "const-declaration": {
      pattern: /\bconst\s+\w+\s*=/,
      inside: {
        keyword: /\bconst\b/,
        "constant-name": {
          pattern: /\b\w+\b(?=\s*=)/,
          alias: "constant",
        },
        operator: /=/,
      },
    },
    "enum-declaration": {
      pattern: /\benum\s+\w+\s*\{[\s\S]*?\}/m,
      inside: {
        keyword: /\benum\b/,
        "enum-name": {
          pattern: /\b\w+\b(?=\s*\{)/,
          alias: "class-name",
        },
        "enum-member": {
          pattern: /\b\w+\b(?=\s*(?:=|,|;))/,
          alias: "constant",
        },
      },
    },
    "struct-declaration": {
      pattern: /(\bstruct\s+)[a-zA-Z_][a-zA-Z0-9_]*/,
      lookbehind: true,
      alias: "class-name",
    },
    "interface-declaration": {
      pattern: /(\binterface\s+)[a-zA-Z_][a-zA-Z0-9_]*/,
      lookbehind: true,
      alias: "class-name",
    },
  });
})(Prism);
