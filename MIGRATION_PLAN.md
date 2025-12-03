# Migration Plan: Gio UI to Native Android (Kotlin/Java)

## Executive Summary

This document outlines the migration of wormhole-william-mobile from a Gio-based UI to a native Android UI using Kotlin/Java, while retaining the Go wormhole-william library for protocol logic. The migration will use gomobile to create bindings instead of the current gogio/Gio approach.

---

## Part 1: Current Architecture Analysis

### 1.1 Project Overview

**Current Stack:**
- **UI Framework:** Gio (gioui.org v0.7.1) - Immediate mode Go UI framework
- **Wormhole Protocol:** wormhole-william v1.0.6 (Go library)
- **JNI Bridge:** Custom JNI via git.wow.st/gmp/jni for Android-specific features
- **Build System:** gogio → AAR → Gradle → APK

**App Functionality:**
1. **Receive:** Accept wormhole codes (typed or QR scanned), download files/text
2. **Send Text:** Send text messages via wormhole
3. **Send File:** Pick and send single files via wormhole
4. **Settings:** Configure rendezvous URL and code length
5. **Share Intent:** Receive shared files/text from other Android apps

### 1.2 Current File Structure

```
wormhole-william-mobile/
├── main.go                      # Entry point (launches Gio)
├── go.mod / go.sum              # Go dependencies
├── Makefile                     # Build orchestration
├── shell.nix                    # Nix development environment
│
├── ui/                          # Gio UI code (~1200 lines total)
│   ├── ui.go                    # Main UI loop, tabs, layouts (~950 lines)
│   ├── platform_android.go      # Android platform handler
│   ├── platform_dummy.go        # Desktop fallback handler
│   ├── richeditor.go            # Custom editor with copy/paste
│   ├── copyable.go              # Read-only copyable widget
│   ├── slider.go                # Tab animation
│   └── proxyreader.go           # Progress tracking reader
│
├── jgo/                         # JNI bridge code
│   └── jgo.go                   # JNI exports for Android callbacks
│
├── internal/picker/             # Shared types
│   └── picker.go                # PickResult, SharedEvent, PermResult
│
├── config/                      # User configuration
│   └── config.go                # JSON persistence for settings
│
└── android/                     # Android-specific code
    ├── build.gradle             # Gradle config
    ├── src/main/
    │   ├── AndroidManifest.xml  # App manifest
    │   └── java/io/sanford/wormholewilliam/
    │       ├── Jni.java         # File picker fragment
    │       ├── Share.java       # Share intent handler
    │       ├── Scan.java        # QR code scanner
    │       ├── Download.java    # Download manager integration
    │       └── WriteFilePerm.java # Permission handler
    └── libs/                    # Contains wormhole-william.aar (generated)
```

### 1.3 Current Build Flow

```
1. make
   └─> gogio -buildmode archive -target android ...
       └─> Generates wormhole-william.aar containing:
           - Go runtime
           - All Go code (including Gio UI)
           - Native .so files for ARM/ARM64

2. Gradle assembleDebug
   └─> Compiles Java code
   └─> Links AAR, AndroidX, ZXing
   └─> Packages APK

Entry Point: GioActivity (from Gio) → main.go → ui.New().Run()
```

### 1.4 Current JNI Communication Pattern

The app uses a bidirectional communication pattern:

**Go → Java (via JNI calls):**
- `jgo.PickFile()` → Launches `Jni.java` Fragment
- `jgo.ScanQRCode()` → Launches `Scan.java` Fragment
- `jgo.NotifyDownloadManager()` → Calls `Download.java`
- `jgo.RequestWriteFilePermission()` → Calls `WriteFilePerm.java`

**Java → Go (via JNI exports):**
- `pickerResult()` → Called by Jni.java after file selection
- `gotSharedItem()` → Called by Share.java on share intent
- `scanResult()` → Called by Scan.java after QR scan
- `permissionResult()` → Called by WriteFilePerm.java

Communication uses Go channels for async results with 10-second timeouts.

### 1.5 Wormhole Client Usage

