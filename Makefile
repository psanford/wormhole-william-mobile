PLATFORM_JAR=$(ANDROID_HOME)/platforms/android-30/android.jar

wormhole-william.apk: $(wildcard *.go) $(wildcard **/*.go)
	go run gioui.org/cmd/gogio -target android -appid io.sanford.wormhole_william .
