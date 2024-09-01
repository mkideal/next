package ast

import (
	"bytes"
	"cmp"
	"fmt"
	"slices"
	"strings"

	"github.com/next/next/src/token"
)

// sortComments sorts the list of comment groups in source order.
func sortComments(list []*CommentGroup) {
	slices.SortFunc(list, func(a, b *CommentGroup) int {
		return cmp.Compare(a.Pos(), b.Pos())
	})
}

// CommentMap maps an AST node to a list of comment groups
// associated with it. See NewCommentMap for a description of
// the association.
type CommentMap map[Node][]*CommentGroup

// addComment adds a comment to the CommentMap.
func (cmap CommentMap) addComment(n Node, c *CommentGroup) {
	cmap[n] = append(cmap[n], c)
}

// byInterval implements sort.Interface for sorting nodes by position.
type byInterval []Node

func (a byInterval) Len() int      { return len(a) }
func (a byInterval) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byInterval) Less(i, j int) bool {
	pi, pj := a[i].Pos(), a[j].Pos()
	return pi < pj || pi == pj && a[i].End() > a[j].End()
}

// nodeList returns the list of nodes of the AST n in source order.
func nodeList(n Node) []Node {
	var list []Node
	Inspect(n, func(n Node) bool {
		// don't collect comments
		switch n.(type) {
		case nil, *CommentGroup, *Comment:
			return false
		}
		list = append(list, n)
		return true
	})

	return list
}

// commentListReader helps iterating through a list of comment groups.
type commentListReader struct {
	fset     *token.FileSet
	list     []*CommentGroup
	index    int
	comment  *CommentGroup
	pos, end token.Position // source interval of comment group at current index
}

// eol returns true if the reader has reached the end of the list.
func (r *commentListReader) eol() bool {
	return r.index >= len(r.list)
}

// next advances the reader to the next comment group.
func (r *commentListReader) next() {
	if !r.eol() {
		r.comment = r.list[r.index]
		r.pos = r.fset.Position(r.comment.Pos())
		r.end = r.fset.Position(r.comment.End())
		r.index++
	}
}

// nodeStack keeps track of nested nodes.
// A node lower on the stack lexically contains the nodes higher on the stack.
type nodeStack []Node

// push pops all nodes that appear lexically before n
// and then pushes n on the stack.
func (s *nodeStack) push(n Node) {
	s.pop(n.Pos())
	*s = append(*s, n)
}

// pop pops all nodes that appear lexically before pos
// (i.e., whose lexical extent has ended before or at pos).
// It returns the last node popped.
func (s *nodeStack) pop(pos token.Pos) (top Node) {
	i := len(*s)
	for i > 0 && (*s)[i-1].End() <= pos {
		top = (*s)[i-1]
		i--
	}
	*s = (*s)[:i]
	return top
}

