{ pkgs ? import <nixpkgs> {} }:

with import <nixpkgs> {
  config.android_sdk.accept_license = true;
  config.allowUnfree = true;
};

let
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
    go_1_20
  ];

  shellHook = ''
    export GRADLE_OPTS="-Dorg.gradle.project.android.aapt2FromMavenOverride=${androidComposition.androidsdk}/libexec/android-sdk/build-tools/${buildToolsVersion}/aapt2";
    export ANDROID_SDK_ROOT="${androidComposition.androidsdk}/libexec/android-sdk"
    export BUILDTOOLS="${androidComposition.androidsdk}/libexec/android-sdk/build-tools/${buildToolsVersion}/"
  '';

}