The Go code uses wormhole-william as follows:

```go
type UI struct {
    wormholeClient wormhole.Client
    conf           *config.Config
}

// Send text
code, status, err := ui.wormholeClient.SendText(ctx, msg)

// Send file (with progress callback)
code, status, err := ui.wormholeClient.SendFile(ctx, filename, file,
    wormhole.WithProgress(func(sent, total int64) { ... }))

// Receive
msg, err := ui.wormholeClient.Receive(ctx, code)
// msg.Type: TransferText | TransferFile | TransferDirectory
// msg implements io.Reader for content

// Configuration
ui.wormholeClient.RendezvousURL = "..."
ui.wormholeClient.PassPhraseComponentLength = 2
```

---

## Part 2: Target Architecture

### 2.1 New Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Native Android UI                     │
│              (Kotlin/Java + Jetpack Compose)            │
│                                                          │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐   │
│  │  Receive │ │Send Text │ │Send File │ │ Settings │   │
│  │   Tab    │ │   Tab    │ │   Tab    │ │   Tab    │   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘   │
│       │            │            │            │          │
│       └────────────┴────────────┴────────────┘          │
│                         │                                │
│                 ┌───────▼───────┐                       │
│                 │ WormholeRepo  │ (Kotlin)              │
│                 │  (Repository) │                       │
│                 └───────┬───────┘                       │
│                         │                                │
└─────────────────────────┼────────────────────────────────┘
                          │ gomobile bindings
┌─────────────────────────▼────────────────────────────────┐
│                   Go Library (AAR)                        │
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │              wormhole.go (new)                      │ │
│  │                                                      │ │
│  │  - SendText(msg) → (code, error)                    │ │
│  │  - SendFile(path, name, progress) → (code, error)   │ │
│  │  - Receive(code, progress) → (type, data, error)    │ │
│  │  - SetRendezvousURL(url)                            │ │
│  │  - SetCodeLength(len)                               │ │
│  │  - Cancel()                                          │ │
│  └────────────────────────────────────────────────────┘ │
│                         │                                │
│  ┌──────────────────────▼─────────────────────────────┐ │
│  │           wormhole-william v1.0.6                   │ │
│  │              (unchanged library)                     │ │
│  └─────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

### 2.2 Key Differences from Current Architecture

| Aspect | Current (Gio) | New (Native) |
|--------|--------------|--------------|
| UI Framework | Gio (Go) | Jetpack Compose (Kotlin) |
| Entry Point | GioActivity | MainActivity (Kotlin) |
| UI Rendering | OpenGL/Vulkan (Gio) | Native Android Views |
| Go Binding Tool | gogio | gomobile |
| AAR Contents | Full app + Gio runtime | Just wormhole logic |
| JNI Bridge | Manual (jgo package) | Auto-generated (gomobile) |
| UI State | Go channels + global vars | ViewModel + StateFlow |

### 2.3 Technology Choices

**UI:** Jetpack Compose (Modern Android UI toolkit)
- Declarative syntax similar to Gio's immediate mode
- First-class Kotlin support
- Material Design 3 built-in
- Easier to maintain than XML layouts

**Architecture:** MVVM with Repository pattern
- ViewModel for UI state management
- Repository to wrap Go library calls
- Coroutines for async operations

**Go Bindings:** gomobile bind
- Generates Java bindings automatically
- No manual JNI code needed
- Supports callbacks via interfaces

---

## Part 3: Go Library Design

### 3.1 New Go Package Structure

```
wormhole-william-mobile/
├── wormhole/                    # NEW: gomobile-compatible library
│   ├── wormhole.go              # Main API for Android
│   ├── callbacks.go             # Callback interfaces
│   ├── types.go                 # Shared types
│   └── config.go                # Configuration (migrated)
│
├── cmd/                         # LEGACY: Keep for desktop testing
│   └── gio-app/                 # Optional: Keep Gio app as reference
│       ├── main.go
│       └── ui/                  # Current ui/ folder moved here
│
└── android/                     # Updated for native UI
    └── app/
        ├── build.gradle.kts     # Modern Kotlin DSL
        └── src/main/
            ├── kotlin/...       # Native Kotlin UI
            └── java/...         # Keep Share.java, etc.
```

