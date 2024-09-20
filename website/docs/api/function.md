# Function

## _ {#user-content-_}

`_` is a no-op function that returns an empty string. It's useful to place a newline in the template. 
Example: 

```
{{- if .Ok}}
{{printf "ok: %v" .Ok}}
{{_}}
{{- end}}
```

## Convert {#user-content-Convert}
### bool {#user-content-Convert_bool}

`bool` converts a value to a boolean. 
Example: 
```
{{bool 1}}
{{bool "false"}}
```

Output: 
```
true
false
```

### float {#user-content-Convert_float}

`float` converts a value to a float. 
Example: 
```
{{float "3.14"}}
{{float 42}}
```

Output: 
```
3.14
42
```

### int {#user-content-Convert_int}

`int` converts a value to an integer. 
Example: 
```
{{int "42"}}
{{int 3.14}}
```

Output: 
```
42
3
```

### string {#user-content-Convert_string}

`string` converts a value to a string. 
Example: 
```
{{string 42}}
{{string true}}
```

Output: 
```
42
true
```

## Date {#user-content-Date}
### now {#user-content-Date_now}

`now` returns the current time. 
Example: 
```
{{now}}
```

Output: 
```
2024-09-12 15:04:05.999999999 +0000 UTC
```

### parseTime {#user-content-Date_parseTime}

`parseTime` parses a time string using the specified layout. 
- **Parameters**: (_layout_: string, _value_: string)

Example: 
```
{{parseTime "2006-01-02" "2024-09-12"}}
```

Output: 
```
2024-09-12 00:00:00 +0000 UTC
```

## Encoding {#user-content-Encoding}
### b64dec {#user-content-Encoding_b64dec}

`b64dec` decodes a base64 encoded string. 
Example: 
```
{{b64dec "SGVsbG8sIFdvcmxkIQ=="}}
```

Output: 
```
Hello, World!
```

### b64enc {#user-content-Encoding_b64enc}

`b64enc` encodes a string to base64. 
Example: 
```
{{b64enc "Hello, World!"}}
```

Output: 
```
SGVsbG8sIFdvcmxkIQ==
```

## List {#user-content-List}
### first {#user-content-List_first}

`first` returns the first element of a list or string. 
Example: 
```
{{first (list 1 2 3)}}
{{first "hello"}}
```

Output: 
```
1
h
```

### includes {#user-content-List_includes}

`includes` checks if an item is present in a list, map, or string. 
- **Parameters**: (_item_: any, _collection_: slice | map | string)
- **Returns**: bool

Example: 
```
{{includes 2 (list 1 2 3)}}
{{includes "world" "hello world"}}
```

Output: 
```
true
true
```

### last {#user-content-List_last}

`last` returns the last element of a list or string. 
Example: 
```
{{last (list 1 2 3)}}
{{last "hello"}}
```

Output: 
```
3
o
```

### list {#user-content-List_list}

`list` creates a list from the given arguments. 
Example: 
```
{{list 1 2 3}}
```

Output: 
```
[1 2 3]
```

### map {#user-content-List_map}

`map` maps a list of values using the given function and returns a list of results. 
- **Parameters**: (_fn_: function, _list_: slice)

Example: 
```
{{list 1 2 3 | map (add 1)}}
{{list "a" "b" "c" | map (upper | replace "A" "X")}}
```

Output: 
```
[2 3 4]
[X B C]
```

### reverse {#user-content-List_reverse}

`reverse` reverses a list or string. 
Example: 
```
{{reverse (list 1 2 3)}}
{{reverse "hello"}}
```

Output: 
```
[3 2 1]
olleh
```

### sort {#user-content-List_sort}

`sort` sorts a list of numbers or strings. 
Example: 
```
{{sort (list 3 1 4 1 5 9)}}
{{sort (list "banana" "apple" "cherry")}}
```

Output: 
```
[1 1 3 4 5 9]
[apple banana cherry]
```

### uniq {#user-content-List_uniq}

