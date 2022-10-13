package layers_widget

import (
	"image/color"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/selected_layer_model"
	"old-school-rpg-map-editor/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
)

type LayersWidget struct {
	widget.List

	mapModel           *map_model.MapModel
	disconnectMapModel utils.Signal0

	visibleIcon   fyne.Resource
	invisibleIcon fyne.Resource

	selectedLayerModel           *selected_layer_model.SelectedLayerModel
	disconnectSelectedLayerModel utils.Signal0
}

func newRow(visibleIcon fyne.Resource) *fyne.Container {
	label := widget.NewLabel("-")
	visible := widget.NewButtonWithIcon("", visibleIcon, func() {})
	return container.New(layout.NewHBoxLayout(), label, layout.NewSpacer(), visible)
}

func dataChanged(container *fyne.Container, locations *map_model.MapModel, layer map_model.LayerInfo, visibleIcon, invisibleIcon fyne.Resource) {
	container.Objects[0].(*widget.Label).SetText(layer.Name)

	visible := container.Objects[2].(*widget.Button)
	if layer.Visible {
		visible.SetIcon(visibleIcon)
	} else {
		visible.SetIcon(invisibleIcon)
	}
	visible.OnTapped = func() {
		locations.SetVisible(locations.LayerIndexById(layer.Uuid), !layer.Visible)
	}
}

func clearListHandlers(w *widget.List) {
	w.Length = func() int { return 0 }
	w.CreateItem = func() fyne.CanvasObject { return canvas.NewCircle(color.Transparent) }
	w.UpdateItem = func(id widget.ListItemID, item fyne.CanvasObject) {}
}

func NewLayersWidget(visibleIcon, invisibleIcon fyne.Resource) *LayersWidget {
	w := &LayersWidget{
		List: widget.List{
			BaseWidget: widget.BaseWidget{},
		},
		visibleIcon:   visibleIcon,
		invisibleIcon: invisibleIcon,
	}
	clearListHandlers(&w.List)

	w.List.OnSelected = func(id widget.ListItemID) {
		if w.selectedLayerModel != nil {
			w.selectedLayerModel.SetSelected(int32(id))
		}
	}

	w.ExtendBaseWidget(w)

	return w
}

func (w *LayersWidget) SetMapModel(mapModel *map_model.MapModel) {
	if w.mapModel == mapModel {
		return
	}

	if w.mapModel != nil {
		w.disconnectMapModel.Emit()
		w.disconnectMapModel.Clear()
	}

	w.mapModel = mapModel

	if mapModel == nil {
		clearListHandlers(&w.List)
	} else {
		w.List.Length = func() int { return mapModel.NumLayers() }
		w.List.CreateItem = func() fyne.CanvasObject {
			return newRow(w.visibleIcon)
		}
		w.List.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
			layer := mapModel.LayerInfo(int32(i))
			if (layer.Uuid == uuid.UUID{}) {
				return
			}

			dataChanged(o.(*fyne.Container), mapModel, layer, w.visibleIcon, w.invisibleIcon)
		}

		w.disconnectMapModel.AddSlot(mapModel.AddDataChangeListener(w.Refresh))

		var activeLayerBeforeDelete uuid.UUID
		var nextActiveLayerBeforeDelete uuid.UUID // на случай, если удалили activeLayerBeforeDelete
		w.disconnectMapModel.AddSlot(mapModel.AddBeforeDeleteLayerListener(func() {
			index := int32(0)
			if w.selectedLayerModel != nil {
				index = w.selectedLayerModel.Selected()
			}

			activeLayerBeforeDelete = mapModel.LayerInfo(index).Uuid

			if index > 0 {
				index = index - 1
			} else {
				index = index + 1
			}

			nextActiveLayerBeforeDelete = mapModel.LayerInfo(index).Uuid
		}))
		w.disconnectMapModel.AddSlot(mapModel.AddAfterDeleteLayerListener(func() {
			index := mapModel.LayerIndexById(activeLayerBeforeDelete)
			if index == -1 {
				index = mapModel.LayerIndexById(nextActiveLayerBeforeDelete)
			}

			if index >= 0 {
				w.Select(int(index))
			}

			activeLayerBeforeDelete = uuid.UUID{}
			nextActiveLayerBeforeDelete = uuid.UUID{}
		}))

		var activeLayerBeforeMove uuid.UUID
		w.disconnectMapModel.AddSlot(mapModel.AddBeforeMoveLayerListener(func() {
			index := int32(0)
			if w.selectedLayerModel != nil {
				index = w.selectedLayerModel.Selected()
			}
			activeLayerBeforeMove = mapModel.LayerInfo(index).Uuid
		}))
		w.disconnectMapModel.AddSlot(mapModel.AddAfterMoveLayerListener(func() {
			index := mapModel.LayerIndexById(activeLayerBeforeMove)
			if index >= 0 {
				w.Select(int(index))
			}

			activeLayerBeforeMove = uuid.UUID{}
		}))
	}

	w.Refresh()
}

func (w *LayersWidget) MapModel() *map_model.MapModel {
	return w.mapModel
}

func (a *LayersWidget) SetSelectedLayerModel(selectedLayerModel *selected_layer_model.SelectedLayerModel) {
	if a.selectedLayerModel == selectedLayerModel {
		return
	}

	if a.selectedLayerModel != nil {
		a.disconnectSelectedLayerModel.Emit()
		a.disconnectSelectedLayerModel.Clear()
	}

	a.selectedLayerModel = selectedLayerModel

	if selectedLayerModel != nil {
		a.disconnectSelectedLayerModel.AddSlot(selectedLayerModel.AddDataChangeListener(func() {
			a.Select(int(selectedLayerModel.Selected()))
		}))
	}
}
