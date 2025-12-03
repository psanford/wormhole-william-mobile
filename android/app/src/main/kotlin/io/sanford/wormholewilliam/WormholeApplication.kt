package io.sanford.wormholewilliam

import android.app.Application

class WormholeApplication : Application() {

    lateinit var wormholeClient: wormhole.Client
        private set

    override fun onCreate() {
        super.onCreate()

        // Initialize the Go wormhole client
        val dataDir = filesDir.absolutePath
        wormholeClient = wormhole.Client(dataDir)

        // Load saved configuration
        val config = wormhole.Wormhole.loadConfig(dataDir)
        if (!config.rendezvousURL.isNullOrEmpty()) {
            wormholeClient.setRendezvousURL(config.rendezvousURL)
        }
        if (config.codeLength > 0) {
            wormholeClient.setCodeLength(config.codeLength.toInt())
        }
    }
}
