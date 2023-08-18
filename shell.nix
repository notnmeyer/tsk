{ pkgs ? import (fetchTarball "channel:nixos-23.05") {} }:

pkgs.mkShell {
  packages = with pkgs; [
    delve
    go
    gopls
    goreleaser
  ];
}