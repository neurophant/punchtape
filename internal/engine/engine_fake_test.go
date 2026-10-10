package engine

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/neurophant/punchtape/internal/execport"
	"github.com/neurophant/punchtape/internal/fsport"
)

// The engine unit stack: the whole machine — store, journals, writer,
// checks, receipts, spawns — runs over the in-memory filesystem and
// the scripted process port. No real disk, no real process, the clock
// is the engine seam. Mock code lives in this test file only.

var errIO = errors.New("io failure")

// --- the in-memory filesystem ---

type memFS struct {
	files       map[string][]byte
	mtimes      map[string]time.Time
	dirs        map[string]bool
	nextTmp     int
	locked      map[string]bool
	appendErr   error  // an injected append failure (the read-only ledger)
	writeErr    error  // an injected atomic-write failure (a read-only zone)
	readErr     error  // an injected read failure (an unreadable zone)
	readErrPath string // the read knob's target path: empty = every read fails
	mu          sync.Mutex

	// the precise fault knobs: a failure aimed at one call or one
	// path, for the branches a global knob cannot reach (an append
	// after the Nth, one zone's write, one lock, one walk child)
	appendCalls   int    // the append counter behind appendFailAt
	appendFailAt  int    // 1-based: the append call that fails; 0 = off
	writeErrPath  string // atomic writes of exactly this path fail
	writeFileErr  error  // every WriteFile fails (the seeded surface)
	readdirErr    error  // every ReadDir fails
	walkErr       error  // the walk refuses at its root
	walkChildErr  error  // the walk errors at one non-root entry
	failWalkAt    string // the entry the walk errors at
	lockErr       error  // every lock attempt fails
	tmpErr        error  // MkdirTemp fails
	removeErr     error  // Remove and RemoveAll fail
	mkdirErr      error  // MkdirAll fails
	removeErrPath string
}

type memInfo struct {
	name string
	dir  bool
	mod  time.Time
}

func (m memInfo) Name() string       { return m.name }
func (m memInfo) Size() int64        { return 0 }
func (m memInfo) Mode() fs.FileMode  { return 0o644 }
func (m memInfo) ModTime() time.Time { return m.mod }
func (m memInfo) IsDir() bool        { return m.dir }
func (m memInfo) Sys() any           { return nil }

type memEntry struct {
	name string
	dir  bool
}

func (e memEntry) Name() string               { return e.name }
func (e memEntry) IsDir() bool                { return e.dir }
func (e memEntry) Type() fs.FileMode          { return 0 }
func (e memEntry) Info() (fs.FileInfo, error) { return memInfo{name: e.name, dir: e.dir}, nil }

func notExist(path string) error {
	return &fs.PathError{Op: "open", Path: path, Err: fs.ErrNotExist}
}

func (m *memFS) hasDir(dir string) bool {
	if m.dirs[dir] {
		return true
	}
	for name := range m.files {
		if strings.HasPrefix(name, dir+"/") {
			return true
		}
	}
	return false
}

