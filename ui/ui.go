package ui

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/psanford/wormhole-william-mobile/config"
	"github.com/psanford/wormhole-william-mobile/internal/picker"
	"github.com/psanford/wormhole-william-mobile/ui/plog"
	"github.com/psanford/wormhole-william/wormhole"
)

type UI struct {
	wormholeClient wormhole.Client
	conf           *config.Config
}

func New() *UI {
	return &UI{}
}

func (ui *UI) Run() error {
	w := new(app.Window)
	w.Option(app.Size(unit.Dp(800), unit.Dp(700)))

	if err := ui.loop(w); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (ui *UI) loop(w *app.Window) error {
	th := material.NewTheme()

	dataDir, _ := app.DataDir()
	conf := config.Load(dataDir)
	ui.conf = conf

	ui.wormholeClient.RendezvousURL = conf.RendezvousURL
	rendezvousEditor.SetText(conf.RendezvousURL)
	ui.wormholeClient.PassPhraseComponentLength = conf.CodeLen
	if conf.CodeLen > 0 {
		codeLenEditor.SetText(strconv.Itoa(conf.CodeLen))
	}

	recvCodeEditor.SingleLine = true

	var (
		pickResult   <-chan picker.PickResult
		qrCodeResult <-chan string
		permResultCh <-chan picker.PermResult

		windowEventCh   = make(chan event.Event)
		windowAckCh     = make(chan struct{})
		ctx             = context.Background()
		platformHandler = newPlatformHandler()
	)

	go func() {
		for {
			ev := w.Event()
			windowEventCh <- ev
			<-windowAckCh
			if _, ok := ev.(app.DestroyEvent); ok {
				return
			}
		}
	}()

	var ops op.Ops
	for {
		select {
		case code := <-qrCodeResult:
			if code != "" {

				parsed, err := parseCodeURI(code)
				if err != nil {
					statusMsg = fmt.Sprintf("invalid code: %s", err)
					continue
				}

				recvCodeEditor.SetText(parsed.code)
				w.Invalidate()
			}
			qrCodeResult = nil
		case shareEvt := <-platformHandler.sharedEventCh():
			if shareEvt.Type == picker.Text {
				textMsgEditor.SetText(shareEvt.Text)
				for idx, tab := range tabs.tabs {
					if tab.Title == "Send Text" {
						tabs.selected = idx
						break
					}
				}
			} else if shareEvt.Type == picker.File {
				for idx, tab := range tabs.tabs {
					if tab.Title == "Send File" {
						tabs.selected = idx
						break
					}
				}

				ui.sendFile(ctx, w, shareEvt.Path, shareEvt.Name)
			}

		case result := <-pickResult:
			pickResult = nil
			plog.Printf("pick result: path=%s name=%s err=%s", result.Path, result.Name, result.Err)
			if result.Err != nil {
				statusMsg = fmt.Sprintf("Pick file err: %s", result.Err)
				w.Invalidate()
				continue
			}

			ui.sendFile(ctx, w, result.Path, result.Name)

		case permResult := <-permResultCh:
			permResultCh = nil
			if permResult.Err != nil {
				statusMsg = fmt.Sprintf("Write file permission not granted, err: %s", permResult.Err)
				w.Invalidate()
				continue
			}
			select {
			case hasPermissionChan <- permResult.Authorized:
			case <-time.After(5 * time.Second):
				plog.Printf("write hasPermissionChan timed out")
			}
		case e := <-windowEventCh:
			switch e := e.(type) {
			case app.DestroyEvent:
				windowAckCh <- struct{}{}
				return e.Err
			case app.FrameEvent:
				gtx := app.NewContext(&ops, e)

				var (
					sendFileClicked bool
					sendTextClicked bool
					recvClicked     bool
					scanQRClicked   bool
					acceptClicked   bool
					cancelClicked   bool
				)
				type btnState struct {
					name    string
					btn     *widget.Clickable
					clicked *bool
				}
				btns := []btnState{
					{"sendFileBtn", sendFileBtn, &sendFileClicked},
					{"sendTextBtn", sendTextBtn, &sendTextClicked},
					{"recvMsgBtn", recvMsgBtn, &recvClicked},
					{"scanQRBtn", scanQRBtn, &scanQRClicked},
					{"acceptBtn", acceptBtn, &acceptClicked},
					{"cancelBtn", cancelBtn, &cancelClicked},
				}
				for _, btn := range btns {
					for btn.btn.Clicked(gtx) {
						*btn.clicked = true
					}
				}

				if rendezvousEditor.Text() != conf.RendezvousURL {
					conf.RendezvousURL = rendezvousEditor.Text()
					conf.Save()
					ui.wormholeClient.RendezvousURL = conf.RendezvousURL
				}
				codeLenStr := strconv.Itoa(conf.CodeLen)
				if codeLenEditor.Text() != codeLenStr {
					n, _ := strconv.Atoi(strings.TrimSpace(codeLenEditor.Text()))
					if n > 0 {
						conf.CodeLen = n
						conf.Save()
						ui.wormholeClient.PassPhraseComponentLength = conf.CodeLen
					}
				}

				if scanQRClicked {
					qrCodeResult = platformHandler.scanQRCode()
				}

				if sendFileClicked {
					pickResult = platformHandler.pickFile()
				}

				if acceptClicked {
					select {
					case confirmChan <- struct{}{}:
					default:
					}
				}

				if cancelClicked {
					select {
					case cancelChan <- struct{}{}:
					default:
					}
				}

				if sendTextClicked {
					func() {
						gtx.Execute(key.FocusCmd{}) // blur textfield

						msg := textMsgEditor.Text()
						if msg == "" {
							return
						}

						sendCtx, cancel := context.WithCancel(ctx)

						go func() {
							select {
							case <-cancelChan:
								cancel()
								statusMsg = "Transfer mid-stream aborted"
								textCodeTxt.SetText("")
								transferInProgress = false
								w.Invalidate()
							case <-ctx.Done():
							}
						}()

						code, status, err := ui.wormholeClient.SendText(sendCtx, msg)
						if err != nil {
							statusMsg = fmt.Sprintf("Send err: %s", err)
							plog.Printf("Send err: %s", err)
							return
						}

						statusMsg = "Waiting for receiver..."
						textCodeTxt.SetText(code)

						go func() {
							transferInProgress = true
							defer func() {
								transferInProgress = false
								recvCodeEditor.SetText("")
							}()

							s := <-status
							if s.Error != nil {
								statusMsg = fmt.Sprintf("Send error: %s", s.Error)
								plog.Printf("Send error: %s", s.Error)
							} else if s.OK {
								statusMsg = "OK!"
							}
							textCodeTxt.SetText("")
							w.Invalidate()
						}()
					}()
				}

				if recvClicked {
					log.Printf("recv clicked")
					func() {
						statusMsg = "Start recv"
						w.Invalidate()

						gtx.Execute(key.FocusCmd{}) // blur textfield
						code := recvCodeEditor.Text()
						if code == "" {
							return
						}

						errf := func(msg string, args ...interface{}) {
							statusMsg = fmt.Sprintf(msg, args...)
							plog.Printf(msg, args...)
						}

						go func() {
							transferInProgress = true
							defer func() {
								transferInProgress = false
								recvCodeEditor.SetText("")
							}()

							defer w.Invalidate()

							recvCtx, cancel := context.WithCancel(ctx)
							defer cancel()

							msg, err := ui.wormholeClient.Receive(recvCtx, code)
							if err != nil {
								errf("Recv msg err: %s", err)
								return
							}
							switch msg.Type {
							case wormhole.TransferText:
								msgBody, err := io.ReadAll(msg)
								if err != nil {
									errf("Recv msg err: %s", err)
									return
								}

								recvTxtMsg.SetText(string(msgBody))
							case wormhole.TransferFile, wormhole.TransferDirectory:
								dataDir, err := app.DataDir()
								if err != nil {
									msg.Reject()

									errf("Recv error, cannot get datadir: %s", err)
									return
								}

								name := msg.Name
								if msg.Type == wormhole.TransferDirectory {
									name += ".zip"
								}

								path := filepath.Join(dataDir, name)
								if _, err := os.Stat(name); err == nil {
									msg.Reject()
									errf("Error refusing to overwrite existing '%s'", name)
									return
								} else if !os.IsNotExist(err) {
									msg.Reject()
									errf("Error stat'ing existing '%s'\n", name)
									return
								}

								confirmInProgress = true
								defer func() {
									confirmInProgress = false
								}()

								statusMsg = fmt.Sprintf("Receiving file (%s) into %s\nAccept or Cancel?", formatBytes(msg.TransferBytes64), msg.Name)

								w.Invalidate()

								select {
								case <-cancelChan:
									msg.Reject()
									statusMsg = "Transfer rejected"
									return
								case <-confirmChan:
								}
								confirmInProgress = false

								permResultCh = platformHandler.requestWriteFilePerm()
								select {
								case <-cancelChan:
									msg.Reject()
									statusMsg = "Transfer rejected"
									return
								case hasPermission := <-hasPermissionChan:
									if !hasPermission {
										msg.Reject()
										statusMsg = "Transfer rejected"
										return
									}
								}

								f, err := os.CreateTemp(dataDir, fmt.Sprintf("%s.tmp", name))
								if err != nil {
									msg.Reject()
									errf("Create tmp file failed: %s", err)
									return
								}

								r := newCountReader(msg)

								go func() {
									select {
									case <-cancelChan:
										cancel()
										statusMsg = "Transfer mid-stream aborted"
									case <-ctx.Done():
									}
								}()

								stopRecvUpdater := make(chan struct{})

								go func() {
									statusMsg = fmt.Sprintf("receiving %d/%s", 0, formatBytes(msg.TransferBytes64))
									for count := range r.countUpdate {
										select {
										case <-ctx.Done():
											return
										case <-stopRecvUpdater:
											return
										default:
										}
										statusMsg = fmt.Sprintf("receiving %s/%s", formatBytes(count), formatBytes(msg.TransferBytes64))
										w.Invalidate()
										time.Sleep(500 * time.Millisecond)
									}
								}()

								_, err = io.Copy(f, r)
								r.Close()
								if err != nil {
									os.Remove(f.Name())
									close(stopRecvUpdater)
									errf("Receive file error: %s", err)
									return
								}

								tmpName := f.Name()
								f.Seek(0, io.SeekStart)
								header := make([]byte, 512)
								io.ReadFull(f, header)
								f.Close()

								err = os.Rename(tmpName, path)
								if err != nil {
									close(stopRecvUpdater)
									errf("Rename file err: %s", err)
									return
								}

								var contentType string
								if msg.Type == wormhole.TransferDirectory {
									contentType = "application/zip"
								} else {
									contentType = http.DetectContentType(header)
								}

								plog.Printf("Call NotifyDownloadManager")
								platformHandler.notifyDownloadManager(name, path, contentType, msg.TransferBytes64)

								close(stopRecvUpdater)
								statusMsg = "Receive complete"
								w.Invalidate()
							}
						}()
					}()
				}

				for moreEvents := true; moreEvents; {
					var e widget.EditorEvent
					e, moreEvents = recvCodeEditor.Update(gtx)
					if _, ok := e.(widget.ChangeEvent); ok {
						orig := recvCodeEditor.Text()
						new := strings.ReplaceAll(orig, " ", "-")
						new = strings.ReplaceAll(new, "\n", "")

						if new != orig {
							_, col := recvCodeEditor.CaretPos()
							recvCodeEditor.SetText(new)
							recvCodeEditor.MoveCaret(col, col)
						}
					}
				}

				drawTabs(gtx, th)
				e.Frame(gtx.Ops)
				windowAckCh <- struct{}{}

			default:
				platformHandler.handleEvent(e)
				windowAckCh <- struct{}{}
			}
		}
	}
}

func (ui *UI) sendFile(ctx context.Context, w *app.Window, path, filename string) {
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			plog.Printf("open file err path=%s err=%s", path, err)
			statusMsg = fmt.Sprintf("open file err: %s", err)
			w.Invalidate()
			return
		}
		defer os.Remove(path)

		progress := func(sentBytes, totalBytes int64) {
			statusMsg = fmt.Sprintf("Send progress %s/%s", formatBytes(sentBytes), formatBytes(totalBytes))
			w.Invalidate()
		}

		sendCtx, cancel := context.WithCancel(ctx)

		go func() {
			select {
			case <-cancelChan:
				cancel()
				statusMsg = "Transfer mid-stream aborted"
				sendFileCodeTxt.SetText("")
				transferInProgress = false
			case <-ctx.Done():
			}
		}()

		code, status, err := ui.wormholeClient.SendFile(sendCtx, filename, f, wormhole.WithProgress(progress))
		if err != nil {
			plog.Printf("wormhole send error err=%s", err)
			statusMsg = fmt.Sprintf("wormhole send err: %s", err)
			w.Invalidate()
			return
		}

		sendFileCodeTxt.SetText(code)
		statusMsg = "Waiting for receiver..."

		go func() {
			transferInProgress = true
			defer func() {
				cancel()
				transferInProgress = false
				sendFileCodeTxt.SetText("")
			}()

			s := <-status
			if s.Error != nil {

				statusMsg = fmt.Sprintf("wormhole send err: %s", s.Error)
			} else {
				statusMsg = "Send Complete!"
				sendFileCodeTxt.SetText("")
			}
			w.Invalidate()
		}()
	}
}

