#!/usr/bin/env node
// notes — a tiny notebook CLI: state in notes.json, zero dependencies.
"use strict";
const fs = require("fs");
const FILE = "notes.json";
const load = () => {
  try {
    return JSON.parse(fs.readFileSync(FILE, "utf8"));
  } catch {
    return [];
  }
};
const save = (notes) => fs.writeFileSync(FILE, JSON.stringify(notes) + "\n");
const notes = load();
const [cmd, arg] = process.argv.slice(2);
if (cmd === "add") {
  notes.push(process.argv.slice(3).join(" "));
  save(notes);
  console.log(notes.length);
} else if (cmd === "list") {
  for (const note of notes) console.log(note);
} else if (cmd === "show") {
  const i = Number(arg);
  if (!Number.isInteger(i) || i < 1 || i > notes.length) {
    console.error("no such note: " + arg);
    process.exit(3);
  }
  console.log(notes[i - 1]);
} else {
  console.error("usage: node notes.js add|list|show");
  process.exit(2);
}
