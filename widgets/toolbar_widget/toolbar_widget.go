package toolbar_widget

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"old-school-rpg-map-editor/common"
	"old-school-rpg-map-editor/common/load_save"
	"old-school-rpg-map-editor/models/center_model"
	"old-school-rpg-map-editor/models/copy_model"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/models/mode_model"
	"old-school-rpg-map-editor/models/notes_model"
	"old-school-rpg-map-editor/models/rot_map_model"
	"old-school-rpg-map-editor/models/rot_select_model"
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/models/select_model"
	"old-school-rpg-map-editor/models/selected_layer_model"
	"old-school-rpg-map-editor/models/selected_map_tab_model"
	"old-school-rpg-map-editor/undo_redo"
	"old-school-rpg-map-editor/utils"
	"old-school-rpg-map-editor/widgets/doc_tabs_widget"
	"old-school-rpg-map-editor/widgets/mode_toolbar_action"
	"old-school-rpg-map-editor/widgets/toolbar_action"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/goki/freetype/truetype"
	"github.com/google/uuid"
)

type ToolbarWidget struct {
	widget.Toolbar

	mapsModel *maps_model.MapsModel

	setModeToolbarAction    *mode_toolbar_action.ModeToolbarAction
	selectModeToolbarAction *mode_toolbar_action.ModeToolbarAction
	moveModeToolbarAction   *mode_toolbar_action.ModeToolbarAction

	saveFile    *toolbar_action.ToolbarAction
	undo        *toolbar_action.ToolbarAction
	redo        *toolbar_action.ToolbarAction
	rotateLeft  *toolbar_action.ToolbarAction
	rotateRight *toolbar_action.ToolbarAction
	copy        *toolbar_action.ToolbarAction
	cut         *toolbar_action.ToolbarAction
	paste       *toolbar_action.ToolbarAction

	currentMapId uuid.UUID
	disconnect   utils.Signal0
}

