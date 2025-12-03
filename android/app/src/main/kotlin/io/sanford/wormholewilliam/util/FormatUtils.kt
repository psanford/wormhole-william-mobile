package io.sanford.wormholewilliam.util

/**
 * Formats a byte count into a human-readable string.
 * Example: 1536 -> "1.5 kB"
 */
fun formatBytes(bytes: Long): String {
    if (bytes < 1000) {
        return "$bytes B"
    }

    var value = bytes.toDouble()
    val units = arrayOf("kB", "MB", "GB", "TB", "PB", "EB")

    for (unit in units) {
        value /= 1000
        if (value < 1000) {
            return "%.1f %s".format(value, unit)
        }
    }

    return "%.1f %s".format(value, units.last())
}

/**
 * Formats transfer progress as a string.
 * Example: "1.5 MB / 10.0 MB"
 */
fun formatProgress(current: Long, total: Long): String {
    return "${formatBytes(current)} / ${formatBytes(total)}"
}

/**
 * Calculates progress as a float between 0 and 1.
 */
fun calculateProgress(current: Long, total: Long): Float {
    return if (total > 0) {
        (current.toFloat() / total.toFloat()).coerceIn(0f, 1f)
    } else {
        0f
    }
}