`uniq` removes duplicate elements from a list. 
Example: 
```
{{uniq (list 1 2 2 3 3 3)}}
```

Output: 
```
[1 2 3]
```

## Math {#user-content-Math}
### add {#user-content-Math_add}

`add` adds two numbers. 
- **Parameters**: (_a_: number, _b_: number)

Example: 
```
{{add 2 3}}
```

Output: 
```
5
```

### ceil {#user-content-Math_ceil}

`ceil` returns the least integer value greater than or equal to the input. 
Example: 
```
{{ceil 3.14}}
```

Output: 
```
4
```

### floor {#user-content-Math_floor}

`floor` returns the greatest integer value less than or equal to the input. 
Example: 
```
{{floor 3.14}}
```

Output: 
```
3
```

### max {#user-content-Math_max}

`max` returns the maximum of a list of numbers. 
- **Parameters**: numbers (variadic)

Example: 
```
{{max 3 1 4 1 5 9}}
```

Output: 
```
9
```

### min {#user-content-Math_min}

`min` returns the minimum of a list of numbers. 
- **Parameters**: numbers (variadic)

Example: 
```
{{min 3 1 4 1 5 9}}
```

Output: 
```
1
```

### mod {#user-content-Math_mod}

`mod` returns the modulus of dividing the first number by the second. 
- **Parameters**: (_a_: number, _b_: number)

Example: 
```
{{mod -7 3}}
```

Output: 
```
2
```

### mul {#user-content-Math_mul}

`mul` multiplies two numbers. 
- **Parameters**: (_a_: number, _b_: number)

Example: 
```
{{mul 2 3}}
```

Output: 
```
6
```

### quo {#user-content-Math_quo}

`quo` divides the first number by the second. 
- **Parameters**: (_a_: number, _b_: number)

Example: 
```
{{quo 6 3}}
```

Output: 
```
2
```

### rem {#user-content-Math_rem}

`rem` returns the remainder of dividing the first number by the second. 
- **Parameters**: (_a_: number, _b_: number)

Example: 
```
{{rem 7 3}}
```

Output: 
```
1
```

### round {#user-content-Math_round}

`round` rounds a number to a specified number of decimal places. 
- **Parameters**: (_precision_: integer, _value_: number)

Example: 
```
{{round 2 3.14159}}
```

Output: 
```
3.14
```

### sub {#user-content-Math_sub}

`sub` subtracts the second number from the first. 
- **Parameters**: (_a_: number, _b_: number)

Example: 
```
{{sub 5 3}}
```

Output: 
```
2
```

## OS {#user-content-OS}
### absPath {#user-content-OS_absPath}

`absPath` returns the absolute path of a file or directory. 
Example: 
```
{{absPath "file.txt"}}
```
Output: 
```
/path/to/file.txt
```

### basename {#user-content-OS_basename}

`basename` returns the last element of a path. 
Example: 
```
{{basename "path/to/file.txt"}}
```
Output: 
```
file.txt
```

### cleanPath {#user-content-OS_cleanPath}

`cleanPath` returns the cleaned path. 
Example: 
```
{{cleanPath "path/to/../file.txt"}}
```
Output: 
```
path/file.txt
```

### dirname {#user-content-OS_dirname}

`dirname` returns the directory of a path. 
Example: 
```
{{dirname "path/to/file.txt"}}
```
Output: 
```
path/to
```

### extname {#user-content-OS_extname}

`extname` returns the extension of a path. 
Example: 
```
{{extname "path/to/file.txt"}}
```
Output: 
```
.txt
```

### glob {#user-content-OS_glob}

`glob` returns the names of all files matching a pattern. 
Example: 
```
{{glob "/path/to/*.txt"}}
```
Output: 
```
[/path/to/file1.txt /path/to/file2.txt]
```

### isAbs {#user-content-OS_isAbs}

`isAbs` reports whether a path is absolute. 
Example: 
```
{{isAbs "/path/to/file.txt"}}
```
Output: 
```
true
```

