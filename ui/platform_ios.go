package ui

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/x/explorer"
	"github.com/psanford/wormhole-william-mobile/internal/picker"
	"log"
)

func newPlatformHandler(fileExplorer *explorer.Explorer) platformHandler {
	return &iosPlatform{
		exp: fileExplorer,
	}
}

type iosPlatform struct {
	exp *explorer.Explorer
}

func (d *iosPlatform) handleEvent(e event.Event) {
}

func (d *iosPlatform) saveFileToDocuments(src *os.File, dataDir, name string) (string, error) {
	w, err := d.exp.CreateFile(name)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(w, src)
	if err != nil {
		return "", fmt.Errorf("save final file err: %w", err)
	}

	return "", w.Close()
}

func (d *iosPlatform) dstFilePathClear(dataDir, name string) error {
	return nil
}

func (d *iosPlatform) pickFile() <-chan picker.PickResult {
	ch := make(chan picker.PickResult, 1)
	go func() {
		ff, err := d.exp.ChooseFile(fileExtensions...)
		if err != nil {
			ch <- picker.PickResult{
				Err: fmt.Errorf("pick file failed: %s", err),
			}
			return
		}
		defer ff.Close()
		f, ok := ff.(fileResult)
		if !ok {
			ch <- picker.PickResult{
				Err: fmt.Errorf("couldn't cast picked file to *os.File, %T", ff),
			}
			return
		}

		u, err := url.Parse(f.URL())
		if err != nil {
			ch <- picker.PickResult{
				Err: fmt.Errorf("couldn't parse file url: %s err=%s", f.URL(), err),
			}
			return
		}

		dataDir, err := app.DataDir()
		if err != nil {
			log.Printf("wormhole: DataDir err: %s", err)
		}

		err = os.MkdirAll(dataDir, 0700)
		if err != nil {
			log.Printf("wormhole: MkdirAll err: %s", err)
		}

		outFile, err := os.CreateTemp(dataDir, "")
		if err != nil {
			ch <- picker.PickResult{
				Err: fmt.Errorf("couldn't create tmp file: %s", err),
			}
			return
		}

		_, err = io.Copy(outFile, f)
		if err != nil {
			ch <- picker.PickResult{
				Err: fmt.Errorf("couldn't populate tmp file: %s", err),
			}
			return
		}

		name := filepath.Base(u.Path)

		path := outFile.Name()
		outFile.Close()

		ch <- picker.PickResult{
			Name: name,
			Path: path,
		}
	}()

	return ch
}

func (d *iosPlatform) notifyDownloadManager(name, path, contentType string, size int64) {
}