func (m *memFS) base(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

func (m *memFS) Stat(path string) (os.FileInfo, error) {
	if _, ok := m.files[path]; ok {
		return memInfo{name: m.base(path), dir: false, mod: m.mtimes[path]}, nil
	}
	if m.hasDir(path) {
		return memInfo{name: m.base(path), dir: true, mod: time.Now()}, nil
	}
	return nil, notExist(path)
}

func (m *memFS) MkdirAll(dir string) error {
	if m.mkdirErr != nil {
		return m.mkdirErr
	}
	m.dirs[dir] = true
	return nil
}

func (m *memFS) ReadFile(path string) ([]byte, error) {
	if (m.readErr != nil || m.readErrPath != "") && (m.readErrPath == "" || m.readErrPath == path) {
		if m.readErr != nil {
			return nil, m.readErr
		}
		return nil, errIO
	}
	if data, ok := m.files[path]; ok {
		return data, nil
	}
	// a file in the ancestor chain: reading under a file is a real
	// error on a real host (ENOTDIR), never a missing directory
	rest := strings.TrimPrefix(path, "/")
	for i := 0; i < len(rest); i++ {
		if rest[i] != '/' {
			continue
		}
		if _, isFile := m.files["/"+rest[:i]]; isFile {
			return nil, &fs.PathError{Op: "open", Path: path, Err: errIO}
		}
	}
	return nil, notExist(path)
}

func (m *memFS) ReadDir(dir string) ([]os.DirEntry, error) {
	if m.readdirErr != nil {
		return nil, m.readdirErr
	}
	seen := map[string]bool{}
	var out []os.DirEntry
	add := func(name string, dir bool) {
		if !seen[name] {
			seen[name] = true
			out = append(out, memEntry{name, dir})
		}
	}
	for d := range m.dirs {
		if strings.HasPrefix(d, dir+"/") {
			rest := d[len(dir)+1:]
			if rest != "" {
				if i := strings.IndexByte(rest, '/'); i >= 0 {
					add(rest[:i], true)
				}
			}
		}
	}
	for name := range m.files {
		if !strings.HasPrefix(name, dir+"/") {
			continue
		}
		rest := name[len(dir)+1:]
		if rest == "" {
			continue
		}
		child, isDir := rest, false
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			child, isDir = rest[:i], true
		}
		add(child, isDir)
	}
	if _, isFile := m.files[dir]; isFile {
		// reading a file as a directory is a real error (ENOTDIR on
		// a real host), not a missing directory
		return nil, &fs.PathError{Op: "readdir", Path: dir, Err: errIO}
	}
	if len(out) == 0 && !m.hasDir(dir) {
		return nil, notExist(dir)
	}
	return out, nil
}

func (m *memFS) WriteAtomic(path string, data []byte) error {
	if m.writeErr != nil {
		return m.writeErr
	}
	if m.writeErrPath != "" && m.writeErrPath == path {
		return errIO
	}
	m.files[path] = data
	m.mtimes[path] = timeNow()
	return nil
}

func (m *memFS) AppendFile(path string, data []byte) error {
	if m.appendErr != nil {
		return m.appendErr
	}
	m.appendCalls++
	if m.appendFailAt > 0 && m.appendCalls >= m.appendFailAt {
		return errIO
	}
	m.files[path] = append(m.files[path], data...)
	m.mtimes[path] = timeNow()
	return nil
}

func (m *memFS) Remove(path string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	if m.removeErrPath != "" && path == m.removeErrPath {
		return errIO
	}
	delete(m.files, path)
	delete(m.dirs, path)
	return nil
}

func (m *memFS) Lstat(path string) (os.FileInfo, error) { return m.Stat(path) }

func (m *memFS) Readlink(string) (string, error) { return "", notExist("link") }

func (m *memFS) Symlink(target, link string) error { return notExist(link) }

func (m *memFS) MkdirTemp(dir, pattern string) (string, error) {
	if m.tmpErr != nil {
		return "", m.tmpErr
	}
	m.nextTmp++
	base := dir
	if base == "" {
		base = "/tmp"
	}
	path := base + "/" + strings.ReplaceAll(pattern, "*", "") + itoaMem(m.nextTmp)
	m.dirs[path] = true
	return path, nil
}

