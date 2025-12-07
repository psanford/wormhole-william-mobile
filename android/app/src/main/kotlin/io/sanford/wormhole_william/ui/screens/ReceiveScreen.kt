package io.sanford.wormhole_william.ui.screens

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.QrCodeScanner
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.FilledTonalButton
import androidx.compose.material3.Icon
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
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.style.TextAlign
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
    val scrollState = rememberScrollState()
    val clipboardManager = LocalClipboardManager.current

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(scrollState)
            .padding(16.dp)
    ) {
        // Code input card
        Card(
            modifier = Modifier.fillMaxWidth(),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.surfaceVariant
            )
        ) {
            Column(
                modifier = Modifier.padding(16.dp)
            ) {
                Text(
                    text = "Enter Code",
                    style = MaterialTheme.typography.titleMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Spacer(modifier = Modifier.height(12.dp))

                OutlinedTextField(
                    value = uiState.code,
                    onValueChange = viewModel::onCodeChanged,
                    modifier = Modifier.fillMaxWidth(),
                    placeholder = { Text("e.g. 7-guitarist-revenge") },
                    singleLine = true,
                    enabled = !uiState.isTransferring,
                    textStyle = MaterialTheme.typography.bodyLarge.copy(
                        fontFamily = FontFamily.Monospace
                    )
                )

                // QR Code button
                if (onScanQR != null) {
                    Spacer(modifier = Modifier.height(12.dp))

                    FilledTonalButton(
                        onClick = onScanQR,
                        enabled = !uiState.isTransferring,
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Icon(
                            Icons.Default.QrCodeScanner,
                            contentDescription = null,
                            modifier = Modifier.size(18.dp)
                        )
                        Spacer(modifier = Modifier.width(6.dp))
                        Text("Scan QR Code")
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        // Receive button
        Button(
            onClick = viewModel::onReceive,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(12.dp),
            enabled = !uiState.isTransferring && uiState.code.isNotEmpty()
        ) {
            Text(
                "Receive",
                style = MaterialTheme.typography.titleMedium
            )
        }

        // Cancel button
        if (uiState.isTransferring) {
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedButton(
                onClick = viewModel::onCancel,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(48.dp),
                shape = RoundedCornerShape(12.dp)
            ) {
                Text("Cancel")
            }
        }

        // Progress indicator
        if (uiState.progress > 0f && uiState.isTransferring) {
            Spacer(modifier = Modifier.height(24.dp))
            LinearProgressIndicator(
                progress = { uiState.progress },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(8.dp),
                trackColor = MaterialTheme.colorScheme.surfaceVariant
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "${(uiState.progress * 100).toInt()}%",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.fillMaxWidth(),
                textAlign = TextAlign.Center
            )
        }

        // Received text display
        if (uiState.receivedText.isNotEmpty()) {
            Spacer(modifier = Modifier.height(24.dp))

            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .clickable {
                        clipboardManager.setText(AnnotatedString(uiState.receivedText))
                        viewModel.onTextCopied()
                    },
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.primaryContainer
                )
            ) {
                Column(
                    modifier = Modifier.padding(16.dp)
                ) {
                    Text(
                        text = "Received Text",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.onPrimaryContainer
                    )

                    Spacer(modifier = Modifier.height(12.dp))

                    Text(
                        text = uiState.receivedText,
                        style = MaterialTheme.typography.bodyLarge,
                        color = MaterialTheme.colorScheme.onPrimaryContainer
                    )

                    Spacer(modifier = Modifier.height(8.dp))

                    Text(
                        text = "Tap to copy",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onPrimaryContainer.copy(alpha = 0.7f)
                    )
                }
            }
        }

        Spacer(modifier = Modifier.weight(1f))

        // Status bar
        if (uiState.status.isNotEmpty()) {
            Spacer(modifier = Modifier.height(16.dp))
            Surface(
                color = StatusYellow,
                shape = RoundedCornerShape(8.dp),
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
