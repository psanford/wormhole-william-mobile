package ui

import (
	"errors"

	"gioui.org/io/event"
	"github.com/psanford/wormhole-william-mobile/internal/picker"
)

func newPlatformHandler() platformHandler {
	return &iosPlatform{}
}

type iosPlatform struct {
}

func (d *iosPlatform) handleEvent(e event.Event) {
}

func (d *iosPlatform) pickFile() <-chan picker.PickResult {
	ch := make(chan picker.PickResult, 1)
	ch <- picker.PickResult{
		Err: errors.New("pick file not implemented on ios"),
	}
	return ch
}

func (d *iosPlatform) notifyDownloadManager(name, path, contentType string, size int64) {
}

func (d *iosPlatform) sharedEventCh() chan picker.SharedEvent {
	return nil
}

func (d *iosPlatform) scanQRCode() <-chan string {
	return nil
}

func (d *iosPlatform) requestWriteFilePerm() <-chan picker.PermResult {
	ch := make(chan picker.PermResult, 1)
	ch <- picker.PermResult{
		Authorized: true,
	}

	return ch
}