### 3.2 Go Library API Design

```go
// wormhole/wormhole.go
package wormhole

import (
    "context"
    "io"
    "os"
    "sync"

    wh "github.com/psanford/wormhole-william/wormhole"
)

// ProgressCallback is implemented by Android code
type ProgressCallback interface {
    OnProgress(sent, total int64)
}

// ReceiveCallback is implemented by Android code
type ReceiveCallback interface {
    OnText(text string)
    OnFileStart(name string, size int64)
    OnFileProgress(received, total int64)
    OnFileComplete(path string)
    OnError(err string)
}

// SendCallback is implemented by Android code
type SendCallback interface {
    OnCode(code string)
    OnProgress(sent, total int64)
    OnComplete()
    OnError(err string)
}

// Client wraps wormhole-william for mobile use
type Client struct {
    mu       sync.Mutex
    client   wh.Client
    cancel   context.CancelFunc
    dataDir  string
}

// NewClient creates a new wormhole client
func NewClient(dataDir string) *Client {
    return &Client{
        dataDir: dataDir,
    }
}

// SetRendezvousURL configures the relay server
func (c *Client) SetRendezvousURL(url string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.client.RendezvousURL = url
}

// SetCodeLength configures passphrase length
func (c *Client) SetCodeLength(length int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.client.PassPhraseComponentLength = length
}

// SendText sends a text message
func (c *Client) SendText(msg string, callback SendCallback) {
    go func() {
        ctx, cancel := context.WithCancel(context.Background())
        c.mu.Lock()
        c.cancel = cancel
        c.mu.Unlock()
        defer cancel()

        code, status, err := c.client.SendText(ctx, msg)
        if err != nil {
            callback.OnError(err.Error())
            return
        }
        callback.OnCode(code)

        s := <-status
        if s.Error != nil {
            callback.OnError(s.Error.Error())
        } else {
            callback.OnComplete()
        }
    }()
}

// SendFile sends a file from the given path
func (c *Client) SendFile(path, name string, callback SendCallback) {
    go func() {
        ctx, cancel := context.WithCancel(context.Background())
        c.mu.Lock()
        c.cancel = cancel
        c.mu.Unlock()
        defer cancel()

        f, err := os.Open(path)
        if err != nil {
            callback.OnError(err.Error())
            return
        }
        defer f.Close()

        progress := func(sent, total int64) {
            callback.OnProgress(sent, total)
        }

        code, status, err := c.client.SendFile(ctx, name, f, wh.WithProgress(progress))
        if err != nil {
            callback.OnError(err.Error())
            return
        }
        callback.OnCode(code)

        s := <-status
        if s.Error != nil {
            callback.OnError(s.Error.Error())
        } else {
            callback.OnComplete()
        }
    }()
}

// Receive receives a wormhole transfer
func (c *Client) Receive(code string, callback ReceiveCallback) {
    go func() {
        ctx, cancel := context.WithCancel(context.Background())
        c.mu.Lock()
        c.cancel = cancel
        c.mu.Unlock()
        defer cancel()

        msg, err := c.client.Receive(ctx, code)
        if err != nil {
            callback.OnError(err.Error())
            return
        }

        switch msg.Type {
        case wh.TransferText:
            data, err := io.ReadAll(msg)
            if err != nil {
                callback.OnError(err.Error())
                return
            }
            callback.OnText(string(data))

        case wh.TransferFile, wh.TransferDirectory:
            name := msg.Name
            if msg.Type == wh.TransferDirectory {
                name += ".zip"
            }

            callback.OnFileStart(name, msg.TransferBytes64)

            // Save to dataDir
            path := filepath.Join(c.dataDir, name)
            f, err := os.Create(path)
            if err != nil {
                msg.Reject()
                callback.OnError(err.Error())
                return
            }

            // Copy with progress
            buf := make([]byte, 32*1024)
            var received int64
            for {
                n, err := msg.Read(buf)
                if n > 0 {
                    f.Write(buf[:n])
                    received += int64(n)
                    callback.OnFileProgress(received, msg.TransferBytes64)
                }
                if err == io.EOF {
                    break
                }
                if err != nil {
                    f.Close()
                    os.Remove(path)
                    callback.OnError(err.Error())
                    return
                }
            }
            f.Close()
            callback.OnFileComplete(path)
        }
    }()
}

// Cancel cancels any ongoing transfer
func (c *Client) Cancel() {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.cancel != nil {
        c.cancel()
        c.cancel = nil
    }
}
```

