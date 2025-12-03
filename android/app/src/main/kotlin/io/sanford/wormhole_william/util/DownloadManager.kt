package io.sanford.wormhole_william.util

import android.app.DownloadManager
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.ContentValues
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import io.sanford.wormhole_william.R
import java.io.File
import java.io.FileInputStream

private const val CHANNEL_ID = "wormhole_downloads"
private const val NOTIFICATION_ID = 1001

/**
 * Registers a file with Android's Download Manager so it appears in the Downloads app.
 * Also copies the file to the public Downloads directory on Android 10+.
 */
fun Context.notifyDownloadManager(
    name: String,
    path: String,
    mimeType: String,
    size: Long
): Result<Uri> {
    return try {
        val uri = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
            // On Android 10+, use MediaStore to copy to Downloads
            copyToDownloadsViaMediaStore(name, path, mimeType, size)
        } else {
            // On older Android, copy to public Downloads and register with DownloadManager
            copyToDownloadsLegacy(name, path, mimeType, size)
        }

        // Show notification
        showDownloadCompleteNotification(name, uri, mimeType)

        Result.success(uri)
    } catch (e: Exception) {
        e.printStackTrace()
        Result.failure(e)
    }
}

private fun Context.copyToDownloadsViaMediaStore(
    name: String,
    sourcePath: String,
    mimeType: String,
    size: Long
): Uri {
    val contentValues = ContentValues().apply {
        put(MediaStore.Downloads.DISPLAY_NAME, name)
        put(MediaStore.Downloads.MIME_TYPE, mimeType)
        put(MediaStore.Downloads.SIZE, size)
        put(MediaStore.Downloads.IS_PENDING, 1)
    }

    val resolver = contentResolver
    val uri = resolver.insert(MediaStore.Downloads.EXTERNAL_CONTENT_URI, contentValues)
        ?: throw IllegalStateException("Failed to create MediaStore entry for Downloads")

    try {
        val outputStream = resolver.openOutputStream(uri)
            ?: throw IllegalStateException("Failed to open output stream for Downloads")

        outputStream.use { out ->
            FileInputStream(sourcePath).use { input ->
                input.copyTo(out)
            }
        }

        // Mark as complete
        contentValues.clear()
        contentValues.put(MediaStore.Downloads.IS_PENDING, 0)
        resolver.update(uri, contentValues, null, null)

        return uri
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
): Uri {
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
    @Suppress("DEPRECATION")
    downloadManager.addCompletedDownload(
        name,
        "Received via Wormhole William",
        true,
        mimeType,
        destFile.absolutePath,
        size,
        true
    )

    return Uri.fromFile(destFile)
}

private fun Context.showDownloadCompleteNotification(name: String, uri: Uri, mimeType: String) {
    // Create notification channel for Android O+
    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
        val channel = NotificationChannel(
            CHANNEL_ID,
            "Downloads",
            NotificationManager.IMPORTANCE_DEFAULT
        ).apply {
            description = "Wormhole file transfer notifications"
        }
        val notificationManager = getSystemService(NotificationManager::class.java)
        notificationManager.createNotificationChannel(channel)
    }

    // Create intent to open the file
    val openIntent = Intent(Intent.ACTION_VIEW).apply {
        setDataAndType(uri, mimeType)
        addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
    }

    val pendingIntent = PendingIntent.getActivity(
        this,
        0,
        openIntent,
        PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
    )

    val notification = NotificationCompat.Builder(this, CHANNEL_ID)
        .setSmallIcon(android.R.drawable.stat_sys_download_done)
        .setContentTitle("Download complete")
        .setContentText(name)
        .setPriority(NotificationCompat.PRIORITY_DEFAULT)
        .setContentIntent(pendingIntent)
        .setAutoCancel(true)
        .build()

    try {
        NotificationManagerCompat.from(this).notify(NOTIFICATION_ID, notification)
    } catch (e: SecurityException) {
        // Notification permission not granted, ignore
    }
}