var (
	textMsgEditor = new(RichEditor)
	textCodeTxt   = new(Copyable)
	sendTextBtn   = new(widget.Clickable)

	rendezvousEditor = &widget.Editor{
		SingleLine: true,
	}

	codeLenEditor = &widget.Editor{
		SingleLine: true,
	}

	acceptBtn = new(widget.Clickable)
	cancelBtn = new(widget.Clickable)

	cancelChan        = make(chan struct{})
	confirmChan       = make(chan struct{})
	hasPermissionChan = make(chan bool)

	statusMsg          string
	transferInProgress bool
	confirmInProgress  bool

	recvCodeEditor = new(RichEditor)
	recvMsgBtn     = new(widget.Clickable)
	scanQRBtn      = new(widget.Clickable)
	recvTxtMsg     = new(Copyable)
	itemList       = &layout.List{
		Axis: layout.Vertical,
	}

	sendFileCodeTxt = new(Copyable)
	sendFileBtn     = new(widget.Clickable)

	topLabel = "Wormhole William"

	tabs = Tabs{
		tabs: []Tab{
			recvTab,
			sendTextTab,
			sendFileTab,
			settingsTab,
		},
	}
)

var slider Slider

type Tabs struct {
	list     layout.List
	tabs     []Tab
	selected int
}

