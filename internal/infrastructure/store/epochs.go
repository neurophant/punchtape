package store

// The epoch files: an append-only stream lives as numbered files
// epoch-NNNNN.yaml inside its directory — the hot layer stays
// small, the whole history stays derivable. This file owns the
// naming law of the storage layer's streams: one spelling, no
// alternatives.

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"

	"github.com/neurophant/punchtape/internal/ports"
)

// EpochNamePattern — an epoch file's name: the number in five digits
// with the yaml suffix.
var EpochNamePattern = regexp.MustCompile(`^epoch-([0-9]{5})\.yaml$`)

// EpochFileName — the epoch file's name for a number.
func EpochFileName(n int) string {
	return fmt.Sprintf("epoch-%05d.yaml", n)
}

// ListEpochs — the directory's epoch files as full paths in number
// order. A missing directory is no files — a stream that never
// wrote yet.
func ListEpochs(dir string) ([]string, error) {
	entries, err := ports.FSBind.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	type numbered struct {
		n    int
		path string
	}
	var found []numbered
	for _, en := range entries {
		m := EpochNamePattern.FindStringSubmatch(en.Name())
		if m == nil || en.IsDir() {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		found = append(found, numbered{n, filepath.Join(dir, en.Name())})
	}
	sort.Slice(found, func(i, j int) bool { return found[i].n < found[j].n })
	out := make([]string, 0, len(found))
	for _, f := range found {
		out = append(out, f.path)
	}
	return out, nil
}

// NextEpochPath — the path a rotation opens: the number after the
// directory's last epoch (no files — the first). The paths come from
// ListEpochs, so every name carries its number.
func NextEpochPath(paths []string, dir string) string {
	if len(paths) == 0 {
		return filepath.Join(dir, EpochFileName(1))
	}
	n, _ := strconv.Atoi(EpochNamePattern.FindStringSubmatch(filepath.Base(paths[len(paths)-1]))[1])
	return filepath.Join(dir, EpochFileName(n+1))
}
