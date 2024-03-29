#!/usr/bin/env bash
set -eux

version="${version:-SPECIFY-A-VERSION}"
platform="${platform:-Darwin}"
arch="${arch:-arm64}"

tmp_dir=$(mktemp -d)
install_dir="${HOME}/bin"
archive_name="tsk_v${version}_${platform}_${arch}.tar.gz"

curl -sL -o "${tmp_dir}/${archive_name}" \
  "https://github.com/notnmeyer/tsk/releases/download/v${version}/${archive_name}"

cd "$tmp_dir"
tar xzf "$archive_name"

cp "tsk" "${install_dir}/"
rm -rf "$tmp_dir"
