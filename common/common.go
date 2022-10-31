package common

import (
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/undo_redo"

	"github.com/google/uuid"
)

// Так и не придумал хорошее название для функции.
//
// Эта функция пытается упростить использование просто Action и Action в контейнерах для Action(UndoRedoContainer).
//
// Если у нас просто Action, то эта функция выполняет Action с помощью Redo. И обновляет ChangeGeneration так как действие было выполнено.
// Но всё сложнее, если мы хотим положить Action в UndoRedoContainer:
//  1. мы выполняем Redo для этого Action, так как последующие действия в коде могут расчитывать, но изменения которые сделал Action. Поэтому их нельзя отложить и в конце сделать UndoRedoContainer.Redo. Кроме того мы не обновляем для Action значение ChangeGeneration, так как все undo/redo будут для UndoRedoContainer, а не Action;
//  2. мы не выполняем Redo для UndoRedoContainer так как все действия мы сделали в пункте 1.;
func MakeAction(action undo_redo.UndoRedoAction, mapsModel *maps_model.MapsModel, mapId uuid.UUID, addToContainer *undo_redo.UndoRedoContainer) error {
	mapElem := mapsModel.GetById(mapId)

	if _, ok := action.(*undo_redo.UndoRedoContainer); !ok {
		action.Redo(undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel, mapElem.SelectedLayerModel))
	}

	if addToContainer == nil {
		changeGeneration, err := mapElem.UndoRedoQueue.AddAction(mapElem.ChangeGeneration, action)
		if err != nil {
			return err
		}

		mapsModel.SetChangeGeneration(mapElem.MapId, changeGeneration)
	} else {
		action.(undo_redo.UndoRedoActionContainer).AddTo(addToContainer)
	}

	return nil
}
