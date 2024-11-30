package ui

import (
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// RichEditor is an extension to widget.Editor with copy and paste buttons.
// This allows users of a mobile device to easily make use of copy and paste
// in the selected text.
type Copyable struct {
	tag  int
	text string
	widget.Label
	copyButton, pasteButton widget.Clickable
}

func (c *Copyable) SetText(s string) {
	c.text = s
}

func (c *Copyable) Text() string {
	return c.text
}

// Layout updates the internal state of the Copyable.
func (r *Copyable) Layout(gtx C) D {
	// if the copy button was clicked, write the contents of the editor
	// into the system clipboard.
	if r.copyButton.Clicked(gtx) {
		gtx.Execute(clipboard.WriteCmd{
			Type: "application/text",
			Data: io.NopCloser(strings.NewReader(r.text)),
		})
	}

	return D{}
}

// CopyableStyle defines how a Copyable is presented.
type CopyableStyle struct {
	*material.Theme
	hint  string
	state *Copyable
	copy  bool
	// Inset around each button
	layout.Inset
}

// Layout renders the editor into the provided gtx.
func (r CopyableStyle) Layout(gtx C) D {
	// update the persistent state of this editor component
	r.state.Layout(gtx)

	children := make([]layout.FlexChild, 0, 2)

	children = append(children, layout.Flexed(1.0, func(gtx C) D {
		// ensure the editor does not try to use all available vertical space
		gtx.Constraints.Min.Y = 0
		return material.Body1(r.Theme, r.state.text).Layout(gtx)
	}))

	if r.copy && r.state.text != "" {
		children = append(children, layout.Rigid(func(gtx C) D {
			return r.Inset.Layout(gtx, material.IconButton(r.Theme, &r.state.copyButton, CopyIcon, "Copy").Layout)
		}))
	}

	// draw the interface after updating state
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
}

func CopyField(th *material.Theme, state *Copyable) CopyableStyle {
	return CopyableStyle{
		Theme: th,
		state: state,
		copy:  true,
		Inset: layout.UniformInset(unit.Dp(4)),
	}
}