### joinPath {#user-content-OS_joinPath}

`joinPath` joins path elements into a single path. 
- **Parameters**: elements (variadic)

Example: 
```
{{joinPath "path" "to" "file.txt"}}
```
Output: 
```
path/to/file.txt
```

### matchPath {#user-content-OS_matchPath}

`matchPath` reports whether a path matches a pattern. 
- **Parameters**: (_pattern_: string, _path_: string)

Example: 
```
{{matchPath "/path/to/*.txt" "/path/to/file.txt"}}
```
Output: 
```
true
```

### relPath {#user-content-OS_relPath}

`relPath` returns the relative path between two paths. 
- **Parameters**: (_base_: string, _target_: string)

Example: 
```
{{relPath "/path/to" "/path/to/file.txt"}}
```
Output: 
```
file.txt
```

### splitPath {#user-content-OS_splitPath}

`splitPath` splits a path into its elements. 
Example: 
```
{{splitPath "path/to/file.txt"}}
```
Output: 
```
[path to file.txt]
```

## Strings {#user-content-Strings}
### camelCase {#user-content-Strings_camelCase}

`camelCase` converts a string to camelCase. 
Example: 
```
{{camelCase "hello world"}}
```

Output: 
```
helloWorld
```

### capitalize {#user-content-Strings_capitalize}

`capitalize` capitalizes the first character of a string. 
Example: 
```
{{capitalize "hello"}}
```

Output: 
```
Hello
```

### center {#user-content-Strings_center}

`center` centers a string in a field of a given width. 
- **Parameters**: (_width_: int, _target_: string)

Example: 
```
{{center 20 "Hello"}}
```

Output: 
```
"       Hello        "
```

### hasPrefix {#user-content-Strings_hasPrefix}

`hasPrefix` checks if a string starts with a given prefix. 
- **Parameters**: (_prefix_: string, _target_: string)
- **Returns**: bool

Example: 
```
{{hasPrefix "Hello" "Hello, World!"}}
```

Output: 
```
true
```

### hasSuffix {#user-content-Strings_hasSuffix}

`hasSuffix` checks if a string ends with a given suffix. 
- **Parameters**: (_suffix_: string, _target_: string)
- **Returns**: bool

Example: 
```
{{hasSuffix "World!" "Hello, World!"}}
```

Output: 
```
true
```

### html {#user-content-Strings_html}

`html` escapes special characters in a string for use in HTML. 
Example: 
```
{{html "<script>alert('XSS')</script>"}}
```

Output: 
```
&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;
```

### join {#user-content-Strings_join}

`join` joins a slice of strings with a separator. 
- **Parameters**: (_separator_: string, _values_: slice of strings)
- **Returns**: string

Example: 
```
{{join "-" (list "apple" "banana" "cherry")}}
```

Output: 
```
apple-banana-cherry
```

### kebabCase {#user-content-Strings_kebabCase}

`kebabCase` converts a string to kebab-case. 
Example: 
```
{{kebabCase "helloWorld"}}
```

Output: 
```
hello-world
```

### lower {#user-content-Strings_lower}

`lower` converts a string to lowercase. 
Example: 
```
{{lower "HELLO"}}
```

Output: 
```
hello
```

### matchRegex {#user-content-Strings_matchRegex}

`matchRegex` checks if a string matches a regular expression. 
- **Parameters**: (_pattern_: string, _target_: string)
- **Returns**: bool

Example: 
```
{{matchRegex "^[a-z]+$" "hello"}}
```

Output: 
```
true
```

### pascalCase {#user-content-Strings_pascalCase}

`pascalCase` converts a string to PascalCase. 
Example: 
```
{{pascalCase "hello world"}}
```

Output: 
```
HelloWorld
```

### quote {#user-content-Strings_quote}

`quote` returns a double-quoted string. 
Example: 
```
{{print "hello"}}
{{quote "hello"}}
```

Output: 
```
hello
"hello"
```

### repeat {#user-content-Strings_repeat}

