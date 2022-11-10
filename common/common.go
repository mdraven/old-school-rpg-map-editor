package common

import (
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/undo_redo"
	"reflect"

	"github.com/google/uuid"
)

// Так и не придумал хорошее название для функции.
//
// Эта функция пытается упростить использование просто Action и Action в контейнерах для Action(UndoRedoContainer).
//
//  1. Эта функция не делает Redo для контейнеров, так как ожидается, что для элементов уже был сделан Redo.
//  2. Если был задан addToContainer, то функция пытается добавить элемент в контейнер addToContainer. И если addToContainer - нет
//     то создаёт его и добавляет элементы.
func MakeAction(action undo_redo.UndoRedoAction, mapsModel *maps_model.MapsModel, mapId uuid.UUID, addToContainer reflect.Type) error {
	mapElem := mapsModel.GetById(mapId)

	if _, ok := action.(undo_redo.UndoRedoActionContainer); !ok {
		action.Redo(undo_redo.NewUndoRedoActionModels(mapElem.Model, mapElem.RotateModel, mapElem.RotMapModel, mapElem.RotSelectModel, mapElem.SelectModel, mapElem.ModeModel, mapElem.SelectedLayerModel, mapElem.CenterModel))
	}

	regularAddAction := func(action undo_redo.UndoRedoAction) error {
		changeGeneration, err := mapElem.UndoRedoQueue.AddAction(mapElem.ChangeGeneration, action)
		if err != nil {
			return err
		}

		mapsModel.SetChangeGeneration(mapElem.MapId, changeGeneration)

		return nil
	}

	if addToContainer == nil {
		return regularAddAction(action)
	} else {
		act := mapElem.UndoRedoQueue.Action(mapElem.ChangeGeneration).Action
		t := reflect.TypeOf(act)
		if container, ok := act.(undo_redo.UndoRedoActionContainer); ok && t == addToContainer {
			if !container.Add(action) {
				return regularAddAction(action)
			}
			return nil
		} else {
			reflectType := reflect.Zero(addToContainer)
			results := reflectType.MethodByName("New").Call([]reflect.Value{})
			container := results[0].Interface().(undo_redo.UndoRedoActionContainer)

			if container.Add(action) {
				return regularAddAction(container)
			} else {
				return regularAddAction(action)
			}
		}
	}
}
