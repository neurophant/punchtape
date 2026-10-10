// Package yamlio — the machine's strict YAML layer: the armor over
// github.com/goccy/go-yaml. The
// library implements YAML 1.2 and is alive; the armor exists
// because history showed format traps arrive through
// permissive parsing, not through the spec. The machine's dialect
// is deliberately narrow — flat dictionaries, lists, block
// scalars — and the layer enforces it:
//
//   - exactly ONE document (a trailing '---' document is a named
//     error, not a silently ignored tail);
//   - anchors, aliases and tags are refused anywhere (the
//     merge-key era is over);
//   - a tab in indentation is refused;
//   - duplicate keys are refused (goccy Strict);
//   - unknown fields are refused (goccy Strict);
//   - the Norway rule: a scalar whose text is on/off/yes/no/true/
//     false (any case) rides ONLY in quotes — the dialect carries
//     no booleans, so a bare trap word is a mistake to catch, not a
//     type to guess;
//   - emptiness has one canonical form: a null value and an empty
//     list decode identically, and Marshal emits the single list
//     form.
//
// Two dialects ride the same armor (DecodeStrict/DecodeDict — the
// texts dialect with no booleans; DecodeDoc — the document dialect
// of every structured surface, where the grammar's real booleans are
// legal and plain on/off/yes/no strings decode as exactly what they
// say). The decode is native throughout: custom forms dispatch
// through the yaml interface methods (UnmarshalYAML/MarshalYAML) —
// no foreign-format bridge exists anywhere in the machine.
//
// Marshal is deterministic: two-space indent, map keys sorted, the
// same data always the same bytes — the roundtrip law of the canon
// (source file == journal effect, byte for byte).
package yamlio

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
	"github.com/neurophant/punchtape/internal/fsport"
)

// trapWords — the YAML 1.1 boolean aliases that 1.2 core leaves as
// strings; in this dialect they are strings ONLY in quotes.
var trapWords = map[string]bool{
	"on": true, "off": true, "yes": true, "no": true,
	"true": true, "false": true,
}

// DecodeStrict decodes exactly one YAML document into a value with
// the full armor: the AST walk first (documents, anchors, aliases,
// tags, tabs, trap words), then goccy's strict decode (unknown
// fields, duplicate keys, types). Every refusal is a one-line
// error, without stack traces in the face.
func DecodeStrict(data []byte, v any) error {
	if err := armor(data); err != nil {
		return err
	}
	if err := yaml.UnmarshalWithOptions(data, v, yaml.Strict()); err != nil {
		return oneLine(err)
	}
	return nil
}

// DecodeDoc — the document dialect of every structured surface:
// manifests, journal frames, state, canon files, checkpoints, cache.
// The boolean-refusal of the texts dialect is off here (the grammar
// carries real booleans: accept-draft, ratify, absent); the
// structural armor stands unchanged — exactly one document, no
// anchors, aliases or tags, no indent tabs, strict fields, duplicate
// keys and types. YAML 1.2 leaves on/off/yes/no as plain strings, so
// no trap-word rule is needed in this dialect: the spelling decodes
// to exactly what it says. Custom forms dispatch through their
// native UnmarshalYAML methods, so a form's own rules (one variant
// per operation, one source per seed) keep validating themselves.
func DecodeDoc(data []byte, v any) error {
	if _, err := parsedDoc(data, true); err != nil {
		return err
	}
	if err := yaml.UnmarshalWithOptions(data, v, yaml.Strict()); err != nil {
		return oneLine(err)
	}
	return nil
}

// KindOf — the top-level "kind" scalar of a document: the live-form
// probe of the inbox door (a file that fails the armor, carries no
// kind, or carries a non-string kind is not a set form). The scan
// is structural — the walked AST, not a byte regex.
func KindOf(data []byte) (string, bool) {
	body, err := parsedDoc(data, true)
	if err != nil {
		return "", false
	}
	m, ok := body.(*ast.MappingNode)
	if !ok {
		return "", false
	}
	for _, v := range m.Values {
		key, ok := v.Key.(*ast.StringNode)
		if !ok || key.Value != "kind" {
			continue
		}
		s, ok := v.Value.(*ast.StringNode)
		if !ok {
			return "", false
		}
		return s.Value, true
	}
	return "", false
}