### 3.3 gomobile Binding Generation

The new Makefile target:

```makefile
# Generate Go bindings AAR
wormhole.aar: $(shell find wormhole -name '*.go')
    gomobile bind -v -target=android -androidapi 24 \
        -o $@ ./wormhole
```

---

## Part 4: Native Android UI Design

### 4.1 Kotlin/Compose Project Structure

```
android/app/src/main/
├── kotlin/io/sanford/wormholewilliam/
│   ├── MainActivity.kt              # Single activity, hosts Compose
│   ├── WormholeApp.kt               # Main Compose app structure
│   │
│   ├── ui/
│   │   ├── theme/
│   │   │   ├── Color.kt
│   │   │   ├── Theme.kt
│   │   │   └── Type.kt
│   │   │
│   │   ├── screens/
│   │   │   ├── ReceiveScreen.kt    # Receive tab
│   │   │   ├── SendTextScreen.kt   # Send text tab
│   │   │   ├── SendFileScreen.kt   # Send file tab
│   │   │   └── SettingsScreen.kt   # Settings tab
│   │   │
│   │   └── components/
│   │       ├── WormholeCodeInput.kt  # Code input with paste
│   │       ├── ProgressIndicator.kt  # Transfer progress
│   │       └── StatusBar.kt          # Status messages
│   │
│   ├── viewmodel/
│   │   ├── ReceiveViewModel.kt
│   │   ├── SendTextViewModel.kt
│   │   ├── SendFileViewModel.kt
│   │   └── SettingsViewModel.kt
│   │
│   ├── repository/
│   │   └── WormholeRepository.kt    # Wraps Go library
│   │
│   └── util/
│       ├── QRScanner.kt             # ZXing integration
│       └── FileUtils.kt             # File operations
│
├── java/io/sanford/wormholewilliam/
│   └── Share.java                    # Keep for share intent (or migrate to Kotlin)
│
└── res/
    ├── values/
    │   ├── strings.xml
    │   └── themes.xml
    └── mipmap-*/                     # App icons (keep existing)
```

### 4.2 ViewModel Example

```kotlin
// viewmodel/ReceiveViewModel.kt
class ReceiveViewModel(
    private val repository: WormholeRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(ReceiveUiState())
    val uiState: StateFlow<ReceiveUiState> = _uiState.asStateFlow()

    fun onCodeChanged(code: String) {
        _uiState.update { it.copy(code = code.replace(" ", "-").replace("\n", "")) }
    }

    fun onReceive() {
        val code = _uiState.value.code
        if (code.isBlank()) return

        _uiState.update { it.copy(isTransferring = true, status = "Connecting...") }

        repository.receive(code, object : ReceiveCallback {
            override fun onText(text: String) {
                _uiState.update {
                    it.copy(
                        isTransferring = false,
                        receivedText = text,
                        status = "Text received"
                    )
                }
            }

            override fun onFileStart(name: String, size: Long) {
                _uiState.update {
                    it.copy(status = "Receiving $name (${formatBytes(size)})")
                }
            }

            override fun onFileProgress(received: Long, total: Long) {
                _uiState.update {
                    it.copy(
                        status = "Receiving ${formatBytes(received)}/${formatBytes(total)}",
                        progress = received.toFloat() / total.toFloat()
                    )
                }
            }

            override fun onFileComplete(path: String) {
                _uiState.update {
                    it.copy(
                        isTransferring = false,
                        status = "File saved: $path"
                    )
                }
            }

            override fun onError(err: String) {
                _uiState.update {
                    it.copy(isTransferring = false, status = "Error: $err")
                }
            }
        })
    }

    fun onCancel() {
        repository.cancel()
        _uiState.update { it.copy(isTransferring = false, status = "Cancelled") }
    }
}

data class ReceiveUiState(
    val code: String = "",
    val isTransferring: Boolean = false,
    val status: String = "",
    val receivedText: String = "",
    val progress: Float = 0f
)
```

