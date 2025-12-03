# wormhole-william-mobile

This is a Magic Wormhole client for Android and iOS.

Some current limitations:
- Receiving directories are kept in zip form.
- Send only supports sending a single file.

## Installing the APK on Android

Available from the Play store:
https://play.google.com/store/apps/details?id=io.sanford.wormhole_william

Prebuilt APKs are provided with each release. You can install this to an android device
that has developer mode enabled by running:

```
adb install wormhole-william.release.apk
```

## Building for Android

### Prerequisites

This project uses Nix for reproducible builds. Install Nix and enable flakes:

```bash
# Install Nix (if not already installed)
curl -L https://nixos.org/nix/install | sh

# Enable flakes (add to ~/.config/nix/nix.conf)
experimental-features = nix-command flakes
```

Alternatively, you can manually set up:
- Go 1.22+
- Android SDK (API level 34)
- Android NDK 26
- JDK 17
- gomobile

### Building

```bash
# Enter the development environment
nix develop

# One-time setup: initialize gomobile
make init

# Build debug APK
make

# Build release APK (requires signing key)
make release
```

The debug APK will be created at `wormhole-william.debug.apk`.

### Project Structure

```
wormhole-william-mobile/
├── wormhole/              # Go library (gomobile-compatible API)
│   ├── wormhole.go        # Main client API
│   ├── callbacks.go       # Callback interfaces for Android
│   ├── pending.go         # File transfer acceptance handling
│   └── config.go          # Configuration persistence
│
├── android/               # Android application
│   ├── app/               # Main Android module
│   │   └── src/main/
│   │       ├── kotlin/    # Kotlin/Compose UI
│   │       └── res/       # Android resources
│   └── build.gradle.kts   # Gradle configuration
│
├── Makefile               # Build orchestration
├── flake.nix              # Nix flake for dev environment
└── shell.nix              # Legacy Nix shell (deprecated)
```

### Architecture

The app uses a native Android UI built with Jetpack Compose, with the wormhole
protocol logic implemented in Go and bound via gomobile:

```
┌─────────────────────────────────────┐
│         Kotlin/Compose UI           │
│  (Screens, ViewModels, Repository)  │
└──────────────┬──────────────────────┘
               │ gomobile bindings
┌──────────────▼──────────────────────┐
│          Go wormhole package        │
│    (Wraps wormhole-william lib)     │
└─────────────────────────────────────┘
```

## iOS

Currently iOS development is happening on the ios branch.

## Video Demo

This [demo](https://www.youtube.com/watch/FOY4vhUoikU?t=210s) was done as part of a larger talk on the development of Wormhole William Mobile and its use of [Gio](https://gioui.org/):

[![Wormhole William Mobile Demo](https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/wormhole-william-mobile-youtube.png?raw=true)](https://www.youtube.com/watch/FOY4vhUoikU?t=210s "Wormhole William Demo")

## Screenshots

<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/recv1.png?raw=true" alt="Receive 1" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/recv2.png?raw=true" alt="Receive 2" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send_text1.png?raw=true" alt="Send Text 1" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send_text2.png?raw=true" alt="Send Text 2" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send_file1.png?raw=true" alt="Send File" width="200" />
