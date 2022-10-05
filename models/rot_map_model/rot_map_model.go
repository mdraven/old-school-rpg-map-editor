package rot_map_model

import (
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/utils"
	"sync"

	"github.com/google/uuid"
)

type RotMapModel struct {
	mutex  sync.Mutex
	model  *map_model.MapModel
	rotate *rotate_model.RotateModel

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewRotMapMode(model *map_model.MapModel, rotate *rotate_model.RotateModel) *RotMapModel {
	m := &RotMapModel{model: model, rotate: rotate, listeners: utils.NewSignal0()}
	model.AddDataChangeListener(m.listeners.Emit)
	rotate.AddDataChangeListener(m.listeners.Emit)
	return m
}

func (m *RotMapModel) Model() *map_model.MapModel {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.model
}

func (m *RotMapModel) VisibleFloor(x, y int) (layerUuid uuid.UUID, floor uint32) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	return model.VisibleFloor(rotate.TransformToRot(x, y))
}

func (m *RotMapModel) Floor(x, y int, layerIndex int32) uint32 {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	return model.Floor(x, y, layerIndex)
}

func (m *RotMapModel) SetFloor(x, y int, layerIndex int32, value uint32) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	model.SetFloor(x, y, layerIndex, value)
}

func (m *RotMapModel) VisibleWall(x, y int, isRight bool) (layerUuid uuid.UUID, wall uint32) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	x, y, isRight = rotate.TranslateWallToRot(x, y, isRight)

	return model.VisibleWall(x, y, isRight)
}

func (m *RotMapModel) Wall(x, y int, layerIndex int32, isRight bool) (wall uint32) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	x, y, isRight = rotate.TranslateWallToRot(x, y, isRight)

	return model.Wall(x, y, layerIndex, isRight)
}

func (m *RotMapModel) SetWall(x, y int, layerIndex int32, isRight bool, value uint32) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	x, y, isRight = rotate.TranslateWallToRot(x, y, isRight)

	model.SetWall(x, y, layerIndex, isRight, value)
}

func (m *RotMapModel) VisibleNoteId(x, y int) (layerUuid uuid.UUID, noteId string) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	return model.VisibleNoteId(rotate.TransformToRot(x, y))
}

func (m *RotMapModel) NoteId(x, y int, layerIndex int32) string {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	return model.NoteId(x, y, layerIndex)
}

func (m *RotMapModel) SetNoteId(x, y int, layerIndex int32, value string) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	model.SetNoteId(x, y, layerIndex, value)
}

func (m *RotMapModel) Bounds(layerIndex int32) (leftTop, rightBottom utils.Int2) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	leftTop, rightBottom = model.Bounds(layerIndex)
	if leftTop == (utils.Int2{}) && rightBottom == (utils.Int2{}) {
		return leftTop, rightBottom
	}

	leftTopX, leftTopY := rotate.TransformFromRot(leftTop.X, leftTop.Y)
	rightBottomX, rightBottomY := rotate.TransformFromRot(rightBottom.X, rightBottom.Y)
	rightBottomX--
	rightBottomY--

	leftTop.X = utils.Min(leftTopX, rightBottomX)
	leftTop.Y = utils.Min(leftTopY, rightBottomY)
	rightBottom.X = utils.Max(leftTopX, rightBottomX)
	rightBottom.Y = utils.Max(leftTopY, rightBottomY)
	rightBottom.X++
	rightBottom.Y++

	return leftTop, rightBottom
}

func (m *RotMapModel) MoveTo(layerIndex int32, offsetX, offsetY int) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	offsetX, offsetY = rotate.TransformToRot(offsetX, offsetY)

	model.MoveTo(layerIndex, offsetX, offsetY)
}

func (m *RotMapModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