### 4.3 Compose Screen Example

```kotlin
// ui/screens/ReceiveScreen.kt
@Composable
fun ReceiveScreen(
    viewModel: ReceiveViewModel = viewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        // Code input
        Text("Code", style = MaterialTheme.typography.titleLarge)

        OutlinedTextField(
            value = uiState.code,
            onValueChange = viewModel::onCodeChanged,
            modifier = Modifier.fillMaxWidth(),
            placeholder = { Text("Enter wormhole code") },
            singleLine = true,
            trailingIcon = {
                IconButton(onClick = { /* paste from clipboard */ }) {
                    Icon(Icons.Default.ContentPaste, "Paste")
                }
            }
        )

        Spacer(modifier = Modifier.height(16.dp))

        // QR Scan button
        Button(
            onClick = { /* launch QR scanner */ },
            enabled = !uiState.isTransferring && uiState.code.isEmpty()
        ) {
            Text("Scan QR Code")
        }

        Spacer(modifier = Modifier.height(8.dp))

        // Receive button
        Button(
            onClick = viewModel::onReceive,
            enabled = !uiState.isTransferring
        ) {
            Text("Receive")
        }

        // Cancel button (shown during transfer)
        if (uiState.isTransferring) {
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedButton(onClick = viewModel::onCancel) {
                Text("Cancel")
            }
        }

        // Progress indicator
        if (uiState.progress > 0f) {
            Spacer(modifier = Modifier.height(16.dp))
            LinearProgressIndicator(
                progress = uiState.progress,
                modifier = Modifier.fillMaxWidth()
            )
        }

        // Received text display
        if (uiState.receivedText.isNotEmpty()) {
            Spacer(modifier = Modifier.height(16.dp))
            OutlinedTextField(
                value = uiState.receivedText,
                onValueChange = {},
                modifier = Modifier.fillMaxWidth(),
                readOnly = true,
                maxLines = 5,
                trailingIcon = {
                    IconButton(onClick = { /* copy to clipboard */ }) {
                        Icon(Icons.Default.ContentCopy, "Copy")
                    }
                }
            )
        }

        Spacer(modifier = Modifier.weight(1f))

        // Status bar
        if (uiState.status.isNotEmpty()) {
            Surface(
                color = MaterialTheme.colorScheme.secondaryContainer,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(
                    text = uiState.status,
                    modifier = Modifier.padding(16.dp)
                )
            }
        }
    }
}
```

---

## Part 5: Build System Migration

### 5.1 New Nix Flake

```nix
# flake.nix
{
  description = "Wormhole William Mobile development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-24.05";
    flake-utils.url = "github:numtide/flake-utils";
    android-nixpkgs = {
      url = "github:nickel/nix-android";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { self, nixpkgs, flake-utils, android-nixpkgs }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
          config = {
            android_sdk.accept_license = true;
            allowUnfree = true;
          };
        };

        buildToolsVersion = "34.0.0";
        androidComposition = pkgs.androidenv.composeAndroidPackages {
          platformVersions = [ "34" ];
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
            gomobile

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
            export ANDROID_NDK_HOME="$ANDROID_SDK_ROOT/ndk-bundle"
            export GRADLE_OPTS="-Dorg.gradle.project.android.aapt2FromMavenOverride=${androidSdk}/libexec/android-sdk/build-tools/${buildToolsVersion}/aapt2"

            # Initialize gomobile if needed
            if [ ! -d "$HOME/go/pkg/gomobile" ]; then
              echo "Initializing gomobile..."
              gomobile init
            fi
          '';
        };

        packages.default = pkgs.stdenv.mkDerivation {
          name = "wormhole-william-mobile";
          src = ./.;
          # ... build steps
        };
      }
    );
}
```

