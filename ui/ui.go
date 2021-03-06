package ui

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"log"
	"time"

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
)

type UI struct {
}

func New() *UI {
	return &UI{}
}

func (ui *UI) Run() error {
	w := app.NewWindow(app.Size(unit.Dp(800), unit.Dp(700)))
	dataDir, err := app.DataDir()
	if err != nil {
		logF("DataDir err: %s", err)
	} else {
		logF("DataDir: %s", dataDir)
	}

	if err := loop(w); err != nil {
		log.Fatal(err)
	}

	return nil
}

func logF(format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	log.Print(str)
	logText.Insert(fmt.Sprintf("[%s] %s\n", time.Now().Format(time.RFC3339), str))
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())

	// var viewEvent app.ViewEvent

	var ops op.Ops
	for {
		select {
		case e := <-w.Events():
			switch e := e.(type) {
			case system.DestroyEvent:
				return e.Err
			// case app.ViewEvent:
			// 	logF("Got view event")
			// 	viewEvent = e
			case system.FrameEvent:
				gtx := layout.NewContext(&ops, e)

				var sendTextClicked bool
				for sendTextBtn.Clicked() {
					sendTextClicked = true
				}

				if sendTextClicked {
					logF("Should send text: %s", textEditor.Text())
				}

				if enabledToggle.Changed() {
					state := "enabled"
					if !enabledToggle.Value {
						state = "disabled"
					}

					url := textEditor.Text()
					username := usernameEditor.Text()
					passwd := "<unset>"
					if passwordEditor.Text() != "" {
						passwd = "<redacted>"
					}

					logText.Insert(fmt.Sprintf("[%s] service state=%s url=%s username=%s password=%s\n", time.Now().Format(time.RFC3339), state, url, username, passwd))

					if state == "enabled" {
						files, err := ioutil.ReadDir("/sdcard/DCIM/Camera")
						if err != nil {
							logF("read sdcard err: %s", err)
						} else {
							logF("sdcard pictures: %+v", files)
						}
					}
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
	logText    = new(widget.Editor)
	textEditor = &widget.Editor{
		Submit: true,
	}
	usernameEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	passwordEditor = &widget.Editor{
		SingleLine: true,
		Submit:     true,
	}
	disableBtn   = new(widget.Clickable)
	sendTextBtn  = new(widget.Clickable)
	settingsList = &layout.List{
		Axis: layout.Vertical,
	}

	topLabel      = "Wormhole William"
	enabledToggle = new(widget.Bool)

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
					return drawSettings(gtx, th)
				case "Recv":
					return drawSettings(gtx, th)
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

func drawSendText(gtx layout.Context, th *material.Theme) layout.Dimensions {
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
					border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Px(2)}
					return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
					})
				}),
			)
		}
	}

	widgets := []layout.Widget{
		textField("Text", "Message", textEditor),

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, sendTextBtn, "Send")
						return btn.Layout(gtx)
					})
				}),
			)
		},
	}

	return settingsList.Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
		return layout.UniformInset(unit.Dp(16)).Layout(gtx, widgets[i])
	})
}

func drawSettings(gtx layout.Context, th *material.Theme) layout.Dimensions {
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
					border := widget.Border{Color: color.NRGBA{A: 0xff}, CornerRadius: unit.Dp(8), Width: unit.Px(2)}
					return border.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layout.UniformInset(unit.Dp(8)).Layout(gtx, e.Layout)
					})
				}),
			)
		}
	}

	widgets := []layout.Widget{
		textField("Username", "Username", usernameEditor),
		textField("Password", "Password", passwordEditor),

		func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx,
						material.Switch(th, enabledToggle).Layout,
					)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						text := "enabled"
						if !enabledToggle.Value {
							text = "disabled"
							gtx = gtx.Disabled()
						}

						btn := material.Button(th, disableBtn, text)
						return btn.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						if !enabledToggle.Value {
							return layout.Dimensions{}
						}
						return material.Loader(th).Layout(gtx)
					})
				}),
			)
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
