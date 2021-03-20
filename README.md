# wormhole-william-mobile

This is a Magic Wormhole client for Android. (Perhaps someday this will also support iOS).

Some current limitations:
- Receiving directories are kept in zip form.
- Send only supports sending a single file.

## Installing the APK

Prebuilt APKs are provided with each release. You can install this to an android device
that has developer mode enabled by running:

```
apk install wormhole-william.release.apk
```

## Building

In order to build this you will need a local install of the android SDK. Set the environment
variable `ANDROID_SDK_ROOT` AND `ANDROID_ROOT` to the path of the android SDK. Currently
this project is hard coded to use platform `android-30` (in the make file), so you will need
to have that installed (or edit the make file for whatever you have). You will also need
a modern version of Go. Probably >= 1.16.

Run `make` and see what happens!

This project uses https://gioui.org/ for its UI. It uses https://github.com/psanford/wormhole-william
for the underlying wormhole implementation.

## Screenshots

<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/recv1.png?raw=true" alt="Receive 1" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/recv2.png?raw=true" alt="Receive 2" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send_text1.png?raw=true" alt="Send Text 1" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send_text2.png?raw=true" alt="Send Text 2" width="200" />
<img src="https://raw.githubusercontent.com/psanford/wormhole-william-mobile/main/screenshots/send_file1.png?raw=true" alt="Send File" width="200" />
