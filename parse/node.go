package parse

import (
    "bytes"
    "fmt"
    "strconv"
    "strings"
)

var textFormat = "%s" // Changed to "%q" in tests for better error messages.

// A Node is an element in the parse tree. The interface is trivial.
// The interface contains an unexported method so that only
// types local to this package can satisfy it.
type Node interface {
    Type() NodeType
    String() string
    // Copy does a deep copy of the Node and all its components.
    // To avoid type assertions, some XxxNodes also have specialized
    // CopyXxx methods that return *XxxNode.
    Copy() Node
    Position() Pos // byte position of start of node in full original input string
    // tree returns the containing *Tree.
    // It is unexported so all implementations of Node are in this package.
    tree() *Tree
    withFallback(other Node) Node
}

// NodeType identifies the type of a parse tree node.
type NodeType int

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
    return p
}

// Type returns itself and provides an easy default implementation
// for embedding in a Node. Embedded in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
    return t
}

const (
    NodeText       NodeType = iota // Plain text.
    NodeField                      // A field or method name.
    NodeList                       // A list of Nodes.
    NodeMap                        // A map of Nodes.
    NodeNil                        // An untyped nil constant.
    NodeBool                       // A boolean constant.
    NodeNumber                     // A numerical constant.
    NodeString                     // A string constant.
)

// Nodes.

// MapNode holds a map of nodes.
type MapNode struct {
    NodeType
    Pos
    tr  *Tree
    Nodes map[string]Node
}

func (t *Tree) newMap(pos Pos) *MapNode {
    return &MapNode{tr: t, NodeType: NodeMap, Pos: pos, Nodes: make(map[string]Node)}
}

func (m *MapNode) put(key string, n Node) {
    m.Nodes[key] = n
}

func (m *MapNode) tree() *Tree {
    return m.tr
}

func (m *MapNode) String() string {
    b := new(bytes.Buffer)
    for k, n := range m.Nodes {
        fmt.Fprint(b, k, " = (", n, ")")
    }
    return b.String()
}

func (m *MapNode) CopyMap() *MapNode {
    if m == nil {
        return m
    }
    n := m.tr.newMap(m.Pos)
    for key, elem := range m.Nodes {
        n.put(key, elem.Copy())
    }
    return n
}

func (m *MapNode) Copy() Node {
    return m.CopyMap()
}

func (m *MapNode) withFallback(other Node) Node {
    if o, ok := other.(*MapNode); ok {
        for k, v := range o.Nodes {
            if _, ok := m.Nodes[k]; !ok {
                m.Nodes[k] = v
            } else {
                m.Nodes[k] = m.Nodes[k].withFallback(v)
            }
        }
    }
    return m
}

// ListNode holds a sequence of nodes.
type ListNode struct {
    NodeType
    Pos
    tr    *Tree
    Nodes []Node // The element nodes in lexical order.
}

func (t *Tree) newList(pos Pos) *ListNode {
    return &ListNode{tr: t, NodeType: NodeList, Pos: pos}
}

func (l *ListNode) append(n Node) {
    l.Nodes = append(l.Nodes, n)
}

func (l *ListNode) tree() *Tree {
    return l.tr
}

func (l *ListNode) String() string {
    b := new(bytes.Buffer)
    for _, n := range l.Nodes {
        fmt.Fprint(b, n)
    }
    return b.String()
}

func (l *ListNode) CopyList() *ListNode {
    if l == nil {
        return l
    }
    n := l.tr.newList(l.Pos)
    for _, elem := range l.Nodes {
        n.append(elem.Copy())
    }
    return n
}

func (l *ListNode) Copy() Node {
    return l.CopyList()
}

func (m *ListNode) withFallback(other Node) Node {
    return m
}

// TextNode holds plain text.
type TextNode struct {
    NodeType
    Pos
    tr   *Tree
    Text []byte // The text; may span newlines.
}

func (t *Tree) newText(pos Pos, text string) *TextNode {
    return &TextNode{tr: t, NodeType: NodeText, Pos: pos, Text: []byte(text)}
}

func (t *TextNode) String() string {
    return fmt.Sprintf(textFormat, t.Text)
}

func (t *TextNode) tree() *Tree {
    return t.tr
}

func (t *TextNode) Copy() Node {
    return &TextNode{tr: t.tr, NodeType: NodeText, Pos: t.Pos, Text: append([]byte{}, t.Text...)}
}

func (m *TextNode) withFallback(other Node) Node {
    return m
}

// NilNode holds the special identifier 'nil' representing an untyped nil constant.
type NilNode struct {
    NodeType
    Pos
    tr *Tree
}

func (t *Tree) newNil(pos Pos) *NilNode {
    return &NilNode{tr: t, NodeType: NodeNil, Pos: pos}
}

func (n *NilNode) Type() NodeType {
    // Override method on embedded NodeType for API compatibility.
    // TODO: Not really a problem; could change API without effect but
    // api tool complains.
    return NodeNil
}

func (n *NilNode) String() string {
    return "nil"
}

func (n *NilNode) tree() *Tree {
    return n.tr
}

func (n *NilNode) Copy() Node {
    return n.tr.newNil(n.Pos)
}

func (m *NilNode) withFallback(other Node) Node {
    return m
}

// FieldNode holds a field (identifier starting with '.').
// The names may be chained ('.x.y').
// The period is dropped from each ident.
type FieldNode struct {
    NodeType
    Pos
    tr    *Tree
    Ident []string // The identifiers in lexical order.
}

func (t *Tree) newField(pos Pos, ident string) *FieldNode {
    return &FieldNode{tr: t, NodeType: NodeField, Pos: pos, Ident: strings.Split(ident[1:], ".")} // [1:] to drop leading period
}

