#!/bin/sh
set -eu

config=$1
work_dir=$(mktemp -d)
trap 'rm -rf "$work_dir"' EXIT HUP INT TERM

jq -er '.scalaVersions[]' "$config" > "$work_dir/versions"
jq -r '.dependencies[] | "//> using dep \(.)"' "$config" > "$work_dir/dependencies"

while IFS= read -r version; do
  printf 'Caching Scala %s and configured dependencies\n' "$version"
  {
    printf '//> using scala %s\n' "$version"
    cat "$work_dir/dependencies"
    printf '\nprintln("Scala dependencies ready")\n'
  } > "$work_dir/cache.sc"
  scala-cli run "$work_dir/cache.sc" --server=false --jvm system
done < "$work_dir/versions"