func (d *iosPlatform) supportedFeatures() platformFeature {
	return 0
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

type fileResult interface {
	io.ReadCloser
	URL() string
}

var fileExtensions = []string{
	".aiff",
	".aliasFile",
	".appleArchive",
	".appleProtectedMPEG4Audio",
	".appleProtectedMPEG4Video",
	".appleScript",
	".application",
	".applicationBundle",
	".applicationExtension",
	".arReferenceObject",
	".archive",
	".assemblyLanguageSource",
	".audio",
	".audiovisualContent",
	".avi",
	".binaryPropertyList",
	".bmp",
	".bookmark",
	".bundle",
	".bz2",
	".cHeader",
	".cPlusPlusHeader",
	".cPlusPlusSource",
	".cSource",
	".calendarEvent",
	".commaSeparatedText",
	".compositeContent",
	".contact",
	".content",
	".data",
	".database",
	".delimitedText",
	".directory",
	".diskImage",
	".emailMessage",
	".epub",
	".exe",
	".executable",
	".fileURL",
	".flatRTFD",
	".folder",
	".font",
	".framework",
	".gif",
	".gzip",
	".heic",
	".heif",
	".html",
	".icns",
	".ico",
	".image",
	".internetLocation",
	".internetShortcut",
	".item",
	".javaScript",
	".jpeg",
	".json",
	".livePhoto",
	".log",
	".m3uPlaylist",
	".makefile",
	".message",
	".midi",
	".mountPoint",
	".movie",
	".mp3",
	".mpeg",
	".mpeg2TransportStream",
	".mpeg2Video",
	".mpeg4Audio",
	".mpeg4Movie",
	".objectiveCPlusPlusSource",
	".objectiveCSource",
	".osaScript",
	".osaScriptBundle",
	".package",
	".pdf",
	".perlScript",
	".phpScript",
	".pkcs12",
	".plainText",
	".playlist",
	".pluginBundle",
	".png",
	".presentation",
	".propertyList",
	".pythonScript",
	".quickLookGenerator",
	".quickTimeMovie",
	".rawImage",
	".realityFile",
	".resolvable",
	".rtf",
	".rtfd",
	".rubyScript",
	".sceneKitScene",
	".script",
	".shellScript",
	".sourceCode",
	".spotlightImporter",
	".spreadsheet",
	".svg",
	".swiftSource",
	".symbolicLink",
	".systemPreferencesPane",
	".tabSeparatedText",
	".text",
	".threeDContent",
	".tiff",
	".toDoItem",
	".unixExecutable",
	".url",
	".urlBookmarkData",
	".usd",
	".usdz",
	".utf16ExternalPlainText",
	".utf16PlainText",
	".utf8PlainText",
	".utf8TabSeparatedText",
	".vCard",
	".video",
	".volume",
	".wav",
	".webArchive",
	".webP",
	".x509Certificate",
	".xml",
	".xmlPropertyList",
	".xpcService",
	".yaml",
	".zip",
	".item",
	".content",
	".compositeContent",
	".diskImage",
	".data",
	".directory",
	".resolvable",
	".symbolicLink",
	".executable",
	".mountPoint",
	".aliasFile",
	".urlBookmarkData",
	".url",
	".fileURL",
	".text",
	".plainText",
	".utf8PlainText",
	".utf16ExternalPlainText",
	".utf16PlainText",
	".delimitedText",
	".commaSeparatedText",
	".tabSeparatedText",
	".utf8TabSeparatedText",
	".rtf",
	".html",
	".xml",
	".yaml",
	".sourceCode",
	".assemblyLanguageSource",
	".cSource",
	".objectiveCSource",
	".swiftSource",
	".cPlusPlusSource",
	".objectiveCPlusPlusSource",
	".cHeader",
	".cPlusPlusHeader",
	".script",
	".appleScript",
	".osaScript",
	".osaScriptBundle",
	".javaScript",
	".shellScript",
	".perlScript",
	".pythonScript",
	".rubyScript",
	".phpScript",
	".makefile",
	".json",
	".propertyList",
	".xmlPropertyList",
	".binaryPropertyList",
	".pdf",
	".rtfd",
	".flatRTFD",
	".webArchive",
	".image",
	".jpeg",
	".tiff",
	".gif",
	".png",
	".icns",
	".bmp",
	".ico",
	".rawImage",
	".svg",
	".livePhoto",
	".heif",
	".heic",
	".webP",
	".threeDContent",
	".usd",
	".usdz",
	".realityFile",
	".sceneKitScene",
	".arReferenceObject",
	".audiovisualContent",
	".movie",
	".video",
	".audio",
	".quickTimeMovie",
	".mpeg",
	".mpeg2Video",
	".mpeg2TransportStream",
	".mp3",
	".mpeg4Movie",
	".mpeg4Audio",
	".appleProtectedMPEG4Audio",
	".appleProtectedMPEG4Video",
	".avi",
	".aiff",
	".wav",
	".midi",
	".playlist",
	".m3uPlaylist",
	".folder",
	".volume",
	".package",
	".bundle",
	".pluginBundle",
	".spotlightImporter",
	".quickLookGenerator",
	".xpcService",
	".framework",
	".application",
	".applicationBundle",
	".applicationExtension",
	".unixExecutable",
	".exe",
	".systemPreferencesPane",
	".archive",
	".gzip",
	".bz2",
	".zip",
	".appleArchive",
	".spreadsheet",
	".presentation",
	".database",
	".message",
	".contact",
	".vCard",
	".toDoItem",
	".calendarEvent",
	".emailMessage",
	".internetLocation",
	".internetShortcut",
	".font",
	".bookmark",
	".pkcs12",
	".x509Certificate",
	".epub",
	".log",
}
