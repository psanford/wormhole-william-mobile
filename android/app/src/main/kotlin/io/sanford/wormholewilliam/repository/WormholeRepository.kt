package io.sanford.wormholewilliam.repository

import android.content.Context
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.callbackFlow
import wormhole.Client
import wormhole.PendingTransfer
import wormhole.ReceiveCallback
import wormhole.ReceiveOfferCallback
import wormhole.SendCallback
import wormhole.Wormhole

/**
 * Repository that wraps the Go wormhole library and provides Kotlin-friendly APIs.
 * Converts callback-based Go APIs to Flow-based Kotlin APIs.
 */
class WormholeRepository(
    private val context: Context,
    private val client: Client
) {
    private val dataDir: String
        get() = context.filesDir.absolutePath

    // ==================== Send Text ====================

    sealed class SendState {
        data object Idle : SendState()
        data class WaitingForReceiver(val code: String) : SendState()
        data class Progress(val sent: Long, val total: Long) : SendState()
        data object Complete : SendState()
        data class Error(val message: String) : SendState()
    }

    fun sendText(message: String): Flow<SendState> = callbackFlow {
        trySend(SendState.Idle)

        client.sendText(message, object : SendCallback {
            override fun onCode(code: String) {
                trySend(SendState.WaitingForReceiver(code))
            }

            override fun onProgress(sent: Long, total: Long) {
                trySend(SendState.Progress(sent, total))
            }

            override fun onComplete() {
                trySend(SendState.Complete)
                close()
            }

            override fun onError(err: String) {
                trySend(SendState.Error(err))
                close()
            }
        })

        awaitClose {
            client.cancel()
        }
    }

    // ==================== Send File ====================

    fun sendFile(path: String, name: String): Flow<SendState> = callbackFlow {
        trySend(SendState.Idle)

        client.sendFile(path, name, object : SendCallback {
            override fun onCode(code: String) {
                trySend(SendState.WaitingForReceiver(code))
            }

            override fun onProgress(sent: Long, total: Long) {
                trySend(SendState.Progress(sent, total))
            }

            override fun onComplete() {
                trySend(SendState.Complete)
                close()
            }

            override fun onError(err: String) {
                trySend(SendState.Error(err))
                close()
            }
        })

        awaitClose {
            client.cancel()
        }
    }

    // ==================== Receive (auto-accept) ====================

    sealed class ReceiveState {
        data object Connecting : ReceiveState()
        data class ReceivedText(val text: String) : ReceiveState()
        data class FileStart(val name: String, val size: Long) : ReceiveState()
        data class FileProgress(val received: Long, val total: Long) : ReceiveState()
        data class FileComplete(val path: String) : ReceiveState()
        data class Error(val message: String) : ReceiveState()
    }

    fun receive(code: String): Flow<ReceiveState> = callbackFlow {
        trySend(ReceiveState.Connecting)

        client.receive(code, object : ReceiveCallback {
            override fun onText(text: String) {
                trySend(ReceiveState.ReceivedText(text))
                close()
            }

            override fun onFileStart(name: String, size: Long) {
                trySend(ReceiveState.FileStart(name, size))
            }

            override fun onFileProgress(received: Long, total: Long) {
                trySend(ReceiveState.FileProgress(received, total))
            }

            override fun onFileComplete(path: String) {
                trySend(ReceiveState.FileComplete(path))
                close()
            }

            override fun onError(err: String) {
                trySend(ReceiveState.Error(err))
                close()
            }
        })

        awaitClose {
            client.cancel()
        }
    }

    // ==================== Receive with Accept/Reject ====================

    sealed class ReceiveOfferState {
        data object Connecting : ReceiveOfferState()
        data class ReceivedText(val text: String) : ReceiveOfferState()
        data class FileOffer(val name: String, val size: Long, val transfer: PendingTransfer) : ReceiveOfferState()
        data class FileProgress(val received: Long, val total: Long) : ReceiveOfferState()
        data class FileComplete(val path: String) : ReceiveOfferState()
        data class Error(val message: String) : ReceiveOfferState()
    }

    fun receiveWithAccept(code: String): Flow<ReceiveOfferState> = callbackFlow {
        trySend(ReceiveOfferState.Connecting)

        client.receiveWithAccept(code, object : ReceiveOfferCallback {
            override fun onText(text: String) {
                trySend(ReceiveOfferState.ReceivedText(text))
                close()
            }

            override fun onFileOffer(pending: PendingTransfer) {
                trySend(ReceiveOfferState.FileOffer(pending.name(), pending.size(), pending))
            }

            override fun onFileProgress(received: Long, total: Long) {
                trySend(ReceiveOfferState.FileProgress(received, total))
            }

            override fun onFileComplete(path: String) {
                trySend(ReceiveOfferState.FileComplete(path))
                close()
            }

            override fun onError(err: String) {
                trySend(ReceiveOfferState.Error(err))
                close()
            }
        })

        awaitClose {
            client.cancel()
        }
    }

    // ==================== Cancel ====================

    fun cancel() {
        client.cancel()
    }

    // ==================== Configuration ====================

    fun getRendezvousURL(): String = client.rendezvousURL

    fun setRendezvousURL(url: String) {
        client.setRendezvousURL(url)
        saveConfig()
    }

    fun getCodeLength(): Int = client.codeLength.toInt()

    fun setCodeLength(length: Int) {
        client.setCodeLength(length.toLong())
        saveConfig()
    }

    private fun saveConfig() {
        val config = wormhole.Config().apply {
            rendezvousURL = client.rendezvousURL
            codeLength = client.codeLength
        }
        Wormhole.saveConfig(dataDir, config)
    }

    companion object {
        @Volatile
        private var instance: WormholeRepository? = null

        fun getInstance(context: Context): WormholeRepository {
            return instance ?: synchronized(this) {
                instance ?: run {
                    val app = context.applicationContext as io.sanford.wormholewilliam.WormholeApplication
                    WormholeRepository(context.applicationContext, app.wormholeClient).also {
                        instance = it
                    }
                }
            }
        }
    }
}
