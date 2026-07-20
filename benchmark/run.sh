#!/bin/sh
# 2026-07-20：在临时目录构建并运行验证集，避免修改仓库或污染已有产物。
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_dir=$(CDPATH= cd -- "$script_dir/.." && pwd)
temp_dir=$(mktemp -d "${TMPDIR:-/tmp}/buildsift-benchmark.XXXXXX")
trap 'rm -rf "$temp_dir"' EXIT HUP INT TERM

GOCACHE="$temp_dir/go-cache"
GOMODCACHE="$temp_dir/go-mod-cache"
export GOCACHE GOMODCACHE
(cd "$repo_dir" && go build -o "$temp_dir/buildsift" ./cmd/buildsift)

total=0
hits=0
drift=0
tab=$(printf '\t')

printf '%-30s %-10s %-10s %-8s %-8s\n' "CASE" "ECOSYSTEM" "EXPECTED" "RESULT" "BASELINE"
while IFS="$tab" read -r case_id ecosystem reported_tool expected_fragment baseline; do
	[ "$case_id" = "id" ] && continue
	total=$((total + 1))
	output=$("$temp_dir/buildsift" "$script_dir/testdata/$case_id.log")
	tool=$(printf '%s' "$reported_tool" | tr '[:lower:]' '[:upper:]')
	result=miss
	if printf '%s\n' "$output" | grep -F "ROOT CAUSE [" >/dev/null \
		&& printf '%s\n' "$output" | grep -F "· $tool]" >/dev/null \
		&& printf '%s\n' "$output" | grep -F "$expected_fragment" >/dev/null; then
		result=hit
		hits=$((hits + 1))
	fi
	if [ "$result" != "$baseline" ]; then
		drift=$((drift + 1))
	fi
	printf '%-30s %-10s %-10s %-8s %-8s\n' "$case_id" "$ecosystem" "$reported_tool" "$result" "$baseline"
done < "$script_dir/cases.tsv"

printf '\nObserved strict hits: %d/%d\n' "$hits" "$total"
printf 'Baseline drift: %d\n' "$drift"
[ "$drift" -eq 0 ]