type Tab struct {
	btn   widget.Clickable
	Title string
	draw  func(gtx layout.Context, th *material.Theme) layout.Dimensions
}

type (
	C = layout.Context
	D = layout.Dimensions
)

func drawTabs(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.UniformInset(unit.Dp(12)).Layout(gtx,
				material.H4(th, topLabel).Layout,
			)
		}),
		layout.Rigid(func(gtx C) D {
			return tabs.list.Layout(gtx, len(tabs.tabs), func(gtx C, tabIdx int) D {
				t := &tabs.tabs[tabIdx]
				if t.btn.Clicked(gtx) {
					if tabs.selected < tabIdx {
						slider.PushLeft()
					} else if tabs.selected > tabIdx {
						slider.PushRight()
					}
					tabs.selected = tabIdx
				}
				var tabWidth int
				return layout.Stack{Alignment: layout.S}.Layout(gtx,
					layout.Stacked(func(gtx C) D {
						dims := material.Clickable(gtx, &t.btn, func(gtx C) D {
							return layout.UniformInset(unit.Dp(12)).Layout(gtx,
								material.H6(th, t.Title).Layout,
							)
						})
						tabWidth = dims.Size.X
						return dims
					}),
					layout.Stacked(func(gtx C) D {
						if tabs.selected != tabIdx {
							return layout.Dimensions{}
						}
						tabHeight := gtx.Dp(4)
						tabRect := image.Rect(0, 0, tabWidth, tabHeight)
						paint.FillShape(gtx.Ops, th.Palette.ContrastBg, clip.Rect(tabRect).Op())
						return layout.Dimensions{
							Size: image.Point{X: tabWidth, Y: tabHeight},
						}
					}),
				)
			})
		}),
		layout.Flexed(1, func(gtx C) D {
			return slider.Layout(gtx, func(gtx C) D {
				selectedTab := tabs.tabs[tabs.selected]
				return selectedTab.draw(gtx, th)
			})
		}),
	)
}