### 5.2 Updated Makefile

```makefile
# Makefile (updated)

# Directories
GO_WORMHOLE_PKG = ./wormhole
ANDROID_DIR = android
AAR_OUTPUT = $(ANDROID_DIR)/app/libs/wormhole.aar

# Tools
GOMOBILE = gomobile
GRADLE = $(ANDROID_DIR)/gradlew

# Build tools from environment
ZIPALIGN = $(ANDROID_SDK_ROOT)/build-tools/34.0.0/zipalign
APKSIGNER = $(ANDROID_SDK_ROOT)/build-tools/34.0.0/apksigner
SIGNKEY = $(HOME)/.android-release/wormhole-william-release.keystore

# Default target
.PHONY: all
all: debug

# Debug APK
.PHONY: debug
debug: $(AAR_OUTPUT)
	cd $(ANDROID_DIR) && ./gradlew assembleDebug
	cp $(ANDROID_DIR)/app/build/outputs/apk/debug/app-debug.apk wormhole-william.debug.apk

# Release APK
.PHONY: release
release: $(AAR_OUTPUT)
	cd $(ANDROID_DIR) && ./gradlew assembleRelease
	$(ZIPALIGN) -v -p 4 \
		$(ANDROID_DIR)/app/build/outputs/apk/release/app-release-unsigned.apk \
		wormhole-william-unsigned-aligned.apk
	$(APKSIGNER) sign --ks $(SIGNKEY) \
		--out wormhole-william.release.apk \
		wormhole-william-unsigned-aligned.apk
	rm wormhole-william-unsigned-aligned.apk

# Generate Go bindings AAR
$(AAR_OUTPUT): $(shell find $(GO_WORMHOLE_PKG) -name '*.go' -type f)
	mkdir -p $(dir $@)
	$(GOMOBILE) bind -v \
		-target=android \
		-androidapi 24 \
		-o $@ \
		$(GO_WORMHOLE_PKG)

# Initialize gomobile (one-time setup)
.PHONY: init
init:
	$(GOMOBILE) init

# Clean build artifacts
.PHONY: clean
clean:
	rm -f $(AAR_OUTPUT)
	rm -f wormhole-william.debug.apk
	rm -f wormhole-william.release.apk
	rm -f wormhole-william-unsigned-aligned.apk
	cd $(ANDROID_DIR) && ./gradlew clean

# Run tests
.PHONY: test
test:
	go test -v ./...

# Format code
.PHONY: fmt
fmt:
	go fmt ./...
	cd $(ANDROID_DIR) && ./gradlew ktlintFormat
```

### 5.3 Updated Gradle Configuration

```kotlin
// android/app/build.gradle.kts
plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

android {
    namespace = "io.sanford.wormhole_william"
    compileSdk = 34

    defaultConfig {
        applicationId = "io.sanford.wormhole_william"
        minSdk = 24
        targetSdk = 34
        versionCode = 19
        versionName = "2.0.0"
    }

    buildFeatures {
        compose = true
    }

    composeOptions {
        kotlinCompilerExtensionVersion = "1.5.8"
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = "17"
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
}

dependencies {
    // Go wormhole library
    implementation(files("libs/wormhole.aar"))

    // Compose
    implementation(platform("androidx.compose:compose-bom:2024.02.00"))
    implementation("androidx.compose.ui:ui")
    implementation("androidx.compose.ui:ui-graphics")
    implementation("androidx.compose.ui:ui-tooling-preview")
    implementation("androidx.compose.material3:material3")
    implementation("androidx.activity:activity-compose:1.8.2")
    implementation("androidx.lifecycle:lifecycle-viewmodel-compose:2.7.0")
    implementation("androidx.navigation:navigation-compose:2.7.7")

    // QR code scanning
    implementation("com.journeyapps:zxing-android-embedded:4.3.0")

    // Preferences
    implementation("androidx.datastore:datastore-preferences:1.0.0")

    // Debug
    debugImplementation("androidx.compose.ui:ui-tooling")
}
```

