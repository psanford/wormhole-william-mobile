package ui

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"github.com/psanford/wormhole-william-mobile/internal/picker"
	"github.com/psanford/wormhole-william-mobile/jgo"
)

type androidPlatform struct {
	viewEvent app.ViewEvent
}

func newPlatformHandler() platformHandler {
	return &androidPlatform{}
}

func (a *androidPlatform) handleEvent(e event.Event) {
	switch e := e.(type) {
	case app.ViewEvent:
		a.viewEvent = e
	}
}
func (a *androidPlatform) pickFile() <-chan picker.PickResult {
	return jgo.PickFile(a.viewEvent)
}

func (a *androidPlatform) notifyDownloadManager(name, path, contentType string, size int64) {
	jgo.NotifyDownloadManager(a.viewEvent, name, path, contentType, size)

}
func (a *androidPlatform) sharedEventCh() chan picker.SharedEvent {
	return jgo.GetSharedEventCh()
}

func (a *androidPlatform) scanQRCode() <-chan string {
	return jgo.ScanQRCode(a.viewEvent)
}

func (a *androidPlatform) requestWriteFilePerm() <-chan picker.PermResult {
	return jgo.RequestWriteFilePermission(a.viewEvent)
}
