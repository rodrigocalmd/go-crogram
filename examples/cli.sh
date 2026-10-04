#!/usr/bin/env sh
# Every way of using the crogram command line.
# Run from the repository root:  sh examples/cli.sh
set -eu

# Build once: "go run" would hide the real exit code of the program.
bin=$(mktemp)
go build -o "$bin" ./cmd/crogram
trap 'rm -f "$bin"' EXIT
crogram() { "$bin" "$@"; }
step() { printf '\n== %s\n' "$*"; }

step "encode (random cipher; the seed goes to stderr)"
crogram encode "Hello World"

step "-q: do not print the seed"
crogram encode -q "Hello World"

step "-s: reproducible cipher (flag before or after the command)"
crogram -s 42 encode "consistency matters"
crogram encode -s 42 "consistency matters"
crogram --seed=42 encode "consistency matters"

step "decode needs the same seed"
crogram decode -s 42 "9ZhB1BQSh9U ktQQS6B"

step "text from stdin"
echo "consistency matters" | crogram encode -s 42
echo "9ZhB1BQSh9U ktQQS6B" | crogram decode -s 42

step "key: portable alternative to the seed"
KEY=$(crogram key -s 42)
echo "$KEY"
crogram encode -k "$KEY" "consistency matters"
crogram decode -k "$KEY" "9ZhB1BQSh9U ktQQS6B"

step "-f / -o: files"
tmp=$(mktemp -d)
printf 'line one\nline two\n' > "$tmp/in.txt"
crogram encode -s 42 -f "$tmp/in.txt" -o "$tmp/out.txt"
cat "$tmp/out.txt"
crogram decode -s 42 -f "$tmp/out.txt"
rm -r "$tmp"

step "-c: other languages (pt, es, ru or your own characters)"
crogram -c pt -s 42 encode "Ação e emoção"
crogram -c pt -s 42 decode "4wsn 1 1pnwsn"
crogram -c ru -s 42 encode "Привет, мир!"
crogram -c "あいうえお" -s 1 encode "あさ"

step "key of a custom charset (-k alone is enough to decode)"
PKEY=$(crogram key -c pt -s 42)
echo "$PKEY"
crogram decode -k "$PKEY" "4wsn 1 1pnwsn"

step "--version and help"
crogram --version
crogram help | head -3

step "errors (invalid usage exits with code 2)"
crogram decode "no seed" || echo "exit code $?"
crogram -c a encode x || echo "exit code $?"
