# Wormhole William Mobile build system
#
# This Makefile builds the Android app using gomobile to create Go bindings.
# The native Android UI (Kotlin/Compose) is built separately by Gradle.

# Directories
GO_WORMHOLE_PKG = ./wormhole
ANDROID_DIR = android
AAR_OUTPUT = $(ANDROID_DIR)/app/libs/wormhole.aar

# Tools (BUILDTOOLS should be set by nix develop or shell.nix)
GOMOBILE = go run golang.org/x/mobile/cmd/gomobile
GRADLE = $(ANDROID_DIR)/gradlew

# Build tools (set BUILDTOOLS env var from nix shell)
ZIPALIGN = $(BUILDTOOLS)/zipalign
APKSIGNER = $(BUILDTOOLS)/apksigner
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

# Generate Go bindings AAR using gomobile
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
	-cd $(ANDROID_DIR) && ./gradlew clean

# Run Go tests
.PHONY: test
test:
	go test -v ./...

# Format Go code
.PHONY: fmt
fmt:
	go fmt ./...

# --- Legacy targets (for backwards compatibility during migration) ---

# Build using old gogio method (deprecated)
LEGACY_AAR = android/libs/wormhole-william.aar

.PHONY: legacy-debug
legacy-debug: $(LEGACY_AAR)
	(cd android && ./gradlew assembleDebug)
	mv android/build/outputs/apk/debug/android-debug.apk wormhole-william.debug.apk

$(LEGACY_AAR): $(shell find . -name '*.go' -o -name '*.java' -o -name '*.xml' -type f)
	mkdir -p $(@D)
	go run gioui.org/cmd/gogio -buildmode archive -target android -appid io.sanford.wormhole_william -o $@ .
