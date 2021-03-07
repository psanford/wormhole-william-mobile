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
	"sync"

	"gioui.org/app"
	"gioui.org/font/gofont"
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
		case logMsg := <-plog.MsgChan():
			logText.Insert(logMsg)

		case result := <-pickResult:
			pickResult = nil
			// TODO(PMS): display error
			plog.Printf("pick result: path=%s name=%s err=%s", result.Path, result.Name, result.Err)
			if result.Err != nil {
				sendFileCodeTxt.SetText(fmt.Sprintf("Pick file err: %s", result.Err))
				w.Invalidate()
				continue
			}

			if result.Path != "" {
				f, err := os.Open(result.Path)
				if err != nil {
					plog.Printf("open file err path=%s err=%s", result.Path, err)
					sendFileCodeTxt.SetText(fmt.Sprintf("open file err: %s", err))
					w.Invalidate()
					continue
				}

				code, status, err := wh.SendFile(ctx, result.Name, f)
				if err != nil {
					plog.Printf("wormhole send error err=%s", err)
					sendFileCodeTxt.SetText(fmt.Sprintf("wormhole send err: %s", err))
					w.Invalidate()
					continue
				}

				sendFileCodeTxt.SetText(code)

				go func() {
					s := <-status
					if s.Error != nil {
						sendFileCodeTxt.SetText(fmt.Sprintf("wormhole send err: %s", s.Error))
					} else {
						sendFileCodeTxt.SetText("Send Complete!")
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

				var sendFileOnce sync.Once
				for sendFileBtn.Clicked() {
					sendFileOnce.Do(func() {
						pickResult = jgo.PickFile(viewEvent)
					})
				}

				var sendTextOnce sync.Once
				for sendTextBtn.Clicked() {
					sendTextOnce.Do(func() {
						msg := textMsgEditor.Text()
						if msg == "" {
							return
						}

						code, status, err := wh.SendText(ctx, msg)
						if err != nil {
							textStatus.SetText(fmt.Sprintf("Send err: %s", err))
							plog.Printf("Send err: %s", err)
							return
						}

						textStatus.SetText(fmt.Sprintf("Code: %s", code))

						go func() {
							s := <-status
							if s.Error != nil {
								textStatus.SetText(fmt.Sprintf("Send error: %s", s.Error))
								plog.Printf("Send error: %s", s.Error)
							} else if s.OK {
								textStatus.SetText("OK!")
							}
						}()
					})
				}

				var recvOnce sync.Once
				for recvMsgBtn.Clicked() {
					recvOnce.Do(func() {
						code := recvCodeEditor.Text()
						if code == "" {
							return
						}

						errf := func(msg string, args ...interface{}) {
							recvTxtMsg.SetText(fmt.Sprintf(msg, args...))
							plog.Printf(msg, args...)
						}

						go func() {
							msg, err := wh.Receive(ctx, code)
							if err != nil {
								// TODO(PMS): set the color to red
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

								// XXXXX
								// TODO(PMS): confirm file before recving

								f, err := os.CreateTemp(dataDir, fmt.Sprintf("%s.tmp", name))
								if err != nil {
									msg.Reject()
									errf("Create tmp file failed: %s", err)
									return
								}

								_, err = io.Copy(f, msg)
								if err != nil {
									os.Remove(f.Name())
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

								plog.Printf("Call NotifyDownloadManager")
								jgo.NotifyDownloadManager(viewEvent, name, path, contentType, msg.TransferBytes64)
							}
						}()
					})
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
	logText       = new(widget.Editor)
	textMsgEditor = &widget.Editor{
		Submit: true,
	}
	textStatus  = new(widget.Editor)
	sendTextBtn = new(widget.Clickable)

	recvCodeEditor = &widget.Editor{
		Submit: true,
	}
	recvMsgBtn   = new(widget.Clickable)
	recvTxtMsg   = new(widget.Editor)
	settingsList = &layout.List{
		Axis: layout.Vertical,
	}

	sendFileCodeTxt = new(widget.Editor)
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
			{
				Title: "Debug",
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
				case "Debug":
					return drawDebug(gtx, th)
				default:
					return layout.Center.Layout(gtx,
						material.H1(th, fmt.Sprintf("Tab content %s", selected)).Layout,
					)
				}
			})
		}),
	)
}

func textField(gtx layout.Context, th *material.Theme, label, hint string, editor *widget.Editor) func(layout.Context) layout.Dimensions {
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

		material.Button(th, sendTextBtn, "Send").Layout,
		func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(400))
			return material.Editor(th, textStatus, "").Layout(gtx)
		},
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func drawRecv(gtx layout.Context, th *material.Theme) layout.Dimensions {
	widgets := []layout.Widget{
		textField(gtx, th, "Code", "Code", recvCodeEditor),

		material.Button(th, recvMsgBtn, "Receive").Layout,
		func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(400))
			return material.Editor(th, recvTxtMsg, "").Layout(gtx)
		},
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func drawSendFile(gtx layout.Context, th *material.Theme) layout.Dimensions {
	widgets := []layout.Widget{
		material.Button(th, sendFileBtn, "Choose File").Layout,
		func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(400))
			return material.Editor(th, sendFileCodeTxt, "").Layout(gtx)
		},
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func drawDebug(gtx layout.Context, th *material.Theme) layout.Dimensions {
	widgets := []layout.Widget{
		material.H5(th, "Event Log").Layout,
		func(gtx C) D {
			gtx.Constraints.Max.Y = gtx.Px(unit.Dp(200))
			return material.Editor(th, logText, "").Layout(gtx)
		},
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}
