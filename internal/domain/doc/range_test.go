// The demand stream's vocabulary battery: the line universe's own
// law, the items' flattening into it, and the range grammar over it
// — the parse of every legal shape, the refusal of every broken
// one, and the collapse of number sets back into the same spelling.
package doc

import (
	"strconv"
	"strings"
	"testing"
)

func TestParseRangesEveryShape(t *testing.T) {
	cases := []struct {
		ranges []string
		want   string
	}{
		{[]string{"D3"}, "3"},
		{[]string{"D2..D4"}, "2 3 4"},
		{[]string{"D1..D2, D5"}, "1 2 5"},
		{[]string{"D1..D3", "D3..D5"}, "1 2 3 4 5"},
		{[]string{"D1,D2", " D7 "}, "1 2 7"},
		{nil, ""},
	}
	for _, c := range cases {
		nums, err := ParseRanges(c.ranges)
		if err != nil {
			t.Fatalf("%v: %v", c.ranges, err)
		}
		spelled := make([]string, len(nums))
		for i, n := range nums {
			spelled[i] = strconv.Itoa(n)
		}
		if got := strings.Join(spelled, " "); got != c.want {
			t.Fatalf("%v parsed to %q, want %q", c.ranges, got, c.want)
		}
	}
}

func TestParseRangesRefusesBrokenSpellings(t *testing.T) {
	for _, broken := range []string{
		"D5..D2", "D0", "D03", "d3", "D-1", "D+3", "D", "D3..", "..D5",
		"D3...D5", "D3..D5..D7", "D3,,D5", "", "D99999999999999999999",
	} {
		if _, err := ParseRanges([]string{broken}); err == nil {
			t.Fatalf("%q must be refused", broken)
		}
	}
}

func TestSpellRangesCollapsesRuns(t *testing.T) {
	cases := []struct {
		nums []int
		want string
	}{
		{nil, ""},
		{[]int{19}, "D19"},
		{[]int{23, 24}, "D23, D24"},
		{[]int{19, 23, 24, 25}, "D19, D23..D25"},
		{[]int{1, 2, 3, 4, 5}, "D1..D5"},
		{[]int{5, 1, 3, 2, 5}, "D1..D3, D5"},
	}
	for _, c := range cases {
		if got := SpellRanges(c.nums); got != c.want {
			t.Fatalf("%v spelled %q, want %q", c.nums, got, c.want)
		}
	}
}

func TestSpellAndParseAreOneGrammar(t *testing.T) {
	nums := []int{2, 3, 4, 9, 11}
	back, err := ParseRanges([]string{SpellRanges(nums)})
	if err != nil {
		t.Fatal(err)
	}
	if len(back) != len(nums) {
		t.Fatalf("the round trip lost lines: %v", back)
	}
	for i := range nums {
		if back[i] != nums[i] {
			t.Fatalf("the round trip moved a line: %v", back)
		}
	}
}

func TestNonEmptyLinesNumberTheNonEmptyInFileOrder(t *testing.T) {
	if got := NonEmptyLines("\nfirst\n\n  \nsecond\r\n\nthird\n\n"); strings.Join(got, "|") != "first|second|third" {
		t.Fatalf("the universe is the non-empty trimmed lines in file order: %q", got)
	}
	if all := NonEmptyLines("only one line"); len(all) != 1 || all[0] != "only one line" {
		t.Fatalf("a stream without blanks numbers whole: %q", all)
	}
	if empty := NonEmptyLines(""); len(empty) != 0 {
		t.Fatalf("an empty stream numbers nothing: %q", empty)
	}
}
