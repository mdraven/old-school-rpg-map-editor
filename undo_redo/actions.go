package undo_redo

import (
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/mode_model"
	"old-school-rpg-map-editor/models/rot_map_model"
	"old-school-rpg-map-editor/models/rot_select_model"
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/models/select_model"
	"old-school-rpg-map-editor/utils"
	"reflect"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type UndoRedoActionModels struct {
	M  *map_model.MapModel
	R  *rotate_model.RotateModel
	Rm *rot_map_model.RotMapModel
	Rs *rot_select_model.RotSelectModel
	Sm *select_model.SelectModel
	Mm *mode_model.ModeModel
}

func NewUndoRedoActionModels(m *map_model.MapModel, r *rotate_model.RotateModel, rm *rot_map_model.RotMapModel, rs *rot_select_model.RotSelectModel, sm *select_model.SelectModel, mm *mode_model.ModeModel) UndoRedoActionModels {
	return UndoRedoActionModels{
		M:  m,
		R:  r,
		Rm: rm,
		Rs: rs,
		Sm: sm,
		Mm: mm,
	}
}

type UndoRedoAction interface {
	Redo(m UndoRedoActionModels)
	Undo(m UndoRedoActionModels)
}

type UndoRedoActionContainer interface {
	UndoRedoAction
	addTo(container *UndoRedoContainer)
}

type UndoRedoContainer struct {
	actionTypes []reflect.Type
	actions     []UndoRedoActionContainer
}

func (c *UndoRedoContainer) goodActionType(t reflect.Type) bool {
	index := slices.IndexFunc(c.actionTypes, func(x reflect.Type) bool {
		return x == t
	})
	return index != -1
}

func NewUndoRedoContainer(actionTypes ...reflect.Type) *UndoRedoContainer {
	return &UndoRedoContainer{actionTypes: actionTypes}
}

func (c *UndoRedoContainer) Redo(m UndoRedoActionModels) {
	for _, action := range c.actions {
		action.Redo(m)
	}
}

func (c *UndoRedoContainer) Undo(m UndoRedoActionModels) {
	for i := len(c.actions) - 1; i >= 0; i-- {
		c.actions[i].Undo(m)
	}
}

func (c *UndoRedoContainer) Len() int {
	return len(c.actions)
}

var _ UndoRedoAction = &SetFloorAction{}

type SetFloorAction struct {
	pos      utils.Int2
	layerId  uuid.UUID
	value    uint32
	oldValue uint32
}

func NewSetFloorAction(pos utils.Int2, layerId uuid.UUID, value uint32) *SetFloorAction {
	return &SetFloorAction{pos: pos, layerId: layerId, value: value}
}

func (a *SetFloorAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*SetFloorAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *SetFloorAction) Redo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	a.oldValue = m.Rm.Floor(a.pos.X, a.pos.Y, layerIndex)
	m.Rm.SetFloor(a.pos.X, a.pos.Y, layerIndex, a.value)
}

func (a *SetFloorAction) Undo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	m.Rm.SetFloor(a.pos.X, a.pos.Y, layerIndex, a.oldValue)
}

type SetWallAction struct {
	pos      utils.Int2
	layerId  uuid.UUID
	isRight  bool
	value    uint32
	oldValue uint32
}

func NewSetWallAction(pos utils.Int2, layerId uuid.UUID, isRight bool, value uint32) *SetWallAction {
	return &SetWallAction{pos: pos, layerId: layerId, isRight: isRight, value: value}
}

func (a *SetWallAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*SetWallAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *SetWallAction) Redo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	a.oldValue = m.Rm.Wall(a.pos.X, a.pos.Y, layerIndex, a.isRight)
	m.Rm.SetWall(a.pos.X, a.pos.Y, layerIndex, a.isRight, a.value)
}

func (a *SetWallAction) Undo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	m.Rm.SetWall(a.pos.X, a.pos.Y, layerIndex, a.isRight, a.oldValue)
}

type SetNoteIdAction struct {
	pos      utils.Int2
	layerId  uuid.UUID
	value    string
	oldValue string
}

func NewSetNoteIdAction(pos utils.Int2, layerId uuid.UUID, value string) *SetNoteIdAction {
	return &SetNoteIdAction{pos: pos, layerId: layerId, value: value}
}

