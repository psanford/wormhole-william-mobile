package io.sanford.wormhole_william.ui

import android.app.Activity
import android.content.Context
import android.content.Intent
import androidx.activity.result.contract.ActivityResultContract
import com.google.zxing.integration.android.IntentIntegrator
import com.google.zxing.integration.android.IntentResult
import java.net.URLDecoder

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
 * Result of parsing a wormhole URI.
 */
data class ParsedWormholeUri(
    val code: String,
    val rendezvousUrl: String? = null
)

/**
 * Parses a wormhole URI from a QR code.
 *
 * Supported formats:
 * 1. wormhole:<rendezvousUrl>?code=<code>
 *    Example: wormhole:ws://relay.magic-wormhole.io:4000/v1?code=5-souvenir-scallion
 *
 * 2. wormhole://<host>?code=<code>
 *    Example: wormhole://relay.example.com?code=123-456-789
 *
 * 3. wormhole-transfer:<code>?rendezvous=<url-encoded-rendezvous>
 *    Example: wormhole-transfer:4-hurricane-equipment?rendezvous=ws%3A%2F%2Fcustom.relay.com%3A4000
 *
 * 4. Plain code (contains dash, no "://")
 *    Example: 5-souvenir-scallion
 *
 * Returns a ParsedWormholeUri with the code and optional rendezvous URL, or null if invalid.
 */
fun parseWormholeUri(uri: String): ParsedWormholeUri? {
    return try {
        when {
            uri.startsWith("wormhole-transfer:") -> parseWormholeTransferUri(uri)
            uri.startsWith("wormhole://") -> parseWormholeDoubleSlashUri(uri)
            uri.startsWith("wormhole:") -> parseWormholeSingleColonUri(uri)
            // Plain code: contains dash but no "://"
            uri.contains("-") && !uri.contains("://") -> ParsedWormholeUri(code = uri.trim())
            else -> null
        }
    } catch (e: Exception) {
        null
    }
}

/**
 * Parses wormhole-transfer: URI format.
 * Format: wormhole-transfer:<code>?rendezvous=<url-encoded-rendezvous>&role=...&version=...
 */
private fun parseWormholeTransferUri(uri: String): ParsedWormholeUri? {
    val withoutScheme = uri.removePrefix("wormhole-transfer:")

    // Split code from query parameters
    val queryIndex = withoutScheme.indexOf('?')
    val (encodedCode, queryString) = if (queryIndex != -1) {
        withoutScheme.substring(0, queryIndex) to withoutScheme.substring(queryIndex + 1)
    } else {
        withoutScheme to null
    }

    // URL-decode the code
    val code = URLDecoder.decode(encodedCode, "UTF-8")
    if (code.isEmpty()) return null

    // Parse rendezvous from query parameters if present
    val rendezvousUrl = queryString?.let { query ->
        query.split("&")
            .map { it.split("=", limit = 2) }
            .find { it.size == 2 && it[0] == "rendezvous" }
            ?.get(1)
            ?.let { URLDecoder.decode(it, "UTF-8") }
    }

    return ParsedWormholeUri(code = code, rendezvousUrl = rendezvousUrl)
}

/**
 * Parses wormhole:// URI format (double slash).
 * Format: wormhole://<host>?code=<code>&other=...
 */
private fun parseWormholeDoubleSlashUri(uri: String): ParsedWormholeUri? {
    // Validate URL structure - reject URLs with spaces
    if (uri.contains(" ")) return null

    val withoutScheme = uri.removePrefix("wormhole://")

    // Split host from query string
    val queryIndex = withoutScheme.indexOf('?')
    if (queryIndex == -1) return null

    val host = withoutScheme.substring(0, queryIndex)
    if (host.isEmpty()) return null

    val queryString = withoutScheme.substring(queryIndex + 1)

    // Parse query parameters to find code
    val code = queryString.split("&")
        .map { it.split("=", limit = 2) }
        .find { it.size == 2 && it[0] == "code" }
        ?.get(1)

    if (code.isNullOrEmpty()) return null

    return ParsedWormholeUri(code = code, rendezvousUrl = host)
}

/**
 * Parses wormhole: URI format (single colon, not double slash).
 * Format: wormhole:<rendezvousUrl>?code=<code>
 * Example: wormhole:ws://relay.magic-wormhole.io:4000/v1?code=5-word-code
 */
private fun parseWormholeSingleColonUri(uri: String): ParsedWormholeUri? {
    // Find the code parameter - could be ?code= or &code=
    val codeRegex = Regex("[?&]code=([^&]*)")
    val match = codeRegex.find(uri)

    if (match == null) return null

    val code = match.groupValues[1].trim()
    if (code.isEmpty()) return null

    // Extract rendezvous URL: everything between "wormhole:" and the query string
    val withoutScheme = uri.removePrefix("wormhole:")
    val queryIndex = withoutScheme.indexOf('?')
    if (queryIndex == -1) return null

    val rendezvousUrl = withoutScheme.substring(0, queryIndex)

    return ParsedWormholeUri(code = code, rendezvousUrl = rendezvousUrl)
}
