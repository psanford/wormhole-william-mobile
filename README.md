# wormhole-william-mobile

This is a Magic Wormhole client for Android. (Perhaps someday this will also support iOS).

At the moment its in an Alpha state. The basic functionality is there, but the UI/UX is
quite rough.

In order to build this you will need a local install of the android SDK. Set the environment
variable `ANDROID_SDK_ROOT` AND `ANDROID_ROOT` to the path of the android SDK. Currently
this project is hard coded to use platform `android-30` (in the make file), so you will need
to have that installed (or edit the make file for whatever you have). You will also need
a modern version of Go. Probably >= 1.16.

Run `make` and see what happens!

This project uses https://gioui.org/ for its UI. It uses https://github.com/psanford/wormhole-william
for the underlying wormhole implementation.
