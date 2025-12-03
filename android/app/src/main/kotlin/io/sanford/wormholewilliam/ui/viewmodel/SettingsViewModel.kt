package io.sanford.wormholewilliam.ui.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import io.sanford.wormholewilliam.repository.WormholeRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update

data class SettingsUiState(
    val rendezvousUrl: String = "",
    val codeLength: String = ""
)

class SettingsViewModel(application: Application) : AndroidViewModel(application) {

    private val repository = WormholeRepository.getInstance(application)

    private val _uiState = MutableStateFlow(SettingsUiState())
    val uiState: StateFlow<SettingsUiState> = _uiState.asStateFlow()

    init {
        // Load current settings
        _uiState.update {
            it.copy(
                rendezvousUrl = repository.getRendezvousURL(),
                codeLength = repository.getCodeLength().takeIf { len -> len > 0 }?.toString() ?: ""
            )
        }
    }

    fun onRendezvousUrlChanged(url: String) {
        _uiState.update { it.copy(rendezvousUrl = url) }
        repository.setRendezvousURL(url)
    }

    fun onCodeLengthChanged(length: String) {
        _uiState.update { it.copy(codeLength = length) }

        // Only save valid positive integers
        val lengthInt = length.toIntOrNull()
        if (lengthInt != null && lengthInt > 0) {
            repository.setCodeLength(lengthInt)
        }
    }
}