func (f *FieldNode) String() string {
    s := ""
    for _, id := range f.Ident {
        s += "." + id
    }
    return s
}

func (f *FieldNode) tree() *Tree {
    return f.tr
}

func (f *FieldNode) Copy() Node {
    return &FieldNode{tr: f.tr, NodeType: NodeField, Pos: f.Pos, Ident: append([]string{}, f.Ident...)}
}

func (m *FieldNode) withFallback(other Node) Node {
    return m
}

// BoolNode holds a boolean constant.
type BoolNode struct {
    NodeType
    Pos
    tr   *Tree
    True bool // The value of the boolean constant.
}

func (t *Tree) newBool(pos Pos, true bool) *BoolNode {
    return &BoolNode{tr: t, NodeType: NodeBool, Pos: pos, True: true}
}

func (b *BoolNode) String() string {
    if b.True {
        return "true"
    }
    return "false"
}

func (b *BoolNode) tree() *Tree {
    return b.tr
}

func (b *BoolNode) Copy() Node {
    return b.tr.newBool(b.Pos, b.True)
}

func (m *BoolNode) withFallback(other Node) Node {
    return m
}

// NumberNode holds a number: signed or unsigned integer, float, or complex.
// The value is parsed and stored under all the types that can represent the value.
// This simulates in a small amount of code the behavior of Go's ideal constants.
type NumberNode struct {
    NodeType
    Pos
    tr         *Tree
    IsInt      bool       // Number has an integral value.
    IsUint     bool       // Number has an unsigned integral value.
    IsFloat    bool       // Number has a floating-point value.
    IsComplex  bool       // Number is complex.
    Int64      int64      // The signed integer value.
    Uint64     uint64     // The unsigned integer value.
    Float64    float64    // The floating-point value.
    Complex128 complex128 // The complex value.
    Text       string     // The original textual representation from the input.
}

func (t *Tree) newNumber(pos Pos, text string, typ itemType) (*NumberNode, error) {
    n := &NumberNode{tr: t, NodeType: NodeNumber, Pos: pos, Text: text}
    switch typ {
        case itemComplex:
            // fmt.Sscan can parse the pair, so let it do the work.
            if _, err := fmt.Sscan(text, &n.Complex128); err != nil {
                return nil, err
            }
            n.IsComplex = true
            n.simplifyComplex()
            return n, nil
    }
    // Imaginary constants can only be complex unless they are zero.
    if len(text) > 0 && text[len(text)-1] == 'i' {
        f, err := strconv.ParseFloat(text[:len(text)-1], 64)
        if err == nil {
            n.IsComplex = true
            n.Complex128 = complex(0, f)
            n.simplifyComplex()
            return n, nil
        }
    }
    // Do integer test first so we get 0x123 etc.
    u, err := strconv.ParseUint(text, 0, 64) // will fail for -0; fixed below.
    if err == nil {
        n.IsUint = true
        n.Uint64 = u
    }
    i, err := strconv.ParseInt(text, 0, 64)
    if err == nil {
        n.IsInt = true
        n.Int64 = i
        if i == 0 {
            n.IsUint = true // in case of -0.
            n.Uint64 = u
        }
    }
    // If an integer extraction succeeded, promote the float.
    if n.IsInt {
        n.IsFloat = true
        n.Float64 = float64(n.Int64)
    } else if n.IsUint {
        n.IsFloat = true
        n.Float64 = float64(n.Uint64)
    } else {
        f, err := strconv.ParseFloat(text, 64)
        if err == nil {
            n.IsFloat = true
            n.Float64 = f
            // If a floating-point extraction succeeded, extract the int if needed.
            if !n.IsInt && float64(int64(f)) == f {
                n.IsInt = true
                n.Int64 = int64(f)
            }
            if !n.IsUint && float64(uint64(f)) == f {
                n.IsUint = true
                n.Uint64 = uint64(f)
            }
        }
    }
    if !n.IsInt && !n.IsUint && !n.IsFloat {
        return nil, fmt.Errorf("illegal number syntax: %q", text)
    }
    return n, nil
}

// simplifyComplex pulls out any other types that are represented by the complex number.
// These all require that the imaginary part be zero.
func (n *NumberNode) simplifyComplex() {
    n.IsFloat = imag(n.Complex128) == 0
    if n.IsFloat {
        n.Float64 = real(n.Complex128)
        n.IsInt = float64(int64(n.Float64)) == n.Float64
        if n.IsInt {
            n.Int64 = int64(n.Float64)
        }
        n.IsUint = float64(uint64(n.Float64)) == n.Float64
        if n.IsUint {
            n.Uint64 = uint64(n.Float64)
        }
    }
}

func (n *NumberNode) String() string {
    return n.Text
}

func (n *NumberNode) tree() *Tree {
    return n.tr
}

func (n *NumberNode) Copy() Node {
    nn := new(NumberNode)
    *nn = *n // Easy, fast, correct.
    return nn
}

func (m *NumberNode) withFallback(other Node) Node {
    return m
}

// StringNode holds a string constant. The value has been "unquoted".
type StringNode struct {
    NodeType
    Pos
    tr     *Tree
    Quoted string // The original text of the string, with quotes.
    Text   string // The string, after quote processing.
}

func (t *Tree) newString(pos Pos, orig, text string) *StringNode {
    return &StringNode{tr: t, NodeType: NodeString, Pos: pos, Quoted: orig, Text: text}
}

func (s *StringNode) String() string {
    return s.Quoted
}

func (s *StringNode) tree() *Tree {
    return s.tr
}

func (s *StringNode) Copy() Node {
    return s.tr.newString(s.Pos, s.Quoted, s.Text)
}

func (m *StringNode) withFallback(other Node) Node {
    return m
}