#!/usr/bin/env bash
set -eux

tmp_dir=$(mktemp -d)
install_dir="${HOME}/bin"

curl -L -o "${tmp_dir}/tsk_v${version}_${platform}_${arch}.tar.gz" \
  "https://github.com/notnmeyer/tsk/releases/download/v${version}/tsk_v${version}_${platform}_${arch}.tar.gz"

tar -xzf "${tmp_dir}/tsk_v${version}_${platform}_${arch}.tar.gz" -C "$tmp_dir"

cp "${tmp_dir}/tsk" "${install_dir}/"