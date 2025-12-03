package wormhole

// SendCallback is implemented by Android/iOS code to receive send notifications.
// gomobile will generate the corresponding interface in Java/Kotlin/Swift.
type SendCallback interface {
	// OnCode is called when the wormhole code is generated.
	// Share this code with the receiver.
	OnCode(code string)

	// OnProgress is called periodically during file transfers.
	// sent is the number of bytes sent so far, total is the total file size.
	OnProgress(sent, total int64)

	// OnComplete is called when the transfer completes successfully.
	OnComplete()

	// OnError is called if the transfer fails.
	OnError(err string)
}

// ReceiveCallback is implemented by Android/iOS code to receive transfer notifications.
// gomobile will generate the corresponding interface in Java/Kotlin/Swift.
type ReceiveCallback interface {
	// OnText is called when a text message is received.
	OnText(text string)

	// OnFileStart is called when a file transfer begins.
	// name is the filename, size is the total size in bytes.
	OnFileStart(name string, size int64)

	// OnFileProgress is called periodically during file transfers.
	// received is bytes received so far, total is the total file size.
	OnFileProgress(received, total int64)

	// OnFileComplete is called when a file is fully received.
	// path is the full path to the saved file.
	OnFileComplete(path string)

	// OnError is called if the transfer fails.
	OnError(err string)
}

// ReceiveOfferCallback is like ReceiveCallback but allows accepting/rejecting file transfers.
// Use with ReceiveWithAccept() instead of Receive().
type ReceiveOfferCallback interface {
	// OnText is called when a text message is received.
	OnText(text string)

	// OnFileOffer is called when a file transfer is offered.
	// Call pending.Accept() to accept or pending.Reject() to reject.
	OnFileOffer(pending *PendingTransfer)

	// OnFileProgress is called periodically during file transfers.
	OnFileProgress(received, total int64)

	// OnFileComplete is called when a file is fully received.
	OnFileComplete(path string)

	// OnError is called if the transfer fails.
	OnError(err string)
}