`repeat` repeats a string a specified number of times. 
- **Parameters**: (_count_: int, _target_: string)

Example: 
```
{{repeat 3 "abc"}}
```

Output: 
```
abcabcabc
```

### replace {#user-content-Strings_replace}

`replace` replaces all occurrences of a substring with another substring. 
- **Parameters**: (_old_: string, _new_: string, _target_: string)

Example: 
```
{{replace "o" "0" "hello world"}}
```

Output: 
```
hell0 w0rld
```

### replaceN {#user-content-Strings_replaceN}

`replaceN` replaces the first n occurrences of a substring with another substring. 
- **Parameters**: (_old_: string, _new_: string, _n_: int, _target_: string)

Example: 
```
{{replaceN "o" "0" 1 "hello world"}}
```

Output: 
```
hell0 world
```

### snakeCase {#user-content-Strings_snakeCase}

`snakeCase` converts a string to snake_case. 
Example: 
```
{{snakeCase "helloWorld"}}
```

Output: 
```
hello_world
```

### split {#user-content-Strings_split}

`split` splits a string by a separator. 
- **Parameters**: (_separator_: string, _target_: string)
- **Returns**: slice of strings

Example: 
```
{{split "," "apple,banana,cherry"}}
```

Output: 
```
[apple banana cherry]
```

### striptags {#user-content-Strings_striptags}

`striptags` removes HTML tags from a string. 
Example: 
```
{{striptags "<p>Hello <b>World</b>!</p>"}}
```

Output: 
```
Hello World!
```

### substr {#user-content-Strings_substr}

`substr` extracts a substring from a string. 
- **Parameters**: (_start_: int, _length_: int, _target_: string)

Example: 
```
{{substr 0 5 "Hello, World!"}}
```

Output: 
```
Hello
```

### trim {#user-content-Strings_trim}

`trim` removes leading and trailing whitespace from a string. 
Example: 
```
{{trim "  hello  "}}
```

Output: 
```
hello
```

### trimPrefix {#user-content-Strings_trimPrefix}

`trimPrefix` removes a prefix from a string if it exists. 
- **Parameters**: (_prefix_: string, _target_: string)

Example: 
```
{{trimPrefix "Hello, " "Hello, World!"}}
```

Output: 
```
World!
```

### trimSuffix {#user-content-Strings_trimSuffix}

`trimSuffix` removes a suffix from a string if it exists. 
- **Parameters**: (_suffix_: string, _target_: string)

Example: 
```
{{trimSuffix ", World!" "Hello, World!"}}
```

Output: 
```
Hello
```

### truncate {#user-content-Strings_truncate}

`truncate` truncates a string to a specified length and adds a suffix if truncated. 
- **Parameters**: (_length_: int, _suffix_: string, _target_: string)

Example: 
```
{{truncate 10 "..." "This is a long sentence."}}
```

Output: 
```
This is a...
```

### unquote {#user-content-Strings_unquote}

`unquote` returns an unquoted string. 
Example: 
```
{{unquote "\"hello\""}}
```

Output: 
```
hello
```

### upper {#user-content-Strings_upper}

`upper` converts a string to uppercase. 
Example: 
```
{{upper "hello"}}
```

Output: 
```
HELLO
```

### urlEscape {#user-content-Strings_urlEscape}

`urlEscape` escapes a string for use in a URL query. 
Example: 
```
{{urlEscape "hello world"}}
```

Output: 
```
hello+world
```

### urlUnescape {#user-content-Strings_urlUnescape}

`urlUnescape` unescapes a URL query string. 
Example: 
```
{{urlUnescape "hello+world"}}
```

Output: 
```
hello world
```

### wordwrap {#user-content-Strings_wordwrap}

`wordwrap` wraps words in a string to a specified width. 
- **Parameters**: (_width_: int, _target_: string)

Example: 
```
{{wordwrap 10 "This is a long sentence that needs wrapping."}}
```

Output: 
```
This is a
long
sentence
that needs
wrapping.
```