---

## Part 6: Migration Steps (Commit-by-Commit)

### Phase 1: Infrastructure Setup (3 commits)

#### Commit 1: Add Nix flake for reproducible builds
- Create `flake.nix` with Go, gomobile, Android SDK
- Keep `shell.nix` as legacy fallback with deprecation notice
- Add `.envrc` for direnv integration
- Update `.gitignore` for flake outputs

#### Commit 2: Create Go wormhole library package
- Create `wormhole/` directory
- Implement `wormhole.go` with gomobile-compatible API
- Implement callback interfaces
- Move/adapt `config/config.go`
- Add unit tests for the library

#### Commit 3: Update build system for gomobile
- Update `Makefile` for gomobile bind
- Verify AAR generation works
- Document new build process in README

### Phase 2: Android Project Restructure (4 commits)

#### Commit 4: Modernize Android project structure
- Convert `build.gradle` to `build.gradle.kts` (Kotlin DSL)
- Update to latest Android Gradle Plugin (8.x)
- Update target SDK to 34
- Add Compose dependencies
- Restructure to `android/app/` layout

#### Commit 5: Create base Kotlin application structure
- Create `MainActivity.kt` with Compose
- Create basic theme files
- Create navigation structure with tabs
- Verify app launches (empty tabs)

#### Commit 6: Implement WormholeRepository
- Create `repository/WormholeRepository.kt`
- Implement bridge to Go library
- Add dependency injection (Hilt or manual)

#### Commit 7: Migrate Java utilities to Kotlin
- Convert `Share.java` to `Share.kt` (or keep Java)
- Update to modern Activity Result API
- Convert other Java files if desired

### Phase 3: UI Implementation (4 commits)

#### Commit 8: Implement Receive screen
- Create `ReceiveViewModel.kt`
- Create `ReceiveScreen.kt` with Compose
- Implement code input with paste
- Implement QR code scanning integration
- Implement file receive with progress
- Implement text receive display

#### Commit 9: Implement Send Text screen
- Create `SendTextViewModel.kt`
- Create `SendTextScreen.kt`
- Implement text input
- Implement send with code display
- Implement progress and cancellation

#### Commit 10: Implement Send File screen
- Create `SendFileViewModel.kt`
- Create `SendFileScreen.kt`
- Implement file picker integration
- Implement send with progress
- Implement cancellation

#### Commit 11: Implement Settings screen
- Create `SettingsViewModel.kt`
- Create `SettingsScreen.kt`
- Implement rendezvous URL configuration
- Implement code length configuration
- Use DataStore for persistence

### Phase 4: Polish and Cleanup (3 commits)

#### Commit 12: Implement share intent handling
- Handle ACTION_SEND intents
- Route to appropriate screen
- Handle both text and file shares

#### Commit 13: Add download manager integration
- Register received files with system
- Handle storage permissions (Android 11+)
- Show download notifications

#### Commit 14: Cleanup and documentation
- Remove old Gio UI code (move to separate branch if desired)
- Update README with new build instructions
- Add migration notes
- Update version to 2.0.0

---

## Part 7: Testing Strategy

### 7.1 Unit Tests

**Go Library Tests:**
```go
// wormhole/wormhole_test.go
func TestSendReceiveText(t *testing.T) {
    // Test with mock wormhole client
}

func TestSendReceiveFile(t *testing.T) {
    // Test file transfer
}

func TestCancellation(t *testing.T) {
    // Test mid-transfer cancellation
}
```

**Kotlin Tests:**
```kotlin
// WormholeRepositoryTest.kt
@Test
fun `sending text calls go library correctly`() {
    // Mock Go library, verify calls
}

// ReceiveViewModelTest.kt
@Test
fun `code input normalizes spaces to dashes`() {
    val viewModel = ReceiveViewModel(mockRepository)
    viewModel.onCodeChanged("1 foo bar")
    assertEquals("1-foo-bar", viewModel.uiState.value.code)
}
```