// DecodeDict — the canon text-bundle form: a flat dictionary of
// string keys to string texts (block scalars carry the lines; the
// composition owns spacing, so an edge-newline-free text is stored
// exactly as parsed). Keys ride the canon key alphabet (lowercase
// letters, digits, dash, dot); an empty value is a named error —
// the old parseBundle law stands.
func DecodeDict(data []byte) (map[string]string, error) {
	if err := armor(data); err != nil {
		return nil, err
	}
	out := map[string]string{}
	if err := yaml.UnmarshalWithOptions(data, &out, yaml.Strict()); err != nil {
		return nil, oneLine(err)
	}
	for k, v := range out {
		v = strings.Trim(v, "\n") // the bundle law: no newlines at either edge
		out[k] = v
		if k == "" || !keyAlphabet(k) {
			return nil, fmt.Errorf("key %q: not the canon key alphabet", k)
		}
		if v == "" {
			return nil, fmt.Errorf("segment %s is empty", k)
		}
	}
	return out, nil
}

// Marshal emits canonical YAML: two-space indent, map keys sorted,
// struct fields in declaration order, the one empty form. The same
// data gives a bit-for-bit identical output.
//
// Two laws ride the marshal. Exactness: every byte of every decodable
// string survives the marshal→decode round trip — the YAML spec
// forbids carriage returns inside block scalars (a CR is a line break
// there and normalizes away), and the library silently emits raw CR
// into literal blocks, so strings carrying CR or other control bytes
// are wrapped into exactScalar, a self-rendered double-quoted form
// whose escape set the YAML spec defines as byte-preserving.
// Sanitization: a scalar escape is a CODEPOINT escape, so sequences
// that do not decode as UTF-8 have no byte-exact spelling at all —
// they sanitize deterministically to U+FFFD at this boundary (the
// storage layer carries text; raw binary rides content files, and
// the deterministic replacement is the observable contract the red
// windows and replies have always pinned).
func Marshal(v any) ([]byte, error) {
	wrapped, err := Wrap(v)
	if err != nil {
		return nil, err
	}
	buf, err := yaml.MarshalWithOptions(wrapped, yaml.Indent(2))
	if err != nil {
		return nil, oneLine(err)
	}
	return buf, nil
}

// Wrap — the exactness pre-pass over a value tree: plain data is
// rebuilt with its unsafe strings wrapped; values whose types carry
// a custom yaml marshaler pass through untouched (the library
// dispatches the method — and the method's own product must ride
// through Wrap again, or its unsafe strings would reach the library
// unwrapped).
func Wrap(v any) (any, error) {
	return wrapExact(reflect.ValueOf(v))
}

// selfMarshaling — the type renders itself: the walk must not
// rebuild it (a rebuild loses the method); its MarshalYAML product
// is a fresh tree the method sends through Wrap. Only struct types
// reach here (the walk descends pointers before structs); a
// pointer-receiver method counts — the library dispatches it on the
// marshaled value — so rebuilding the struct behind it (TextExpect)
// would drop the custom form and leak the raw fields as a map (the
// self-damage lesson).
func selfMarshaling(t reflect.Type) bool {
	if _, ok := t.MethodByName("MarshalYAML"); ok {
		return true
	}
	_, ok := reflect.PointerTo(t).MethodByName("MarshalYAML")
	return ok
}

// exactScalar — a string that renders itself as a double-quoted
// YAML scalar with full byte escaping: the one style the spec
// guarantees to carry every byte (block scalars normalize line
// breaks; plain scalars cannot carry controls at all).
type exactScalar string