func NewToolbar(window fyne.Window, fnt *truetype.Font, mapsModel *maps_model.MapsModel, selectedMapTabModel *selected_map_tab_model.SelectedMapTabModel, copyModel *copy_model.CopyModel, rotateLeftIcon, rotateRightIcon, setModeIcon, setModeSelectedIcon, selectModeIcon, selectModeSelectedIcon, moveModeIcon, moveModeSelectedIcon fyne.Resource) *ToolbarWidget {
	w := &ToolbarWidget{
		Toolbar:      widget.Toolbar{},
		mapsModel:    mapsModel,
		currentMapId: uuid.UUID{},
	}

	newFile := toolbar_action.NewToolbarAction(theme.FileIcon(), func() {
		mapModel := map_model.NewMapModel()

		/*
			moveLayerIndex := mapModel.AddLayerWithId(uuid.New(), map_model.MoveLayerType)
			mapModel.SetName(moveLayerIndex, "MOVE")
		*/

		layerIndex := mapModel.AddLayerWithId(uuid.New(), map_model.RegularLayerType)
		mapModel.SetName(layerIndex, "Layer1")

		selectedLayerModel := selected_layer_model.NewSelectedLayerModel()
		rotateModel := rotate_model.NewRotateModel(0)
		rotMapModel := rot_map_model.NewRotMapMode(mapModel, rotateModel)
		selectModel := select_model.NewSelectModel(mapModel, selectedLayerModel)
		rotSelectModel := rot_select_model.NewRotSelectModel(selectModel, rotateModel)
		notesModel := notes_model.NewNotesModel(8, fnt)
		undoRedoQueue := undo_redo.NewUndoRedoQueue(100 /*TODO*/)
		centerModel := center_model.NewCenterModel(utils.Int2{})

		mapsModel.Add(mapModel, selectModel, mode_model.NewModeModel(), rotateModel, rotMapModel, rotSelectModel, notesModel, undoRedoQueue, selectedLayerModel, centerModel, "")
	})

	openFile := toolbar_action.NewToolbarAction(theme.FolderOpenIcon(), func() {
		d := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if uc == nil {
				return
			}

			defer uc.Close()

			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}

			mapModel, notesModel, err := load_save.LoadMapFile(uc)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}

			notesModel.SetFont(8, fnt)

			selectedLayerModel := selected_layer_model.NewSelectedLayerModel()
			rotateModel := rotate_model.NewRotateModel(0)
			selectModel := select_model.NewSelectModel(mapModel, selectedLayerModel)
			rotMapModel := rot_map_model.NewRotMapMode(mapModel, rotateModel)
			rotSelectModel := rot_select_model.NewRotSelectModel(selectModel, rotateModel)
			undoRedoQueue := undo_redo.NewUndoRedoQueue(100 /*TODO*/)
			centerModel := center_model.NewCenterModel(utils.Int2{})

			mapsModel.Add(mapModel, selectModel, mode_model.NewModeModel(), rotateModel, rotMapModel, rotSelectModel, notesModel, undoRedoQueue, selectedLayerModel, centerModel, uc.URI().Path())
		}, window)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".map"}))
		d.Resize(window.Canvas().Size())
		d.Show()
	})

	w.saveFile = toolbar_action.NewToolbarAction(theme.DocumentSaveIcon(), func() {
		d := dialog.NewFileSave(func(uc fyne.URIWriteCloser, err error) {
			if uc == nil {
				return
			}

			defer uc.Close()

			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}

			var t struct {
				Version    int                     `json:"version"`
				MapModel   *map_model.MapModel     `json:"map"`
				NotesModel *notes_model.NotesModel `json:"notes"`
			}

			mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
			t.Version = 1
			t.MapModel = mapElem.Model
			t.NotesModel = mapElem.NotesModel

			d, err := json.Marshal(&t)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}

			writer := gzip.NewWriter(uc)
			defer writer.Close()

			_, err = writer.Write(d)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}

			mapsModel.SetFilePath(mapElem.MapId, uc.URI().Path())
			mapsModel.UpdateSaveChangeGeneration(mapElem.MapId)
		}, window)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".map"}))
		d.Resize(window.Canvas().Size())
		d.Show()
	})

	w.setModeToolbarAction = mode_toolbar_action.NewModeToolbarAction(setModeIcon, setModeSelectedIcon, mode_model.SetMode, func(mm *mode_model.ModeModel, m mode_model.Mode) {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		SetMode(mapsModel, mapElem.MapId, mode_model.SetMode)
	})
	w.selectModeToolbarAction = mode_toolbar_action.NewModeToolbarAction(selectModeIcon, selectModeSelectedIcon, mode_model.SelectMode, func(mm *mode_model.ModeModel, m mode_model.Mode) {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		SetMode(mapsModel, mapElem.MapId, mode_model.SelectMode)
	})
	w.moveModeToolbarAction = mode_toolbar_action.NewModeToolbarAction(moveModeIcon, moveModeSelectedIcon, mode_model.MoveMode, func(mm *mode_model.ModeModel, m mode_model.Mode) {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		SetMode(mapsModel, mapElem.MapId, mode_model.MoveMode)
	})

	w.rotateLeft = toolbar_action.NewToolbarAction(rotateLeftIcon, func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		err := common.MakeAction(undo_redo.NewRotateCounterclockwiseAction(), mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})

	w.rotateRight = toolbar_action.NewToolbarAction(rotateRightIcon, func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		err := common.MakeAction(undo_redo.NewRotateClockwiseAction(), mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})

	w.undo = toolbar_action.NewToolbarAction(theme.ContentUndoIcon(), func() {
		Undo(selectedMapTabModel, mapsModel)
	})
	w.redo = toolbar_action.NewToolbarAction(theme.ContentRedoIcon(), func() {
		Redo(selectedMapTabModel, mapsModel)
	})

	w.cut = toolbar_action.NewToolbarAction(theme.ContentCutIcon(), func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		copyResult := Copy(mapElem.Model, mapElem.SelectedLayerModel, mapElem.RotateModel, mapElem.RotSelectModel, mapElem.RotMapModel)
		copyModel.SetCopyResult(copyResult)
		err := common.MakeAction(undo_redo.NewCutAction(copyResult), mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})
	w.copy = toolbar_action.NewToolbarAction(theme.ContentCopyIcon(), func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		copyModel.SetCopyResult(Copy(mapElem.Model, mapElem.SelectedLayerModel, mapElem.RotateModel, mapElem.RotSelectModel, mapElem.RotMapModel))
	})
	w.paste = toolbar_action.NewToolbarAction(theme.ContentPasteIcon(), func() {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())

		if copyModel.IsEmpty() {
			return
		}

		actions := undo_redo.NewUndoRedoContainer()

		actionModels := undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel, mapElem.SelectedLayerModel, mapElem.CenterModel)

		action := undo_redo.NewSetModeAndMergeDownMoveLayerAction(mode_model.MoveMode)
		action.Redo(actionModels)
		actions.Add(action)

		mapWidget := doc_tabs_widget.GetMapWidget(mapElem.ExternalData)

		copyResult := copyModel.CopyResult()

		pastePos := utils.Int2{}
		if mapWidget != nil {
			pastePos = mapWidget.Center()

			leftTop, rightBottom := copyResult.Bounds()
			if leftTop != rightBottom {
				centerX := (rightBottom.X - leftTop.X) / 2
				centerY := (rightBottom.Y - leftTop.Y) / 2
				pastePos.X -= centerX
				pastePos.Y -= centerY
			}
		}

		pasteToMoveLayerAction := undo_redo.NewPasteToMoveLayerAction(pastePos, copyResult)
		pasteToMoveLayerAction.Redo(actionModels)
		actions.Add(pasteToMoveLayerAction)

		err := common.MakeAction(actions, mapsModel, mapElem.MapId, nil)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})

	w.Items = append(w.Items,
		newFile,
		openFile,
		w.saveFile,
		widget.NewToolbarSeparator(),
		w.cut,
		w.copy,
		w.paste,
		widget.NewToolbarSeparator(),
		w.undo,
		w.redo,
		widget.NewToolbarSeparator(),
		w.setModeToolbarAction,
		w.selectModeToolbarAction,
		w.moveModeToolbarAction,
		widget.NewToolbarSeparator(),
		toolbar_action.NewToolbarAction(theme.ZoomInIcon(), func() {}),
		toolbar_action.NewToolbarAction(theme.ZoomOutIcon(), func() {}),
		w.rotateLeft,
		w.rotateRight,
	)

	disableButtons := func() {
		for _, b := range w.Items {
			if b == newFile || b == openFile {
				continue
			}
			if btn, ok := b.ToolbarObject().(*widget.Button); ok {
				btn.Disable()
			}
		}
	}
	disableButtons()

	selectedMapTabModel.AddDataChangeListener(func() {
		if mapsModel.Length() == 0 {
			w.setMapElem(maps_model.MapElem{})
		} else {
			mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
			if (mapElem.MapId != uuid.UUID{}) {
				w.setMapElem(mapElem)
			}
		}
	})

	copyModel.AddDataChangeListener(func() {
		if copyModel.IsEmpty() {
			w.paste.ToolbarObject().(*widget.Button).Disable()
			w.cut.ToolbarObject().(*widget.Button).Disable()
		} else {
			w.paste.ToolbarObject().(*widget.Button).Enable()
			w.cut.ToolbarObject().(*widget.Button).Enable()
		}
	})

	w.ExtendBaseWidget(w)

	return w
}