func textField(gtx layout.Context, th *material.Theme, label, hint string, editor *RichEditor) func(layout.Context) layout.Dimensions {
	return func(gtx layout.Context) layout.Dimensions {
		flex := layout.Flex{
			Axis: layout.Vertical,
		}
		return flex.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.H5(th, label).Layout(gtx)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				e := PasteEditor(th, editor, hint)

				border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Dp(2)}
				return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
				})
			}),
		)
	}
}

var sendTextTab = Tab{
	Title: "Send Text",
	draw: func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		widgets := []layout.Widget{
			textField(gtx, th, "Text", "Message", textMsgEditor),

			func(gtx C) D {
				if transferInProgress {
					gtx = gtx.Disabled()
				}
				return material.Button(th, sendTextBtn, "Send").Layout(gtx)
			},
			func(gtx C) D {
				if textCodeTxt.Text() != "" {
					return material.H6(th, "Code:").Layout(gtx)
				}
				return D{}
			},
			func(gtx C) D {
				if textCodeTxt.Text() != "" {
					gtx.Constraints.Max.Y = gtx.Dp(400)
					return CopyField(th, textCodeTxt).Layout(gtx)
				}
				return D{}
			},
			func(gtx C) D {
				if transferInProgress || confirmInProgress {
					return material.Button(th, cancelBtn, "Cancel").Layout(gtx)
				}
				return D{}
			},
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Flexed(0.9, func(gtx layout.Context) layout.Dimensions {
				return itemList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
				})
			}),
			layout.Rigid(drawStatus(th)),
		)
	},
}

