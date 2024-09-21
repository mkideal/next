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
      lookbehind: true,
    },

    type: {
      pattern:
        /\b(?:int(?:8|16|32|64)?|float(?:32|64)|bool|string|byte|bytes|any|map|vector|array)\b/,
      alias: "keyword",
    },

    function: /\b[a-z_]\w*(?=\s*\()/i,

    constant: {
      pattern: /\b[A-Z][A-Z_\d]+\b/,
      alias: "property",
    },

    builtin: {
      pattern:
        /\b(?:iotafloat|sprint|sprintf|sprintln|print|printf|assert(?:_eq|_ne|_lt|_le|_gt|_ge)?)\b/,
      alias: "keyword",
    },
  };
})(Prism);
