# Aggregates the call edges to the package level: one node per package,
# grouped in its architecture layer, one edge per package pair with the
# call count. This is the call map that fits a page — the function
# level lives in calls.md. Input: the sorted internal edge list.
BEGIN {
	print "digraph packages {"
	print "  graph [fontname=\"monospace\", rankdir=TB, margin=0.3, nodesep=0.6, ranksep=0.7, splines=ortho]"
	print "  node [shape=box, fontname=\"monospace\", fontsize=11]"
	print "  edge [fontname=\"monospace\", fontsize=9, arrowsize=0.7]"
}
{
	line = $0
	sub(/^[ \t]+/, "", line)
	p = index(line, "\" -> \"")
	if (!p) next
	a = substr(line, 2, p - 2)
	rest = substr(line, p + 6)
	q = index(rest, "\"")
	if (!q) next
	b = substr(rest, 1, q - 1)
	pa = pkgof(a)
	pb = pkgof(b)
	if (pa == pb) next
	k = pa "\t" pb
	cnt[k]++
	if (!(pa in seen)) { seen[pa] = 1; shortp(pa) }
	if (!(pb in seen)) { seen[pb] = 1; shortp(pb) }
}
function shortp(p) {
	lab = p
	sub(/^github\.com\/neurophant\/punchtape\//, "", lab)
	sub(/^internal\//, "", lab)
	lay = lab
	sub(/\/.*$/, "", lay)
	if (lay == "cmd") lay = "the composition root"
	if (lay == "application") lay = "the application (the machine)"
	if (lay == "domain") lay = "the domain laws"
	if (lay == "infrastructure") lay = "the adapters"
	node[++nn] = p
	label[p] = lab
	layer[p] = lay
	if (!(lay in layerSeen)) { layerSeen[lay] = 1; layers[++nl] = lay }
}
END {
	for (i = 2; i <= nl; i++) { t = layers[i]; for (j = i - 1; j >= 1 && layers[j] > t; j--) layers[j+1] = layers[j]; layers[j+1] = t }
	for (i = 1; i <= nn; i++) printf "  \"%s\" [label=\"%s\"]\n", node[i], label[node[i]]
	print "  { rank=same;"
	for (j = 1; j <= nn; j++) if (layer[node[j]] == layers[1]) printf " \"%s\";", node[j]
	print " }"
	print "  { rank=same;"
	for (j = 1; j <= nn; j++) if (layer[node[j]] == layers[2]) printf " \"%s\";", node[j]
	print " }"
	print "  { rank=same;"
	for (j = 1; j <= nn; j++) if (layer[node[j]] == layers[3]) printf " \"%s\";", node[j]
	print " }"
	print "  { rank=same;"
	for (j = 1; j <= nn; j++) if (layer[node[j]] == layers[4]) printf " \"%s\";", node[j]
	print " }"
	for (k in cnt) {
		split(k, pair, "\t")
		if (layer[pair[1]] == layer[pair[2]])
			printf "  \"%s\" -> \"%s\" [constraint=false]\n", pair[1], pair[2]
		else
			printf "  \"%s\" -> \"%s\"\n", pair[1], pair[2]
	}
	print "}"
}
# The package of a node: a pointer-method node is "(*PATH/pkg.Type).M"
# — the package is the bracketed type's path minus ".Type"; any other
# node is "PATH/pkg.Entity" — the entity is cut at the last segment's
# first dot. A generic entity's "[type]" tail is dropped first.
function pkgof(full,    i, n, s, last, inner, typ) {
	i = index(full, "[")
	if (i) full = substr(full, 1, i - 1)
	if (substr(full, 1, 2) == "(*") {
		inner = substr(full, 3)
		i = index(inner, ")")
		typ = substr(inner, 1, i - 1)
		sub(/\.[^.]*$/, "", typ)
		return typ
	}
	n = split(full, s, "/")
	last = s[n]
	i = index(last, ".")
	if (i) last = substr(last, 1, i - 1)
	full = s[1]
	for (k = 2; k < n; k++) full = full "/" s[k]
	return full "/" last
}
