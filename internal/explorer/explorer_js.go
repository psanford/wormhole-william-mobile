// SPDX-License-Identifier: Unlicense OR MIT

package explorer

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"io"
	"strings"
	"syscall/js"
)

type explorer struct{}

func newExplorer(_ *app.Window) *explorer {
	return &explorer{}
}

func (e *Explorer) listenEvents(_ event.Event) {
	// NO-OP
}

func (e *Explorer) exportFile(name string) (io.WriteCloser, error) {
	return newFileWriter(name), nil
}

func (e *Explorer) importFile(extensions ...string) (io.ReadCloser, error) {
	// TODO: Replace with "File System Access API" when that becomes available on most browsers.
	// BUG: Not work on iOS/Safari.

	// It's not possible to know if the user closes the file-picker dialog, so an new channerl is needed.
	r := make(chan result)

	document := js.Global().Get("document")
	input := document.Call("createElement", "input")
	input.Call("addEventListener", "change", openCallback(r))
	input.Set("type", "file")
	input.Set("style", "display:none;")
	if len(extensions) > 0 {
		input.Set("accept", strings.Join(extensions, ","))
	}
	document.Get("body").Call("appendChild", input)
	input.Call("click")

	file := <-r
	if file.error != nil {
		return nil, file.error
	}
	return file.file.(io.ReadCloser), nil
}

type FileReader struct {
	buffer                   js.Value
	isClosed                 bool
	index                    uint32
	callback                 chan js.Value
	successFunc, failureFunc js.Func
}

func newFileReader(v js.Value) *FileReader {
	f := &FileReader{
		buffer:   v,
		callback: make(chan js.Value, 1),
	}
	f.successFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		f.callback <- args[0]
		return nil
	})
	f.failureFunc = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		f.callback <- js.Undefined()
		return nil
	})

	return f
}

func (f *FileReader) Read(b []byte) (n int, err error) {
	if f == nil || f.isClosed {
		return 0, io.ErrClosedPipe
	}

	go func() {
		fileSlice(f.index, f.index+uint32(len(b)), f.buffer, f.successFunc, f.failureFunc)
	}()

	buffer := <-f.callback
	n32 := fileRead(buffer, b)
	if n32 == 0 {
		return 0, io.EOF
	}
	f.index += n32

	return int(n32), err
}

func (f *FileReader) Close() error {
	if f == nil || f.isClosed {
		return io.ErrClosedPipe
	}

	f.failureFunc.Release()
	f.successFunc.Release()
	f.isClosed = true
	return nil
}

type FileWriter struct {
	buffer                   js.Value
	isClosed                 bool
	name                     string
	successFunc, failureFunc js.Func
}

func newFileWriter(name string) *FileWriter {
	return &FileWriter{
		name:   name,
		buffer: js.Global().Get("Uint8Array").New(),
	}
}

func (f *FileWriter) Write(b []byte) (n int, err error) {
	if f == nil || f.isClosed {
		return 0, io.ErrClosedPipe
	}
	if len(b) == 0 {
		return 0, nil
	}

	fileWrite(f.buffer, b)
	return len(b), err
}

func (f *FileWriter) Close() error {
	if f == nil || f.isClosed {
		return io.ErrClosedPipe
	}
	f.isClosed = true
	return f.saveFile()
}

func (f *FileWriter) saveFile() error {
	config := js.Global().Get("Object").New()
	config.Set("type", "octet/stream")

	blob := js.Global().Get("Blob").New(
		js.Global().Get("Array").New().Call("concat", f.buffer),
		config,
	)

	document := js.Global().Get("document")
	anchor := document.Call("createElement", "a")
	anchor.Set("download", f.name)
	anchor.Set("href", js.Global().Get("URL").Call("createObjectURL", blob))
	document.Get("body").Call("appendChild", anchor)
	anchor.Call("click")

	return nil
}

// fileRead and fileWrite calls the JS function directly (without syscall/js to avoid double copying).
// The function is defined into explorer_js.s, which calls explorer_js.js.
func fileRead(value js.Value, b []byte) uint32
func fileWrite(value js.Value, b []byte)
func fileSlice(start, end uint32, value js.Value, success, failure js.Func)

func openCallback(r chan result) js.Func {
	// There's no way to detect when the dialog is closed, so we can't re-use the callback.
	return js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		files := args[0].Get("target").Get("files")
		if files.Length() <= 0 {
			r <- result{error: ErrUserDecline}
			return nil
		}
		r <- result{file: newFileReader(files.Index(0))}
		return nil
	})
}

var (
	_ io.ReadCloser  = (*FileReader)(nil)
	_ io.WriteCloser = (*FileWriter)(nil)
)