### 7.2 Integration Tests

- Test actual wormhole transfers between two emulators
- Test QR code scanning
- Test share intent handling
- Test file picker integration

### 7.3 Manual Testing Checklist

- [ ] Send text and receive on another device
- [ ] Send file and receive on another device
- [ ] Receive text and file from command-line wormhole
- [ ] QR code scanning works
- [ ] Share from other app (text and file)
- [ ] Cancel mid-transfer
- [ ] Settings persist across restarts
- [ ] Download manager shows received files
- [ ] Works on Android 7 (API 24)
- [ ] Works on Android 14 (API 34)

---

## Part 8: Risk Assessment and Mitigations

### 8.1 Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| gomobile API limitations | High | Medium | Prototype callback interface early |
| Compose learning curve | Medium | Low | Use well-documented patterns |
| Go goroutine/Android lifecycle | High | Medium | Careful cancellation handling |
| Breaking existing users | High | Low | Maintain same package ID, test upgrade |

### 8.2 Compatibility Concerns

- **Package ID:** Keep `io.sanford.wormhole_william` for upgrade path
- **Settings:** Migrate existing config file format
- **Play Store:** Same signing key required

### 8.3 Rollback Plan

- Keep current Gio code on a `gio-legacy` branch
- Tag current release before migration starts
- Ability to revert to Gio if critical issues found

---

## Part 9: Timeline Estimate

This migration can be done incrementally:

**Phase 1 (Infrastructure):** ~2-3 focused sessions
- Flake setup, Go library, build system

**Phase 2 (Android Structure):** ~3-4 focused sessions
- Project modernization, Kotlin setup, repository

**Phase 3 (UI):** ~4-5 focused sessions
- Four screens, each moderately complex

**Phase 4 (Polish):** ~2-3 focused sessions
- Intents, downloads, cleanup

**Total:** ~11-15 focused development sessions

---

## Appendix A: gomobile Callback Pattern

gomobile supports callbacks via interfaces. Here's the pattern:

```go
// Go side
type Callback interface {
    OnResult(data string)
    OnError(err string)
}

func DoSomethingAsync(callback Callback) {
    go func() {
        // ... work ...
        callback.OnResult("done")
    }()
}
```

```kotlin
// Kotlin side
class MyCallback : wormhole.Callback {
    override fun onResult(data: String) {
        runOnUiThread {
            // Update UI
        }
    }

    override fun onError(err: String) {
        runOnUiThread {
            // Show error
        }
    }
}

// Usage
wormhole.doSomethingAsync(MyCallback())
```

---

## Appendix B: Feature Parity Checklist

| Feature | Current (Gio) | New (Compose) |
|---------|--------------|---------------|
| Receive text | ✅ | 🔲 |
| Receive file | ✅ | 🔲 |
| Receive directory (as zip) | ✅ | 🔲 |
| Send text | ✅ | 🔲 |
| Send file | ✅ | 🔲 |
| QR code scanning | ✅ | 🔲 |
| Share intent (text) | ✅ | 🔲 |
| Share intent (file) | ✅ | 🔲 |
| Settings (rendezvous URL) | ✅ | 🔲 |
| Settings (code length) | ✅ | 🔲 |
| Progress display | ✅ | 🔲 |
| Cancel transfer | ✅ | 🔲 |
| Download manager integration | ✅ | 🔲 |
| Copy/paste in editors | ✅ | 🔲 |
| Tab navigation | ✅ | 🔲 |
| Status messages | ✅ | 🔲 |

---

## Appendix C: References

- [gomobile documentation](https://pkg.go.dev/golang.org/x/mobile/cmd/gomobile)
- [Jetpack Compose documentation](https://developer.android.com/jetpack/compose)
- [wormhole-william library](https://github.com/psanford/wormhole-william)
- [Nix flakes documentation](https://nixos.wiki/wiki/Flakes)
- [Material Design 3 for Compose](https://m3.material.io/)
