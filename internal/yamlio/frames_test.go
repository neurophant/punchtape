// Frame units: the split over separator lines, the offset/rebuild
// pair, the torn-tail report — and the format-critical pins of the
// YAML era: a marshaled document never opens a column-zero "---"
// line of its own, so the split is exact without parsing.
package yamlio

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestSplitFrames(t *testing.T) {
	cases := []struct {
		name         string
		stream       string
		wantPreamble string
		wantFrames   []string
	}{
		{"a stream without separators is all preamble",
			"# header\nsecond\n", "# header\nsecond\n", nil},
		{"an empty stream", "", "", nil},
		{"a preamble and two frames",
			"# head\n---\na: 1\n---\nlist:\n  - one\n  - two\n",
			"# head\n", []string{"a: 1\n", "list:\n  - one\n  - two\n"}},
		{"a frame without a trailing newline",
			"---\na: 1", "", []string{"a: 1"}},
		{"a bare separator at EOF opens one empty frame",
			"pre\n---", "pre\n", []string{""}},
		{"a block-scalar frame stays one frame",
			"---\ntext: |\n  a\n  b\n", "", []string{"text: |\n  a\n  b\n"}},
		{"an indented --- line is not a separator",
			"---\na: 1\n---\ntext: |\n  ---\n  not a doc boundary\n",
			"", []string{"a: 1\n", "text: |\n  ---\n  not a doc boundary\n"}},
		{"an empty frame is kept, not dropped",
			"# c\n---\n---\na: 1\n", "# c\n", []string{"", "a: 1\n"}},
		{"a stream of separators only is all empty frames",
			"---\n---\n", "", []string{"", ""}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			preamble, frames := SplitFrames([]byte(tc.stream))
			if string(preamble) != tc.wantPreamble {
				t.Fatalf("preamble:\n got %q\nwant %q", preamble, tc.wantPreamble)
			}
			if len(frames) != len(tc.wantFrames) {
				t.Fatalf("frame count: got %d (%q), want %d", len(frames), frames, len(tc.wantFrames))
			}
			for i, fr := range frames {
				if string(fr) != tc.wantFrames[i] {
					t.Fatalf("frame %d:\n got %q\nwant %q", i, fr, tc.wantFrames[i])
				}
			}
		})
	}
}

// The one-line and split-exact laws of the marshaled frame: a plain
// multiline string rides an indented block scalar (a line that IS
// "---" after its own indentation lands deeper still, never at
// column zero); a string a block scalar cannot carry — one that
// starts with a space — rides double-quoted \n escapes on ONE line.
// Either way no marshal can open a separator inside itself.
func TestMarshalStaysOneFrame(t *testing.T) {
	block, err := Marshal(map[string]string{"text": "line \"one\"\n  ---\nplain --- line\nsnowman \u2603\n"})
	wantNoErr(t, "marshal block", err)
	quoted, err := Marshal(map[string]string{"text": " line one\nthe \"quotes\"\n  ---\nsnowman \u2603\tctrl \x01 cr \r"})
	wantNoErr(t, "marshal quoted", err)

	if n := bytes.Count(quoted, []byte("\n")); n != 1 {
		t.Fatalf("a double-quoted value must marshal on one line: %d newlines in %q", n, quoted)
	}
	if !strings.Contains(string(quoted), `\n`) || !strings.Contains(string(quoted), `\"`) {
		t.Fatalf("the value must ride quoted escapes: %q", quoted)
	}

	for name, image := range map[string][]byte{"block": block, "quoted": quoted} {
		if _, frames := SplitFrames(image); len(frames) != 0 {
			t.Fatalf("%s marshal leaked a column-zero separator: %q -> %q", name, image, frames)
		}
	}

	stream := Rebuild([]byte("# journal\n"), [][]byte{block, quoted})
	preamble, frames := SplitFrames(stream)
	if string(preamble) != "# journal\n" || len(frames) != 2 ||
		string(frames[0]) != string(block) || string(frames[1]) != string(quoted) {
		t.Fatalf("marshaled frames drifted: %q %q", preamble, frames)
	}
}

func TestFrameRebuildRoundtrip(t *testing.T) {
	stream := []byte("# head\n---\na: 1\n---\nb: 2\n")
	preamble, frames := SplitFrames(stream)
	if got := Rebuild(preamble, frames); !reflect.DeepEqual(got, stream) {
		t.Fatalf("Rebuild: %q", got)
	}
	if got := FrameOffset(preamble, frames, 0); got != len("# head\n") {
		t.Fatalf("FrameOffset of the first frame: %d", got)
	}
	if got := FrameOffset(preamble, frames, 1); got != len("# head\n---\na: 1\n") {
		t.Fatalf("FrameOffset of the second frame: %d", got)
	}
	if got := FrameOffset(preamble, frames, len(frames)); got != len(stream) {
		t.Fatalf("FrameOffset past the last frame: %d", got)
	}
	if !bytes.Equal(stream[FrameOffset(preamble, frames, 1):][:len(separator)], separator) {
		t.Fatal("every frame rides behind its separator line")
	}
}

func TestTornTailError(t *testing.T) {
	tt := &TornTail{Good: 3, Offset: 40, Length: 7}
	got := tt.Error()
	if !strings.Contains(got, "after 3 records") || !strings.Contains(got, "7 bytes") || !strings.Contains(got, "offset 40") {
		t.Fatalf("TornTail text: %q", got)
	}
}