func (a *SetNoteIdAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*SetNoteIdAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *SetNoteIdAction) Redo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	a.oldValue = m.Rm.NoteId(a.pos.X, a.pos.Y, layerIndex)
	m.Rm.SetNoteId(a.pos.X, a.pos.Y, layerIndex, a.value)
}

func (a *SetNoteIdAction) Undo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	m.Rm.SetNoteId(a.pos.X, a.pos.Y, layerIndex, a.oldValue)
}

type AddLayerAction struct {
	layerId uuid.UUID
	name    string
	visible bool
	system  bool
}

func NewAddLayerAction(name string, visible, system bool) *AddLayerAction {
	return &AddLayerAction{layerId: uuid.New(), name: name, visible: visible, system: system}
}

func (a *AddLayerAction) Redo(m UndoRedoActionModels) {
	layerIndex := m.M.AddLayerWithId(a.layerId)
	m.M.SetName(layerIndex, a.name)
	m.M.SetVisible(layerIndex, a.visible)
	m.M.SetSystem(layerIndex, a.system)
}

func (a *AddLayerAction) Undo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	m.M.DeleteLayer(layerIndex)
}

type DeleteLayerAction struct {
	layerId uuid.UUID

	layer      *map_model.Layer
	layerIndex int32
}

func NewDeleteLayerAction(layerId uuid.UUID) *DeleteLayerAction {
	return &DeleteLayerAction{layerId: layerId}
}

func (a *DeleteLayerAction) Redo(m UndoRedoActionModels) {
	a.layerIndex = m.M.LayerIndexById(a.layerId)
	a.layer = m.M.Layer(a.layerIndex)
	m.M.DeleteLayer(a.layerIndex)
}

func (a *DeleteLayerAction) Undo(m UndoRedoActionModels) {
	layerIndex := m.M.AddLayer(a.layer)
	if layerIndex < a.layerIndex {
		m.M.MoveDown(layerIndex, a.layerIndex-layerIndex)
	} else if layerIndex > a.layerIndex {
		m.M.MoveUp(layerIndex, layerIndex-a.layerIndex)
	}
}

type MoveLayerAction struct {
	offset   int
	layerId  uuid.UUID
	oldIndex int32
}

func NewMoveLayerAction(offset int, layerId uuid.UUID) *MoveLayerAction {
	return &MoveLayerAction{offset: offset, layerId: layerId}
}

func (a *MoveLayerAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*MoveLayerAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *MoveLayerAction) Redo(m UndoRedoActionModels) {
	a.oldIndex = m.M.LayerIndexById(a.layerId)
	if a.offset > 0 {
		map_model.MoveDownWithoutSystem(m.M, a.oldIndex, int32(a.offset))
	} else {
		map_model.MoveUpWithoutSystem(m.M, a.oldIndex, -int32(a.offset))
	}
}

func (a *MoveLayerAction) Undo(m UndoRedoActionModels) {
	diff := m.M.LayerIndexById(a.layerId) - a.oldIndex
	if diff > 0 {
		map_model.MoveDownWithoutSystem(m.M, a.oldIndex, diff)
	} else {
		map_model.MoveUpWithoutSystem(m.M, a.oldIndex, -diff)
	}
}

type ClearLayerAction struct {
	layerId   uuid.UUID
	locations map[utils.Int2]map_model.Location
}

func NewClearLayerAction(layerId uuid.UUID) *ClearLayerAction {
	return &ClearLayerAction{layerId: layerId}
}

func (a *ClearLayerAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*ClearLayerAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *ClearLayerAction) Redo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	a.locations = m.M.Locations(layerIndex)
	m.M.ClearLayer(layerIndex)
}

func (a *ClearLayerAction) Undo(m UndoRedoActionModels) {
	m.M.SetLocations(m.M.LayerIndexById(a.layerId), a.locations)
}

var _ UndoRedoActionContainer = &MoveToSelectedAction{}

type MoveToSelectedAction struct {
	offset      utils.Int2
	moveLayerId uuid.UUID
}

func NewMoveToSelectedAction(moveLayerId uuid.UUID, offset utils.Int2) *MoveToSelectedAction {
	return &MoveToSelectedAction{offset: offset, moveLayerId: moveLayerId}
}

