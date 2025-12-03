package io.sanford.wormhole_william.util

import android.content.ContentResolver
import android.content.Context
import android.net.Uri
import android.provider.OpenableColumns
import android.webkit.MimeTypeMap
import java.io.File
import java.io.FileOutputStream

/**
 * Information about a file from a content URI.
 */
data class FileInfo(
    val name: String,
    val size: Long,
    val mimeType: String?
)

/**
 * Gets file information from a content URI.
 */
fun Context.getFileInfo(uri: Uri): FileInfo? {
    return try {
        contentResolver.query(uri, null, null, null, null)?.use { cursor ->
            if (cursor.moveToFirst()) {
                val nameIndex = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                val sizeIndex = cursor.getColumnIndex(OpenableColumns.SIZE)

                val name = if (nameIndex >= 0) cursor.getString(nameIndex) else "unknown"
                val size = if (sizeIndex >= 0) cursor.getLong(sizeIndex) else 0L
                val mimeType = contentResolver.getType(uri)

                FileInfo(name, size, mimeType)
            } else {
                null
            }
        }
    } catch (e: Exception) {
        null
    }
}

/**
 * Copies a content URI to a file in the cache directory.
 * Returns the path to the copied file.
 */
fun Context.copyUriToCache(uri: Uri): String? {
    return try {
        val fileInfo = getFileInfo(uri) ?: return null
        val cacheFile = File(cacheDir, "${System.currentTimeMillis()}_${fileInfo.name}")

        contentResolver.openInputStream(uri)?.use { input ->
            FileOutputStream(cacheFile).use { output ->
                input.copyTo(output)
            }
        }

        cacheFile.absolutePath
    } catch (e: Exception) {
        null
    }
}

/**
 * Gets a filename for a content URI.
 */
fun Context.getFileName(uri: Uri): String? {
    return getFileInfo(uri)?.name
}

/**
 * Detects MIME type from file content (first 512 bytes).
 */
fun detectMimeType(path: String): String {
    return try {
        val file = File(path)
        val extension = file.extension.lowercase()

        // Try extension-based detection first
        MimeTypeMap.getSingleton().getMimeTypeFromExtension(extension)
            ?: "application/octet-stream"
    } catch (e: Exception) {
        "application/octet-stream"
    }
}
