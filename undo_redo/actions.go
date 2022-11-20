package undo_redo

import (
	"old-school-rpg-map-editor/models/center_model"
	"old-school-rpg-map-editor/models/copy_model"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/mode_model"
	"old-school-rpg-map-editor/models/rot_map_model"
	"old-school-rpg-map-editor/models/rot_select_model"
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/models/select_model"
	"old-school-rpg-map-editor/models/selected_layer_model"
	"old-school-rpg-map-editor/utils"
	"reflect"
	"sync"

	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

type UndoRedoActionModels struct {
	M   *map_model.MapModel
	R   *rotate_model.RotateModel
	Rm  *rot_map_model.RotMapModel
	Rs  *rot_select_model.RotSelectModel
	Sm  *select_model.SelectModel
	Mm  *mode_model.ModeModel
	Slm *selected_layer_model.SelectedLayerModel
	Cm  *center_model.CenterModel
}

func NewUndoRedoActionModels(m *map_model.MapModel, r *rotate_model.RotateModel, rm *rot_map_model.RotMapModel, rs *rot_select_model.RotSelectModel, sm *select_model.SelectModel, mm *mode_model.ModeModel, slm *selected_layer_model.SelectedLayerModel, cm *center_model.CenterModel) UndoRedoActionModels {
	return UndoRedoActionModels{
		M:   m,
		R:   r,
		Rm:  rm,
		Rs:  rs,
		Sm:  sm,
		Mm:  mm,
		Slm: slm,
		Cm:  cm,
	}
}

type UndoRedoAction interface {
	Redo(m UndoRedoActionModels)
	Undo(m UndoRedoActionModels)
	//IsChangeSaveFile() bool
}

type UndoRedoActionContainer interface {
	UndoRedoAction
	Add(a UndoRedoAction) bool
	Len() int
	New() UndoRedoActionContainer
}

var _ UndoRedoActionContainer = &UndoRedoContainer{}

type UndoRedoContainer struct {
	mutex   sync.Mutex
	actions []UndoRedoAction
}

func NewUndoRedoContainer() *UndoRedoContainer {
	return &UndoRedoContainer{}
}

func (*UndoRedoContainer) New() UndoRedoActionContainer {
	return NewUndoRedoContainer()
}

func (c *UndoRedoContainer) Add(a UndoRedoAction) bool {
	c.mutex.Lock()
	c.actions = append(c.actions, a)
	c.mutex.Unlock()

	return true
}

func (c *UndoRedoContainer) Redo(m UndoRedoActionModels) {
	c.mutex.Lock()
	actions := slices.Clone(c.actions)
	c.mutex.Unlock()

	for _, action := range actions {
		action.Redo(m)
	}
}

func (c *UndoRedoContainer) Undo(m UndoRedoActionModels) {
	c.mutex.Lock()
	actions := slices.Clone(c.actions)
	c.mutex.Unlock()

	for i := len(actions) - 1; i >= 0; i-- {
		actions[i].Undo(m)
	}
}

func (c *UndoRedoContainer) Len() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.actions)
}

var _ UndoRedoActionContainer = &SetCenterContainer{}

type SetCenterContainer struct {
	UndoRedoContainer
}

func NewSetCenterContainer() *SetCenterContainer {
	return &SetCenterContainer{}
}

func (*SetCenterContainer) New() UndoRedoActionContainer {
	return NewSetCenterContainer()
}

func (c *SetCenterContainer) Add(a UndoRedoAction) bool {
	if reflect.TypeOf(a) != reflect.TypeOf((*SetCenterAction)(nil)) {
		return false
	}

	c.mutex.Lock()
	c.actions = append(c.actions, a)
	c.mutex.Unlock()

	return true
}

var _ UndoRedoActionContainer = &RotateMapContainer{}

type RotateMapContainer struct {
	UndoRedoContainer
}

func NewRotateMapContainer() *RotateMapContainer {
	return &RotateMapContainer{}
}

func (*RotateMapContainer) New() UndoRedoActionContainer {
	return NewRotateMapContainer()
}

func (c *RotateMapContainer) Add(a UndoRedoAction) bool {
	t := reflect.TypeOf(a)
	if t != reflect.TypeOf((*RotateClockwiseAction)(nil)) &&
		t != reflect.TypeOf((*RotateCounterclockwiseAction)(nil)) {
		return false
	}

	c.mutex.Lock()
	c.actions = append(c.actions, a)
	c.mutex.Unlock()

	return true
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
	layerId   uuid.UUID
	name      string
	visible   bool
	layerType map_model.LayerType
}