func (e exactScalar) MarshalYAML() ([]byte, error) {
	return []byte(quoteExact(string(e))), nil
}

// needsExact — the strings the library cannot carry byte-exactly:
// any control byte except the line feed (which block scalars keep)
// or DEL — and the whitespace-only-line family (the root defect): the
// YAML spec makes a block-scalar line of only spaces an empty line
// (the spaces absorb into the line break), and the lone "\n" renders
// as a clip-chomped empty body — both lose bytes through every block
// form, so they ride the exact double-quoted scalar.
func needsExact(s string) bool {
	for i := 0; i < len(s); i++ {
		if b := s[i]; (b < 0x20 && b != '\n') || b == 0x7F {
			return true
		}
	}
	if s == "\n" {
		return true
	}
	start := 0
	for i := 0; i <= len(s); i++ {
		if i != len(s) && s[i] != '\n' {
			continue
		}
		line := s[start:i]
		start = i + 1
		if line == "" {
			continue
		}
		onlySpaces := true
		for j := 0; j < len(line); j++ {
			if line[j] != ' ' {
				onlySpaces = false
				break
			}
		}
		if onlySpaces {
			return true
		}
	}
	return false
}

// sanitizeText — the storage law of the YAML layer: a scalar escape
// is a codepoint escape (\xNN means U+00NN, not the raw byte), so
// sequences that do not decode as UTF-8 have no byte-exact spelling
// at all; they sanitize to U+FFFD deterministically here — one
// replacement per undecodable run, the same observable contract the
// replies and red windows have always pinned.
func sanitizeText(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "\uFFFD")
}

// quoteExact — the double-quoted YAML scalar for any bytes: the
// YAML escape set (\\ \" \n \t \r and \xNN for the other
// controls, DEL and undecodable bytes), every printable rune as is.
func quoteExact(s string) string {
	var sb strings.Builder
	sb.Grow(len(s) + 2)
	sb.WriteByte('"')
	// the caller guarantees valid UTF-8: every rune is printable or
	// one of the escaped controls below
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch r {
		case '\\':
			sb.WriteString("\\\\")
		case '"':
			sb.WriteString("\\\"")
		case '\n':
			sb.WriteString("\\n")
		case '\t':
			sb.WriteString("\\t")
		case '\r':
			sb.WriteString("\\r")
		default:
			if r < 0x20 || r == 0x7F {
				fmt.Fprintf(&sb, "\\x%02X", r)
			} else {
				sb.WriteString(s[i : i+size])
			}
		}
		i += size
	}
	sb.WriteByte('"')
	return sb.String()
}

// wrapExact — the reflect pre-pass: a tree that carries NO unsafe
// string returns unchanged (the original value rides the library
// natively — typed nils, omitempty and the custom forms keep their
// own semantics); a tree that does is rebuilt with exactly those
// strings wrapped into exactScalar. Structures the walk cannot
// rebuild (mixed unexported fields) pass through unchanged when
// they carry no unsafe strings, and fail loudly when they do —
// silent byte loss is the one forbidden outcome. Map keys are
// identifiers by the machine's law; an unsafe key is a loud
// refusal, not a wrapping.
func wrapExact(v reflect.Value) (any, error) {
	w, _, err := wrapValue(v)
	return w, err
}

