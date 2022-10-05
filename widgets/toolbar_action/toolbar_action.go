package toolbar_action

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ widget.ToolbarItem = &ToolbarAction{}

// Эта штука нужна так как родной widget.ToolbarAction не умеет быть Disabled
type ToolbarAction struct {
	icon fyne.Resource

	button *widget.Button
}

func NewToolbarAction(icon fyne.Resource, tapped func()) *ToolbarAction {
	a := &ToolbarAction{
		icon: icon,
	}

	a.button = widget.NewButtonWithIcon("", icon, tapped)
	a.button.Importance = widget.LowImportance

	return a
}

func (a *ToolbarAction) ToolbarObject() fyne.CanvasObject {
	return a.button
}
