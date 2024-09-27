(function (Prism) {
  Prism.languages.map = {
    comment: {
      pattern: /^\s*#.*/m,
      greedy: true,
    },
    keyword: {
      pattern: /\bbox\b/,
      greedy: true,
    },
    replacer: {
      pattern: /%[^%\r\n]*%/,
      greedy: true,
      alias: "selector",
    },
    punctuation: /[=()<>[\]]/,
    string: {
      pattern: /(["])(?:\\(?:\r\n|[\s\S])|(?!\1)[^\\\r\n])*\1/,
      greedy: true,
    },
    parameter: {
      pattern: /(?<=box\()[^)]+(?=\))/,
      greedy: true,
    },
    property: [
      {
        pattern: /^\s*\w+(?=\s*=)/m,
        greedy: true,
      },
    ],
  };
})(Prism);
