// Package brief — the brief compiler: a machine slot turned into a
// request for the executor. Caller shapes differ, the slot structure
// is one; zero business logic here.
package brief

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/neurophant/punchtape/internal/canondata"
)

// Caller shapes. The slot (situation, one task, output contract,
// bounds) is always the same text; the envelope is what the caller
// can carry: a terminal turn, a workspace file pair, or a bare JSON
// request. Adapters translate the envelope, never the slot.
const (
	// ShapeCLI — an interactive CLI agent: one turn on stdin, the delta
	// YAML (plus an optional "# TOKENS <n>" trailer) on stdout.
	ShapeCLI = "cli"
	// ShapeIDE — an IDE agent working the workspace: the request lands
	// in .punchtape/caller/slot.md, the reply in delta.yaml beside it.
	ShapeIDE = "ide"
	// ShapeAPI — a bare API call: a strict JSON request on stdin, a
	// strict JSON response on stdout.
	ShapeAPI = "api"
)

// KnownShape — is the shape one of the catalog.
func KnownShape(shape string) bool {
	switch shape {
	case ShapeCLI, ShapeIDE, ShapeAPI:
		return true
	}
	return false
}

// apiRequest — the strict bare API request form: fields in fixed
// order, deterministic output.
type apiRequest struct {
	Slot         string `json:"slot"`
	LastReply    string `json:"last_reply,omitempty"`
	ReplyFormat  string `json:"reply_format"`
	ReplyExample string `json:"reply_example"`
}

// apiResponse — the strict bare API response form.
type apiResponse struct {
	Delta  string `json:"delta"`
	Tokens int64  `json:"tokens"`
}

// Make builds the executor brief from the slot (the next text) and,
// on retries, the reply to the previous submission. The brief is the
// goal, the verbs and the slot — nothing more. The default shape is
// the interactive CLI turn.
func Make(slot string, lastReply string) string {
	return MakeFor(ShapeCLI, slot, lastReply)
}

// MakeFor builds the request envelope for the caller shape. The slot
// text is carried verbatim in every shape.
func MakeFor(shape, slot, lastReply string) string {
	switch shape {
	case ShapeIDE:
		var sb strings.Builder
		sb.WriteString(canondata.T("brief.ide.head"))
		sb.WriteString("\n\n")
		sb.WriteString(canondata.T("brief.ide.context-head"))
		sb.WriteString("\n\n")
		sb.WriteString(slot)
		sb.WriteString("\n\n")
		if strings.TrimSpace(lastReply) != "" {
			sb.WriteString(canondata.T("brief.ide.reply-head"))
			sb.WriteString("\n\n")
			sb.WriteString(lastReply)
			sb.WriteString("\n\n")
		}
		sb.WriteString(canondata.T("brief.ide.contract-head"))
		sb.WriteString("\n\n")
		sb.WriteString(canondata.T("brief.ide.contract"))
		sb.WriteString("\n\n")
		sb.WriteString(canondata.T("brief.ide.bounds-head"))
		sb.WriteString("\n\n")
		sb.WriteString(canondata.T("brief.ide.bounds"))
		sb.WriteString("\n")
		return sb.String()
	case ShapeAPI:
		req := apiRequest{
			Slot:         slot,
			LastReply:    lastReply,
			ReplyFormat:  canondata.T("brief.api.reply-format"),
			ReplyExample: canondata.T("brief.api.reply-example"),
		}
		data, err := json.Marshal(req)
		if err != nil {
			// marshaling a struct with string fields does not fail
			return "{}"
		}
		return string(data) + "\n"
	default: // ShapeCLI
		var sb strings.Builder
		sb.WriteString(canondata.T("brief.cli.head"))
		sb.WriteString("\n\n")
		sb.WriteString(canondata.T("brief.cli.context-head"))
		sb.WriteString("\n")
		sb.WriteString(slot)
		if strings.TrimSpace(lastReply) != "" {
			sb.WriteString("\n\n")
			sb.WriteString(canondata.T("brief.cli.reply-head"))
			sb.WriteString("\n")
			sb.WriteString(lastReply)
		}
		sb.WriteString("\n\n")
		sb.WriteString(canondata.T("brief.cli.contract-head"))
		sb.WriteString("\n")
		sb.WriteString(canondata.T("brief.cli.contract"))
		sb.WriteString("\n")
		sb.WriteString("\n")
		sb.WriteString(canondata.T("brief.cli.bounds-head"))
		sb.WriteString("\n")
		sb.WriteString(canondata.T("brief.cli.bounds"))
		sb.WriteString("\n")
		return sb.String()
	}
}

// IDE exchange: the slot and reply file names in the caller
// directory.
const (
	IDEChallengeFile = "slot.md"
	IDEReplyFile     = "delta.yaml"
)

// DecodeAPIReply parses the bare API's strict JSON reply: a delta
// and the token counter. Anything but the two fields — an error.
// Normalization is only a guarantee: the delta's bytes are
// untouchable (trailing newlines are block scalar content); the
// envelope only ensures one final newline when it is missing.
func DecodeAPIReply(out string) (string, int64, error) {
	var resp apiResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		return "", 0, fmt.Errorf("%s: %w", canondata.T("brief.error.api-json"), err)
	}
	return ensureTrailingBreak(strings.TrimLeft(resp.Delta, " \t\r\n")), resp.Tokens, nil
}

// SplitTokenReport separates the trailing "# TOKENS <n>" line from
// the executor's reply; the third return — whether the trailer was
// there, the fourth — whether it reported a number. "# TOKENS n/a" —
// the caller honestly declared the absence of a counter: the trailer
// is there, the number is not ("no data" is booked, not zero). A
// missing trailer is indistinguishable from a declared zero
// ("# TOKENS 0" is a legal report), so the presence flag is
// explicit. The envelope is shared by all shapes: the driver strips
// it from the executor's reply, a manual CLI submission from the
// delta file before the strict YAML decode. The document's bytes
// before the trailer are not touched: only the trailer line is cut;
// the newline before it stays with the document (the final \n of
// the last block scalar is delta content, not decorative
// emptiness).
func SplitTokenReport(out string) (string, int64, bool, bool) {
	body := strings.TrimRight(out, "\n")
	tail, doc := body, ""
	if i := strings.LastIndexByte(body, '\n'); i >= 0 {
		tail, doc = body[i+1:], out[:i+1]
	}
	last := strings.TrimSpace(tail)
	if strings.HasPrefix(last, "# TOKENS ") {
		rest := strings.TrimSpace(strings.TrimPrefix(last, "# TOKENS "))
		if rest == "n/a" {
			return doc, 0, true, false
		}
		var tokens int64
		if _, err := fmt.Sscanf(last, "# TOKENS %d", &tokens); err == nil {
			return doc, tokens, true, true
		}
	}
	return out, 0, false, false
}

// ExtractDelta pulls the YAML document out of an executor reply:
// unwraps code fences, takes the rest as is. The document's
// trailing bytes are content (block scalars); only the envelope
// wrappers are cut.
func ExtractDelta(out string) string {
	s := strings.TrimSpace(out)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```yaml")
		s = strings.TrimPrefix(s, "```yml")
		s = strings.TrimPrefix(s, "```")
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = s[:i]
		}
	}
	return ensureTrailingBreak(strings.TrimLeft(s, " \t\r\n"))
}

// ensureTrailingBreak guarantees one final newline without touching
// the other bytes: trailing \n\n is author content (the keep
// chomping of block scalars); the envelope does not cut it.
func ensureTrailingBreak(s string) string {
	if s == "" || strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}
