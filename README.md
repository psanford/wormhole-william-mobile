# wormhole-william-mobile

A Magic Wormhole client for Android.

Current limitations:
- Received directories are kept in zip form
- Send only supports sending a single file

## Installing

Available from the Play Store:
https://play.google.com/store/apps/details?id=io.sanford.wormhole_william

Prebuilt APKs are provided with each release. Install to an Android device with developer mode enabled:

```
adb install wormhole-william.release.apk
```

## Building

### Prerequisites

This project uses Nix for reproducible builds. Install Nix and enable flakes:

```bash
# Install Nix (if not already installed)
curl -L https://nixos.org/nix/install | sh

# Enable flakes (add to ~/.config/nix/nix.conf)
experimental-features = nix-command flakes
```

Alternatively, manually set up:
- Go 1.22+
- Android SDK (API level 34)
- Android NDK
- JDK 17

### Build Commands

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

The debug APK will be output to `wormhole-william.debug.apk`.

## Project Structure

```
wormhole-william-mobile/
├── wormhole/              # Go library (gomobile bindings)
│   ├── wormhole.go        # Main client API
│   ├── callbacks.go       # Callback interfaces for Android
│   ├── pending.go         # File transfer acceptance handling
│   └── config.go          # Configuration persistence
│
├── android/               # Android application
│   ├── app/src/main/
│   │   ├── kotlin/        # Kotlin/Compose UI
│   │   └── res/           # Android resources
│   └── build.gradle.kts   # Gradle build config
│
├── Makefile               # Build orchestration
└── flake.nix              # Nix flake for dev environment
```

## Architecture

The app uses a native Android UI built with Jetpack Compose. The wormhole protocol logic is implemented in Go and bound via gomobile:

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

## Screenshots

<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/recv.png" alt="Receive" width="250" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send-text.png" alt="Send Text" width="250" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send-file.png" alt="Send File" width="250" />

## iOS

iOS development is on the `ios` branch.
