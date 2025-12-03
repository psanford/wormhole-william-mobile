# Proguard rules for Wormhole William Mobile

# Keep Go library classes
-keep class wormhole.** { *; }
-keep class go.** { *; }

# Keep callback interfaces that Go code calls
-keep interface wormhole.SendCallback { *; }
-keep interface wormhole.ReceiveCallback { *; }
-keep interface wormhole.ReceiveOfferCallback { *; }

# ZXing
-keep class com.google.zxing.** { *; }
-keep class com.journeyapps.** { *; }