// wrapValue — the walk proper: the bool says the tree changed (the
// caller hands the original to the library when it did not).
func wrapValue(v reflect.Value) (any, bool, error) {
	switch v.Kind() {
	case reflect.Invalid:
		return nil, false, nil
	case reflect.String:
		s := v.String()
		safe := sanitizeText(s)
		if needsExact(safe) {
			return exactScalar(safe), true, nil
		}
		if safe != s {
			return safe, true, nil
		}
		return v.Interface(), false, nil
	case reflect.Interface, reflect.Ptr:
		if v.IsNil() {
			return v.Interface(), false, nil
		}
		w, changed, err := wrapValue(v.Elem())
		if err != nil {
			return nil, false, err
		}
		if !changed {
			return v.Interface(), false, nil
		}
		return w, true, nil
	case reflect.Slice, reflect.Array:
		if v.Kind() == reflect.Slice && v.IsNil() {
			// nil rides the null form: "not filled" and "explicitly
			// emptied" are different facts of the partial-update
			// semantics, and the canonicalization round trip (the
			// door marshals what the pipeline decodes back) must not
			// merge them — the library emits [] for a typed nil
			return nil, true, nil
		}
		out := make([]any, v.Len())
		changed := false
		for i := 0; i < v.Len(); i++ {
			w, c, err := wrapValue(v.Index(i))
			if err != nil {
				return nil, false, err
			}
			if c {
				changed = true
			}
			out[i] = w
		}
		if !changed {
			return v.Interface(), false, nil
		}
		return out, true, nil
	case reflect.Map:
		if v.IsNil() {
			return nil, true, nil
		}
		out := make(map[string]any, v.Len())
		changed := false
		for _, k := range v.MapKeys() {
			raw := k.String()
			key := sanitizeText(raw)
			if needsExact(key) {
				return nil, false, fmt.Errorf("map key %q carries control bytes the YAML layer cannot carry: rename the key", key)
			}
			if key != raw {
				changed = true
			}
			w, c, err := wrapValue(v.MapIndex(k))
			if err != nil {
				return nil, false, err
			}
			if c {
				changed = true
			}
			out[key] = w
		}
		if !changed {
			return v.Interface(), false, nil
		}
		return out, true, nil
	case reflect.Struct:
		if selfMarshaling(v.Type()) {
			return v.Interface(), false, nil
		}
		t := v.Type()
		hasUnexported := false
		unsafeInside := false
		for i := 0; i < v.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				hasUnexported = true
				if f.Type.Kind() == reflect.String && needsExact(sanitizeText(v.Field(i).String())) {
					unsafeInside = true
				}
				continue
			}
		}
		if hasUnexported && unsafeInside {
			return nil, false, fmt.Errorf("type %s carries control bytes in a field the YAML layer cannot wrap: store the bytes in a content file", t)
		}
		if hasUnexported {
			return v.Interface(), false, nil // e.g. time.Time: no strings to lose
		}
		fields := []reflect.StructField{}
		values := []any{}
		changed := false
		for i := 0; i < v.NumField(); i++ {
			// every field is exported here: the unexported guard above
			// already returned
			w, c, err := wrapValue(v.Field(i))
			if err != nil {
				return nil, false, err
			}
			if c {
				changed = true
			}
			fd := v.Type().Field(i)
			fields = append(fields, reflect.StructField{
				Name: fd.Name, Type: anyType, Tag: fd.Tag, Index: fd.Index,
			})
			values = append(values, w)
		}
		if !changed {
			return v.Interface(), false, nil
		}
		out := reflect.New(reflect.StructOf(fields)).Elem()
		for i, w := range values {
			if w == nil {
				out.Field(i).Set(reflect.Zero(anyType))
			} else {
				out.Field(i).Set(reflect.ValueOf(w))
			}
		}
		return out.Interface(), true, nil
	}
	// numbers, bools and the other leaves carry no bytes to lose;
	// a non-data value (a channel, a function) lands here and the
	// library's own marshal refuses it
	return v.Interface(), false, nil
}

// anyType — the rebuilt struct fields hold interface{} leaves so a
// wrapped string and a plain one ride the same field.
var anyType = reflect.TypeOf((*any)(nil)).Elem()

// WriteAtomic writes data to a file in one action: a temp file next
// to the target, then a rename — a reader at any moment sees either
// the whole old content or the new one. The write rides the
// machine's filesystem port (fsport): units bind an in-memory fake
// there.
func WriteAtomic(path string, data []byte) error {
	return fsport.Bind.WriteAtomic(path, data)
}

