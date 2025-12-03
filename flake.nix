{
  description = "Wormhole William Mobile - Magic Wormhole client for Android";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            android_sdk.accept_license = true;
            allowUnfree = true;
          };
        };

        buildToolsVersion = "35.0.0";
        androidComposition = pkgs.androidenv.composeAndroidPackages {
          platformVersions = [ "35" ];
          buildToolsVersions = [ buildToolsVersion ];
          includeNDK = true;
          ndkVersions = [ "26.1.10909125" ];
        };

        androidSdk = androidComposition.androidsdk;
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            # Go toolchain
            go_1_22

            # Android toolchain
            androidSdk
            openjdk17
            kotlin

            # Build tools
            gnumake
          ];

          shellHook = ''
            export ANDROID_SDK_ROOT="${androidSdk}/libexec/android-sdk"
            export ANDROID_HOME="$ANDROID_SDK_ROOT"
            export ANDROID_NDK_HOME="$ANDROID_SDK_ROOT/ndk/26.1.10909125"
            export PATH="$ANDROID_NDK_HOME:$PATH"
            export BUILDTOOLS="${androidSdk}/libexec/android-sdk/build-tools/${buildToolsVersion}"
            export GRADLE_OPTS="-Dorg.gradle.project.android.aapt2FromMavenOverride=$BUILDTOOLS/aapt2"

            echo "Wormhole William Mobile development environment"
            echo "  Go: $(go version)"
            echo "  Android SDK: $ANDROID_SDK_ROOT"
            echo "  NDK: $ANDROID_NDK_HOME"
            echo ""
            echo "Run 'make init' to initialize gomobile (first time only)"
            echo "Run 'make' to build the debug APK"
          '';
        };
      }
    );
}
