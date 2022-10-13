package common

import (
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/undo_redo"

	"github.com/google/uuid"
)

func MakeAction(action undo_redo.UndoRedoAction, mapsModel *maps_model.MapsModel, mapId uuid.UUID, addToContainer bool) (changeGeneration uint64, err error) {
	mapElem := mapsModel.GetById(mapId)

	action.Redo(undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel, mapElem.SelectedLayerModel))

	changeGeneration, err = mapElem.UndoRedoQueue.AddAction(mapElem.ChangeGeneration, action, addToContainer)
	if err != nil {
		return 0, err
	}

	mapsModel.SetChangeGeneration(mapElem.MapId, changeGeneration)

	return
}
