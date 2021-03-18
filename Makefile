PLATFORM_JAR=$(ANDROID_HOME)/platforms/android-30/android.jar

wormhole-william.debug.apk: $(wildcard *.go) $(wildcard **/*.go) jgo.jar
	go run gioui.org/cmd/gogio -target android -icon wormhole-william-icon.png -appid io.sanford.wormhole_william -o $@ .

wormhole-william.release.apk: $(wildcard *.go) $(wildcard **/*.go) jgo.jar
	go run gioui.org/cmd/gogio -target android -icon wormhole-william-icon.png -appid io.sanford.wormhole_william -signkey ~/.android-release/wormhole-william-release.keystore -o $@ .


jgo.jar: $(wildcard jgo/*.java)
	mkdir -p classes
	javac -cp "$(PLATFORM_JAR)" -d classes $^
	jar cf $@ -C classes .
	rm -rf classes