func (a *MoveToSelectedAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*MoveToSelectedAction)")
	}

	if a.offset.X == 0 && a.offset.Y == 0 {
		return
	}

	act := container.actions

	if len(act) == 0 {
		act = append(act, a)
	} else {
		prevAction := act[len(act)-1].(*MoveToSelectedAction)
		prevAction.offset.X += a.offset.X
		prevAction.offset.Y += a.offset.Y
	}

	container.actions = act
}

func (a *MoveToSelectedAction) Redo(m UndoRedoActionModels) {
	moveLayerIndex := m.M.LayerIndexById(a.moveLayerId)

	m.Rm.MoveTo(moveLayerIndex, a.offset.X, a.offset.Y)
	m.Rs.MoveTo(a.offset.X, a.offset.Y)
}

func (a *MoveToSelectedAction) Undo(m UndoRedoActionModels) {
	moveLayerIndex := m.M.LayerIndexById(a.moveLayerId)

	m.Rm.MoveTo(moveLayerIndex, -a.offset.X, -a.offset.Y)
	m.Rs.MoveTo(-a.offset.X, -a.offset.Y)
}

type SelectType int

const (
	Floor      SelectType = 0
	RightWall  SelectType = 1
	BottomWall SelectType = 2
)

type SelectAction struct {
	pos        utils.Int2
	selectType SelectType
}

func NewSelectAction(pos utils.Int2, selectType SelectType) *SelectAction {
	return &SelectAction{pos: pos, selectType: selectType}
}

func (a *SelectAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*SelectAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *SelectAction) Redo(m UndoRedoActionModels) {
	if a.selectType == Floor {
		m.Rs.SelectFloor(a.pos.X, a.pos.Y)
	} else if a.selectType == RightWall {
		m.Rs.SelectWall(a.pos.X, a.pos.Y, true)
	} else if a.selectType == BottomWall {
		m.Rs.SelectWall(a.pos.X, a.pos.Y, false)
	}
}

func (a *SelectAction) Undo(m UndoRedoActionModels) {
	if a.selectType == Floor {
		m.Rs.UnselectFloor(a.pos.X, a.pos.Y)
	} else if a.selectType == RightWall {
		m.Rs.UnselectWall(a.pos.X, a.pos.Y, true)
	} else if a.selectType == BottomWall {
		m.Rs.UnselectWall(a.pos.X, a.pos.Y, false)
	}
}

type UnselectAllAction struct {
	selected map[utils.Int2]select_model.Selected
}

func NewUnselectAllAction() *UnselectAllAction {
	return &UnselectAllAction{}
}

func (a *UnselectAllAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*UnselectAllAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *UnselectAllAction) Redo(m UndoRedoActionModels) {
	a.selected = m.Sm.Selected()
	m.Sm.UnselectAll()
}

func (a *UnselectAllAction) Undo(m UndoRedoActionModels) {
	m.Sm.SetSelected(a.selected)
}

type MoveFromMoveLayerAction struct {
	moveLayerId uuid.UUID
	leftTop     utils.Int2
	rightBottom utils.Int2
	actions     *UndoRedoContainer
}

func NewMoveFromMoveLayerAction(moveLayerId uuid.UUID, leftTop, rightBottom utils.Int2) *MoveFromMoveLayerAction {
	return &MoveFromMoveLayerAction{moveLayerId: moveLayerId, leftTop: leftTop, rightBottom: rightBottom, actions: NewUndoRedoContainer(reflect.TypeOf((*SetFloorAction)(nil)), reflect.TypeOf((*SetWallAction)(nil)))}
}

