// Package zone — the layout law of the machine's zone: the one
// directory that holds every file the machine owns, and the law of
// the paths inside a document. The zone lives at the root of the
// working tree the machine verifies; everything the machine reads
// or writes is a file of this layout or a file of that tree under a
// root-relative path.
package zone

import (
	"errors"
	"path/filepath"
	"strings"
)

// Dirname — the zone directory's name inside the working root.
const Dirname = ".punchtape"

// The zone's files, relative to the zone directory.
const (
	ContractFile   = "contract.md"
	ContractRUFile = "contract-ru.md"
	ContractZHFile = "contract-zh.md"
	TemplateFile   = "template.yaml"
	WishSlot       = "wish.md"
	InboxSlot      = "inbox.yaml"
	EnvSlot        = "env"
	CorpusFile     = "corpus.yaml"
	StateFile      = "state.md"
	ReportFile     = "report.md"
	KnowledgeFile  = "knowledge.md"
	SpecDir        = "spec"
	JournalDir     = "journal"
	ObjectsDir     = "journal/objects"
	CacheDir       = "cache"
	TmpDir         = "tmp"
)

// Path — a file of the zone as a root-relative slash path.
func Path(elem ...string) string {
	return Dirname + "/" + filepath.Join(elem...)
}

// ErrBadPath — the refusal family of the path law: the path is not
// a legal root-relative path.
var ErrBadPath = errors.New("path must be relative to the working root")

// ValidateRel — the path law of documents and effects: a path is a
// slash path relative to the working root — no drive, no scheme, no
// leading slash, no parent hop, no empty or dot-only segment. The
// law is the whole defense: what cannot be spelled cannot point
// outside the tree.
func ValidateRel(p string) error {
	if p == "" {
		return ErrBadPath
	}
	if strings.Contains(p, "\x00") {
		return ErrBadPath
	}
	if strings.HasPrefix(p, "/") || strings.Contains(p, "\\") {
		return ErrBadPath
	}
	for _, seg := range strings.Split(p, "/") {
		if seg == "" || seg == "." || seg == ".." {
			return ErrBadPath
		}
	}
	return nil
}