func NewAddLayerAction(name string, visible bool, layerType map_model.LayerType) *AddLayerAction {
	return &AddLayerAction{layerId: uuid.New(), name: name, visible: visible, layerType: layerType}
}

func (a *AddLayerAction) Redo(m UndoRedoActionModels) {
	layerIndex := m.M.AddLayerWithId(a.layerId, a.layerType)
	m.M.SetName(layerIndex, a.name)
	m.M.SetVisible(layerIndex, a.visible)
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

func (a *MoveLayerAction) Redo(m UndoRedoActionModels) {
	a.oldIndex = m.M.LayerIndexById(a.layerId)
	if a.offset > 0 {
		m.M.MoveDown(a.oldIndex, int32(a.offset))
	} else {
		m.M.MoveUp(a.oldIndex, -int32(a.offset))
	}
}

func (a *MoveLayerAction) Undo(m UndoRedoActionModels) {
	diff := m.M.LayerIndexById(a.layerId) - a.oldIndex
	if diff > 0 {
		m.M.MoveDown(a.oldIndex, diff)
	} else {
		m.M.MoveUp(a.oldIndex, -diff)
	}
}

type ClearLayerAction struct {
	layerId   uuid.UUID
	locations map[utils.Int2]map_model.Location
}

func NewClearLayerAction(layerId uuid.UUID) *ClearLayerAction {
	return &ClearLayerAction{layerId: layerId}
}

func (a *ClearLayerAction) Redo(m UndoRedoActionModels) {
	layerIndex := m.M.LayerIndexById(a.layerId)
	a.locations = m.M.Locations(layerIndex)
	m.M.ClearLayer(layerIndex)
}

func (a *ClearLayerAction) Undo(m UndoRedoActionModels) {
	m.M.SetLocations(m.M.LayerIndexById(a.layerId), a.locations)
}

type MoveToSelectedAction struct {
	offset      utils.Int2
	moveLayerId uuid.UUID
}

func NewMoveToSelectedAction(moveLayerId uuid.UUID, offset utils.Int2) *MoveToSelectedAction {
	return &MoveToSelectedAction{offset: offset, moveLayerId: moveLayerId}
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

func (a *UnselectAllAction) Redo(m UndoRedoActionModels) {
	a.selected = m.Sm.Selected()
	m.Sm.UnselectAll()
}

func (a *UnselectAllAction) Undo(m UndoRedoActionModels) {
	m.Sm.SetSelected(a.selected)
}

type SetSelectedAction struct {
	selected    map[utils.Int2]select_model.Selected
	oldSelected map[utils.Int2]select_model.Selected
}

func NewSetSelectedAction(selected map[utils.Int2]select_model.Selected) *SetSelectedAction {
	return &SetSelectedAction{selected: selected}
}

func (a *SetSelectedAction) Redo(m UndoRedoActionModels) {
	a.oldSelected = m.Sm.Selected()
	m.Sm.SetSelected(a.selected)
}

func (a *SetSelectedAction) Undo(m UndoRedoActionModels) {
	m.Sm.SetSelected(a.oldSelected)
}

type MergeLayersAction struct {
	fromLayerId uuid.UUID
	toLayerId   uuid.UUID
	actions     *UndoRedoContainer
}

func NewMergeLayersAction(fromLayerId uuid.UUID, toLayerId uuid.UUID) *MergeLayersAction {
	return &MergeLayersAction{fromLayerId: fromLayerId, toLayerId: toLayerId, actions: NewUndoRedoContainer()}
}

func (a *MergeLayersAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		fromLayerIndex := m.M.LayerIndexById(a.fromLayerId)

		leftTop, rightBottom := m.Rm.Bounds(fromLayerIndex)

		for y := leftTop.Y; y < rightBottom.Y; y++ {
			for x := leftTop.X; x < rightBottom.X; x++ {
				if v := m.Rm.Floor(x, y, fromLayerIndex); v > 0 {
					action := NewSetFloorAction(utils.NewInt2(x, y), a.toLayerId, m.Rm.Floor(x, y, fromLayerIndex))
					action.Redo(m)
					a.actions.Add(action)
				}
				if v := m.Rm.Wall(x, y, fromLayerIndex, true); v > 0 {
					action := NewSetWallAction(utils.NewInt2(x, y), a.toLayerId, true, v)
					action.Redo(m)
					a.actions.Add(action)
				}
				if v := m.Rm.Wall(x, y, fromLayerIndex, false); v > 0 {
					action := NewSetWallAction(utils.NewInt2(x, y), a.toLayerId, false, v)
					action.Redo(m)
					a.actions.Add(action)
				}
			}
		}

		deleteLayerAction := NewDeleteLayerAction(a.fromLayerId)
		deleteLayerAction.Redo(m)
		a.actions.Add(deleteLayerAction)
	} else {
		a.actions.Redo(m)
	}
}

func (a *MergeLayersAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

// Делает MergeLayersAction на слой ниже
type MergeLayerDownAction struct {
	fromLayerId uuid.UUID
	actions     *UndoRedoContainer
}

func NewMergeLayerDownAction(fromLayerId uuid.UUID) *MergeLayerDownAction {
	return &MergeLayerDownAction{fromLayerId: fromLayerId, actions: NewUndoRedoContainer()}
}

func (a *MergeLayerDownAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		fromLayerIndex := m.M.LayerIndexById(a.fromLayerId)
		toLayerId := m.M.Layer(fromLayerIndex + 1).Uuid

		action := NewMergeLayersAction(a.fromLayerId, toLayerId)
		action.Redo(m)
		a.actions.Add(action)
	} else {
		a.actions.Redo(m)
	}
}

func (a *MergeLayerDownAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type SetModeAndMergeDownMoveLayerAction struct {
	mode    mode_model.Mode
	actions *UndoRedoContainer
}

func NewSetModeAndMergeDownMoveLayerAction(mode mode_model.Mode) *SetModeAndMergeDownMoveLayerAction {
	return &SetModeAndMergeDownMoveLayerAction{mode: mode, actions: NewUndoRedoContainer()}
}

func (a *SetModeAndMergeDownMoveLayerAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		moveLayerIndex := pie.FirstOr(m.M.LayerIndexByType(map_model.MoveLayerType), -1)

		if moveLayerIndex != -1 {
			moveLayerId := m.M.Layer(moveLayerIndex).Uuid

			leftTop, rightBottom := m.Rm.Bounds(moveLayerIndex)

			if leftTop != rightBottom {
				moveAction := NewMergeLayerDownAction(moveLayerId)
				moveAction.Redo(m)
				a.actions.Add(moveAction)

				setSelectedLayerAction := NewSetSelectedLayerAction(moveLayerIndex)
				setSelectedLayerAction.Redo(m)
				a.actions.Add(setSelectedLayerAction)
			}
		}

		if m.Mm.Mode() != a.mode {
			setModeAction := NewSetModeAction(a.mode)
			setModeAction.Redo(m)
			a.actions.Add(setModeAction)
		}
	} else {
		a.actions.Redo(m)
	}
}

func (a *SetModeAndMergeDownMoveLayerAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type SetModeAction struct {
	mode    mode_model.Mode
	oldMode mode_model.Mode
}

func NewSetModeAction(mode mode_model.Mode) *SetModeAction {
	return &SetModeAction{mode: mode}
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

type CutAction struct {
	copyResult copy_model.CopyResult
	actions    *UndoRedoContainer
}

func NewCutAction(copyResult copy_model.CopyResult) *CutAction {
	return &CutAction{copyResult: copyResult, actions: NewUndoRedoContainer()}
}

func (a *CutAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		unselectAction := NewUnselectAllAction()
		unselectAction.Redo(m)
		a.actions.Add(unselectAction)

		for pos, location := range a.copyResult.Locations {
			x, y := m.R.TransformToRot(pos.X, pos.Y)

			if location.Floor > 0 {
				action := NewSetFloorAction(utils.NewInt2(x, y), a.copyResult.LayerId, 0)
				action.Redo(m)
				a.actions.Add(action)
			}
			if location.RightWall > 0 {
				action := NewSetWallAction(utils.NewInt2(x, y), a.copyResult.LayerId, true, 0)
				action.Redo(m)
				a.actions.Add(action)
			}
			if location.BottomWall > 0 {
				action := NewSetWallAction(utils.NewInt2(x, y), a.copyResult.LayerId, false, 0)
				action.Redo(m)
				a.actions.Add(action)
			}
		}
	} else {
		a.actions.Redo(m)
	}
}

func (a *CutAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type PasteToMoveLayerAction struct {
	pos        utils.Int2
	copyResult copy_model.CopyResult
	actions    *UndoRedoContainer
}

func NewPasteToMoveLayerAction(pos utils.Int2, copyResult copy_model.CopyResult) *PasteToMoveLayerAction {
	return &PasteToMoveLayerAction{
		pos:        pos,
		copyResult: copyResult,
		actions:    NewUndoRedoContainer(),
	}
}

func (a *PasteToMoveLayerAction) Redo(m UndoRedoActionModels) {
	if a.actions.Len() == 0 {
		leftTop, rightBottom := a.copyResult.Bounds()
		if leftTop == rightBottom {
			return
		}

		unselectAction := NewUnselectAllAction()
		unselectAction.Redo(m)
		a.actions.Add(unselectAction)

		leftTop.X, leftTop.Y = m.R.TransformToRot(leftTop.X, leftTop.Y)

		leftTop.X -= a.pos.X
		leftTop.Y -= a.pos.Y

		selectedLayerBefore := m.Slm.Selected()
		selectedLayerIdBefore := m.M.Layer(selectedLayerBefore).Uuid

		addPastedLayerAction := NewAddLayerAction("Pasted layer", true, map_model.MoveLayerType)
		addPastedLayerAction.Redo(m)
		a.actions.Add(addPastedLayerAction)

		moveLayersIndices := m.M.LayerIndexByType(map_model.MoveLayerType)
		if len(moveLayersIndices) != 1 {
			panic("len(moveLayersIndices) != 1")
		}
		moveLayerIndex := moveLayersIndices[0]
		moveLayerId := m.M.LayerInfo(moveLayerIndex).Uuid

		selectedLayerAfter := m.M.LayerIndexById(selectedLayerIdBefore)

		moveUpAction := NewMoveLayerAction(int(selectedLayerAfter-moveLayerIndex), moveLayerId)
		moveUpAction.Redo(m)
		a.actions.Add(moveUpAction)

		setSelectedLayerAction := NewSetSelectedLayerAction(selectedLayerAfter)
		setSelectedLayerAction.Redo(m)
		a.actions.Add(setSelectedLayerAction)

		selected := make(map[utils.Int2]select_model.Selected)

		// если какие-то блоки были выделены, то переносим их в moveLayer
		for pos, location := range a.copyResult.Locations {
			pos.X, pos.Y = m.R.TransformToRot(pos.X, pos.Y)

			pos.X -= leftTop.X
			pos.Y -= leftTop.Y

			if location.Floor > 0 {
				action := NewSetFloorAction(pos, moveLayerId, location.Floor)
				action.Redo(m)
				a.actions.Add(action)

				s := selected[pos]
				s.Floor = true
				selected[pos] = s
			}
			if location.RightWall > 0 {
				action := NewSetWallAction(pos, moveLayerId, true, location.RightWall)
				action.Redo(m)
				a.actions.Add(action)

				s := selected[pos]
				s.RightWall = true
				selected[pos] = s
			}
			if location.BottomWall > 0 {
				action := NewSetWallAction(pos, moveLayerId, false, location.BottomWall)
				action.Redo(m)
				a.actions.Add(action)

				s := selected[pos]
				s.BottomWall = true
				selected[pos] = s
			}
		}

		action := NewSetSelectedAction(selected)
		action.Redo(m)
		a.actions.Add(action)

		setModeAction := NewSetModeAction(mode_model.MoveMode)
		setModeAction.Redo(m)
		a.actions.Add(setModeAction)
	} else {
		a.actions.Redo(m)
	}
}

func (a *PasteToMoveLayerAction) Undo(m UndoRedoActionModels) {
	a.actions.Undo(m)
}

type SetSelectedLayerAction struct {
	value    int32
	oldValue int32
}

func NewSetSelectedLayerAction(value int32) *SetSelectedLayerAction {
	return &SetSelectedLayerAction{value: value}
}

func (a *SetSelectedLayerAction) Redo(m UndoRedoActionModels) {
	a.oldValue = m.Slm.Selected()
	m.Slm.SetSelected(a.value)
}

func (a *SetSelectedLayerAction) Undo(m UndoRedoActionModels) {
	m.Slm.SetSelected(a.oldValue)
}

type SetCenterAction struct {
	pos    utils.Int2
	oldPos utils.Int2
}

func NewSetCenterAction(pos utils.Int2) *SetCenterAction {
	return &SetCenterAction{pos: pos}
}

func (a *SetCenterAction) Redo(m UndoRedoActionModels) {
	a.oldPos = m.Cm.Get()
	m.Cm.Set(a.pos)
}

func (a *SetCenterAction) Undo(m UndoRedoActionModels) {
	m.Cm.Set(a.oldPos)
}
