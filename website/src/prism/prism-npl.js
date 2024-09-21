(function (Prism) {
  const delimiter = "tag";
  Prism.languages.npl = {
    // Template actions {{ ... }}
    "template-action": {
      pattern: /\{\{[\s\S]+?\}\}/,
      inside: {
        // Delimiters {{ and }}
        delimiter: {
          pattern: /^\{\{|\}\}$/,
          alias: delimiter,
        },
        // Comments within template actions
        comment: {
          pattern: /^\/\*[\s\S]*?\*\/$/,
          alias: "comment",
        },
        // Strings within template actions
        string: /"(?:\\.|[^"\\])*"/,
        // Keywords within template actions
        keyword:
          /\b(?:if|else|range|with|end|block|define|template|and|or|not)\b/,
        // Variables within template actions
        variable: /\.(\w+(?:\.\w+)*)?/,
        // Functions and variables within template actions
        function: {
          //pattern: /\b\w+(?:\.\w+)*\b/,
          pattern: /\b(?<!\.)(?!(?:\w+\.)+\w*\b)\w+(?!\.)\b/,
        },
        // Numbers within template actions
        number: /\b\d+(?:\.\d+)?\b/,
        // Operators within template actions, including pipe '|'
        operator: /\||=|:?=/,
        // Punctuation within template actions
        punctuation: /[(),]/,
      },
    },
  };
})(Prism);
