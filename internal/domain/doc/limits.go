package doc

// The document's own bounds: one submission is a bounded document —
// its bytes, its registry lists, its behaviors' claims and rows, its
// expectations' files. The ceilings keep a single document from
// eating the machine's memory, time and disk; the values are
// measured from the real usage and printed in the reference's
// constants — an honest giant asks the ceiling to move, a runaway
// loop meets it.

// InboxBytes — one document's whole byte ceiling.
const InboxBytes = 512 * 1024

// ClaimsPerBehavior — one behavior's claim ceiling.
const ClaimsPerBehavior = 32

// RowsPerBehavior — one behavior's row ceiling.
const RowsPerBehavior = 16

// ListPerSubmission — one submission's ceiling per registry list:
// intents, notes, requirements, markers, orders, exclusions.
const ListPerSubmission = 64

// FilesPerExpect — one row's expected-file ceiling.
const FilesPerExpect = 8