// armor — the pre-decode walk over the parsed AST plus the byte
// scan for tab indentation. A refusal here is the armor speaking:
// the library would accept several of these silently (a trailing
// document is dropped, a type-compatible alias resolves, a bare
// 'no' becomes the string "no" with no one the wiser).
func armor(data []byte) error {
	_, err := parsedDoc(data, false)
	return err
}

// parsedDoc — the shared armor parse: tab scan, one document, the
// node walk in the dialect asked for. The texts dialect (doc=false)
// refuses booleans and bare trap words; the document dialect
// (doc=true) carries real booleans and plain trap-word strings.
func parsedDoc(data []byte, doc bool) (ast.Node, error) {
	if err := noIndentTabs(data); err != nil {
		return nil, err
	}
	file, err := parser.ParseBytes(data, parser.ParseComments)
	if err != nil {
		return nil, oneLine(err)
	}
	if len(file.Docs) != 1 {
		return nil, fmt.Errorf("exactly one document expected, found %d", len(file.Docs))
	}
	if err := walk(file.Docs[0].Body, doc); err != nil {
		return nil, err
	}
	return file.Docs[0].Body, nil
}

// walk — the node tree: anchors, aliases and tags are refused
// wherever they sit; in the texts dialect so are bare trap words
// and booleans. StringNode covers both plain and quoted scalars;
// the origin carries the quotes, so the Norway rule bites only on
// the unquoted form.
func walk(n ast.Node, doc bool) error {
	switch t := n.(type) {
	case *ast.AnchorNode:
		return fmt.Errorf("anchor &%s is not in the dialect: anchors are refused", t.Name.String())
	case *ast.AliasNode:
		return fmt.Errorf("alias *%s is not in the dialect: aliases are refused", t.Value.String())
	case *ast.TagNode:
		return fmt.Errorf("tag %s is not in the dialect: tags are refused", t.Start.Value)
	case *ast.StringNode:
		if !doc {
			if tok := t.GetToken(); tok != nil && !quoted(tok.Origin) && trapWords[strings.ToLower(t.Value)] {
				return fmt.Errorf("%q is a trap word: quote it or rename it (the dialect has no booleans)", t.Value)
			}
		}
		return nil
	case *ast.BoolNode:
		if doc {
			return nil
		}
		return fmt.Errorf("bare boolean %s: the dialect has no booleans — quote it or rename it", t.String())
	case *ast.LiteralNode:
		return nil
	case *ast.MappingNode:
		for _, v := range t.Values {
			if err := walk(v.Key, doc); err != nil {
				return err
			}
			if err := walk(v.Value, doc); err != nil {
				return err
			}
		}
		return nil
	case *ast.SequenceNode:
		for _, item := range t.Values {
			if err := walk(item, doc); err != nil {
				return err
			}
		}
		return nil
	default:
		return nil
	}
}

// quoted — the token's raw origin carries the quotes when there
// were any.
func quoted(origin string) bool {
	origin = strings.TrimSpace(origin)
	return len(origin) >= 2 &&
		((origin[0] == '"' && origin[len(origin)-1] == '"') ||
			(origin[0] == '\'' && origin[len(origin)-1] == '\''))
}

// noIndentTabs — a tab byte in a line's leading whitespace: the
// library's scanner also refuses it, but the armor names the line.
func noIndentTabs(data []byte) error {
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
	return nil
}

// keyAlphabet — lowercase letters, digits, dash, dot: the canon
// key alphabet of the bundle era, unchanged by the format swap.
func keyAlphabet(k string) bool {
	for _, c := range k {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-', c == '.':
		default:
			return false
		}
	}
	return true
}

// oneLine — errors carry source excerpts on their own lines; the
// machine's face wants one line per refusal.
func oneLine(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(err.Error())
	if i := strings.IndexByte(msg, '\n'); i >= 0 {
		msg = msg[:i]
	}
	return fmt.Errorf("%s", msg)
}