func (a *MoveFromMoveLayerAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*MoveFromMoveLayerAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *MoveFromMoveLayerAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		moveLayerIndex := m.M.LayerIndexById(a.moveLayerId)

		for y := a.leftTop.Y; y < a.rightBottom.Y; y++ {
			for x := a.leftTop.X; x < a.rightBottom.X; x++ {
				selected := m.Rs.At(x, y)

				if layerIndex := m.M.LayerIndexById(selected.Floor); layerIndex != -1 {
					action := NewSetFloorAction(utils.NewInt2(x, y), m.M.Layer(layerIndex).Uuid, m.Rm.Floor(x, y, moveLayerIndex))
					action.Redo(m)
					action.addTo(a.actions)

					action = NewSetFloorAction(utils.NewInt2(x, y), m.M.Layer(moveLayerIndex).Uuid, 0)
					action.Redo(m)
					action.addTo(a.actions)
				}
				if layerIndex := m.M.LayerIndexById(selected.RightWall); layerIndex != -1 {
					right := m.Rm.Wall(x, y, moveLayerIndex, true)

					action := NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(layerIndex).Uuid, true, right)
					action.Redo(m)
					action.addTo(a.actions)

					action = NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(moveLayerIndex).Uuid, true, 0)
					action.Redo(m)
					action.addTo(a.actions)
				}
				if layerIndex := m.M.LayerIndexById(selected.BottomWall); layerIndex != -1 {
					bottom := m.Rm.Wall(x, y, moveLayerIndex, false)

					action := NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(layerIndex).Uuid, false, bottom)
					action.Redo(m)
					action.addTo(a.actions)

					action = NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(moveLayerIndex).Uuid, false, 0)
					action.Redo(m)
					action.addTo(a.actions)
				}
			}
		}
	} else {
		a.actions.Redo(m)
	}
}

func (a *MoveFromMoveLayerAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type MoveToMoveLayerAction struct {
	moveLayerId uuid.UUID
	leftTop     utils.Int2
	rightBottom utils.Int2
	actions     *UndoRedoContainer
}

func NewMoveToMoveLayerAction(moveLayerId uuid.UUID, leftTop, rightBottom utils.Int2) *MoveToMoveLayerAction {
	return &MoveToMoveLayerAction{moveLayerId: moveLayerId, leftTop: leftTop, rightBottom: rightBottom, actions: NewUndoRedoContainer(reflect.TypeOf((*SetFloorAction)(nil)), reflect.TypeOf((*SetWallAction)(nil)))}
}

func (a *MoveToMoveLayerAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*MoveFromMoveLayerAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *MoveToMoveLayerAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		moveLayerIndex := m.M.LayerIndexById(a.moveLayerId)

		for y := a.leftTop.Y; y < a.rightBottom.Y; y++ {
			for x := a.leftTop.X; x < a.rightBottom.X; x++ {
				selected := m.Rs.At(x, y)

				if layerIndex := m.M.LayerIndexById(selected.Floor); layerIndex != -1 {
					action := NewSetFloorAction(utils.NewInt2(x, y), m.M.Layer(moveLayerIndex).Uuid, m.Rm.Floor(x, y, layerIndex))
					action.Redo(m)
					action.addTo(a.actions)

					action = NewSetFloorAction(utils.NewInt2(x, y), m.M.Layer(layerIndex).Uuid, 0)
					action.Redo(m)
					action.addTo(a.actions)
				}
				if layerIndex := m.M.LayerIndexById(selected.RightWall); layerIndex != -1 {
					right := m.Rm.Wall(x, y, layerIndex, true)

					action := NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(moveLayerIndex).Uuid, true, right)
					action.Redo(m)
					action.addTo(a.actions)

					action = NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(layerIndex).Uuid, true, 0)
					action.Redo(m)
					action.addTo(a.actions)
				}
				if layerIndex := m.M.LayerIndexById(selected.BottomWall); layerIndex != -1 {
					bottom := m.Rm.Wall(x, y, layerIndex, false)

					action := NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(moveLayerIndex).Uuid, false, bottom)
					action.Redo(m)
					action.addTo(a.actions)

					action = NewSetWallAction(utils.NewInt2(x, y), m.M.Layer(layerIndex).Uuid, false, 0)
					action.Redo(m)
					action.addTo(a.actions)
				}
			}
		}
	} else {
		a.actions.Redo(m)
	}
}

