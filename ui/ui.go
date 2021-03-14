package ui

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/psanford/wormhole-william-mobile/jgo"
	"github.com/psanford/wormhole-william-mobile/ui/plog"
	"github.com/psanford/wormhole-william/wormhole"
)

type UI struct {
}

func New() *UI {
	return &UI{}
}

func (ui *UI) Run() error {
	w := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(700)))

	if err := ui.loop(w); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (ui *UI) loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())

	var (
		wh  wormhole.Client
		ctx = context.Background()
	)

	var (
		pickResult <-chan jgo.PickResult
		viewEvent  app.ViewEvent
	)

	var ops op.Ops
	for {
		select {
		case result := <-pickResult:
			pickResult = nil
			plog.Printf("pick result: path=%s name=%s err=%s", result.Path, result.Name, result.Err)
			if result.Err != nil {
				statusMsg.SetText(fmt.Sprintf("Pick file err: %s", result.Err))
				w.Invalidate()
				continue
			}

			if result.Path != "" {
				f, err := os.Open(result.Path)
				if err != nil {
					plog.Printf("open file err path=%s err=%s", result.Path, err)
					statusMsg.SetText(fmt.Sprintf("open file err: %s", err))
					w.Invalidate()
					continue
				}

				progress := func(sentBytes, totalBytes int64) {
					statusMsg.SetText(fmt.Sprintf("Send progress %s/%s", formatBytes(sentBytes), formatBytes(totalBytes)))
					w.Invalidate()
				}
				code, status, err := wh.SendFile(ctx, result.Name, f, wormhole.WithProgress(progress))
				if err != nil {
					plog.Printf("wormhole send error err=%s", err)
					statusMsg.SetText(fmt.Sprintf("wormhole send err: %s", err))
					w.Invalidate()
					continue
				}

				sendFileCodeTxt.SetText(code)
				statusMsg.SetText("Waiting for receiver...")

				go func() {
					transferInProgress = true
					defer func() {
						transferInProgress = false
						recvCodeEditor.SetText("")
					}()

					s := <-status
					if s.Error != nil {

						statusMsg.SetText(fmt.Sprintf("wormhole send err: %s", s.Error))
					} else {
						statusMsg.SetText("Send Complete!")
						sendFileCodeTxt.SetText("")
					}
					w.Invalidate()
				}()
			}
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			case app.ViewEvent:
				viewEvent = e
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				var (
					sendFileClicked bool
					sendTextClicked bool
					recvClicked     bool
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
					{"acceptBtn", acceptBtn, &acceptClicked},
					{"cancelBtn", cancelBtn, &cancelClicked},
				}
				for _, btn := range btns {
					for btn.btn.Clicked() {
						*btn.clicked = true
					}
				}

				if sendFileClicked {
					pickResult = jgo.PickFile(viewEvent)
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
						key.FocusOp{}.Add(&ops) // blur textfield

						msg := textMsgEditor.Text()
						if msg == "" {
							return
						}

						code, status, err := wh.SendText(ctx, msg)
						if err != nil {
							statusMsg.SetText(fmt.Sprintf("Send err: %s", err))
							plog.Printf("Send err: %s", err)
							return
						}

						statusMsg.SetText("Waiting for receiver...")
						textCodeTxt.SetText(code)

						go func() {
							transferInProgress = true
							defer func() {
								transferInProgress = false
								recvCodeEditor.SetText("")
							}()

							s := <-status
							if s.Error != nil {
								statusMsg.SetText(fmt.Sprintf("Send error: %s", s.Error))
								plog.Printf("Send error: %s", s.Error)
							} else if s.OK {
								statusMsg.SetText("OK!")
							}
							textCodeTxt.SetText("")
							w.Invalidate()
						}()
					}()
				}

				if recvClicked {
					log.Printf("recv clicked")
					func() {
						statusMsg.SetText("Start recv")
						w.Invalidate()

						key.FocusOp{}.Add(&ops) // blur textfield
						code := recvCodeEditor.Text()
						if code == "" {
							return
						}

						errf := func(msg string, args ...interface{}) {
							statusMsg.SetText(fmt.Sprintf(msg, args...))
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

							msg, err := wh.Receive(recvCtx, code)
							if err != nil {
								errf("Recv msg err: %s", err)
								return
							}
							switch msg.Type {
							case wormhole.TransferText:
								msgBody, err := ioutil.ReadAll(msg)
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

								statusMsg.SetText(fmt.Sprintf("Receiving file (%s)  into %s\nAccept or Cancel?", formatBytes(msg.TransferBytes64), msg.Name))

								w.Invalidate()

								select {
								case <-cancelChan:
									msg.Reject()
									statusMsg.SetText("Transfer rejected")
									return
								case <-confirmChan:
								}
								confirmInProgress = false

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
										statusMsg.SetText("Transfer mid-stream aborted")
									case <-ctx.Done():
									}
								}()

								go func() {
									statusMsg.SetText(fmt.Sprintf("receiving %d/%s", 0, formatBytes(msg.TransferBytes64)))
									for count := range r.countUpdate {
										select {
										case <-ctx.Done():
											return
										default:
										}
										statusMsg.SetText(fmt.Sprintf("receiving %s/%s", formatBytes(count), formatBytes(msg.TransferBytes64)))
										w.Invalidate()
										time.Sleep(500 * time.Millisecond)
									}
								}()

								_, err = io.Copy(f, r)
								r.Close()
								if err != nil {
									os.Remove(f.Name())
									statusMsg.SetText(fmt.Sprintf("Receive file error: %s", err))
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
									errf("Rename file err: %s", err)
									return
								}

								var contentType string
								if msg.Type == wormhole.TransferDirectory {
									contentType = "application/zip"
								} else {
									contentType = http.DetectContentType(header)
								}

								statusMsg.SetText("Receive complete")
								w.Invalidate()

								plog.Printf("Call NotifyDownloadManager")
								jgo.NotifyDownloadManager(viewEvent, name, path, contentType, msg.TransferBytes64)
							}
						}()
					}()
				}

				layout.Inset{
					Bottom: e.Insets.Bottom,
					Left:   e.Insets.Left,
					Right:  e.Insets.Right,
					Top:    e.Insets.Top,
				}.Layout(gtx, func(gtx C) D {
					return drawTabs(gtx, th)
				})
				e.Frame(gtx.Ops)
			}
		}
	}
}

