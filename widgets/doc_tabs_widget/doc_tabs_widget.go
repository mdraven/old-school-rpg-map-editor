package doc_tabs_widget

import (
	"fmt"
	"image"
	"old-school-rpg-map-editor/common"
	"old-school-rpg-map-editor/configuration"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/undo_redo"
	"old-school-rpg-map-editor/utils"
	"old-school-rpg-map-editor/widgets/layers_widget"
	"old-school-rpg-map-editor/widgets/map_widget"
	"old-school-rpg-map-editor/widgets/notes_widget"
	"old-school-rpg-map-editor/widgets/palette_widget"
	"path/filepath"
	"reflect"

	"fyne.io/fyne/v2/container"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

func newMapWidget(mapsModel *maps_model.MapsModel, mapId uuid.UUID, moveLayerId uuid.UUID, isClickFloor bool, floorPaletteWidget *palette_widget.PaletteWidget, wallPaletteWidget *palette_widget.PaletteWidget, notesWidget *notes_widget.NotesWidget, paletteTabFloors *container.TabItem, paletteTabNotes *container.TabItem, paletteTabs *container.AppTabs, layersWidget *layers_widget.LayersWidget, floorImage, wallImage, floorSelectedImage, wallSelectedImage image.Image, imageConfig configuration.ImageConfig) *map_widget.MapWidget {
	mapElem := mapsModel.GetById(mapId)
	model := mapElem.Model
	rotModel := mapElem.RotMapModel

	mapWidget := map_widget.NewMapWidget(floorImage, wallImage, floorSelectedImage, wallSelectedImage,
		imageConfig, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.ModeModel, mapElem.NotesModel, func(x, y int) {
			selectedTab := paletteTabs.Selected()
			if selectedTab == nil {
				return
			}

			if selectedTab == paletteTabFloors {
				activeLayer := map_model.LayerIndexWithoutSystemToWithSystem(model, int32(layersWidget.Selected))
				layerId := model.LayerInfo(activeLayer).Uuid

				value := uint32(floorPaletteWidget.Selected())
				if rotModel.Floor(x, y, model.LayerIndexById(layerId)) == value {
					return
				}

				_, err := common.MakeAction(undo_redo.NewSetFloorAction(utils.NewInt2(x, y), layerId, value), mapsModel, mapId, false)
				if err != nil {
					// TODO
					fmt.Println(err)
					return
				}
			} else if selectedTab == paletteTabNotes {
				activeLayer := map_model.LayerIndexWithoutSystemToWithSystem(model, int32(layersWidget.Selected))
				layerId := model.LayerInfo(activeLayer).Uuid

				value := notesWidget.Selected()
				if rotModel.Model().NoteId(x, y, model.LayerIndexById(layerId)) == value {
					return
				}

				_, err := common.MakeAction(undo_redo.NewSetNoteIdAction(utils.NewInt2(x, y), layerId, value), mapsModel, mapId, false)
				if err != nil {
					// TODO
					fmt.Println(err)
					return
				}
			}
		}, func(x, y int, isRight bool) {
			activeLayer := map_model.LayerIndexWithoutSystemToWithSystem(model, int32(layersWidget.Selected))
			layerId := model.LayerInfo(activeLayer).Uuid

			value := uint32(wallPaletteWidget.Selected())
			{
				wall := rotModel.Wall(x, y, model.LayerIndexById(layerId), isRight)
				if wall == value {
					return
				}
			}

			_, err := common.MakeAction(undo_redo.NewSetWallAction(utils.NewInt2(x, y), layerId, isRight, value), mapsModel, mapId, false)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}
		}, func(offsetX, offsetY int, startDrag bool) {
			if startDrag {
				_, err := common.MakeAction(undo_redo.NewUndoRedoContainer(reflect.TypeOf((*undo_redo.MoveToSelectedAction)(nil))), mapsModel, mapId, false)
				if err != nil {
					// TODO
					fmt.Println(err)
					return
				}
			}

			_, err := common.MakeAction(undo_redo.NewMoveToSelectedAction(moveLayerId, utils.NewInt2(offsetX, offsetY)), mapsModel, mapId, true)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}
		}, func(floors, rightWalls, bottomWalls []utils.Int2) {
			floors = pie.Filter(floors, func(pos utils.Int2) bool {
				return !mapElem.RotSelectModel.IsFloorSelected(pos.X, pos.Y)
			})
			rightWalls = pie.Filter(rightWalls, func(pos utils.Int2) bool {
				return !mapElem.RotSelectModel.IsWallSelected(pos.X, pos.Y, true)
			})
			bottomWalls = pie.Filter(bottomWalls, func(pos utils.Int2) bool {
				return !mapElem.RotSelectModel.IsWallSelected(pos.X, pos.Y, false)
			})

			if len(floors) == 0 && len(rightWalls) == 0 && len(bottomWalls) == 0 {
				return
			}

			_, err := common.MakeAction(undo_redo.NewUndoRedoContainer(reflect.TypeOf((*undo_redo.SelectAction)(nil))), mapsModel, mapId, false)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}

			addNewAction := func(pos utils.Int2, selectType undo_redo.SelectType) {
				_, err := common.MakeAction(undo_redo.NewSelectAction(pos, selectType), mapsModel, mapId, true)
				if err != nil {
					// TODO
					fmt.Println(err)
					return
				}
			}

			for _, pos := range floors {
				addNewAction(pos, undo_redo.Floor)
			}

			for _, pos := range rightWalls {
				addNewAction(pos, undo_redo.RightWall)
			}

			for _, pos := range bottomWalls {
				addNewAction(pos, undo_redo.BottomWall)
			}
		}, func() {
			leftTop, rightBottom := mapElem.SelectModel.Bounds()

			if leftTop != rightBottom {
				_, err := common.MakeAction(undo_redo.NewUnselectAllAction(), mapsModel, mapId, false)
				if err != nil {
					// TODO
					fmt.Println(err)
					return
				}
			}
		})

	mapWidget.SetIsClickFloor(isClickFloor)

	return mapWidget
}

