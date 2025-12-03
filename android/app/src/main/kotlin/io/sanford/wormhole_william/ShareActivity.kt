package io.sanford.wormhole_william

import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.os.Parcelable
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import io.sanford.wormhole_william.ui.theme.WormholeTheme

/**
 * Activity that handles ACTION_SEND intents from other apps.
 * Receives shared text or files and launches the appropriate send flow.
 */
class ShareActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val sharedData = handleShareIntent(intent)

        setContent {
            WormholeTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    WormholeApp(initialShare = sharedData)
                }
            }
        }
    }

    private fun handleShareIntent(intent: Intent): SharedData? {
        if (intent.action != Intent.ACTION_SEND) {
            return null
        }

        val type = intent.type ?: return null

        return when {
            type == "text/plain" -> {
                val text = intent.getStringExtra(Intent.EXTRA_TEXT)
                text?.let { SharedData.Text(it) }
            }
            intent.hasExtra(Intent.EXTRA_STREAM) -> {
                @Suppress("DEPRECATION")
                val uri = intent.getParcelableExtra<Parcelable>(Intent.EXTRA_STREAM) as? Uri
                uri?.let { SharedData.File(it) }
            }
            else -> null
        }
    }
}

sealed class SharedData {
    data class Text(val content: String) : SharedData()
    data class File(val uri: Uri) : SharedData()
}
