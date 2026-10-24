// The wish-line range grammar's battery: the parse of every legal
// shape, the refusal of every broken one, and the collapse of number
// sets back into the same spelling.
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
		{[]string{"W3"}, "3"},
		{[]string{"W2..W4"}, "2 3 4"},
		{[]string{"W1..W2, W5"}, "1 2 5"},
		{[]string{"W1..W3", "W3..W5"}, "1 2 3 4 5"},
		{[]string{"W1,W2", " W7 "}, "1 2 7"},
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
		"W5..W2", "W0", "W03", "w3", "W-1", "W+3", "W", "W3..", "..W5",
		"W3...W5", "W3..W5..W7", "W3,,W5", "", "W99999999999999999999",
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
		{[]int{19}, "W19"},
		{[]int{23, 24}, "W23, W24"},
		{[]int{19, 23, 24, 25}, "W19, W23..W25"},
		{[]int{1, 2, 3, 4, 5}, "W1..W5"},
		{[]int{5, 1, 3, 2, 5}, "W1..W3, W5"},
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