type DocTabsWidget struct {
	container *container.DocTabs
	mapsModel *maps_model.MapsModel

	onClosed           utils.Signal1[maps_model.MapElem]
	onSelected         utils.Signal1[maps_model.MapElem]
	IsFloorTabSelected func() bool
}

func NewDocTabsWidget(mapsModel *maps_model.MapsModel, floorPaletteWidget *palette_widget.PaletteWidget, wallPaletteWidget *palette_widget.PaletteWidget, notesWidget *notes_widget.NotesWidget, paletteTabFloors *container.TabItem, paletteTabNotes *container.TabItem, paletteTabs *container.AppTabs, layersWidget *layers_widget.LayersWidget, floorImage, wallImage, floorSelectedImage, wallSelectedImage image.Image, imageConfig configuration.ImageConfig) *DocTabsWidget {
	w := &DocTabsWidget{}
	w.container = container.NewDocTabs()
	w.mapsModel = mapsModel

	w.container.OnClosed = func(ti *container.TabItem) {
		mapWidget := ti.Content.(*map_widget.MapWidget)
		mapWidget.Destroy()

		mapElem := mapsModel.GetByExternalData(ti)
		mapsModel.Delete(mapElem.MapId)

		w.onClosed.Emit(mapElem)
	}

	w.container.OnSelected = func(ti *container.TabItem) {
		mapElem := mapsModel.GetByExternalData(ti)

		w.onSelected.Emit(mapElem)
	}

	mapsModel.AddDataChangeListener(func() {
		// обновляем строчку с табами с открытыми файлами
		{
			data := mapsModel.GetIdAndExternalData()
			tabs := slices.Clone(w.container.Items)

			for _, d := range data {
				var tabName string

				m := mapsModel.GetById(d.MapId)
				if len(m.FilePath) > 0 {
					tabName = filepath.Base(m.FilePath)
				} else {
					tabName = "Unknown"
				}

				if m.ChangeGeneration != m.SavedChangeGeneration {
					tabName += "*"
				}

				index := -1
				if d.ExternalData != nil {
					index = slices.Index(tabs, d.ExternalData.(*container.TabItem))
				}

				if index > -1 {
					tabs[index].Text = tabName
					if w.container.Selected() == d.ExternalData {
						w.container.OnSelected(w.container.Selected())
					}
					tabs = slices.Delete(tabs, index, index+1)
				} else {
					moveLayerId := m.Model.LayerInfo(m.Model.LayerIndexByName("MOVE", map_model.MoveLayerType)).Uuid

					mapWidget := newMapWidget(mapsModel, m.MapId, moveLayerId, w.IsFloorTabSelected(), floorPaletteWidget, wallPaletteWidget, notesWidget, paletteTabFloors, paletteTabNotes, paletteTabs, layersWidget, floorImage, wallImage, floorSelectedImage, wallSelectedImage, imageConfig)
					item := container.NewTabItem(tabName, mapWidget)

					w.container.Append(item)
					mapsModel.SetExternalData(d.MapId, item)
				}
			}

			for _, t := range tabs {
				w.container.Remove(t)
			}

			w.container.Refresh()
		}
	})

	return w
}

func (w *DocTabsWidget) Container() *container.DocTabs {
	return w.container
}

func (w *DocTabsWidget) Selected() maps_model.MapElem {
	selected := w.container.Selected()
	if selected == nil {
		return maps_model.MapElem{}
	}

	return w.mapsModel.GetByExternalData(selected)
}

func (w *DocTabsWidget) MapWidget() *map_widget.MapWidget {
	selectedTab := w.container.Selected()
	if selectedTab == nil {
		return nil
	}
	return selectedTab.Content.(*map_widget.MapWidget)
}

func (m *DocTabsWidget) AddOnClosedListener(listener func(el maps_model.MapElem)) func() {
	return m.onClosed.AddSlot(listener)
}

func (m *DocTabsWidget) AddOnSelectedListener(listener func(el maps_model.MapElem)) func() {
	return m.onSelected.AddSlot(listener)
}