var (
	textMsgEditor = new(RichEditor)
	statusMsg     = new(widget.Editor)
	textCodeTxt   = new(RichEditor)
	textStat      = new(widget.Editor)
	sendTextBtn   = new(widget.Clickable)

	acceptBtn = new(widget.Clickable)
	cancelBtn = new(widget.Clickable)

	cancelChan  = make(chan struct{})
	confirmChan = make(chan struct{})

	transferInProgress bool
	confirmInProgress  bool

	recvCodeEditor = new(RichEditor)
	recvMsgBtn     = new(widget.Clickable)
	recvTxtMsg     = new(RichEditor)
	settingsList   = &layout.List{
		Axis: layout.Vertical,
	}

	sendFileCodeTxt = new(RichEditor)
	sendFileBtn     = new(widget.Clickable)

	topLabel = "Wormhole William"

	tabs = Tabs{
		tabs: []Tab{
			{
				Title: "Recv",
			},
			{
				Title: "Send Text",
			},
			{
				Title: "Send File",
			},
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
}

type (
	C = layout.Context
	D = layout.Dimensions
)

func drawTabs(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return material.H4(th, topLabel).Layout(gtx)
		}),
		layout.Rigid(func(gtx C) D {
			return tabs.list.Layout(gtx, len(tabs.tabs), func(gtx C, tabIdx int) D {
				t := &tabs.tabs[tabIdx]
				if t.btn.Clicked() {
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
							return layout.UniformInset(unit.Sp(12)).Layout(gtx,
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
						tabHeight := gtx.Px(unit.Dp(4))
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
				selected := tabs.tabs[tabs.selected].Title
				switch selected {
				case "Send Text":
					return drawSendText(gtx, th)
				case "Send File":
					return drawSendFile(gtx, th)
				case "Recv":
					return drawRecv(gtx, th)
				default:
					return layout.Center.Layout(gtx,
						material.H1(th, fmt.Sprintf("Tab content %s", selected)).Layout,
					)
				}
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
				border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Px(2)}
				return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
				})
			}),
		)
	}
}

func drawSendText(gtx layout.Context, th *material.Theme) layout.Dimensions {
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
				gtx.Constraints.Max.Y = gtx.Px(unit.Dp(400))
				return CopyEditor(th, textCodeTxt).Layout(gtx)
			}
			return D{}
		},
		material.Editor(th, statusMsg, "").Layout,
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func drawRecv(gtx layout.Context, th *material.Theme) layout.Dimensions {
	widgets := []layout.Widget{
		textField(gtx, th, "Code", "Code", recvCodeEditor),

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
			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(400))
			return CopyEditor(th, recvTxtMsg).Layout(gtx)
		},
		material.Editor(th, statusMsg, "").Layout,
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func drawSendFile(gtx layout.Context, th *material.Theme) layout.Dimensions {
	widgets := []layout.Widget{

		func(gtx C) D {
			if transferInProgress {
				gtx = gtx.Disabled()
			}
			return material.Button(th, sendFileBtn, "Choose File").Layout(gtx)
		},
		func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(400))
			return CopyEditor(th, sendFileCodeTxt).Layout(gtx)
		},
		material.Editor(th, statusMsg, "").Layout,
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
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
