package io.sanford.wormholewilliam.util

import android.app.DownloadManager
import android.content.ContentValues
import android.content.Context
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import java.io.File
import java.io.FileInputStream

/**
 * Registers a file with Android's Download Manager so it appears in the Downloads app.
 * Also copies the file to the public Downloads directory on Android 10+.
 */
fun Context.notifyDownloadManager(
    name: String,
    path: String,
    mimeType: String,
    size: Long
): Boolean {
    return try {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            // On Android 10+, use MediaStore to copy to Downloads
            copyToDownloadsViaMediaStore(name, path, mimeType, size)
        } else {
            // On older Android, copy to public Downloads and register with DownloadManager
            copyToDownloadsLegacy(name, path, mimeType, size)
        }
    } catch (e: Exception) {
        e.printStackTrace()
        false
    }
}

private fun Context.copyToDownloadsViaMediaStore(
    name: String,
    sourcePath: String,
    mimeType: String,
    size: Long
): Boolean {
    val contentValues = ContentValues().apply {
        put(MediaStore.Downloads.DISPLAY_NAME, name)
        put(MediaStore.Downloads.MIME_TYPE, mimeType)
        put(MediaStore.Downloads.SIZE, size)
        put(MediaStore.Downloads.IS_PENDING, 1)
    }

    val resolver = contentResolver
    val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, contentValues)
        ?: return false

    return try {
        resolver.openOutputStream(uri)?.use { outputStream ->
            FileInputStream(sourcePath).use { inputStream ->
                inputStream.copyTo(outputStream)
            }
        }

        // Mark as complete
        contentValues.clear()
        contentValues.put(MediaStore.Downloads.IS_PENDING, 0)
        resolver.update(uri, contentValues, null, null)

        true
    } catch (e: Exception) {
        // Delete the incomplete file
        resolver.delete(uri, null, null)
        throw e
    }
}

private fun Context.copyToDownloadsLegacy(
    name: String,
    sourcePath: String,
    mimeType: String,
    size: Long
): Boolean {
    // Copy to public Downloads directory
    @Suppress("DEPRECATION")
    val downloadsDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
    val destFile = File(downloadsDir, name)

    FileInputStream(sourcePath).use { input ->
        destFile.outputStream().use { output ->
            input.copyTo(output)
        }
    }

    // Register with DownloadManager
    val downloadManager = getSystemService(Context.DOWNLOAD_SERVICE) as DownloadManager
    downloadManager.addCompletedDownload(
        name,
        "Received via Wormhole William",
        true,
        mimeType,
        destFile.absolutePath,
        size,
        true
    )

    return true
}