var recvTab = Tab{
	Title: "Recv",
	draw: func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		widgets := []layout.Widget{
			textField(gtx, th, "Code", "Code", recvCodeEditor),

			func(gtx C) D {
				if transferInProgress || recvCodeEditor.Text() != "" {
					gtx = gtx.Disabled()
				}
				return material.Button(th, scanQRBtn, "Scan QR Code").Layout(gtx)
			},
			func(gtx C) D {
				if transferInProgress {
					gtx = gtx.Disabled()
				}
				return material.Button(th, recvMsgBtn, "Receive").Layout(gtx)
			},
			func(gtx C) D {
				if confirmInProgress {
					return material.Button(th, acceptBtn, "Accept").Layout(gtx)
				}
				return D{}
			},
			func(gtx C) D {
				if transferInProgress || confirmInProgress {
					return material.Button(th, cancelBtn, "Cancel").Layout(gtx)
				}
				return D{}
			},
			func(gtx C) D {
				gtx.Constraints.Max.Y = gtx.Dp(200)
				return CopyField(th, recvTxtMsg).Layout(gtx)
			},
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Flexed(0.9, func(gtx layout.Context) layout.Dimensions {
				return itemList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
				})
			}),
			layout.Rigid(drawStatus(th)),
		)
	},
}

