package toolbar

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"old-school-rpg-map-editor/common"
	"old-school-rpg-map-editor/common/load_save"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/models/mode_model"
	"old-school-rpg-map-editor/models/notes_model"
	"old-school-rpg-map-editor/models/rot_map_model"
	"old-school-rpg-map-editor/models/rot_select_model"
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/models/select_model"
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

type Toolbar struct {
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

	currentMapId uuid.UUID
	disconnect   utils.Signal0
}

func NewToolbar(window fyne.Window, fnt *truetype.Font, mapsModel *maps_model.MapsModel, mapTabs *doc_tabs_widget.DocTabsWidget, rotateLeftIcon, rotateRightIcon, setModeIcon, setModeSelectedIcon, selectModeIcon, selectModeSelectedIcon, moveModeIcon, moveModeSelectedIcon fyne.Resource) *Toolbar {
	w := &Toolbar{
		Toolbar:      widget.Toolbar{},
		mapsModel:    mapsModel,
		currentMapId: uuid.UUID{},
	}

	newFile := toolbar_action.NewToolbarAction(theme.FileIcon(), func() {
		mapModel := map_model.NewMapModel()

		moveLayerIndex := mapModel.AddLayerWithId(uuid.New())
		mapModel.SetSystem(moveLayerIndex, true)
		mapModel.SetName(moveLayerIndex, "MOVE")

		layerIndex := mapModel.AddLayerWithId(uuid.New())
		mapModel.SetName(layerIndex, "Layer1")

		rotateModel := rotate_model.NewRotateModel(0)
		rotMapModel := rot_map_model.NewRotMapMode(mapModel, rotateModel)
		selectModel := select_model.NewSelectModel(mapModel)
		rotSelectModel := rot_select_model.NewRotSelectModel(selectModel, rotateModel)
		notesModel := notes_model.NewNotesModel(8, fnt)
		undoRedoQueue := undo_redo.NewUndoRedoQueue(100 /*TODO*/)

		mapsModel.Add(mapModel, selectModel, mode_model.NewModeModel(), rotateModel, rotMapModel, rotSelectModel, notesModel, undoRedoQueue, "")
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

			rotateModel := rotate_model.NewRotateModel(0)
			selectModel := select_model.NewSelectModel(mapModel)
			rotMapModel := rot_map_model.NewRotMapMode(mapModel, rotateModel)
			rotSelectModel := rot_select_model.NewRotSelectModel(selectModel, rotateModel)
			undoRedoQueue := undo_redo.NewUndoRedoQueue(100 /*TODO*/)

			mapsModel.Add(mapModel, selectModel, mode_model.NewModeModel(), rotateModel, rotMapModel, rotSelectModel, notesModel, undoRedoQueue, uc.URI().Path())
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

			mapElem := mapTabs.Selected()
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
		mapElem := mapTabs.Selected()
		SetMode(mapsModel, mapElem.MapId, mode_model.SetMode)
	})
	w.selectModeToolbarAction = mode_toolbar_action.NewModeToolbarAction(selectModeIcon, selectModeSelectedIcon, mode_model.SelectMode, func(mm *mode_model.ModeModel, m mode_model.Mode) {
		mapElem := mapTabs.Selected()
		SetMode(mapsModel, mapElem.MapId, mode_model.SelectMode)
	})
	w.moveModeToolbarAction = mode_toolbar_action.NewModeToolbarAction(moveModeIcon, moveModeSelectedIcon, mode_model.MoveMode, func(mm *mode_model.ModeModel, m mode_model.Mode) {
		mapElem := mapTabs.Selected()
		SetMode(mapsModel, mapElem.MapId, mode_model.MoveMode)
	})

	w.rotateLeft = toolbar_action.NewToolbarAction(rotateLeftIcon, func() {
		mapElem := mapTabs.Selected()
		_, err := common.MakeAction(undo_redo.NewRotateCounterclockwiseAction(), mapsModel, mapElem.MapId, false)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})

	w.rotateRight = toolbar_action.NewToolbarAction(rotateRightIcon, func() {
		mapElem := mapTabs.Selected()
		_, err := common.MakeAction(undo_redo.NewRotateClockwiseAction(), mapsModel, mapElem.MapId, false)
		if err != nil {
			// TODO
			fmt.Println(err)
			return
		}
	})

	w.undo = toolbar_action.NewToolbarAction(theme.ContentUndoIcon(), func() {
		Undo(mapTabs, mapsModel)
	})
	w.redo = toolbar_action.NewToolbarAction(theme.ContentRedoIcon(), func() {
		Redo(mapTabs, mapsModel)
	})

	w.Items = append(w.Items,
		newFile,
		openFile,
		w.saveFile,
		widget.NewToolbarSeparator(),
		toolbar_action.NewToolbarAction(theme.ContentCutIcon(), func() {}),
		toolbar_action.NewToolbarAction(theme.ContentCopyIcon(), func() {}),
		toolbar_action.NewToolbarAction(theme.ContentPasteIcon(), func() {}),
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

	mapTabs.AddOnClosedListener(func(mapElem maps_model.MapElem) {
		if w.setModeToolbarAction.ModeModel() == mapElem.ModeModel {
			w.setMapElem(maps_model.MapElem{})
		}
	})

	mapTabs.AddOnSelectedListener(func(mapElem maps_model.MapElem) {
		if (mapElem.MapId != uuid.UUID{}) {
			w.setMapElem(mapElem)
		}
	})

	w.ExtendBaseWidget(w)

	return w
}

func (w *Toolbar) setMapElem(mapElem maps_model.MapElem) {
	if mapElem.MapId == w.currentMapId {
		return
	}

	if w.currentMapId != (uuid.UUID{}) {
		w.disconnect.Emit()
		w.disconnect.Clear()
	}

	w.currentMapId = mapElem.MapId

	if mapElem.MapId == (uuid.UUID{}) {
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
	} else {
		w.setModeToolbarAction.SetModeModel(mapElem.ModeModel)
		w.setModeToolbarAction.ToolbarObject().(*widget.Button).Enable()

		w.selectModeToolbarAction.SetModeModel(mapElem.ModeModel)
		w.selectModeToolbarAction.ToolbarObject().(*widget.Button).Enable()

		rotMapModelChangeListener := func() {
			leftTop, rightBottom := mapElem.RotSelectModel.Bounds()

			if leftTop != rightBottom {
				w.moveModeToolbarAction.ToolbarObject().(*widget.Button).Enable()
			} else {
				w.moveModeToolbarAction.ToolbarObject().(*widget.Button).Disable()
			}
		}
		w.moveModeToolbarAction.SetModeModel(mapElem.ModeModel)
		w.disconnect.AddSlot(mapElem.RotSelectModel.AddDataChangeListener(rotMapModelChangeListener))
		rotMapModelChangeListener()

		saveChangeGenerationListener := func(mapElem maps_model.MapElem) {
			if len(mapElem.FilePath) == 0 || mapElem.SavedChangeGeneration != mapElem.ChangeGeneration {
				w.saveFile.ToolbarObject().(*widget.Button).Enable()
			} else {
				w.saveFile.ToolbarObject().(*widget.Button).Disable()
			}
		}
		w.disconnect.AddSlot(w.mapsModel.AddDataChangeListener(func() {
			mapElem := w.mapsModel.GetById(w.currentMapId)
			if mapElem.MapId != (uuid.UUID{}) {
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

func Undo(mapTabs *doc_tabs_widget.DocTabsWidget, mapsModel *maps_model.MapsModel) {
	mapElem := mapTabs.Selected()
	action := mapElem.UndoRedoQueue.Action(mapElem.ChangeGeneration)
	if action.Action != nil {
		action.Action.Undo(undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel))
		actionBefore := mapElem.UndoRedoQueue.ActionBefore(mapElem.ChangeGeneration)
		mapsModel.SetChangeGeneration(mapElem.MapId, actionBefore.ChangeGeneration)
	}
}

func Redo(mapTabs *doc_tabs_widget.DocTabsWidget, mapsModel *maps_model.MapsModel) {
	mapElem := mapTabs.Selected()
	actionAfter := mapElem.UndoRedoQueue.ActionAfter(mapElem.ChangeGeneration)
	if actionAfter.Action != nil {
		actionAfter.Action.Redo(undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel))
		mapsModel.SetChangeGeneration(mapElem.MapId, actionAfter.ChangeGeneration)
	}
}

func SetMode(mapsModel *maps_model.MapsModel, mapId uuid.UUID, mode mode_model.Mode) {
	mapElem := mapsModel.GetById(mapId)

	if mode == mode_model.SetMode || mode == mode_model.SelectMode {
		locations := mapElem.Model
		moveLayerIndex := locations.LayerIndexByName("MOVE", true)

		if mapElem.ModeModel.Mode() != mode {
			_, err := common.MakeAction(undo_redo.NewMoveFromSelectModelAction(locations.LayerInfo(moveLayerIndex).Uuid, mode), mapsModel, mapElem.MapId, false)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}
		}
	} else if mode == mode_model.MoveMode {
		modeModel := mapElem.ModeModel
		locations := mapElem.Model
		moveLayerIndex := locations.LayerIndexByName("MOVE", true)

		if modeModel.Mode() != mode_model.MoveMode {
			_, err := common.MakeAction(undo_redo.NewMoveToSelectModelAction(locations.LayerInfo(moveLayerIndex).Uuid), mapsModel, mapElem.MapId, false)
			if err != nil {
				// TODO
				fmt.Println(err)
				return
			}
		}
	}
}
