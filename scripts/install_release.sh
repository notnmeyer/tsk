#!/usr/bin/env bash
set -eux

tmp_dir=$(mktemp -d)
install_dir="${HOME}/bin"
archive_name="tsk_v${version}_${platform}_${arch}.zip"

curl -L -o "${tmp_dir}/${archive_name}" \
  "https://github.com/notnmeyer/tsk/releases/download/v${version}/${archive_name}"

unzip "${tmp_dir}/${archive_name}" -d "$tmp_dir"

cp "${tmp_dir}/tsk" "${install_dir}/"