func (a *MoveToMoveLayerAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type MoveFromSelectModelAction struct {
	moveLayerId uuid.UUID
	mode        mode_model.Mode
	actions     *UndoRedoContainer
}

func NewMoveFromSelectModelAction(moveLayerId uuid.UUID, mode mode_model.Mode) *MoveFromSelectModelAction {
	if mode != mode_model.SetMode && mode != mode_model.SelectMode {
		panic("mode != mode_model.SetMode && mode != mode_model.SelectMode")
	}
	return &MoveFromSelectModelAction{moveLayerId: moveLayerId, mode: mode, actions: NewUndoRedoContainer(reflect.TypeOf((*MoveFromMoveLayerAction)(nil)), reflect.TypeOf((*ClearLayerAction)(nil)), reflect.TypeOf((*UnselectAllAction)(nil)), reflect.TypeOf((*SetModeAction)(nil)))}
}

func (a *MoveFromSelectModelAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		leftTop, rightBottom := m.Rm.Bounds(m.M.LayerIndexById(a.moveLayerId))

		if leftTop != rightBottom {
			moveAction := NewMoveFromMoveLayerAction(a.moveLayerId, leftTop, rightBottom)
			moveAction.Redo(m)
			moveAction.addTo(a.actions)

			clearAction := NewClearLayerAction(a.moveLayerId)
			clearAction.Redo(m)
			clearAction.addTo(a.actions)
		}

		unselectAction := NewUnselectAllAction()
		unselectAction.Redo(m)
		unselectAction.addTo(a.actions)

		if m.Mm.Mode() != a.mode {
			setModeAction := NewSetModeAction(a.mode)
			setModeAction.Redo(m)
			setModeAction.addTo(a.actions)
		}
	} else {
		a.actions.Redo(m)
	}
}

func (a *MoveFromSelectModelAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type MoveToSelectModelAction struct {
	moveLayerId uuid.UUID
	actions     *UndoRedoContainer
}

func NewMoveToSelectModelAction(moveLayerId uuid.UUID) *MoveToSelectModelAction {
	return &MoveToSelectModelAction{moveLayerId: moveLayerId, actions: NewUndoRedoContainer(reflect.TypeOf((*MoveLayerAction)(nil)), reflect.TypeOf((*MoveToMoveLayerAction)(nil)), reflect.TypeOf((*SetModeAction)(nil)))}
}

func (a *MoveToSelectModelAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		moveLayerIndex := m.M.LayerIndexById(a.moveLayerId)

		// Делаем слой для перемещения самым верхним
		moveUpAction := NewMoveLayerAction(-int(moveLayerIndex), a.moveLayerId)
		moveUpAction.Redo(m)
		moveUpAction.addTo(a.actions)

		leftTop, rightBottom := m.Rs.Bounds()
		// если какие-то блоки были выделены, то переносим их в moveLayer
		if leftTop != rightBottom {
			moveAction := NewMoveToMoveLayerAction(a.moveLayerId, leftTop, rightBottom)
			moveAction.Redo(m)
			moveAction.addTo(a.actions)
		}

		setModeAction := NewSetModeAction(mode_model.MoveMode)
		setModeAction.Redo(m)
		setModeAction.addTo(a.actions)
	} else {
		a.actions.Redo(m)
	}
}

func (a *MoveToSelectModelAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type SetModeAction struct {
	mode    mode_model.Mode
	oldMode mode_model.Mode
}

func NewSetModeAction(mode mode_model.Mode) *SetModeAction {
	return &SetModeAction{mode: mode}
}

func (a *SetModeAction) addTo(container *UndoRedoContainer) {
	if !container.goodActionType(reflect.TypeOf(a)) {
		panic("container.actionType != reflect.TypeOf(*SetModeAction)")
	}

	container.actions = append(container.actions, a)
}

func (a *SetModeAction) Redo(m UndoRedoActionModels) {
	a.oldMode = m.Mm.Mode()
	m.Mm.SetMode(a.mode)
}

func (a *SetModeAction) Undo(m UndoRedoActionModels) {
	m.Mm.SetMode(a.oldMode)
}

type RotateClockwiseAction struct{}

func NewRotateClockwiseAction() *RotateClockwiseAction {
	return &RotateClockwiseAction{}
}

func (a *RotateClockwiseAction) Redo(m UndoRedoActionModels) {
	m.R.RotateClockwise()
}

func (a *RotateClockwiseAction) Undo(m UndoRedoActionModels) {
	m.R.RotateCounterclockwise()
}

type RotateCounterclockwiseAction struct{}

func NewRotateCounterclockwiseAction() *RotateCounterclockwiseAction {
	return &RotateCounterclockwiseAction{}
}

func (a *RotateCounterclockwiseAction) Redo(m UndoRedoActionModels) {
	m.R.RotateCounterclockwise()
}

func (a *RotateCounterclockwiseAction) Undo(m UndoRedoActionModels) {
	m.R.RotateClockwise()
}