func (w *ToolbarWidget) setMapElem(mapElem maps_model.MapElem) {
	if mapElem.MapId == w.currentMapId {
		return
	}

	if (w.currentMapId != uuid.UUID{}) {
		w.disconnect.Emit()
		w.disconnect.Clear()
	}

	w.currentMapId = mapElem.MapId

	if (mapElem.MapId == uuid.UUID{}) {
		w.setModeToolbarAction.SetModeModel(nil)
		w.selectModeToolbarAction.SetModeModel(nil)
		w.moveModeToolbarAction.SetModeModel(nil)

		w.setModeToolbarAction.ToolbarObject().(*widget.Button).Disable()
		w.selectModeToolbarAction.ToolbarObject().(*widget.Button).Disable()
		w.moveModeToolbarAction.ToolbarObject().(*widget.Button).Disable()
		w.saveFile.ToolbarObject().(*widget.Button).Disable()
		w.undo.ToolbarObject().(*widget.Button).Disable()
		w.redo.ToolbarObject().(*widget.Button).Disable()
		w.rotateLeft.ToolbarObject().(*widget.Button).Disable()
		w.rotateRight.ToolbarObject().(*widget.Button).Disable()
		w.copy.ToolbarObject().(*widget.Button).Disable()
		w.cut.ToolbarObject().(*widget.Button).Disable()
		w.paste.ToolbarObject().(*widget.Button).Disable()
	} else {
		w.setModeToolbarAction.SetModeModel(mapElem.ModeModel)
		w.setModeToolbarAction.ToolbarObject().(*widget.Button).Enable()

		w.selectModeToolbarAction.SetModeModel(mapElem.ModeModel)
		w.selectModeToolbarAction.ToolbarObject().(*widget.Button).Enable()

		w.moveModeToolbarAction.SetModeModel(mapElem.ModeModel)

		mapModelDataChangeListener := func() {
			hasMoveLayer := len(mapElem.Model.LayerIndexByType(map_model.MoveLayerType)) > 0

			if hasMoveLayer {
				w.moveModeToolbarAction.ToolbarObject().(*widget.Button).Enable()
				w.copy.ToolbarObject().(*widget.Button).Disable()
				w.cut.ToolbarObject().(*widget.Button).Disable()
			} else {
				w.moveModeToolbarAction.ToolbarObject().(*widget.Button).Disable()
				w.copy.ToolbarObject().(*widget.Button).Enable()
				w.cut.ToolbarObject().(*widget.Button).Enable()
			}
		}
		w.disconnect.AddSlot(mapElem.Model.AddDataChangeListener(mapModelDataChangeListener))
		mapModelDataChangeListener()

		saveChangeGenerationListener := func(mapElem maps_model.MapElem) {
			if len(mapElem.FilePath) == 0 || mapElem.SavedChangeGeneration != mapElem.ChangeGeneration {
				w.saveFile.ToolbarObject().(*widget.Button).Enable()
			} else {
				w.saveFile.ToolbarObject().(*widget.Button).Disable()
			}
		}
		w.disconnect.AddSlot(w.mapsModel.AddDataChangeListener(func() {
			mapElem := w.mapsModel.GetById(w.currentMapId)
			if (mapElem.MapId != uuid.UUID{}) {
				saveChangeGenerationListener(mapElem)

				if mapElem.UndoRedoQueue.Action(mapElem.ChangeGeneration) == (undo_redo.UndoRedoElement{}) {
					w.undo.ToolbarObject().(*widget.Button).Disable()
				} else {
					w.undo.ToolbarObject().(*widget.Button).Enable()
				}

				if mapElem.UndoRedoQueue.ActionAfter(mapElem.ChangeGeneration) == (undo_redo.UndoRedoElement{}) {
					w.redo.ToolbarObject().(*widget.Button).Disable()
				} else {
					w.redo.ToolbarObject().(*widget.Button).Enable()
				}
			}
		}))
		saveChangeGenerationListener(mapElem)

		w.rotateLeft.ToolbarObject().(*widget.Button).Enable()
		w.rotateRight.ToolbarObject().(*widget.Button).Enable()
	}
}

