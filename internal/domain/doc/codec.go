package doc

// The document's bytes: one strict decode and one canonical
// marshal. The decode refuses everything the machine does not write
// or mean — a second document, anchors, aliases, tags, merge keys,
// unknown fields, duplicate keys — so a trap cannot ride in through
// the format. The marshal is deterministic: the same document gives
// the same bytes, every time, in every session; reformatting a
// submission changes nothing because comparison happens after
// normalization (decode → canonical bytes).

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

// Decode — one document from bytes, strict.
func Decode(data []byte) (Document, error) {
	if err := armor(data); err != nil {
		return Document{}, err
	}
	var d Document
	if err := yaml.UnmarshalWithOptions(data, &d, yaml.Strict()); err != nil {
		return Document{}, oneLine(err)
	}
	return d, nil
}

// Marshal — the document's canonical bytes: two-space indent, map
// keys sorted, struct fields in declaration order. The same document
// always marshals to the same bytes.
func Marshal(d Document) []byte {
	// The model is plain scalar/slice/map data: marshal cannot fail.
	out, _ := yaml.MarshalWithOptions(d, yaml.Indent(2))
	return out
}

// Normalize — a submission's bytes to its canonical spelling and
// model: what the machine compares and journals.
func Normalize(data []byte) (Document, []byte, error) {
	d, err := Decode(data)
	if err != nil {
		return Document{}, nil, err
	}
	return d, Marshal(d), nil
}

// UnmarshalYAML — the command's two spellings: a string (a shell
// line) or a list of strings (an argv). Anything else is a refusal.
func (c *Cmd) UnmarshalYAML(u func(interface{}) error) error {
	var v any
	_ = u(&v) // a node that yields nothing is the default refusal below
	switch t := v.(type) {
	case string:
		c.Line, c.Argv = t, nil
		return nil
	case []any:
		argv := make([]string, 0, len(t))
		for _, item := range t {
			word, ok := item.(string)
			if !ok {
				return fmt.Errorf("run: an argv list carries strings only")
			}
			argv = append(argv, word)
		}
		if len(argv) == 0 {
			return fmt.Errorf("run: an argv list must carry at least one word")
		}
		c.Line, c.Argv = "", argv
		return nil
	default:
		return fmt.Errorf("run: one command — a string line or an argv list")
	}
}

// MarshalYAML — the command's canonical spelling: the author's own
// form rides unchanged.
func (c Cmd) MarshalYAML() ([]byte, error) {
	if c.Argv != nil {
		return yaml.Marshal(c.Argv)
	}
	return yaml.Marshal(c.Line)
}

// Key — the command's identity as a row's merge key: its canonical
// bytes. A re-spelled run is a new run — the standing one stands.
func (c Cmd) Key() string {
	b, _ := c.MarshalYAML()
	return string(b)
}

// armor — the pre-decode walk: exactly one document, no anchors,
// aliases, tags or merge keys anywhere, no tab in indentation. The
// library would accept several of these silently; the machine does
// not.
func armor(data []byte) error {
	line := 1
	for _, l := range bytes.Split(data, []byte{'\n'}) {
		for i := 0; i < len(l); i++ {
			if l[i] == ' ' {
				continue
			}
			if l[i] == '\t' {
				return fmt.Errorf("tab in indentation (line %d)", line)
			}
			break
		}
		line++
	}
	file, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		return oneLine(err)
	}
	if len(file.Docs) != 1 {
		return fmt.Errorf("exactly one document expected, found %d", len(file.Docs))
	}
	return walk(file.Docs[0].Body)
}

// walk — the node tree of the one document.
func walk(n ast.Node) error {
	switch t := n.(type) {
	case *ast.AnchorNode:
		return fmt.Errorf("anchor &%s: anchors are refused", t.Name)
	case *ast.AliasNode:
		return fmt.Errorf("alias *%s: aliases are refused", t.Value)
	case *ast.TagNode:
		return fmt.Errorf("tag %s: tags are refused", t.Start.Value)
	case *ast.MergeKeyNode:
		return fmt.Errorf("merge key <<: merge keys are refused")
	case *ast.MappingNode:
		for _, v := range t.Values {
			if err := walk(v.Key); err != nil {
				return err
			}
			if err := walk(v.Value); err != nil {
				return err
			}
		}
		return nil
	case *ast.SequenceNode:
		for _, item := range t.Values {
			if err := walk(item); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

// oneLine — errors carry source excerpts on their own lines; the
// machine's face wants one line per refusal. The caller passes a
// non-nil error.
func oneLine(err error) error {
	msg := strings.TrimSpace(err.Error())
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	return fmt.Errorf("%s", msg)
}
