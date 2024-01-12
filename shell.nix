{ pkgs ? import (fetchTarball "channel:nixos-unstable") {} }:

pkgs.mkShell {
  packages = with pkgs; [
    delve
    go
    gopls
    goreleaser
  ];
}