func Undo(selectedMapTabModel *selected_map_tab_model.SelectedMapTabModel, mapsModel *maps_model.MapsModel) {
	mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
	action := mapElem.UndoRedoQueue.Action(mapElem.ChangeGeneration)
	if action.Action != nil {
		action.Action.Undo(undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel, mapElem.SelectedLayerModel, mapElem.CenterModel))
		actionBefore := mapElem.UndoRedoQueue.ActionBefore(mapElem.ChangeGeneration)
		mapsModel.SetChangeGeneration(mapElem.MapId, actionBefore.ChangeGeneration)
	}
}

func Redo(selectedMapTabModel *selected_map_tab_model.SelectedMapTabModel, mapsModel *maps_model.MapsModel) {
	mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
	actionAfter := mapElem.UndoRedoQueue.ActionAfter(mapElem.ChangeGeneration)
	if actionAfter.Action != nil {
		actionAfter.Action.Redo(undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel, mapElem.SelectedLayerModel, mapElem.CenterModel))
		mapsModel.SetChangeGeneration(mapElem.MapId, actionAfter.ChangeGeneration)
	}
}

func SetMode(mapsModel *maps_model.MapsModel, mapId uuid.UUID, mode mode_model.Mode) {
	mapElem := mapsModel.GetById(mapId)

	if mode == mode_model.SetMode || mode == mode_model.SelectMode {
		if mapElem.ModeModel.Mode() != mode {
			actions := undo_redo.NewUndoRedoContainer()

			actionModels := undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel, mapElem.SelectedLayerModel, mapElem.CenterModel)

			unselectAllAction := undo_redo.NewUnselectAllAction()
			unselectAllAction.Redo(actionModels)
			actions.Add(unselectAllAction)

			setModeAndMergeDownMoveLayerAction := undo_redo.NewSetModeAndMergeDownMoveLayerAction(mode)
			setModeAndMergeDownMoveLayerAction.Redo(actionModels)
			actions.Add(setModeAndMergeDownMoveLayerAction)

			err := common.MakeAction(actions, mapsModel, mapElem.MapId, nil)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}
		}
	} else if mode == mode_model.MoveMode {

	}
}