func drawStatus(th *material.Theme) func(gtx layout.Context) layout.Dimensions {
	return func(gtx layout.Context) layout.Dimensions {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx layout.Context) layout.Dimensions {
				if statusMsg == "" {
					return layout.Dimensions{}
				}

				size := image.Point{
					X: gtx.Constraints.Max.X,
					Y: gtx.Constraints.Min.Y,
				}

				return ColorBox(gtx, size, lightYellow)
			}),
			layout.Stacked(func(gtx layout.Context) layout.Dimensions {
				// gtx.Constraints.Min.Y = gtx.Px(unit.Dp(20))
				return layout.UniformInset(unit.Dp(16)).Layout(gtx,
					material.Body1(th, statusMsg).Layout,
				)
			}),
		)
	}
}

var sendFileTab = Tab{
	Title: "Send File",
	draw: func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		widgets := []layout.Widget{

			func(gtx C) D {
				if transferInProgress {
					gtx = gtx.Disabled()
				}
				return material.Button(th, sendFileBtn, "Choose File").Layout(gtx)
			},
			func(gtx C) D {
				gtx.Constraints.Max.Y = gtx.Dp(400)
				return CopyField(th, sendFileCodeTxt).Layout(gtx)
			},
			func(gtx C) D {
				if transferInProgress || confirmInProgress {
					return material.Button(th, cancelBtn, "Cancel").Layout(gtx)
				}
				return D{}
			},
		}

		return layout.Flex{
			Axis: layout.Vertical,
		}.Layout(gtx,
			layout.Flexed(0.9, func(gtx layout.Context) layout.Dimensions {
				return itemList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
				})
			}),
			layout.Rigid(drawStatus(th)),
		)
	},
}

var settingsTab = Tab{
	Title: "Settings",
	draw: func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		textField := func(label, hint string, editor *widget.Editor) func(layout.Context) layout.Dimensions {
			return func(gtx layout.Context) layout.Dimensions {
				flex := layout.Flex{
					Axis: layout.Vertical,
				}
				return flex.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.H5(th, label).Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						e := material.Editor(th, editor, hint)
						border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Dp(2)}
						return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
						})
					}),
				)
			}
		}

		widgets := []layout.Widget{
			textField("Rendezvous URL", wormhole.DefaultRendezvousURL, rendezvousEditor),
			textField("Code Length", "2", codeLenEditor),
		}

		return itemList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
			return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
		})
	},
}

func formatBytes(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

type platformHandler interface {
	handleEvent(event.Event)
	pickFile() <-chan picker.PickResult
	sharedEventCh() chan picker.SharedEvent
	notifyDownloadManager(name, path, contentType string, size int64)
	scanQRCode() <-chan string
	requestWriteFilePerm() <-chan picker.PermResult
}

// Test colors.
var (
	background  = color.NRGBA{R: 0xC0, G: 0xC0, B: 0xC0, A: 0xFF}
	red         = color.NRGBA{R: 0xC0, G: 0x40, B: 0x40, A: 0xFF}
	green       = color.NRGBA{R: 0x40, G: 0xC0, B: 0x40, A: 0xFF}
	blue        = color.NRGBA{R: 0x40, G: 0x40, B: 0xC0, A: 0xFF}
	lightYellow = color.NRGBA{R: 0xFF, G: 0xFE, B: 0xE3, A: 0xFF}
)

// ColorBox creates a widget with the specified dimensions and color.
func ColorBox(gtx layout.Context, size image.Point, color color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: color}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

type parsedCode struct {
	relay string
	code  string
}

func parseCodeURI(codeStr string) (*parsedCode, error) {
	if !strings.HasPrefix(codeStr, "wormhole:") {
		return nil, errors.New("not a wormhole code")
	}

	u := strings.TrimPrefix(codeStr, "wormhole:")

	url, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	code := url.Query().Get("code")
	if code == "" {
		return nil, errors.New("no code")
	}

	return &parsedCode{
		relay: url.Host,
		code:  code,
	}, nil
}
