#!/bin/sh
# the campaign's network probe: the bytes
# ride the cell's input/ as staged material; the HAND stages them
# beside its submission in the machine's inbox and the row's
# procedural launch mode (launch: nonet) runs them — the machine
# reads rc/stdout only, what this checks is the campaign's business.
#
# rc 0 + the fixed line when no external route answers (a network-
# empty namespace carries loopback only); rc 1 when a route or a
# name resolves — the network is on.
if ip route get 192.0.2.1 >/dev/null 2>&1 || getent hosts example.com >/dev/null 2>&1; then
  echo "external network reachable: the network is NOT off" >&2
  exit 1
fi
echo "nonet: no external network"
