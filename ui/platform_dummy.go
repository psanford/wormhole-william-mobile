// +build !android

package ui

import (
	"errors"

	"gioui.org/io/event"
	"github.com/psanford/wormhole-william-mobile/internal/picker"
)

func newPlatformHandler() platformHandler {
	return &dummyPlatform{}
}

type dummyPlatform struct {
}

func (d *dummyPlatform) handleEvent(e event.Event) {
}

func (d *dummyPlatform) pickFile() <-chan picker.PickResult {
	ch := make(chan picker.PickResult, 1)
	ch <- picker.PickResult{
		Err: errors.New("pick file not implemented"),
	}
	return ch
}

func (d *dummyPlatform) notifyDownloadManager(name, path, contentType string, size int64) {
}

func (d *dummyPlatform) sharedEventCh() chan picker.SharedEvent {
	return nil
}
