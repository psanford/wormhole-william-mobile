// Package wormhole provides a gomobile-compatible API for the wormhole-william library.
// This package is designed to be bound to Android/iOS using gomobile bind.
package wormhole

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	wh "github.com/psanford/wormhole-william/wormhole"
)

// Client wraps wormhole-william for mobile use.
// Create a new client with NewClient() and configure it with
// SetRendezvousURL() and SetCodeLength() before use.
type Client struct {
	mu      sync.Mutex
	client  wh.Client
	cancel  context.CancelFunc
	dataDir string
}

// NewClient creates a new wormhole client.
// dataDir is the directory where received files will be saved.
func NewClient(dataDir string) *Client {
	return &Client{
		dataDir: dataDir,
	}
}

// SetRendezvousURL configures the relay server URL.
// If not set, the default wormhole relay is used.
func (c *Client) SetRendezvousURL(url string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.client.RendezvousURL = url
}

// GetRendezvousURL returns the current relay server URL.
func (c *Client) GetRendezvousURL() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.RendezvousURL
}

// SetCodeLength configures the passphrase word count.
// Default is 2 words.
func (c *Client) SetCodeLength(length int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.client.PassPhraseComponentLength = length
}

// GetCodeLength returns the current passphrase word count.
func (c *Client) GetCodeLength() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.PassPhraseComponentLength
}

// Cancel cancels any ongoing transfer.
// Safe to call even if no transfer is in progress.
func (c *Client) Cancel() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}
}

// SendText sends a text message through the wormhole.
// The callback receives the wormhole code, progress updates, and completion/error notifications.
// This method returns immediately; the transfer happens asynchronously.
func (c *Client) SendText(msg string, callback SendCallback) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		c.mu.Lock()
		c.cancel = cancel
		c.mu.Unlock()
		defer cancel()

		code, status, err := c.client.SendText(ctx, msg)
		if err != nil {
			callback.OnError(err.Error())
			return
		}
		callback.OnCode(code)

		s := <-status
		if s.Error != nil {
			callback.OnError(s.Error.Error())
		} else {
			callback.OnComplete()
		}
	}()
}

// SendFile sends a file through the wormhole.
// path is the full path to the file, name is the filename to present to the receiver.
// The callback receives the wormhole code, progress updates, and completion/error notifications.
// This method returns immediately; the transfer happens asynchronously.
func (c *Client) SendFile(path, name string, callback SendCallback) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		c.mu.Lock()
		c.cancel = cancel
		c.mu.Unlock()
		defer cancel()

		f, err := os.Open(path)
		if err != nil {
			callback.OnError(err.Error())
			return
		}
		defer f.Close()

		progress := func(sent, total int64) {
			callback.OnProgress(sent, total)
		}

		code, status, err := c.client.SendFile(ctx, name, f, wh.WithProgress(progress))
		if err != nil {
			callback.OnError(err.Error())
			return
		}
		callback.OnCode(code)

		s := <-status
		if s.Error != nil {
			callback.OnError(s.Error.Error())
		} else {
			callback.OnComplete()
		}
	}()
}

// Receive receives a wormhole transfer using the given code.
// The callback receives notifications about the transfer type, progress, and completion.
// For text transfers, OnText is called with the message content.
// For file transfers, OnFileStart, OnFileProgress, and OnFileComplete are called.
// This method returns immediately; the transfer happens asynchronously.
func (c *Client) Receive(code string, callback ReceiveCallback) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		c.mu.Lock()
		c.cancel = cancel
		c.mu.Unlock()
		defer cancel()

		msg, err := c.client.Receive(ctx, code)
		if err != nil {
			callback.OnError(err.Error())
			return
		}

		switch msg.Type {
		case wh.TransferText:
			data, err := io.ReadAll(msg)
			if err != nil {
				callback.OnError(err.Error())
				return
			}
			callback.OnText(string(data))

		case wh.TransferFile, wh.TransferDirectory:
			name := msg.Name
			if msg.Type == wh.TransferDirectory {
				name += ".zip"
			}

			callback.OnFileStart(name, msg.TransferBytes64)

			// Save to dataDir
			path := filepath.Join(c.dataDir, name)

			// Check if file exists
			if _, err := os.Stat(path); err == nil {
				msg.Reject()
				callback.OnError("file already exists: " + name)
				return
			}

			f, err := os.Create(path)
			if err != nil {
				msg.Reject()
				callback.OnError(err.Error())
				return
			}

			// Copy with progress
			buf := make([]byte, 32*1024)
			var received int64
			for {
				n, readErr := msg.Read(buf)
				if n > 0 {
					_, writeErr := f.Write(buf[:n])
					if writeErr != nil {
						f.Close()
						os.Remove(path)
						callback.OnError(writeErr.Error())
						return
					}
					received += int64(n)
					callback.OnFileProgress(received, msg.TransferBytes64)
				}
				if readErr == io.EOF {
					break
				}
				if readErr != nil {
					f.Close()
					os.Remove(path)
					callback.OnError(readErr.Error())
					return
				}
			}
			f.Close()
			callback.OnFileComplete(path)
		}
	}()
}

// ReceiveWithAccept is like Receive but requires explicit acceptance of file transfers.
// After OnFileOffer is called, the transfer will wait until Accept() or Reject() is called
// on the returned PendingTransfer.
func (c *Client) ReceiveWithAccept(code string, callback ReceiveOfferCallback) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		c.mu.Lock()
		c.cancel = cancel
		c.mu.Unlock()

		msg, err := c.client.Receive(ctx, code)
		if err != nil {
			cancel()
			callback.OnError(err.Error())
			return
		}

		switch msg.Type {
		case wh.TransferText:
			data, err := io.ReadAll(msg)
			cancel()
			if err != nil {
				callback.OnError(err.Error())
				return
			}
			callback.OnText(string(data))

		case wh.TransferFile, wh.TransferDirectory:
			name := msg.Name
			if msg.Type == wh.TransferDirectory {
				name += ".zip"
			}

			pending := &PendingTransfer{
				client:   c,
				msg:      msg,
				name:     name,
				size:     msg.TransferBytes64,
				callback: callback,
				cancel:   cancel,
			}

			callback.OnFileOffer(pending)
		}
	}()
}
