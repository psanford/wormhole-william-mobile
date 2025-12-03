package io.sanford.wormholewilliam.ui.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import io.sanford.wormholewilliam.repository.WormholeRepository
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class SendTextUiState(
    val message: String = "",
    val code: String = "",
    val isTransferring: Boolean = false,
    val status: String = ""
)

class SendTextViewModel(application: Application) : AndroidViewModel(application) {

    private val repository = WormholeRepository.getInstance(application)

    private val _uiState = MutableStateFlow(SendTextUiState())
    val uiState: StateFlow<SendTextUiState> = _uiState.asStateFlow()

    private var currentTransfer: Job? = null

    fun onMessageChanged(message: String) {
        _uiState.update { it.copy(message = message) }
    }

    fun setInitialMessage(message: String) {
        if (_uiState.value.message.isEmpty()) {
            _uiState.update { it.copy(message = message) }
        }
    }

    fun onSend() {
        val message = _uiState.value.message
        if (message.isBlank()) return

        _uiState.update {
            it.copy(
                isTransferring = true,
                status = "Generating code...",
                code = ""
            )
        }

        currentTransfer = viewModelScope.launch {
            repository.sendText(message).collect { state ->
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
                        // Text sends don't really have progress, but handle it anyway
                    }

                    is WormholeRepository.SendState.Complete -> {
                        _uiState.update {
                            it.copy(
                                isTransferring = false,
                                status = "Sent!",
                                code = "",
                                message = ""
                            )
                        }
                    }

                    is WormholeRepository.SendState.Error -> {
                        _uiState.update {
                            it.copy(
                                isTransferring = false,
                                status = "Error: ${state.message}",
                                code = ""
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
                code = ""
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
