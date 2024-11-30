{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    pkg-config
    wayland
    vulkan-headers
    libxkbcommon
    libGL
    go_1_22
  ];
}
