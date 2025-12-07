package io.sanford.wormhole_william

import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Message
import androidx.compose.material.icons.automirrored.filled.CallReceived
import androidx.compose.material.icons.filled.AttachFile
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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import io.sanford.wormhole_william.ui.ScanQRCodeContract
import io.sanford.wormhole_william.ui.parseWormholeUri
import io.sanford.wormhole_william.ui.screens.ReceiveScreen
import io.sanford.wormhole_william.ui.screens.SendFileScreen
import io.sanford.wormhole_william.ui.screens.SendTextScreen
import io.sanford.wormhole_william.ui.screens.SettingsScreen
import io.sanford.wormhole_william.ui.viewmodel.ReceiveViewModel

data class TabItem(
    val title: String,
    val icon: ImageVector
)

val tabs = listOf(
    TabItem("Receive", Icons.AutoMirrored.Filled.CallReceived),
    TabItem("Text", Icons.AutoMirrored.Filled.Message),
    TabItem("File", Icons.Default.AttachFile),
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
            val parsed = parseWormholeUri(scannedContent)
            if (parsed != null) {
                parsed.rendezvousUrl?.let { receiveViewModel.setRendezvousUrl(it) }
                receiveViewModel.setCode(parsed.code)
            } else {
                // Fall back to using raw content as code
                receiveViewModel.setCode(scannedContent)
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Row(
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Box(
                            modifier = Modifier
                                .size(48.dp)
                                .clip(CircleShape)
                                .background(Color.Black),
                            contentAlignment = Alignment.Center
                        ) {
                            Image(
                                painter = painterResource(id = R.mipmap.ic_launcher_foreground),
                                contentDescription = null,
                                modifier = Modifier.size(48.dp)
                            )
                        }
                        Spacer(modifier = Modifier.width(12.dp))
                        Text("Wormhole William")
                    }
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