func Copy(m *map_model.MapModel, slm *selected_layer_model.SelectedLayerModel, r *rotate_model.RotateModel, rS *rot_select_model.RotSelectModel, rM *rot_map_model.RotMapModel) copy_model.CopyResult {
	result := copy_model.CopyResult{}
	result.Locations = make(map[utils.Int2]map_model.Location)
	result.LayerId = m.Layer(slm.Selected()).Uuid

	leftTop, rightBottom := rS.Bounds()

	for y := leftTop.Y; y < rightBottom.Y; y++ {
		for x := leftTop.X; x < rightBottom.X; x++ {
			selected := rS.At(x, y)
			tX, tY := r.TransformFromRot(x, y)

			setLocation := func(f func(l *map_model.Location)) {
				location := result.Locations[utils.NewInt2(tX, tY)]
				f(&location)
				result.Locations[utils.NewInt2(tX, tY)] = location
			}

			if selected.Floor {
				if v := rM.Floor(x, y, slm.Selected()); v > 0 {
					setLocation(func(l *map_model.Location) {
						l.Floor = v
					})
				}
			}
			if selected.RightWall {
				if v := rM.Wall(x, y, slm.Selected(), true); v > 0 {
					setLocation(func(l *map_model.Location) {
						l.RightWall = v
					})
				}
			}
			if selected.BottomWall {
				if v := rM.Wall(x, y, slm.Selected(), false); v > 0 {
					setLocation(func(l *map_model.Location) {
						l.BottomWall = v
					})
				}
			}
		}
	}

	return result
}
