package io.sanford.wormhole_william.ui.screens

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material.icons.filled.ContentPaste
import androidx.compose.material.icons.filled.QrCodeScanner
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import io.sanford.wormhole_william.ui.theme.StatusYellow
import io.sanford.wormhole_william.ui.theme.StatusYellowText
import io.sanford.wormhole_william.ui.viewmodel.ReceiveViewModel
import io.sanford.wormhole_william.util.formatBytes

@Composable
fun ReceiveScreen(
    viewModel: ReceiveViewModel = viewModel(),
    onScanQR: (() -> Unit)? = null
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val scrollState = rememberScrollState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(scrollState)
            .padding(16.dp)
    ) {
        // Code input section
        Text(
            text = "Code",
            style = MaterialTheme.typography.titleLarge
        )

        Spacer(modifier = Modifier.height(8.dp))

        OutlinedTextField(
            value = uiState.code,
            onValueChange = viewModel::onCodeChanged,
            modifier = Modifier.fillMaxWidth(),
            placeholder = { Text("Enter wormhole code") },
            singleLine = true,
            trailingIcon = {
                IconButton(onClick = {
                    val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                    clipboard.primaryClip?.getItemAt(0)?.text?.toString()?.let { text ->
                        viewModel.onCodeChanged(text)
                    }
                }) {
                    Icon(Icons.Default.ContentPaste, contentDescription = "Paste")
                }
            },
            enabled = !uiState.isTransferring
        )

        Spacer(modifier = Modifier.height(16.dp))

        // QR Code button
        if (onScanQR != null) {
            Button(
                onClick = onScanQR,
                modifier = Modifier.fillMaxWidth(),
                enabled = !uiState.isTransferring && uiState.code.isEmpty()
            ) {
                Icon(
                    Icons.Default.QrCodeScanner,
                    contentDescription = null,
                    modifier = Modifier.padding(end = 8.dp)
                )
                Text("Scan QR Code")
            }

            Spacer(modifier = Modifier.height(8.dp))
        }

        // Receive button
        Button(
            onClick = viewModel::onReceive,
            modifier = Modifier.fillMaxWidth(),
            enabled = !uiState.isTransferring && uiState.code.isNotEmpty()
        ) {
            Text("Receive")
        }

        // Cancel button
        if (uiState.isTransferring) {
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedButton(
                onClick = viewModel::onCancel,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("Cancel")
            }
        }

        // Progress indicator
        if (uiState.progress > 0f && uiState.isTransferring) {
            Spacer(modifier = Modifier.height(16.dp))
            LinearProgressIndicator(
                progress = { uiState.progress },
                modifier = Modifier.fillMaxWidth()
            )
        }

        // Received text display
        if (uiState.receivedText.isNotEmpty()) {
            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Received Text",
                style = MaterialTheme.typography.titleMedium
            )

            Spacer(modifier = Modifier.height(8.dp))

            OutlinedTextField(
                value = uiState.receivedText,
                onValueChange = {},
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp),
                readOnly = true,
                trailingIcon = {
                    IconButton(onClick = {
                        val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                        val clip = ClipData.newPlainText("Received text", uiState.receivedText)
                        clipboard.setPrimaryClip(clip)
                    }) {
                        Icon(Icons.Default.ContentCopy, contentDescription = "Copy")
                    }
                }
            )
        }

        Spacer(modifier = Modifier.weight(1f))

        // Status bar
        if (uiState.status.isNotEmpty()) {
            Spacer(modifier = Modifier.height(16.dp))
            Surface(
                color = StatusYellow,
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(
                    text = uiState.status,
                    modifier = Modifier.padding(16.dp),
                    style = MaterialTheme.typography.bodyMedium,
                    color = StatusYellowText
                )
            }
        }
    }

    // Accept/Reject Dialog
    if (uiState.showAcceptDialog) {
        AlertDialog(
            onDismissRequest = { /* Don't allow dismissing */ },
            title = { Text("Accept File?") },
            text = {
                Column {
                    Text("Filename: ${uiState.pendingFileName}")
                    Text("Size: ${formatBytes(uiState.pendingFileSize)}")
                }
            },
            confirmButton = {
                Button(onClick = viewModel::onAcceptFile) {
                    Text("Accept")
                }
            },
            dismissButton = {
                TextButton(onClick = viewModel::onRejectFile) {
                    Text("Reject")
                }
            }
        )
    }
}
