package wormhole

import (
	"context"
	"io"
	"os"
	"path/filepath"

	wh "github.com/psanford/wormhole-william/wormhole"
)

// PendingTransfer represents a file transfer waiting for acceptance.
// Call Accept() to receive the file or Reject() to decline.
type PendingTransfer struct {
	client   *Client
	msg      *wh.IncomingMessage
	name     string
	size     int64
	callback ReceiveOfferCallback
	cancel   context.CancelFunc
}

// Name returns the filename being offered.
func (p *PendingTransfer) Name() string {
	return p.name
}

// Size returns the file size in bytes.
func (p *PendingTransfer) Size() int64 {
	return p.size
}

// Accept accepts the file transfer and begins receiving.
// Progress and completion are reported via the callback.
func (p *PendingTransfer) Accept() {
	go func() {
		defer p.cancel()

		path := filepath.Join(p.client.dataDir, p.name)

		// Check if file exists
		if _, err := os.Stat(path); err == nil {
			p.msg.Reject()
			p.callback.OnError("file already exists: " + p.name)
			return
		}

		f, err := os.Create(path)
		if err != nil {
			p.msg.Reject()
			p.callback.OnError(err.Error())
			return
		}

		// Copy with progress
		buf := make([]byte, 32*1024)
		var received int64
		for {
			n, readErr := p.msg.Read(buf)
			if n > 0 {
				_, writeErr := f.Write(buf[:n])
				if writeErr != nil {
					f.Close()
					os.Remove(path)
					p.callback.OnError(writeErr.Error())
					return
				}
				received += int64(n)
				p.callback.OnFileProgress(received, p.size)
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				f.Close()
				os.Remove(path)
				p.callback.OnError(readErr.Error())
				return
			}
		}
		f.Close()
		p.callback.OnFileComplete(path)
	}()
}

// Reject rejects the file transfer.
func (p *PendingTransfer) Reject() {
	p.msg.Reject()
	p.cancel()
}
