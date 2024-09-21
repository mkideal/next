(function (Prism) {
  const delimiter = "prolog";
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
        // Trim operators - at start or end
        trim: {
          pattern: /^-|-$/,
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
          /\b(?:if|else|range|with|end|block|define|template|and|or|not|render)\b/,
        // Functions within template actions
        function: /\b\w+(?=\s|\()/,
        // Variables within template actions
        variable: /\.\w+(?:\.\w+)*/,
        // Numbers within template actions
        number: /\b\d+(\.\d+)?\b/,
        // Operators within template actions, including pipe '|'
        operator: /[|=!<>]=?|\|\|/,
        // Punctuation within template actions
        punctuation: /[(),.-]/,
      },
    },
  };
})(Prism);
