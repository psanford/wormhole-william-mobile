PLATFORM_JAR=$(ANDROID_HOME)/platforms/android-30/android.jar

wormhole-william.apk: $(wildcard *.go) $(wildcard **/*.go) jgo.jar
	go run gioui.org/cmd/gogio -target android -appid io.sanford.wormhole_william .

jgo.jar: jgo/Jni.java
	mkdir -p classes
	javac -cp "$(PLATFORM_JAR)" -d classes $^
	jar cf $@ -C classes .
	rm -rf classes