func itoaMem(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func (m *memFS) WriteFile(path string, data []byte, _ os.FileMode) error {
	if m.writeFileErr != nil {
		return m.writeFileErr
	}
	m.files[path] = data
	m.mtimes[path] = timeNow()
	return nil
}

func (m *memFS) Chmod(string, os.FileMode) error { return nil }

func (m *memFS) RemoveAll(path string) error {
	if m.removeErr != nil {
		return m.removeErr
	}
	for name := range m.files {
		if name == path || strings.HasPrefix(name, path+"/") {
			delete(m.files, name)
		}
	}
	for d := range m.dirs {
		if d == path || strings.HasPrefix(d, path+"/") {
			delete(m.dirs, d)
		}
	}
	return nil
}

func (m *memFS) Walk(root string, fn filepath.WalkFunc) error {
	if m.walkErr != nil {
		return m.walkErr
	}
	if _, ok := m.files[root]; !ok && !m.hasDir(root) {
		return fn(root, memInfo{name: m.base(root)}, notExist(root))
	}
	paths := []string{root}
	for d := range m.dirs {
		if strings.HasPrefix(d, root+"/") {
			paths = append(paths, d)
		}
	}
	for name := range m.files {
		if name != root && strings.HasPrefix(name, root+"/") {
			paths = append(paths, name)
		}
	}
	sort.Slice(paths, func(i, j int) bool {
		if (paths[i] == root) != (paths[j] == root) {
			return paths[i] == root
		}
		return paths[i] < paths[j]
	})
	for i := 0; i < len(paths); i++ {
		p := paths[i]
		isDir := p == root || m.dirs[p]
		if p != root && m.failWalkAt == p {
			if err := fn(p, memInfo{name: m.base(p), dir: isDir}, m.walkChildErr); err != nil {
				return err
			}
			continue
		}
		if err := fn(p, memInfo{name: m.base(p), dir: isDir, mod: m.mtimes[p]}, nil); err != nil {
			if err == filepath.SkipDir {
				if !isDir {
					continue
				}
				// skip the whole subtree under this directory
				prefix := p + "/"
				for i+1 < len(paths) && strings.HasPrefix(paths[i+1], prefix) {
					i++
				}
				continue
			}
			return err
		}
	}
	return nil
}

func (m *memFS) LockExclusive(path string) (func(), error) {
	if m.lockErr != nil {
		return nil, m.lockErr
	}
	release, held, err := m.TryLockExclusive(path)
	if err != nil {
		return nil, err
	}
	if !held {
		// the blocking form over memory: contention within one test
		// process is a bug, not a state to wait on
		return nil, errIO
	}
	return release, nil
}

func (m *memFS) TryLockExclusive(path string) (func(), bool, error) {
	if m.lockErr != nil {
		return nil, false, m.lockErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locked == nil {
		m.locked = map[string]bool{}
	}
	if m.locked[path] {
		return nil, false, nil
	}
	m.locked[path] = true
	release := func() {
		m.mu.Lock()
		delete(m.locked, path)
		m.mu.Unlock()
	}
	return release, true, nil
}

// --- the scripted process port ---

// procPlan — what a run of one command line answers. Matching: the
// full "name arg..." line first, then the bare name.
type procPort struct {
	mu     sync.Mutex
	calls  int
	plan   map[string]execport.Outcome
	starts [][]string
	// lineCalls — launches per command line (the flip convention's
	// ground: "flip:<line>" answers every SECOND launch of the line,
	// modeling a product whose repeat disagrees)
	lineCalls map[string]int
	// netless — the launches that carried the network-less namespace
	// discipline (the launch mode rides the launch itself).
	netless int
	// fs — the in-memory tree the run directory lives in: a launched
	// script's own bytes tell the fake what it does (a stub written
	// by the coverage probe exits 1; the product script's outcome
	// comes from the plan)
	fs *memFS
}

type procInst struct {
	out execport.Outcome
}

func (p *procInst) Wait(time.Duration) execport.Outcome { return p.out }
func (p *procInst) KillGroup()                          {}

// LookPath — the environment's own baseline tools are on PATH;
// everything else is a name this environment does not know (the
// compositeness candidates and the env doctor depend on it).
func (p *procPort) LookPath(name string) (string, error) {
	switch name {
	case "echo", "cat", "true", "false", "sh":
		return "/bin/" + name, nil
	}
	return "", &fs.PathError{Op: "lookpath", Path: name, Err: fs.ErrNotExist}
}

func (p *procPort) outcomeFor(argv []string) (execport.Outcome, bool) {
	line := strings.Join(argv, " ")
	if out, ok := p.plan[line]; ok {
		return out, true
	}
	name := filepath.Base(argv[0])
	if out, ok := p.plan[name]; ok {
		return out, true
	}
	if out, ok := p.plan[name+" "+strings.Join(argv[1:], " ")]; ok {
		return out, true
	}
	return execport.Outcome{Exit: 0}, false
}

// Start models an honest host. The environment's own PATH tools
// (echo, cat, true, false — the baseline every real host carries)
// answer with their real semantics; everything else is either a file
// the row carried into its run directory (the surface: its own bytes
// tell the fake what it does — an "exit 1" script fails, the probe
// stub included) or a name this environment does not know: an honest
// start failure of the not-found class. A relative argv[0] (the row
// runner's "./name" launch form) resolves against the run directory
// first — the host's own cwd discipline.
func (p *procPort) Start(argv []string, dir string, stdin *string) (execport.Process, error) {
	p.mu.Lock()
	p.calls++
	p.mu.Unlock()
	p.starts = append(p.starts, argv)
	if argv0 := argv[0]; argv0 != "" && argv0[0] != '/' && strings.Contains(argv0, "/") {
		if resolved := filepath.Join(dir, argv0); resolved != argv0 {
			argv = append([]string{resolved}, argv[1:]...)
		}
	}
	if out, ok := p.plan["call2:"+strings.Join(argv, " ")]; ok && p.calls >= 2 {
		return &procInst{out: out}, nil
	}
	// the flip convention: every second launch of the same line
	// answers the flip outcome when one is planned (a product whose
	// repeat disagrees with itself)
	if out, ok := p.plan["flip:"+strings.Join(argv, " ")]; ok {
		p.mu.Lock()
		n := p.lineCalls[strings.Join(argv, " ")]
		p.lineCalls[strings.Join(argv, " ")] = n + 1
		p.mu.Unlock()
		if n%2 == 1 {
			return &procInst{out: out}, nil
		}
	}
	if out, ok := p.baselineTool(argv, dir, stdin); ok {
		return &procInst{out: out}, nil
	}
	if p.fs == nil {
		return &procInst{}, nil
	}
	body, ok := p.fs.files[argv[0]]
	if !ok {
		return nil, &fs.PathError{Op: "exec", Path: argv[0], Err: fs.ErrNotExist}
	}
	script := string(body)
	if !strings.HasPrefix(script, "#!") {
		// no interpreter line and not a binary: the honest host
		// refuses the launch (ENOEXEC — a seeded-over surface file
		// fails its own run)
		return nil, &fs.PathError{Op: "exec", Path: argv[0], Err: errors.New("exec format error")}
	}
	if strings.Contains(script, "exit 1") {
		return &procInst{out: execport.Outcome{Exit: 1}}, nil
	}
	out, _ := p.outcomeFor(argv)
	return &procInst{out: out}, nil
}

// baselineTool — the environment's own tools with real semantics.
func (p *procPort) baselineTool(argv []string, dir string, stdin *string) (execport.Outcome, bool) {
	if strings.ContainsAny(argv[0], "/") {
		return execport.Outcome{}, false
	}
	switch argv[0] {
	case "echo":
		return execport.Outcome{Exit: 0, Stdout: strings.Join(argv[1:], " ") + "\n"}, true
	case "true":
		return execport.Outcome{Exit: 0}, true
	case "false":
		return execport.Outcome{Exit: 1}, true
	case "sh":
		if len(argv) >= 3 && argv[1] == "-c" {
			// The plan first: a scripted product launch rides inside the
			// sh -c string (the field row form — the product is not a
			// literal argv path); the fake answers it as the real shell
			// would have run the product.
			if out, ok := p.outcomeFor(argv); ok {
				return out, true
			}
			return p.shScript(strings.Join(argv[2:], " "), dir), true
		}
		return execport.Outcome{Exit: 0}, true
	case "cat":
		var sb strings.Builder
		for _, name := range argv[1:] {
			path := name
			if !strings.HasPrefix(name, "/") {
				path = dir + "/" + name
			}
			if body, ok := p.fs.files[path]; ok {
				sb.Write(body)
			} else {
				return execport.Outcome{Exit: 1, Stderr: "cat: " + name + ": No such file"}, false
			}
		}
		return execport.Outcome{Exit: 0, Stdout: sb.String()}, true
	}
	return execport.Outcome{}, false
}

// shScript — the echo/exit grammar of the corpus's shell one-liners:
// ';'-separated segments, each an echo (a plain line, a redirect
// into a file of the run directory, the stderr twist) or an exit
// that ends the script. Everything else is an empty success — the
// baseline sh of the environment.
func (p *procPort) shScript(script, dir string) execport.Outcome {
	out := execport.Outcome{}
	for _, part := range strings.Split(script, ";") {
		seg := strings.TrimSpace(part)
		var n int
		if _, err := fmt.Sscanf(seg, "exit %d", &n); err == nil {
			out.Exit = n
			return out
		}
		if !strings.HasPrefix(seg, "echo ") {
			continue
		}
		body := strings.TrimPrefix(seg, "echo ")
		if strings.HasSuffix(body, " 1>&2") {
			out.Stderr += unquote(strings.TrimSpace(strings.TrimSuffix(body, " 1>&2"))) + "\n"
			continue
		}
		if i := strings.Index(body, " > "); i >= 0 {
			text, name := unquote(body[:i]), unquote(strings.TrimSpace(body[i+3:]))
			if p.fs != nil {
				p.fs.files[dir+"/"+name] = []byte(text + "\n")
				p.fs.mtimes[dir+"/"+name] = timeNow()
			}
			continue
		}
		out.Stdout += unquote(strings.TrimSpace(body)) + "\n"
	}
	return out
}

// unquote — one surrounding quote pair of the shell.
func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}

func (p *procPort) StartCtx(_ context.Context, argv []string, dir string, stdin *string) (execport.Process, error) {
	return p.Start(argv, dir, stdin)
}

// StartOpts — Start plus the launch disciplines; the fake records the
// mode (the discipline is a launch fact the caller can assert) and
// answers exactly as Start would: the mode changes nothing the
// scripted plan answers.
func (p *procPort) StartOpts(argv []string, dir string, stdin *string, opts execport.LaunchOpts) (execport.Process, error) {
	if opts.Netns {
		p.mu.Lock()
		p.netless++
		p.mu.Unlock()
	}
	return p.Start(argv, dir, stdin)
}

// sawNetlessLaunch — whether at least one launch carried the mode.
func (p *procPort) sawNetlessLaunch() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.netless > 0
}

