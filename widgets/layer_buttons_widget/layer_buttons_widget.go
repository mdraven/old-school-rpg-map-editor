package layer_buttons_widget

import (
	"fmt"
	"old-school-rpg-map-editor/common"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/models/selected_map_tab_model"
	"old-school-rpg-map-editor/rename_layer_dialog"
	"old-school-rpg-map-editor/undo_redo"
	"old-school-rpg-map-editor/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

type LayerButtonsWidget struct {
	mapsModel           *maps_model.MapsModel
	selectedMapTabModel *selected_map_tab_model.SelectedMapTabModel

	container           *fyne.Container
	moveUpLayerButtom   *widget.Button
	moveDownLayerButtom *widget.Button
	addLayerButtom      *widget.Button
	removeLayerButtom   *widget.Button
	renameLayerButtom   *widget.Button
}

func NewLayerButtonsWidget(parent fyne.Window, mapsModel *maps_model.MapsModel, selectedMapTabModel *selected_map_tab_model.SelectedMapTabModel) *LayerButtonsWidget {
	w := &LayerButtonsWidget{
		mapsModel:           mapsModel,
		selectedMapTabModel: selectedMapTabModel,
	}

	// инструменты слева(слои и палитра)
	w.addLayerButtom = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())

		err := common.MakeAction(undo_redo.NewAddLayerAction("Layer", true, map_model.RegularLayerType), w.mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})
	w.removeLayerButtom = widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		locations := mapElem.Model

		activeLayer := mapElem.SelectedLayerModel.Selected()
		layerId := locations.LayerInfo(activeLayer).Uuid

		err := common.MakeAction(undo_redo.NewDeleteLayerAction(layerId), w.mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})
	w.moveUpLayerButtom = widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		locations := mapElem.Model

		activeLayer := mapElem.SelectedLayerModel.Selected()
		layerId := locations.LayerInfo(activeLayer).Uuid

		err := common.MakeAction(undo_redo.NewMoveLayerAction(-1, layerId), w.mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})
	w.moveDownLayerButtom = widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		locations := mapElem.Model

		activeLayer := mapElem.SelectedLayerModel.Selected()
		layerId := locations.LayerInfo(activeLayer).Uuid

		err := common.MakeAction(undo_redo.NewMoveLayerAction(1, layerId), w.mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})
	w.renameLayerButtom = widget.NewButton("Rename", func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		locations := mapElem.Model
		activeLayer := mapElem.SelectedLayerModel.Selected()

		dialog := rename_layer_dialog.NewRenameLayerDialog(parent, locations, locations.LayerInfo(activeLayer).Uuid)
		dialog.Show()
	})

	w.container = container.New(layout.NewHBoxLayout(), w.moveUpLayerButtom, w.moveDownLayerButtom, w.addLayerButtom, w.removeLayerButtom, w.renameLayerButtom)

	disableAllButtons := func() {
		w.moveUpLayerButtom.Disable()
		w.moveDownLayerButtom.Disable()
		w.addLayerButtom.Disable()
		w.removeLayerButtom.Disable()
		w.renameLayerButtom.Disable()
	}
	disableAllButtons()

	mapsModel.AddDataChangeListener(func() {
		if mapsModel.Length() == 0 {
			disableAllButtons()
		}
	})

	var disconnectSelectedMapTab utils.Signal0
	selectedMapTabModel.AddDataChangeListener(func() {
		disconnectSelectedMapTab.Emit()
		disconnectSelectedMapTab.Clear()

		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		if (mapElem.MapId == uuid.UUID{}) {
			disableAllButtons()
			return
		}

		updateButtonState := func() {
			indices := mapElem.Model.LayerIndexByType(map_model.MoveLayerType)
			if len(indices) > 0 {
				disableAllButtons()
				return
			}

			layerIndex := mapElem.SelectedLayerModel.Selected()

			if layerIndex == 0 {
				w.moveUpLayerButtom.Disable()
			} else {
				w.moveUpLayerButtom.Enable()
			}
			if layerIndex == int32(mapElem.Model.NumLayers()-1) {
				w.moveDownLayerButtom.Disable()
			} else {
				w.moveDownLayerButtom.Enable()
			}
			w.addLayerButtom.Enable()
			w.removeLayerButtom.Enable()
			w.renameLayerButtom.Enable()
		}
		updateButtonState()

		disconnectSelectedMapTab.AddSlot(mapElem.Model.AddDataChangeListener(func() {
			updateButtonState()
		}))
		disconnectSelectedMapTab.AddSlot(mapElem.SelectedLayerModel.AddDataChangeListener(func() {
			updateButtonState()
		}))
	})

	return w
}

func (w *LayerButtonsWidget) Container() *fyne.Container {
	return w.container
}
