{ pkgs ? import <nixpkgs> {} }:

with import <nixpkgs> {
  config.android_sdk.accept_license = true;
  config.allowUnfree = true;
};

let
  pinPkgsFetch = pkgs.fetchFromGitHub {
    owner  = "NixOS";
    repo   = "nixpkgs";
    rev    = "a9858885e197f984d92d7fe64e9fff6b2e488d40";
    # Hash obtained using `nix-prefetch-url --unpack --unpack https://github.com/nixos/nixpkgs/archive/<rev>.tar.gz`
    sha256 = "0a55lp827bfx102czy0bp5d6pbp5lh6l0ysp3zs0m1gyniy2jck9";
  };
  pinPkgs = import pinPkgsFetch {
    config.android_sdk.accept_license = true;
    config.allowUnfree = true;
  };

  buildToolsVersion = "33.0.2";
  androidComposition = androidenv.composeAndroidPackages {
    platformVersions = [ "30" "33" ];
    buildToolsVersions = [ "${buildToolsVersion}" ];
    includeNDK = true;
  };
in

pkgs.mkShell {
  nativeBuildInputs = with pkgs.buildPackages; [
    openjdk17
    androidComposition.androidsdk
    pinPkgs.go_1_22
  ];

  shellHook = ''
    export GRADLE_OPTS="-Dorg.gradle.project.android.aapt2FromMavenOverride=${androidComposition.androidsdk}/libexec/android-sdk/build-tools/${buildToolsVersion}/aapt2";
    export ANDROID_SDK_ROOT="${androidComposition.androidsdk}/libexec/android-sdk"
    export BUILDTOOLS="${androidComposition.androidsdk}/libexec/android-sdk/build-tools/${buildToolsVersion}/"
  '';

}
