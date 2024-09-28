(function (Prism) {
  const delimiter = "tag";
  Prism.languages.npl = {
    "comment-action": {
      pattern: /\{\{\-?\s*\/\*[\s\S]*?\*\/\s*\-?\}\}/,
      alias: "comment",
      greedy: true,
    },
    // Template actions {{ ... }} or {{- ... -}}
    "template-action": {
      pattern: /\{\{[-]?[\s\S]*?[-]?\}\}/,
      inside: {
        // Delimiters {{ and }} with optional -
        delimiter: {
          pattern: /^\{\{[-]?|[-]?\}\}$/,
          alias: delimiter,
        },
        // Trim operators - at start or end
        trim: {
          pattern: /^-|-$/,
          alias: delimiter,
        },
        // Strings within template actions
        string: /"(?:\\.|[^"\\])*"/,
        // Booleans within template actions
        boolean: /\b(?:true|false)\b/,
        // Numbers within template actions (including floats)
        number: /\b\d+(?:\.\d+)?\b/,
        // Keywords within template actions
        keyword:
          /\b(?:if|else|range|with|end|block|define|template|and|or|not)\b/,
        // Variables within template actions, including $ symbol
        variable: {
          pattern: /\.(?:\w+(?:\.\w+)*)?|\$(?:\w+)?/,
          greedy: true,
        },
        // Functions and variables within template actions
        function: {
          pattern:
            /\b(?<!\.)(?!(?:\w+\.)+\w*\b)(?![0-9])(?!true\b)(?!false\b)\w+(?!\.)\b/,
        },
        // Operators within template actions, including pipe '|'
        operator: /\||=|:?=/,
        // Punctuation within template actions
        punctuation: /[(),]/,
      },
    },
  };
  Prism.languages.tmpl = Prism.languages.npl;
})(Prism);
