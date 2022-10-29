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

func (d *androidPlatform) supportedFeatures() platformFeature {
	return supportsQRScanning
}

func (d *androidPlatform) dstFilePathClear(dataDir, name string) error {
	path := filepath.Join(dataDir, name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("Error refusing to overwrite existing '%s'", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Error stat'ing existing '%s'\n", path)
	}

	return nil
}

func (d *androidPlatform) saveFileToDocuments(src *os.File, dataDir, name string) (string, error) {
	path := filepath.Join(dataDir, name)

	err = os.Rename(src.Name(), path)
	return path, err
}