// NewCommentMap creates a new comment map by associating comment groups
// of the comments list with the nodes of the AST specified by node.
//
// A comment group g is associated with a node n if:
//
//   - g starts on the same line as n ends
//   - g starts on the line immediately following n, and there is
//     at least one empty line after g and before the next node
//   - g starts before n and is not associated to the node before n
//     via the previous rules
//
// NewCommentMap tries to associate a comment group to the "largest"
// node possible: For instance, if the comment is a line comment
// trailing an assignment, the comment is associated with the entire
// assignment rather than just the last operand in the assignment.
func NewCommentMap(fset *token.FileSet, node Node, comments []*CommentGroup) CommentMap {
	if len(comments) == 0 {
		return nil // no comments to map
	}

	cmap := make(CommentMap)

	// set up comment reader r
	tmp := make([]*CommentGroup, len(comments))
	copy(tmp, comments) // don't change incoming comments
	sortComments(tmp)
	r := commentListReader{fset: fset, list: tmp}
	r.next()

	// create node list in lexical order
	nodes := nodeList(node)
	nodes = append(nodes, nil) // append sentinel

	// set up iteration variables
	var (
		p     Node
		pend  token.Position // end of p
		pg    Node           // previous node group (enclosing nodes of "importance")
		pgend token.Position // end of pg
		stack nodeStack      // stack of node groups
	)

	for _, q := range nodes {
		var qpos token.Position
		if q != nil {
			qpos = fset.Position(q.Pos())
		} else {
			// set fake sentinel position to infinity so that
			// all comments get processed before the sentinel
			qpos.Offset = 1 << 30
			qpos.Line = 1 << 30
		}

		// process comments before current node
		for r.end.Offset <= qpos.Offset {
			// determine recent node group
			if top := stack.pop(r.comment.Pos()); top != nil {
				pg = top
				pgend = fset.Position(pg.End())
			}
			// Try to associate a comment first with a node group
			// (i.e., a node of "importance" such as a declaration);
			// if that fails, try to associate it with the most recent
			// node.
			var assoc Node
			switch {
			case pg != nil &&
				(pgend.Line == r.pos.Line ||
					pgend.Line+1 == r.pos.Line && r.end.Line+1 < qpos.Line):
				// comment starts on same line as previous node group ends, or
				// comment starts on the line immediately after the
				// previous node group and there is an empty line before
				// the current node
				assoc = pg
			case p != nil &&
				(pend.Line == r.pos.Line ||
					pend.Line+1 == r.pos.Line && r.end.Line+1 < qpos.Line ||
					q == nil):
				// same rules apply as above for p rather than pg,
				// but also associate with p if we are at the end (q == nil)
				assoc = p
			default:
				// otherwise, associate comment with current node
				if q == nil {
					panic("internal error: no comments should be associated with sentinel")
				}
				assoc = q
			}
			cmap.addComment(assoc, r.comment)
			if r.eol() {
				return cmap
			}
			r.next()
		}

		// update previous node
		p = q
		pend = fset.Position(p.End())

		// update previous node group if we see an "important" node
		switch q.(type) {
		case *File, *EnumMember, *StructField, *Method, Decl, Type:
			stack.push(q)
		}
	}

	return cmap
}

// Update replaces an old node in the comment map with the new node
// and returns the new node. Comments that were associated with the
// old node are associated with the new node.
func (cmap CommentMap) Update(old, new Node) Node {
	if list := cmap[old]; len(list) > 0 {
		delete(cmap, old)
		cmap[new] = append(cmap[new], list...)
	}
	return new
}

// Filter returns a new comment map consisting of only those
// entries of cmap for which a corresponding node exists in
// the AST specified by node.
func (cmap CommentMap) Filter(node Node) CommentMap {
	umap := make(CommentMap)
	Inspect(node, func(n Node) bool {
		if g := cmap[n]; len(g) > 0 {
			umap[n] = g
		}
		return true
	})
	return umap
}

// Comments returns the list of comment groups in the comment map.
// The result is sorted in source order.
func (cmap CommentMap) Comments() []*CommentGroup {
	list := make([]*CommentGroup, 0, len(cmap))
	for _, e := range cmap {
		list = append(list, e...)
	}
	sortComments(list)
	return list
}

// summary returns a condensed string representation of the given comment groups.
func summary(list []*CommentGroup) string {
	const maxLen = 40
	var buf bytes.Buffer

	for _, group := range list {
		for _, comment := range group.List {
			if buf.Len() >= maxLen {
				break
			}
			buf.WriteString(comment.Text)
		}
		if buf.Len() >= maxLen {
			break
		}
	}

	if buf.Len() > maxLen {
		buf.Truncate(maxLen - 3)
		buf.WriteString("...")
	}

	bytes := buf.Bytes()
	for i, b := range bytes {
		switch b {
		case '\t', '\n', '\r':
			bytes[i] = ' '
		}
	}

	return string(bytes)
}

// String returns a string representation of the CommentMap.
func (cmap CommentMap) String() string {
	var nodes []Node
	for node := range cmap {
		nodes = append(nodes, node)
	}
	slices.SortFunc(nodes, func(a, b Node) int {
		r := cmp.Compare(a.Pos(), b.Pos())
		if r != 0 {
			return r
		}
		return cmp.Compare(a.End(), b.End())
	})

	var buf strings.Builder
	fmt.Fprintln(&buf, "CommentMap {")
	for _, node := range nodes {
		comment := cmap[node]
		var s string
		if ident, ok := node.(*Ident); ok {
			s = ident.Name
		} else {
			s = fmt.Sprintf("%T", node)
		}
		fmt.Fprintf(&buf, "\t%p  %20s:  %s\n", node, s, summary(comment))
	}
	fmt.Fprintln(&buf, "}")
	return buf.String()
}
