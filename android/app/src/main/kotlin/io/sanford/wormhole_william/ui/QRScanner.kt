package io.sanford.wormhole_william.ui

import android.app.Activity
import android.content.Context
import android.content.Intent
import androidx.activity.result.contract.ActivityResultContract
import com.google.zxing.integration.android.IntentIntegrator
import com.google.zxing.integration.android.IntentResult

/**
 * Activity Result Contract for ZXing QR code scanning.
 */
class ScanQRCodeContract : ActivityResultContract<Unit, String?>() {

    override fun createIntent(context: Context, input: Unit): Intent {
        val integrator = IntentIntegrator(context as Activity).apply {
            setDesiredBarcodeFormats(IntentIntegrator.QR_CODE)
            setPrompt("Scan a wormhole QR code")
            setBeepEnabled(false)
            setOrientationLocked(true)
        }
        return integrator.createScanIntent()
    }

    override fun parseResult(resultCode: Int, intent: Intent?): String? {
        val result: IntentResult? = IntentIntegrator.parseActivityResult(resultCode, intent)
        return result?.contents
    }
}

/**
 * Parses a wormhole URI from a QR code.
 *
 * Format: wormhole://?code=<code>
 * Returns just the code portion, or null if invalid.
 */
fun parseWormholeUri(uri: String): String? {
    // Handle direct code (not a URI)
    if (!uri.startsWith("wormhole:")) {
        // Might be a plain code, check if it looks valid
        return if (uri.contains("-") && !uri.contains("://")) {
            uri.trim()
        } else {
            null
        }
    }

    // Parse wormhole URI
    val withoutScheme = uri.removePrefix("wormhole:")

    // Try to extract code parameter
    return try {
        val url = android.net.Uri.parse("http://dummy$withoutScheme")
        url.getQueryParameter("code")
    } catch (e: Exception) {
        null
    }
}
