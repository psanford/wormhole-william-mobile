package io.sanford.wormhole_william.ui.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import io.sanford.wormhole_william.repository.WormholeRepository
import io.sanford.wormhole_william.util.detectMimeType
import io.sanford.wormhole_william.util.formatBytes
import io.sanford.wormhole_william.util.notifyDownloadManager
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import wormhole.PendingTransfer

data class ReceiveUiState(
    val code: String = "",
    val isTransferring: Boolean = false,
    val status: String = "",
    val receivedText: String = "",
    val progress: Float = 0f,
    val showAcceptDialog: Boolean = false,
    val pendingFileName: String = "",
    val pendingFileSize: Long = 0L
)

class ReceiveViewModel(application: Application) : AndroidViewModel(application) {

    private val repository = WormholeRepository.getInstance(application)

    private val _uiState = MutableStateFlow(ReceiveUiState())
    val uiState: StateFlow<ReceiveUiState> = _uiState.asStateFlow()

    private var currentTransfer: Job? = null
    private var pendingTransfer: PendingTransfer? = null

    fun onCodeChanged(code: String) {
        // Normalize code: replace spaces with dashes, remove newlines
        val normalized = code
            .replace(" ", "-")
            .replace("\n", "")
            .replace("\r", "")
        _uiState.update { it.copy(code = normalized) }
    }

    fun setCode(code: String) {
        _uiState.update { it.copy(code = code) }
    }

    fun onReceive() {
        val code = _uiState.value.code
        if (code.isBlank()) return

        _uiState.update {
            it.copy(
                isTransferring = true,
                status = "Connecting...",
                receivedText = "",
                progress = 0f
            )
        }

        currentTransfer = viewModelScope.launch {
            repository.receiveWithAccept(code).collect { state ->
                when (state) {
                    is WormholeRepository.ReceiveOfferState.Connecting -> {
                        _uiState.update { it.copy(status = "Connecting...") }
                    }

                    is WormholeRepository.ReceiveOfferState.ReceivedText -> {
                        _uiState.update {
                            it.copy(
                                isTransferring = false,
                                status = "Text received",
                                receivedText = state.text,
                                code = ""
                            )
                        }
                    }

                    is WormholeRepository.ReceiveOfferState.FileOffer -> {
                        pendingTransfer = state.transfer
                        _uiState.update {
                            it.copy(
                                showAcceptDialog = true,
                                pendingFileName = state.name,
                                pendingFileSize = state.size,
                                status = "File offer: ${state.name} (${formatBytes(state.size)})"
                            )
                        }
                    }

                    is WormholeRepository.ReceiveOfferState.FileProgress -> {
                        val progressFloat = if (state.total > 0) {
                            state.received.toFloat() / state.total.toFloat()
                        } else 0f
                        _uiState.update {
                            it.copy(
                                status = "Receiving ${formatBytes(state.received)} / ${formatBytes(state.total)}",
                                progress = progressFloat
                            )
                        }
                    }

                    is WormholeRepository.ReceiveOfferState.FileComplete -> {
                        // Register with download manager
                        val context = getApplication<Application>()
                        val fileName = _uiState.value.pendingFileName
                        val fileSize = _uiState.value.pendingFileSize
                        val mimeType = detectMimeType(state.path)

                        val result = context.notifyDownloadManager(
                            name = fileName,
                            path = state.path,
                            mimeType = mimeType,
                            size = fileSize
                        )

                        val statusMsg = result.fold(
                            onSuccess = { _ -> "File saved to Downloads: $fileName" },
                            onFailure = { e -> "File received but failed to copy to Downloads: ${e.message}" }
                        )

                        _uiState.update {
                            it.copy(
                                isTransferring = false,
                                status = statusMsg,
                                progress = 1f,
                                code = ""
                            )
                        }
                    }

                    is WormholeRepository.ReceiveOfferState.Error -> {
                        _uiState.update {
                            it.copy(
                                isTransferring = false,
                                status = "Error: ${state.message}",
                                progress = 0f
                            )
                        }
                    }
                }
            }
        }
    }

    fun onAcceptFile() {
        _uiState.update {
            it.copy(
                showAcceptDialog = false,
                status = "Receiving ${it.pendingFileName}..."
            )
        }
        pendingTransfer?.accept()
    }

    fun onRejectFile() {
        pendingTransfer?.reject()
        pendingTransfer = null
        _uiState.update {
            it.copy(
                showAcceptDialog = false,
                isTransferring = false,
                status = "Transfer rejected"
            )
        }
    }

    fun onCancel() {
        currentTransfer?.cancel()
        currentTransfer = null
        pendingTransfer?.reject()
        pendingTransfer = null
        repository.cancel()
        _uiState.update {
            it.copy(
                isTransferring = false,
                showAcceptDialog = false,
                status = "Cancelled",
                progress = 0f
            )
        }
    }

    fun clearStatus() {
        _uiState.update { it.copy(status = "") }
    }

    fun clearReceivedText() {
        _uiState.update { it.copy(receivedText = "") }
    }

    override fun onCleared() {
        super.onCleared()
        currentTransfer?.cancel()
        pendingTransfer?.reject()
    }
}
