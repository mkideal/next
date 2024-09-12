package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type TemplateItem struct {
	Name     string
	Content  string
	Children map[string]*TemplateItem
	Links    []string
}

func (t *TemplateItem) IsProperty() bool {
	return strings.HasPrefix(t.Name, ".")
}

func (t *TemplateItem) NumProperties() int {
	n := 0
	for _, child := range t.Children {
		if child.IsProperty() {
			n++
		}
	}
	return n
}

func trimSpace(s string) string {
	s = strings.TrimLeft(s, " \r\n")
	s = strings.TrimRight(s, " \t\r\n")
	for i := 0; i < len(s); i++ {
		if s[i] == '\t' {
			return s[i+1:]
		}
		return s[i:]
	}
	return ""
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: program <input_directory> <output_file>")
		return
	}

	inputDir := os.Args[1]
	outputFile := os.Args[2]

	root := &TemplateItem{Children: make(map[string]*TemplateItem)}

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			err := processFile(path, root)
			if err != nil {
				return fmt.Errorf("error processing file %s: %v", path, err)
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking through directory: %v\n", err)
		return
	}

	err = generateMarkdown(root, outputFile)
	if err != nil {
		fmt.Printf("Error generating Markdown: %v\n", err)
		return
	}

	fmt.Println("Markdown file generated successfully.")
}

func processFile(filename string, root *TemplateItem) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, commentGroup := range node.Comments {
		processCommentGroup(commentGroup, root)
	}

	return nil
}

func processCommentGroup(group *ast.CommentGroup, root *TemplateItem) {
	if len(group.List) == 0 {
		return
	}

	firstComment := group.List[0].Text
	if !strings.HasPrefix(firstComment, "// @api(") {
		return
	}

	path := strings.TrimPrefix(firstComment, "// @api(")
	end := strings.Index(path, ")")
	currentItem := addToTree(root, path[:end])
	group.List[0].Text = "// _" + currentItem.Name + "_ " + strings.TrimSpace(path[end+1:])

	var content bytes.Buffer
	var coding int
	const (
		codingNone = iota
		codingStart
		codingEnd
	)
	for _, comment := range group.List { // Start from the second comment
		text := strings.TrimPrefix(comment.Text, "//")
		text = trimSpace(text)
		if strings.HasPrefix(text, "```") {
			if coding == codingStart {
				coding = codingEnd
				if content.Len() > 1 && bytes.Equal(content.Bytes()[content.Len()-2:], []byte("\n\n")) {
					content.Truncate(content.Len() - 1)
				}
			} else {
				coding = codingStart
				content.WriteString("\n")
			}
		}
		content.WriteString(text)
		if coding != codingNone || (text == "" || isListItem(text)) {
			content.WriteString("\n")
		} else {
			content.WriteString(" ")
		}
		if coding == codingEnd {
			coding = codingNone
		}
	}
	content.WriteString("\n")

	currentItem.Content = content.String()
	currentItem.Links = extractLinks(currentItem.Content)
}

func isListItem(text string) bool {
	return len(text) >= 2 && (text[0] == '*' || text[0] == '-' || text[0] == '+') && text[1] == ' '
}

func addToTree(root *TemplateItem, path string) *TemplateItem {
	parts := strings.Split(path, "/")
	current := root

	if len(parts) > 0 {
		if index := strings.Index(parts[len(parts)-1], "."); index != -1 {
			newPart := parts[len(parts)-1][index:]
			parts[len(parts)-1] = parts[len(parts)-1][:index]
			parts = append(parts, newPart)
		}
	}

	for _, part := range parts {
		if current.Children == nil {
			current.Children = make(map[string]*TemplateItem)
		}
		if _, exists := current.Children[part]; !exists {
			current.Children[part] = &TemplateItem{Name: part}
		}
		current = current.Children[part]
	}
	return current
}

func extractLinks(content string) []string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(#([^)]+)\)`)
	matches := re.FindAllStringSubmatch(content, -1)
	links := make([]string, len(matches))
	for i, match := range matches {
		links[i] = match[2] // The anchor name is in the second capture group
	}
	return links
}

func generateMarkdown(item *TemplateItem, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var toc, content bytes.Buffer
	writeMarkdownTree(&toc, &content, item, 0, "")

	file.WriteString("# API Reference\n\n")
	file.WriteString(toc.String())
	file.WriteString("\n")
	file.WriteString(content.String())
	return nil
}

func linkName(path string) string {
	return "user-content-" + strings.ReplaceAll(strings.ReplaceAll(path, ".", "-"), "/", "_")
}

func writeMarkdownTree(toc, content io.Writer, item *TemplateItem, depth int, parentPath string) {
	if item.Name != "" {
		level := depth + 1
		fullPath := parentPath + item.Name
		if item.IsProperty() {
			level = 5 // <h5>
		}
		fmt.Fprintf(content, "<h%d><a id=\"%s\" target=\"_self\">%s</a></h%d>\n", level, linkName(fullPath), item.Name, level)
		if !item.IsProperty() {
			fmt.Fprintf(toc, "%s<li><a href=\"#%s\">%s</a></li>\n", strings.Repeat("  ", depth), linkName(fullPath), item.Name)
		}

		if item.Content != "" {
			text := processLinks(item.Content)
			fmt.Fprintf(content, "\n%s\n\n", trimSpace(text))
		}
	}

	children := sortedChildren(item)
	parentPath = parentPath + item.Name
	if parentPath != "" {
		parentPath = parentPath + "/"
	}
	if len(children) > item.NumProperties() {
		fmt.Fprintf(toc, "<ul>\n")
	}
	for _, child := range children {
		writeMarkdownTree(toc, content, child, depth+1, parentPath)
	}
	if len(children) > item.NumProperties() {
		fmt.Fprintf(toc, "</ul>\n")
	}

}

func processLinks(content string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(#([^)]+)\)`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			linkText := parts[1]
			linkPath := parts[2]
			anchorName := linkName(linkPath)
			return fmt.Sprintf("[%s](#%s)", linkText, anchorName)
		}
		return match
	})
}

func specialOrder(name string) int {
	if len(name) == 0 {
		return 0
	}
	switch name[0] {
	case '.':
		return -1
	case '-':
		return -2
	case '_':
		return -3
	default:
		return 0
	}
}

func sortedChildren(item *TemplateItem) []*TemplateItem {
	var children []*TemplateItem
	for _, child := range item.Children {
		children = append(children, child)
	}
	sort.Slice(children, func(i, j int) bool {
		oi := specialOrder(children[i].Name)
		oj := specialOrder(children[j].Name)
		if oi != oj {
			return oi < oj
		}
		return children[i].Name < children[j].Name
	})
	return children
}
