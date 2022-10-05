package layers_widget

import (
	"image/color"
	"old-school-rpg-map-editor/models/map_model"
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
	OnSelected func(id int)
	Selected   int

	mapModel      *map_model.MapModel
	disconnect    utils.Signal0
	visibleIcon   fyne.Resource
	invisibleIcon fyne.Resource
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

func NewLayersWidget(mapModel *map_model.MapModel, onSelected func(id int), visibleIcon, invisibleIcon fyne.Resource) *LayersWidget {
	w := &LayersWidget{
		List: widget.List{
			BaseWidget: widget.BaseWidget{},
		},
		OnSelected:    onSelected,
		Selected:      0,
		visibleIcon:   visibleIcon,
		invisibleIcon: invisibleIcon,
	}
	clearListHandlers(&w.List)

	w.SetMapModel(mapModel)

	w.List.OnSelected = func(id widget.ListItemID) {
		w.Selected = id
		if w.OnSelected != nil {
			w.OnSelected(w.Selected)
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
		w.disconnect.Emit()
		w.disconnect.Clear()
	}

	w.mapModel = mapModel

	if mapModel == nil {
		clearListHandlers(&w.List)
	} else {
		w.List.Length = func() int { return map_model.LengthWithoutSystem(mapModel) }
		w.List.CreateItem = func() fyne.CanvasObject {
			return newRow(w.visibleIcon)
		}
		w.List.UpdateItem = func(i widget.ListItemID, o fyne.CanvasObject) {
			layer := mapModel.LayerInfo(map_model.LayerIndexWithoutSystemToWithSystem(mapModel, int32(i)))
			if (layer.Uuid == uuid.UUID{}) {
				return
			}

			dataChanged(o.(*fyne.Container), mapModel, layer, w.visibleIcon, w.invisibleIcon)
		}

		w.disconnect.AddSlot(mapModel.AddDataChangeListener(w.Refresh))

		var activeLayerBeforeDelete uuid.UUID
		var nextActiveLayerBeforeDelete uuid.UUID // на случай, если удалили activeLayerBeforeDelete
		w.disconnect.AddSlot(mapModel.AddBeforeDeleteLayerListener(func() {
			index := map_model.LayerIndexWithoutSystemToWithSystem(mapModel, int32(w.Selected))
			activeLayerBeforeDelete = mapModel.LayerInfo(index).Uuid

			if w.Selected > 0 {
				index = map_model.LayerIndexWithoutSystemToWithSystem(mapModel, int32(w.Selected)-1)
			} else {
				index = map_model.LayerIndexWithoutSystemToWithSystem(mapModel, int32(w.Selected)+1)
			}
			nextActiveLayerBeforeDelete = mapModel.LayerInfo(index).Uuid
		}))
		w.disconnect.AddSlot(mapModel.AddAfterDeleteLayerListener(func() {
			index := mapModel.LayerIndexById(activeLayerBeforeDelete)
			if index == -1 {
				index = mapModel.LayerIndexById(nextActiveLayerBeforeDelete)
			}

			if index >= 0 {
				w.Select(int(map_model.LayerIndexWithSystemToWithoutSystem(mapModel, index)))
			}

			activeLayerBeforeDelete = uuid.UUID{}
			nextActiveLayerBeforeDelete = uuid.UUID{}
		}))

		var activeLayerBeforeMove uuid.UUID
		w.disconnect.AddSlot(mapModel.AddBeforeMoveLayerListener(func() {
			index := map_model.LayerIndexWithoutSystemToWithSystem(mapModel, int32(w.Selected))
			activeLayerBeforeMove = mapModel.LayerInfo(index).Uuid
		}))
		w.disconnect.AddSlot(mapModel.AddAfterMoveLayerListener(func() {
			index := mapModel.LayerIndexById(activeLayerBeforeMove)
			if index >= 0 {
				w.Select(int(map_model.LayerIndexWithSystemToWithoutSystem(mapModel, index)))
			}

			activeLayerBeforeMove = uuid.UUID{}
		}))
	}

	w.Refresh()
}

func (w *LayersWidget) MapModel() *map_model.MapModel {
	return w.mapModel
}
