PLATFORM_JAR=$(ANDROID_HOME)/platforms/android-30/android.jar
AAR=android/libs/wormhole-william.aar

wormhole-william.debug.apk: $(AAR)
	(cd android && ./gradlew assembleDebug)
	mv android/build/outputs/apk/debug/android-debug.apk $@

$(AAR): $(shell find . -name '*.go' -o -name '*.java' -type f) jgo.jar
	mkdir -p $(@D)
	go run gioui.org/cmd/gogio -buildmode archive -target android -appid io.sanford.wormhole_william -o $@ .

jgo.jar: $(wildcard jgo/*.java)
	mkdir -p classes
	javac -cp "$(PLATFORM_JAR)" -d classes $^
	jar cf $@ -C classes .
	rm -rf classes
