package rot_select_model

import (
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/models/select_model"
	"old-school-rpg-map-editor/utils"
	"sync"
)

type RotSelectModel struct {
	mutex  sync.Mutex
	model  *select_model.SelectModel
	rotate *rotate_model.RotateModel

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewRotSelectModel(model *select_model.SelectModel, rotate *rotate_model.RotateModel) *RotSelectModel {
	m := &RotSelectModel{model: model, rotate: rotate, listeners: utils.NewSignal0()}
	model.AddDataChangeListener(m.listeners.Emit)
	rotate.AddDataChangeListener(m.listeners.Emit)
	return m
}

func (m *RotSelectModel) Model() *select_model.SelectModel {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.model
}

func (m *RotSelectModel) SelectFloor(x int, y int) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	model.SelectFloor(x, y)
}

func (m *RotSelectModel) SelectWall(x int, y int, isRight bool) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	x, y, isRight = rotate.TranslateWallToRot(x, y, isRight)

	model.SelectWall(x, y, isRight)
}

func (m *RotSelectModel) UnselectFloor(x int, y int) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	model.UnselectFloor(x, y)
}

func (m *RotSelectModel) UnselectWall(x int, y int, isRight bool) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	x, y, isRight = rotate.TranslateWallToRot(x, y, isRight)
	model.UnselectWall(x, y, isRight)
}

func (m *RotSelectModel) IsFloorSelected(x int, y int) bool {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	return model.IsFloorSelected(x, y)
}

func (m *RotSelectModel) IsWallSelected(x int, y int, isRight bool) bool {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)
	x, y, isRight = rotate.TranslateWallToRot(x, y, isRight)
	return model.IsWallSelected(x, y, isRight)
}

func (m *RotSelectModel) Bounds() (leftTop, rightBottom utils.Int2) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	leftTop, rightBottom = model.Bounds()
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

func (m *RotSelectModel) At(x, y int) select_model.Selected {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	x, y = rotate.TransformToRot(x, y)

	var selected select_model.Selected
	selected.Floor = model.At(x, y).Floor
	{
		x, y, isRight := rotate.TranslateWallToRot(x, y, true)
		at := model.At(x, y)
		if isRight {
			selected.RightWall = at.RightWall
		} else {
			selected.RightWall = at.BottomWall
		}
	}
	{
		x, y, isRight := rotate.TranslateWallToRot(x, y, false)
		at := model.At(x, y)
		if isRight {
			selected.BottomWall = at.RightWall
		} else {
			selected.BottomWall = at.BottomWall
		}
	}

	return selected
}

func (m *RotSelectModel) MoveTo(offsetX, offsetY int) {
	m.mutex.Lock()
	model := m.model
	rotate := m.rotate
	m.mutex.Unlock()

	offsetX, offsetY = rotate.TransformToRot(offsetX, offsetY)

	model.MoveTo(offsetX, offsetY)
}

func (m *RotSelectModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
