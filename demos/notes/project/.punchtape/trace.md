# punchtape trace — run 2

D1 «Make a notes utility: notes add <text> saves a note and prints its number;»
├─ intent: work — «notes add saves a note and prints its number»
│  └─ R1 (span D1)
│     └─ notes/add (covers R1)
│        claims: notes.js:1-19 — token e89f52e44429
│        row 0 PROOF: rm -f notes.json && node notes.js add hello && node notes.js add world → green
D2 «notes list prints the notes one per line, oldest first;»
├─ intent: work — «notes list prints the notes oldest first»
│  └─ R2 (span D2)
│     └─ notes/list (covers R2)
│        claims: notes.js:20-21 — token 46832763864d
│        row 0 PROOF: rm -f notes.json && node notes.js add alpha && node notes.js add beta && node notes.js list → green
D3 «notes show <number> prints one note;»
├─ intent: work — «notes show prints one note»
│  └─ R3 (span D3)
│     └─ notes/show (covers R3)
│        claims: notes.js:22-29 — token 128de6075a9a
│        row 0 PROOF: rm -f notes.json && node notes.js add first && node notes.js show 1 → green
D4 «state lives in notes.json and survives a restart.»
├─ intent: work — «state lives in notes.json and survives a restart»
│  └─ R4 (span D4)
│     └─ notes/persist (covers R4)
│        claims: notes.js:30-32 — token db3428fe8d95
│        row 0 PROOF: rm -f notes.json && node notes.js add pinned && node notes.js list → green
