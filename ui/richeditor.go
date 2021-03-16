package ui

import (
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

var CopyIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentContentCopy)
	return icon
}()

var PasteIcon *widget.Icon = func() *widget.Icon {
	icon, _ := widget.NewIcon(icons.ContentContentPaste)
	return icon
}()

// RichEditor is an extension to widget.Editor with copy and paste buttons.
// This allows users of a mobile device to easily make use of copy and paste
// in the selected text.
type RichEditor struct {
	tag int
	widget.Editor
	copyButton, pasteButton widget.Clickable
}

// Layout updates the internal state of the RichEditor.
func (r *RichEditor) Layout(gtx C) D {
	// if the copy button was clicked, write the contents of the editor
	// into the system clipboard.
	if r.copyButton.Clicked() {
		clipboard.WriteOp{
			Text: r.Editor.Text(),
		}.Add(gtx.Ops)
	}
	// if the paste button was clicked, request the contents of the system
	// clipboard. This is asynchronous, and the results will be delivered
	// in a future frame.
	if r.pasteButton.Clicked() {
		clipboard.ReadOp{
			Tag: &r.tag,
		}.Add(gtx.Ops)
	}
	// check for the results of a requested paste operation and insert them
	// into the editor if they arrive.
	for _, e := range gtx.Events(&r.tag) {
		switch e := e.(type) {
		case clipboard.Event:
			r.Editor.Insert(e.Text)
		}
	}
	return D{}
}

// RichEditorStyle defines how a RichEditor is presented.
type RichEditorStyle struct {
	*material.Theme
	hint  string
	state *RichEditor
	copy  bool
	paste bool
	// Inset around each button
	layout.Inset
}

// Layout renders the editor into the provided gtx.
func (r RichEditorStyle) Layout(gtx C) D {
	// update the persistent state of this editor component
	r.state.Layout(gtx)

	children := make([]layout.FlexChild, 0, 2)
	if r.copy && r.state.Editor.Text() != "" {
		children = append(children, layout.Rigid(func(gtx C) D {
			return r.Inset.Layout(gtx, material.IconButton(r.Theme, &r.state.copyButton, CopyIcon).Layout)
		}))
	}

	children = append(children, layout.Flexed(1.0, func(gtx C) D {
		// ensure the editor does not try to use all available vertical space
		gtx.Constraints.Min.Y = 0
		return material.Editor(r.Theme, &r.state.Editor, r.hint).Layout(gtx)
	}))

	if r.paste {
		children = append(children, layout.Rigid(func(gtx C) D {
			return r.Inset.Layout(gtx, material.IconButton(r.Theme, &r.state.pasteButton, PasteIcon).Layout)
		}))
	}

	// draw the interface after updating state
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
}

func PasteEditor(th *material.Theme, state *RichEditor, hint string) RichEditorStyle {
	return RichEditorStyle{
		Theme: th,
		state: state,
		paste: true,
		hint:  hint,
		Inset: layout.UniformInset(unit.Dp(4)),
	}
}