func (p *procPort) SpawnDetached(argv []string, dir string) error {
	_, err := p.Start(argv, dir, nil)
	return err
}

func (p *procPort) Combined(ctx context.Context, argv []string, dir string, timeout time.Duration) (int, string, bool, error) {
	proc, err := p.StartCtx(ctx, argv, dir, nil)
	if err != nil {
		return 0, "", false, err
	}
	out := proc.Wait(timeout)
	return out.Exit, out.Stdout + out.Stderr, out.TimedOut, nil
}

// --- the binding helpers ---

type fakeStack struct {
	fs   *memFS
	proc *procPort
	dir  string
}

// lastStack — the most recently bound stack (test convenience).
var lastStack *fakeStack

// bindUnits binds the in-memory stack for one test and returns the
// working directory path. Sleeps are free, and the clock is the seam.
func bindUnits(t *testing.T) *fakeStack {
	t.Helper()
	stack := &fakeStack{
		fs:  &memFS{files: map[string][]byte{}, mtimes: map[string]time.Time{}, dirs: map[string]bool{}},
		dir: "/w",
	}
	stack.proc = &procPort{plan: map[string]execport.Outcome{}, lineCalls: map[string]int{}, fs: stack.fs}
	oldFS, oldProc := fsport.Bind, execport.Bind
	oldSleep, oldSince := sleep, timeSince
	lastStack = stack
	fsport.Bind = stack.fs
	execport.Bind = stack.proc
	sleep = func(time.Duration) {}
	timeSince = func(time.Time) time.Duration { return 0 }
	t.Cleanup(func() {
		fsport.Bind, execport.Bind = oldFS, oldProc
		sleep, timeSince = oldSleep, oldSince
	})
	return stack
}

// walkToVerdict — the standard green walk into a fresh in-memory
// instance; the finished instance is returned opened (the stack stays
// bound for the test's direct use).
func memDir(t *testing.T) string {
	t.Helper()
	memSeqMu.Lock()
	memSeq++
	n := memSeq
	memSeqMu.Unlock()
	path := "/dir" + itoaMem(n)
	if m, ok := fsport.Bind.(*memFS); ok {
		m.dirs[path] = true
	}
	return path
}

var (
	memSeqMu sync.Mutex
	memSeq   int
)
