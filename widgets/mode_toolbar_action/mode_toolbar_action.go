package mode_toolbar_action

import (
	"old-school-rpg-map-editor/models/mode_model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ widget.ToolbarItem = &ModeToolbarAction{}

type ModeToolbarAction struct {
	inactiveIcon fyne.Resource
	activeIcon   fyne.Resource
	mode         mode_model.Mode

	button *widget.Button

	modeModel           *mode_model.ModeModel
	disconnectModeModel func()
}

func NewModeToolbarAction(inactiveIcon fyne.Resource, activeIcon fyne.Resource, mode mode_model.Mode, tapped func(*mode_model.ModeModel, mode_model.Mode)) *ModeToolbarAction {
	a := &ModeToolbarAction{
		inactiveIcon: inactiveIcon,
		activeIcon:   activeIcon,
		mode:         mode,
	}

	a.button = widget.NewButtonWithIcon("", inactiveIcon, func() {
		if a.modeModel != nil {
			tapped(a.modeModel, a.mode)
		}
	})
	a.button.Importance = widget.LowImportance

	return a
}

func (a *ModeToolbarAction) ToolbarObject() fyne.CanvasObject {
	return a.button
}

func (a *ModeToolbarAction) SetModeModel(modeModel *mode_model.ModeModel) {
	if a.modeModel == modeModel {
		return
	}

	if a.modeModel != nil {
		a.disconnectModeModel()
	}

	a.modeModel = modeModel

	if modeModel != nil {
		a.disconnectModeModel = modeModel.AddDataChangeListener(func() {
			a.update()
		})
		a.update()
	}
}

func (a *ModeToolbarAction) ModeModel() *mode_model.ModeModel {
	return a.modeModel
}

func (a *ModeToolbarAction) update() {
	if a.modeModel.Mode() == a.mode {
		a.button.SetIcon(a.activeIcon)
	} else {
		a.button.SetIcon(a.inactiveIcon)
	}
}
