#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "$script_dir/.." && pwd)"
lexicons_dir="$repo_root/lexicons"
work_dir="$(mktemp -d "$repo_root/.lexicons-update.XXXXXX")"
checkout_dir="$work_dir/upstream"
source_dir="$checkout_dir/packages/editor/lexicons"
staged_dir="$work_dir/lexicons"
backup_dir="$work_dir/previous-lexicons"

cleanup() {
    if [[ -d "$backup_dir" && ! -e "$lexicons_dir" ]]; then
        mv "$backup_dir" "$lexicons_dir" || return 1
    fi
    rm -rf "$work_dir"
}
trap cleanup EXIT

git clone --depth 1 --filter=blob:none --sparse --branch main \
    https://github.com/nus25/gyoka.git "$checkout_dir"
git -C "$checkout_dir" sparse-checkout set packages/editor/lexicons

if [[ ! -d "$source_dir/net/nusno/gyoka" ]] ||
    [[ -z "$(find "$source_dir" -type f -name '*.json' -print -quit)" ]]; then
    printf 'Upstream lexicons are missing or empty: %s\n' "$source_dir" >&2
    exit 1
fi

cp -a "$source_dir" "$staged_dir"
mv "$lexicons_dir" "$backup_dir"
mv "$staged_dir" "$lexicons_dir"

commit="$(git -C "$checkout_dir" rev-parse --short HEAD)"
printf 'Updated lexicons from nus25/gyoka@%s\n' "$commit"
