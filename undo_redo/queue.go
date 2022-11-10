package undo_redo

import (
	"errors"
	"old-school-rpg-map-editor/utils"

	"github.com/elliotchance/pie/v2"
	"golang.org/x/exp/slices"
)

type UndoRedoElement struct {
	Action           UndoRedoAction
	ChangeGeneration uint64
}

type UndoRedoQueue struct {
	actions              []UndoRedoElement
	nextChangeGeneration uint64
	maxElements          int

	lastRemovedChangeGeneration uint64 // если мы сделали undo всех элементов, то надо знать какой теперь generation
}

func NewUndoRedoQueue(maxElements int) *UndoRedoQueue {
	return &UndoRedoQueue{nextChangeGeneration: 1, maxElements: maxElements}
}

func (q *UndoRedoQueue) AddAction(currentChangeGeneration uint64, action UndoRedoAction) (changeGeneration uint64, err error) {
	if action == nil {
		panic("action == nil")
	}

	if currentChangeGeneration > 0 {
		index := slices.IndexFunc(q.actions, func(e UndoRedoElement) bool {
			return e.ChangeGeneration == currentChangeGeneration
		})
		if index == -1 {
			return 0, errors.New("cannot find generation")
		}
		q.actions = q.actions[:index+1]
	}

	changeGeneration = q.nextChangeGeneration
	q.nextChangeGeneration++

	q.actions = append(q.actions, UndoRedoElement{Action: action, ChangeGeneration: changeGeneration})

	removeOldFrom := utils.Max(0, len(q.actions)-q.maxElements)
	if removeOldFrom != 0 {
		q.lastRemovedChangeGeneration = q.actions[removeOldFrom-1].ChangeGeneration
		q.actions = q.actions[removeOldFrom:]
	}

	// Удаляем UndoRedoContainer которые пустые. Не удаляем последний UndoRedoContainer, так как
	// в него могли ещё ничего не положить
	newActions := pie.Filter(q.actions[:len(q.actions)-1], func(a UndoRedoElement) bool {
		if container, ok := a.Action.(UndoRedoActionContainer); ok {
			return container.Len() > 0
		}
		return true
	})
	newActions = append(newActions, q.actions[len(q.actions)-1])
	q.actions = newActions

	return
}

func (q *UndoRedoQueue) Action(changeGeneration uint64) UndoRedoElement {
	if changeGeneration == q.lastRemovedChangeGeneration {
		return UndoRedoElement{Action: nil, ChangeGeneration: changeGeneration}
	}

	index := slices.IndexFunc(q.actions, func(e UndoRedoElement) bool {
		return e.ChangeGeneration == changeGeneration
	})
	if index == -1 {
		return UndoRedoElement{}
	}

	return q.actions[index]
}

func (q *UndoRedoQueue) ActionBefore(changeGeneration uint64) UndoRedoElement {
	index := slices.IndexFunc(q.actions, func(e UndoRedoElement) bool {
		return e.ChangeGeneration == changeGeneration
	})
	if index == -1 {
		return UndoRedoElement{}
	}

	if index == 0 {
		return UndoRedoElement{Action: nil, ChangeGeneration: q.lastRemovedChangeGeneration}
	}

	return q.actions[index-1]
}

func (q *UndoRedoQueue) ActionAfter(changeGeneration uint64) UndoRedoElement {
	var index int

	if changeGeneration == q.lastRemovedChangeGeneration {
		index = -1
	} else {
		index = slices.IndexFunc(q.actions, func(e UndoRedoElement) bool {
			return e.ChangeGeneration == changeGeneration
		})
		if index == -1 {
			return UndoRedoElement{}
		}
	}

	if index == len(q.actions)-1 {
		return UndoRedoElement{}
	}

	return q.actions[index+1]
}
