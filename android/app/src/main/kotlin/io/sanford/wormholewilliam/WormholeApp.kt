package io.sanford.wormholewilliam

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CallReceived
import androidx.compose.material.icons.filled.InsertDriveFile
import androidx.compose.material.icons.filled.Message
import androidx.compose.material.icons.filled.Settings
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.lifecycle.viewmodel.compose.viewModel
import io.sanford.wormholewilliam.ui.ScanQRCodeContract
import io.sanford.wormholewilliam.ui.parseWormholeUri
import io.sanford.wormholewilliam.ui.screens.ReceiveScreen
import io.sanford.wormholewilliam.ui.screens.SendFileScreen
import io.sanford.wormholewilliam.ui.screens.SendTextScreen
import io.sanford.wormholewilliam.ui.screens.SettingsScreen
import io.sanford.wormholewilliam.ui.viewmodel.ReceiveViewModel

data class TabItem(
    val title: String,
    val icon: ImageVector
)

val tabs = listOf(
    TabItem("Receive", Icons.Default.CallReceived),
    TabItem("Send Text", Icons.Default.Message),
    TabItem("Send File", Icons.Default.InsertDriveFile),
    TabItem("Settings", Icons.Default.Settings)
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WormholeApp(
    initialShare: SharedData? = null
) {
    // Determine initial tab based on shared data
    val initialTab = when (initialShare) {
        is SharedData.Text -> 1 // Send Text tab
        is SharedData.File -> 2 // Send File tab
        null -> 0 // Receive tab
    }

    var selectedTab by remember { mutableIntStateOf(initialTab) }

    // Shared ReceiveViewModel so QR scanner can set the code
    val receiveViewModel: ReceiveViewModel = viewModel()

    // QR code scanner launcher
    val qrScannerLauncher = rememberLauncherForActivityResult(
        contract = ScanQRCodeContract()
    ) { result ->
        result?.let { scannedContent ->
            // Parse the QR code content
            val code = parseWormholeUri(scannedContent) ?: scannedContent
            receiveViewModel.setCode(code)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text("Wormhole William")
                },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primary,
                    titleContentColor = MaterialTheme.colorScheme.onPrimary
                )
            )
        },
        bottomBar = {
            NavigationBar {
                tabs.forEachIndexed { index, tab ->
                    NavigationBarItem(
                        icon = { Icon(tab.icon, contentDescription = tab.title) },
                        label = { Text(tab.title) },
                        selected = selectedTab == index,
                        onClick = { selectedTab = index }
                    )
                }
            }
        }
    ) { paddingValues ->
        Surface(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
            color = MaterialTheme.colorScheme.background
        ) {
            when (selectedTab) {
                0 -> ReceiveScreen(
                    viewModel = receiveViewModel,
                    onScanQR = { qrScannerLauncher.launch(Unit) }
                )
                1 -> SendTextScreen(initialText = (initialShare as? SharedData.Text)?.content)
                2 -> SendFileScreen(initialFileUri = (initialShare as? SharedData.File)?.uri)
                3 -> SettingsScreen()
            }
        }
    }
}
