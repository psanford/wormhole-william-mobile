package io.sanford.wormhole_william.ui.screens

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import io.sanford.wormhole_william.ui.viewmodel.SettingsViewModel

// Default rendezvous URL from wormhole-william
private const val DEFAULT_RENDEZVOUS_URL = "wss://mailbox.mw.leastauthority.com/v1"

@Composable
fun SettingsScreen(
    viewModel: SettingsViewModel = viewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val scrollState = rememberScrollState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(scrollState)
            .padding(16.dp)
    ) {
        // Rendezvous URL
        Text(
            text = "Rendezvous URL",
            style = MaterialTheme.typography.titleLarge
        )

        Spacer(modifier = Modifier.height(8.dp))

        OutlinedTextField(
            value = uiState.rendezvousUrl,
            onValueChange = viewModel::onRendezvousUrlChanged,
            modifier = Modifier.fillMaxWidth(),
            placeholder = { Text(DEFAULT_RENDEZVOUS_URL) },
            singleLine = true
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "The relay server used to coordinate transfers. Leave empty to use the default.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )

        Spacer(modifier = Modifier.height(24.dp))

        // Code Length
        Text(
            text = "Code Length",
            style = MaterialTheme.typography.titleLarge
        )

        Spacer(modifier = Modifier.height(8.dp))

        OutlinedTextField(
            value = uiState.codeLength,
            onValueChange = { value ->
                // Only allow digits
                if (value.all { it.isDigit() }) {
                    viewModel.onCodeLengthChanged(value)
                }
            },
            modifier = Modifier.fillMaxWidth(),
            placeholder = { Text("2") },
            singleLine = true,
            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number)
        )

        Spacer(modifier = Modifier.height(8.dp))

        Text(
            text = "Number of words in the generated code. Default is 2.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}
