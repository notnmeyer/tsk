apple_id {
  password = "@env:AC_PASSWORD"
}

bundle_id = "com.notnmeyer.tsk"

source = [
  "./dist/macos_darwin_arm64/tsk",
]

sign {
  application_identity = "Developer ID Application: Nathan Meyer"
}

# https://github.com/mitchellh/gon/issues/64
# dmg {
#   output_path = "./dist/tsk.dmg"
#   volume_name = "tsk"
# }

zip {
  output_path = "./dist/tsk_macos_arm64.zip"
}

