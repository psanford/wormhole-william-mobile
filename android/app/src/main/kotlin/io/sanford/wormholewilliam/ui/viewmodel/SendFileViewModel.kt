package io.sanford.wormholewilliam.ui.viewmodel

import android.app.Application
import android.net.Uri
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import io.sanford.wormholewilliam.repository.WormholeRepository
import io.sanford.wormholewilliam.util.copyUriToCache
import io.sanford.wormholewilliam.util.formatBytes
import io.sanford.wormholewilliam.util.getFileName
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

data class SendFileUiState(
    val fileName: String = "",
    val filePath: String = "",
    val code: String = "",
    val isTransferring: Boolean = false,
    val isPreparing: Boolean = false,
    val status: String = "",
    val progress: Float = 0f
)

class SendFileViewModel(application: Application) : AndroidViewModel(application) {

    private val repository = WormholeRepository.getInstance(application)

    private val _uiState = MutableStateFlow(SendFileUiState())
    val uiState: StateFlow<SendFileUiState> = _uiState.asStateFlow()

    private var currentTransfer: Job? = null

    fun onFileSelected(uri: Uri) {
        _uiState.update { it.copy(isPreparing = true, status = "Preparing file...") }

        viewModelScope.launch {
            val context = getApplication<Application>()
            val fileName = context.getFileName(uri) ?: "unknown"

            // Copy file to cache (required for Go to read it)
            val cachedPath = withContext(Dispatchers.IO) {
                context.copyUriToCache(uri)
            }

            if (cachedPath != null) {
                _uiState.update {
                    it.copy(
                        fileName = fileName,
                        filePath = cachedPath,
                        isPreparing = false,
                        status = "Ready to send: $fileName"
                    )
                }
            } else {
                _uiState.update {
                    it.copy(
                        isPreparing = false,
                        status = "Error: Could not prepare file"
                    )
                }
            }
        }
    }

    fun setInitialFile(uri: Uri) {
        if (_uiState.value.filePath.isEmpty()) {
            onFileSelected(uri)
        }
    }

    fun onSend() {
        val path = _uiState.value.filePath
        val name = _uiState.value.fileName
        if (path.isBlank() || name.isBlank()) return

        _uiState.update {
            it.copy(
                isTransferring = true,
                status = "Generating code...",
                code = "",
                progress = 0f
            )
        }

        currentTransfer = viewModelScope.launch {
            repository.sendFile(path, name).collect { state ->
                when (state) {
                    is WormholeRepository.SendState.Idle -> {
                        // Initial state
                    }

                    is WormholeRepository.SendState.WaitingForReceiver -> {
                        _uiState.update {
                            it.copy(
                                code = state.code,
                                status = "Waiting for receiver..."
                            )
                        }
                    }

                    is WormholeRepository.SendState.Progress -> {
                        val progressFloat = if (state.total > 0) {
                            state.sent.toFloat() / state.total.toFloat()
                        } else 0f
                        _uiState.update {
                            it.copy(
                                status = "Sending ${formatBytes(state.sent)} / ${formatBytes(state.total)}",
                                progress = progressFloat
                            )
                        }
                    }

                    is WormholeRepository.SendState.Complete -> {
                        _uiState.update {
                            it.copy(
                                isTransferring = false,
                                status = "Send Complete!",
                                code = "",
                                fileName = "",
                                filePath = "",
                                progress = 1f
                            )
                        }
                    }

                    is WormholeRepository.SendState.Error -> {
                        _uiState.update {
                            it.copy(
                                isTransferring = false,
                                status = "Error: ${state.message}",
                                code = "",
                                progress = 0f
                            )
                        }
                    }
                }
            }
        }
    }

    fun onCancel() {
        currentTransfer?.cancel()
        currentTransfer = null
        repository.cancel()
        _uiState.update {
            it.copy(
                isTransferring = false,
                status = "Cancelled",
                code = "",
                progress = 0f
            )
        }
    }

    fun clearFile() {
        _uiState.update {
            it.copy(
                fileName = "",
                filePath = "",
                status = ""
            )
        }
    }

    fun clearStatus() {
        _uiState.update { it.copy(status = "") }
    }

    override fun onCleared() {
        super.onCleared()
        currentTransfer?.cancel()
    }
}
