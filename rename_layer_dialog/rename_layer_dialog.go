package rename_layer_dialog

import (
	"old-school-rpg-map-editor/models/map_model"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

var _ dialog.Dialog = renameLayerDialog{}

type renameLayerDialog struct {
	dialog.Dialog
	parent fyne.Window
	entry  *widget.Entry
}

func NewRenameLayerDialog(parent fyne.Window, locations *map_model.MapModel, layerUuid uuid.UUID) renameLayerDialog {
	entry := widget.NewEntry()
	entry.SetText(locations.LayerInfo(locations.LayerIndexById(layerUuid)).Name)

	items := []*widget.FormItem{
		widget.NewFormItem("Name", entry),
	}

	d := dialog.NewForm("Rename layer", "Ok", "Cancel", items, func(b bool) {
		if b {
			locations.SetName(locations.LayerIndexById(layerUuid), entry.Text)
		}
	}, parent)

	return renameLayerDialog{d, parent, entry}
}

func (d renameLayerDialog) Show() {
	d.Dialog.Show()
	d.parent.Canvas().Focus(d.entry)
}
