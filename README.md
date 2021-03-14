# wormhole-william-mobile

This is a Magic Wormhole client for Android. (Perhaps someday this will also support iOS).

At the moment its in beta state. Transfers work although you might experience some rough edges.

## Screenshots

![recv1.png](screenshots/recv1.png?raw=true "Receive 1")
![recv2.png](screenshots/recv2.png?raw=true "Receive 2")
![send_text1.png](screenshots/send_text1.png?raw=true "Send Text 1")
![send_text2.png](screenshots/send_text2.png?raw=true "Send Text 2")
![send_file1.png](screenshots/send_file1.png?raw=true "Send File")

## Building

In order to build this you will need a local install of the android SDK. Set the environment
variable `ANDROID_SDK_ROOT` AND `ANDROID_ROOT` to the path of the android SDK. Currently
this project is hard coded to use platform `android-30` (in the make file), so you will need
to have that installed (or edit the make file for whatever you have). You will also need
a modern version of Go. Probably >= 1.16.

Run `make` and see what happens!

This project uses https://gioui.org/ for its UI. It uses https://github.com/psanford/wormhole-william
for the underlying wormhole implementation.
