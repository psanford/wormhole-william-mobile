//go:build !(android || ios)
// +build !android,!ios

package ui

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

func (d *dummyPlatform) scanQRCode() <-chan string {
	return nil
}

func (d *dummyPlatform) requestWriteFilePerm() <-chan picker.PermResult {
	ch := make(chan picker.PermResult, 1)
	ch <- picker.PermResult{
		Authorized: true,
	}

	return ch
}

func (d *dummyPlatform) supportedFeatures() platformFeature {
	return 0
}

func (d *dummyPlatform) dstFilePathClear(dataDir, name string) error {
	path := filepath.Join(dataDir, name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("Error refusing to overwrite existing '%s'", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Error stat'ing existing '%s'\n", path)
	}

	return nil
}

func (d *dummyPlatform) saveFileToDocuments(src *os.File, dataDir, name string) (string, error) {
	path := filepath.Join(dataDir, name)

	err := os.Rename(src.Name(), path)
	return path, err
}
