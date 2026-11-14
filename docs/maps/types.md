# The type sheets

The machine's types as goplantuml reads them from the code;
the source of truth is the code itself.

```plantuml
@startuml
namespace application {
    class "CommandResult" << (S,Aquamarine) >> {
        + Exit int
        + Stdout string
        + Stderr string
        + TimedOut bool
        + Err error

    }
    class "application.refusalList" << (T, #FF7700) >>  {
    }
    class "briefMap" << (S,Aquamarine) >> {
        + Lines int
        + Claimed int
        + Holes []int
        + Requirements int
        + Covered int
        + Markers int

    }
    class "claimRange" << (S,Aquamarine) >> {
        + Claim string
        + Digest12 string

    }
    class "claimsFacts" << (S,Aquamarine) >> {
        - digest string
        - ranges []claimRange

    }
    class "historyEvent" << (S,Aquamarine) >> {
        - run int
        - kind string
        - text string

    }
    class "machine" << (S,Aquamarine) >> {
        - root string
        - cwd string
        - zoneDir string
        - stream *store.Stream
        - objects *cas.Store
        - standing *journal.Standing
        - universe []string
        - covered *fprint.Coverage
        - pins []string
        - rowTimeout time.Duration
        - runBudget time.Duration
        - spent time.Duration
        - completed bool
        - buildFailed bool
        - records []journal.Record
        - submitted *doc.Document
        - execFacts []journal.ExecutionFact
        - treeBefore string
        - warnings []string
        - tree *treeScope

        - ensureAgentsBlock() error
        - checkChain() error
        - demandTail() []string
        - unresolvedLines() []string
        - briefAccounting(corpus doc.Document) briefMap
        - slot(name string) string
        - corpusDoc() doc.Document
        - run(terminal io.Writer) int
        - treeLockArmed(d doc.Document) bool
        - handleInbox(inbox string, d doc.Document, canonical []byte, normErr error) error
        - streamRangeRefusals(d doc.Document) []doc.Refusal
        - askCloseRefusals(d doc.Document, preview doc.Registry) []doc.Refusal
        - demandLine(id string) (string, bool)
        - danglingCovers(d doc.Document, preview doc.Registry) []doc.Refusal
        - accept(d doc.Document, canonical []byte, registry doc.Registry) error
        - append(rec journal.Record) error
        - runExecutions() error
        - claimsFactsOf(b doc.Behavior) claimsFacts
        - claimsMap(corpus doc.Document) <font color=blue>map</font>[string]claimsFacts
        - executeStale(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts, envHash string) ([]journal.RowFact, []rowReport)
        - unpinnedBehaviors(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts, envHash string) []string
        - greenPin(b doc.Behavior, claimsDigest string, envHash string) bool
        - driftNamedBySubmission(since int) bool
        - proven(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts, envHash string) bool
        - lessons(executed []journal.RowFact, greenRun int) []journal.Lesson
        - rowBytesOf(rowDigest string) []byte
        - verdictBlockers(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts, claimsAfter <font color=blue>map</font>[string]claimsFacts) ([]verdictBlocker, error)
        - closeVerdict(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts, claimsAfter <font color=blue>map</font>[string]claimsFacts) ([]string, error)
        - orphanRequirements(corpus doc.Document) []string
        - coverage(corpus doc.Document) (fprint.Coverage, error)
        - uncoveredLines(corpus doc.Document) (string, error)
        - registryDigest() string
        - signature(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts) string
        - finish(terminal io.Writer, instruction string) int
        - buildRedInstruction() string
        - nextInstruction(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts, claimsAfter <font color=blue>map</font>[string]claimsFacts) string
        - readFileOrEmpty(path string) []byte
        - writeState(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts, instruction string) error
        - treeHoles(corpus doc.Document) int
        - writeHoles(cov fprint.Coverage) error
        - writeExclusionsProposal(corpus doc.Document) error
        - writeTrace(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts) error
        - markerReasonOf(n int) string
        - orderOn(n int) (doc.OpenOrder, string)
        - behaviorRows(b doc.Behavior) (int, int, int)
        - countRows(corpus doc.Document) (int, int)
        - writeReport(n int, build *journal.Build, rows []rowReport, notes []string) error
        - reportRefusals(refusals []doc.Refusal) error
        - writeKnowledge() error
        - writeSpec(corpus doc.Document, claims <font color=blue>map</font>[string]claimsFacts) error
        - behaviorStatus(b doc.Behavior) string
        - rowStatus(id string, i int, b doc.Behavior) string
        - writeIfChanged(path string, data []byte) error

        + FileLines(rel string) (int, bool)
        + Hole(claims []string, rel string, first int, last int) bool

    }
    class "refusalList" << (S,Aquamarine) >> {
        + Error() string

    }
    class "rowReport" << (S,Aquamarine) >> {
        + Key string
        + Run doc.Cmd
        + Exit int
        + Stdout string
        + Facts []string

    }
    class "treeEntry" << (S,Aquamarine) >> {
        - size int64
        - mtime int64
        - lines int
        - digest string

    }
    class "treeFacts" << (S,Aquamarine) >> {
        - files []fprint.FileLineCount
        - digest string
        - inside int
        - outside int

    }
    class "treeScope" << (S,Aquamarine) >> {
        - root string
        - roots []string
        - cache <font color=blue>map</font>[string]treeEntry
        - lastFacts treeFacts
        - walked bool

        - declared() bool
        - loadCache(path string) error
        - cacheBytes() []byte
        - walk() (treeFacts, error)

    }
    class "verdictBlocker" << (S,Aquamarine) >> {
        - note string
        - instruction string

    }
    class "[]doc.Refusal" as docRefusal {
        'This class was created so that we can correctly have an alias pointing to this name. Since it contains dots that can break namespaces
    }
}
namespace domain {
}
namespace canon {
    class "LawRef" << (S,Aquamarine) >> {
        + Rule string
        + Law string

    }
    class "Source" << (S,Aquamarine) >> {
        + Contract string
        + ContractRU string
        + ContractZH string
        + Template string
        + Reference string
        + ReferenceRU string
        + ReferenceZH string
        + Example string
        + LawMap []LawRef
        + Env []string

    }
}
namespace doc {
    class "Behavior" << (S,Aquamarine) >> {
        + ID string
        + Name string
        + Prose string
        + Claims []string
        + Covers []string
        + Rows []Row

    }
    class "Cmd" << (S,Aquamarine) >> {
        + Line string
        + Argv []string

        + UnmarshalYAML(u <font color=blue>func</font>(<font color=blue>interface</font>{}) error) error
        + MarshalYAML() ([]byte, error)
        + Key() string
        + IsZero() bool

    }
    class "Document" << (S,Aquamarine) >> {
        + Done bool
        + Intents []Intent
        + Work string
        + Build *string
        + Remove []string
        + Requirements []Requirement
        + Markers []Marker
        + Exclusions []Exclusion
        + Open []OpenOrder
        + Behaviors []Behavior
        + Notes []Note
        + Mandate []MandateRef

        + BuildCmd() string

    }
    class "Exclusion" << (S,Aquamarine) >> {
        + Range string
        + Reason string
        + Refs []string

    }
    class "Expect" << (S,Aquamarine) >> {
        + RC *int
        + Out string
        + Files <font color=blue>map</font>[string]string

    }
    class "Intent" << (S,Aquamarine) >> {
        + Lines []string
        + Outcome string
        + Prose string
        + Text string
        + Run Cmd

    }
    class "MandateRef" << (S,Aquamarine) >> {
        + Line string
        + Target string

    }
    class "Marker" << (S,Aquamarine) >> {
        + Lines []string
        + Reason string

    }
    class "Note" << (S,Aquamarine) >> {
        + Kind string
        + Text string
        + Refs []string

    }
    class "OpenOrder" << (S,Aquamarine) >> {
        + Demand string
        + Question string
        + State string
        + Reason string
        + OpenedAtLines int

    }
    class "RangeError" << (S,Aquamarine) >> {
        + Line int
        + Max int

        + Error() string

    }
    class "Refusal" << (S,Aquamarine) >> {
        + Where string
        + Law string
        + Fact string

        + Error() string

    }
    class "Registry" << (S,Aquamarine) >> {
        + Requirements <font color=blue>map</font>[string]Requirement
        + Order []string
        + Markers []Marker
        + Exclusions []Exclusion
        + Open []OpenOrder

    }
    class "Removals" << (S,Aquamarine) >> {
        + Requirements <font color=blue>map</font>[string]bool
        + Lines <font color=blue>map</font>[int]bool
        + Ranges <font color=blue>map</font>[string]bool

    }
    class "Requirement" << (S,Aquamarine) >> {
        + ID string
        + Lines []string
        + Prose string
        + Refines string

    }
    class "Resolvers" << (S,Aquamarine) >> {
        + Stands <font color=blue>func</font>(string) bool
        + Requirement <font color=blue>func</font>(string) bool
        + DemandLine <font color=blue>func</font>(string) (string, bool)
        + FileLines <font color=blue>func</font>(string) (int, bool)
        + Law <font color=blue>func</font>(string) bool
        + Hole <font color=blue>func</font>(string, int, int) bool
        + MarkerStands <font color=blue>func</font>(int) bool
        + OrderStands <font color=blue>func</font>(int) bool
        + ExclusionStands <font color=blue>func</font>(string) bool
        + RequirementLines <font color=blue>func</font>(string) ([]int, bool)
        + RequirementIDs <font color=blue>func</font>() []string
        + DemandsTailLines <font color=blue>func</font>() []int
        + UniverseSize <font color=blue>func</font>() int
        + LineOutcome <font color=blue>func</font>(string) (string, bool)
        + RunsInUse <font color=blue>func</font>(string) (string, bool)

    }
    class "Row" << (S,Aquamarine) >> {
        + Run Cmd
        + Expect Expect

    }
    class "internal.domain.doc.RemoveKeyKind" << (T, #FF7700) >>  {
    }
}
namespace fprint {
    class "Coverage" << (S,Aquamarine) >> {
        + Total int
        + Claimed int
        + Holes <font color=blue>map</font>[string][]Range
        + Excluded []FileExclusion
        + ExcludedLines int

    }
    class "EnvIdentity" << (S,Aquamarine) >> {
        + Version string
        + Arch string
        + Kernel string
        + Env []string

    }
    class "FileDigest" << (S,Aquamarine) >> {
        + Path string
        + Digest string

    }
    class "FileExclusion" << (S,Aquamarine) >> {
        + Path string
        + Reason string

    }
    class "FileLineCount" << (S,Aquamarine) >> {
        + Path string
        + Lines int
        + Outside string

    }
    class "Range" << (S,Aquamarine) >> {
        + First int
        + Last int

    }
}
namespace journal {
    class "BehaviorOp" << (S,Aquamarine) >> {
        + ID string
        + Addr string

    }
    class "Birth" << (S,Aquamarine) >> {
        + Addr string
        + Digest string
        + Version string

    }
    class "Build" << (S,Aquamarine) >> {
        + Exit int
        + TimedOut bool
        + Reason string

    }
    class "Completion" << (S,Aquamarine) >> {
        + Run int

    }
    class "DemandAppend" << (S,Aquamarine) >> {
        + Lines int
        + Chain string

    }
    class "ExecutionFact" << (S,Aquamarine) >> {
        + At int
        + Lines []string
        + Run doc.Cmd
        + Exit int
        + StdoutDigest string
        + Stdout string
        + StdoutAddr string

    }
    class "Lesson" << (S,Aquamarine) >> {
        + Reason string
        + Key string
        + Facts []string
        + RedRow string
        + GreenRow string
        + RedRun int
        + GreenRun int

    }
    class "LineOutcome" << (S,Aquamarine) >> {
        + Outcome string
        + Prose string
        + At int

    }
    class "NoNewDemands" << (S,Aquamarine) >> {
        + Lines int
        + Run int

    }
    class "Record" << (S,Aquamarine) >> {
        + Kind string
        + Birth *Birth
        + Submission *Submission
        + DemandAppend *DemandAppend
        + NoNewDemands *NoNewDemands
        + Run *Run
        + Verdict *Verdict
        + Refusals []doc.Refusal
        + Completed *Completion
        + Execution *ExecutionFact
        + TreeLock *TreeLock

    }
    class "RowFact" << (S,Aquamarine) >> {
        + Key string
        + Row string
        + Status string
        + Exit int
        + Facts []string
        + Proof string

    }
    class "Run" << (S,Aquamarine) >> {
        + N int
        + Build *Build
        + Rows []RowFact
        + Lessons []Lesson

    }
    class "Standing" << (S,Aquamarine) >> {
        + Born bool
        + BirthVersion string
        + ChainAppends []DemandAppend
        + ChainHead string
        + PastLines int
        + LastNoNewDemands int
        + CorpusVersion int
        + Build string
        + Registry doc.Registry
        + Behaviors <font color=blue>map</font>[string]doc.Behavior
        + Order []string
        + Rows <font color=blue>map</font>[string]RowFact
        + RunNumber int
        + Verdict *Verdict
        + Accepted int
        + Refused int
        + Completions int
        + Lessons []Lesson
        + LineOutcomes <font color=blue>map</font>[int]LineOutcome
        + AskedLines <font color=blue>map</font>[int]bool
        + Answers <font color=blue>map</font>[int]string
        + Executions <font color=blue>map</font>[int]doc.Cmd
        + LastExecution *ExecutionFact
        + TreeLockViolation *TreeLock
        + LastSubmissionAt int
        + Notes []doc.Note
        + RowBytes <font color=blue>map</font>[string]string

    }
    class "Submission" << (S,Aquamarine) >> {
        + CorpusVersion int
        + Build *string
        + Intents []doc.Intent
        + Notes []doc.Note
        + Requirements []doc.Requirement
        + Markers []doc.Marker
        + Exclusions []doc.Exclusion
        + Open []doc.OpenOrder
        + Added []BehaviorOp
        + Removed []string
        + DocAddr string

    }
    class "TreeLock" << (S,Aquamarine) >> {
        + Before string
        + After string

    }
    class "Verdict" << (S,Aquamarine) >> {
        + N int
        + CorpusVersion int
        + Signature string
        + RegistryDigest string
        + Scopes []VerdictScope
        + Run int

    }
    class "VerdictScope" << (S,Aquamarine) >> {
        + ID string
        + Pointer string

    }
}
namespace zone {
}
namespace fsys {
    class "OS" << (S,Aquamarine) >> {
        + Flock(path string) (<font color=blue>func</font>() error, error)
        + Flock(path string) (<font color=blue>func</font>() error, error)
        + Stat(path string) (os.FileInfo, error)
        + MkdirAll(dir string) error
        + Mkdir(dir string) error
        + ReadFile(path string) ([]byte, error)
        + ReadDir(dir string) ([]os.DirEntry, error)
        + WriteAtomic(path string, data []byte) error
        + AppendFile(path string, data []byte) error
        + Remove(path string) error
        + RemoveAll(path string) error
        + Walk(root string, fn filepath.WalkFunc) error

    }
}
namespace process {
    class "OS" << (S,Aquamarine) >> {
        + StartOpts(argv []string, dir string, opts ports.LaunchOpts) (ports.Process, error)
        + StartOpts(argv []string, dir string, opts ports.LaunchOpts) (ports.Process, error)

    }
    class "osProcess" << (S,Aquamarine) >> {
        - cmd *exec.Cmd
        - stdout *bytes.Buffer
        - stderr *bytes.Buffer
        - once sync.Once

        + Wait(timeout time.Duration) ports.Outcome
        + KillGroup() 

    }
}
namespace store {
    class "Stream" << (S,Aquamarine) >> {
        - dir string
        - files []string

        - loadFile(path string) ([]journal.Record, error)
        - lastIsEmpty() bool

        + Records() ([]journal.Record, error)
        + FirstRecord() (journal.Record, error)
        + Append(rec journal.Record) error
        + Rotate() error

    }
    namespace cas {
    class "Store" << (S,Aquamarine) >> {
        - root string

        - path(addr string) string

        + Put(data []byte) (string, error)
        + Get(addr string) ([]byte, error)

    }
    }
}
namespace ports {
    interface "Exec"  {
        + StartOpts(argv []string, dir string, opts LaunchOpts) (Process, error)

    }
    interface "FS"  {
        + Stat(path string) (os.FileInfo, error)
        + MkdirAll(dir string) error
        + Mkdir(dir string) error
        + ReadFile(path string) ([]byte, error)
        + ReadDir(dir string) ([]os.DirEntry, error)
        + WriteAtomic(path string, data []byte) error
        + AppendFile(path string, data []byte) error
        + Remove(path string) error
        + RemoveAll(path string) error
        + Walk(root string, fn filepath.WalkFunc) error
        + Flock(path string) (<font color=blue>func</font>() error, error)

    }
    class "LaunchOpts" << (S,Aquamarine) >> {
        + Env []string

    }
    class "Outcome" << (S,Aquamarine) >> {
        + Exit int
        + Stdout string
        + Stderr string
        + WaitErr string
        + TimedOut bool

    }
    interface "Process"  {
        + Wait(timeout time.Duration) Outcome
        + KillGroup() 

    }
    class "unwiredExec" << (S,Aquamarine) >> {
        + StartOpts( []string,  string,  LaunchOpts) (Process, error)

    }
    class "unwiredFS" << (S,Aquamarine) >> {
        + Stat( string) (os.FileInfo, error)
        + MkdirAll( string) error
        + Mkdir( string) error
        + ReadFile( string) ([]byte, error)
        + ReadDir( string) ([]os.DirEntry, error)
        + WriteAtomic( string,  []byte) error
        + AppendFile( string,  []byte) error
        + Remove( string) error
        + RemoveAll( string) error
        + Walk( string,  filepath.WalkFunc) error
        + Flock( string) (<font color=blue>func</font>() error, error)

    }
}
"__builtin__.int" #.. "internal.domain.doc.RemoveKeyKind"
"application.docRefusal" #.. "application.refusalList"

"ports.Exec" <|-- "ports.unwiredExec"
"ports.FS" <|-- "ports.unwiredFS"


@enduml
```
