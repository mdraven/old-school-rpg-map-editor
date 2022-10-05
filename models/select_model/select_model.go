package select_model

import (
	"math"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/utils"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type Selected struct {
	Floor      uuid.UUID
	RightWall  uuid.UUID
	BottomWall uuid.UUID
}

type SelectModel struct {
	mutex    sync.Mutex
	selected map[utils.Int2]Selected
	mapModel *map_model.MapModel

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewSelectModel(mapModel *map_model.MapModel) *SelectModel {
	return &SelectModel{selected: make(map[utils.Int2]Selected), mapModel: mapModel}
}

func (m *SelectModel) Clear() {
	func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.selected = make(map[utils.Int2]Selected)
	}()

	m.listeners.Emit()
}

func (m *SelectModel) selectFloor(x int, y int) bool {
	layerUuid, v := m.mapModel.VisibleFloor(x, y)
	if v > 0 {
		pos := utils.NewInt2(x, y)

		selected := m.selected[pos]
		selected.Floor = layerUuid
		m.selected[pos] = selected

		return true
	}

	return false
}

func (m *SelectModel) SelectFloor(x int, y int) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		return m.selectFloor(x, y)
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *SelectModel) selectWall(x int, y int, isRight bool) bool {
	pos := utils.NewInt2(x, y)

	layerUuid, v := m.mapModel.VisibleWall(x, y, isRight)
	if isRight {
		if v > 0 {
			selected := m.selected[pos]
			selected.RightWall = layerUuid
			m.selected[pos] = selected

			return true
		}
	} else {
		if v > 0 {
			selected := m.selected[pos]
			selected.BottomWall = layerUuid
			m.selected[pos] = selected

			return true
		}
	}

	return false
}

func (m *SelectModel) SelectWall(x int, y int, isRight bool) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		return m.selectWall(x, y, isRight)
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *SelectModel) UnselectAll() {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		l := len(m.selected)
		if l > 0 {
			maps.Clear(m.selected)
			return true
		} else {
			return false
		}
	}()

	if send {
		m.listeners.Emit()
	}
}

func isEmptySelected(s Selected) bool {
	return s.Floor == uuid.UUID{} && s.RightWall == uuid.UUID{} && s.BottomWall == uuid.UUID{}
}

func (m *SelectModel) unselectFloor(x int, y int) bool {
	pos := utils.NewInt2(x, y)

	v, exists := m.selected[pos]
	if exists {
		v.Floor = uuid.UUID{}
		m.selected[pos] = v

		if isEmptySelected(v) {
			delete(m.selected, pos)
		}

		return true
	}

	return false
}

func (m *SelectModel) UnselectFloor(x int, y int) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		return m.unselectFloor(x, y)
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *SelectModel) unselectWall(x int, y int, isRight bool) bool {
	pos := utils.NewInt2(x, y)

	v, exists := m.selected[pos]
	if exists {
		if isRight {
			v.RightWall = uuid.UUID{}
		} else {
			v.BottomWall = uuid.UUID{}
		}
		m.selected[pos] = v

		if isEmptySelected(v) {
			delete(m.selected, pos)
		}

		return true
	}

	return false
}

func (m *SelectModel) UnselectWall(x int, y int, isRight bool) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		return m.unselectWall(x, y, isRight)
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *SelectModel) IsFloorSelected(x int, y int) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pos := utils.NewInt2(x, y)

	v, exists := m.selected[pos]
	return exists && v.Floor != uuid.UUID{}
}

func (m *SelectModel) IsWallSelected(x int, y int, isRight bool) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	pos := utils.NewInt2(x, y)

	v, exists := m.selected[pos]
	if isRight {
		return exists && v.RightWall != uuid.UUID{}
	} else {
		return exists && v.BottomWall != uuid.UUID{}
	}
}

func (m *SelectModel) bounds() (leftTop, rightBottom utils.Int2) {
	leftTop = utils.NewInt2(math.MaxInt32, math.MaxInt32)
	rightBottom = utils.NewInt2(math.MinInt32, math.MinInt32)

	for k, v := range m.selected {
		if !isEmptySelected(v) {
			if k.X < leftTop.X {
				leftTop.X = k.X
			}
			if k.Y < leftTop.Y {
				leftTop.Y = k.Y
			}
			if k.X > rightBottom.X {
				rightBottom.X = k.X
			}
			if k.Y > rightBottom.Y {
				rightBottom.Y = k.Y
			}
		}
	}

	// В случае, если или m.selected пустой, или в нём нет ни одного выделенного элемента
	if leftTop.X == math.MaxInt32 {
		leftTop = utils.Int2{}
		rightBottom = utils.Int2{}
	} else {
		rightBottom.X++
		rightBottom.Y++
	}

	return
}

func (m *SelectModel) Bounds() (leftTop, rightBottom utils.Int2) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.bounds()
}

func (m *SelectModel) at(x, y int) Selected {
	pos := utils.NewInt2(x, y)

	v, exists := m.selected[pos]
	if exists {
		return v
	}

	return Selected{}
}

func (m *SelectModel) At(x, y int) Selected {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.at(x, y)
}

func (m *SelectModel) moveTo(offsetX, offsetY int) bool {
	leftTop, rightBottom := m.bounds()
	if leftTop == rightBottom {
		return false
	}

	newSelected := make(map[utils.Int2]Selected)

	for y := leftTop.Y; y < rightBottom.Y; y++ {
		for x := leftTop.X; x < rightBottom.X; x++ {
			offsetPos := utils.NewInt2(x+offsetX, y+offsetY)
			newSelected[offsetPos] = m.at(x, y)
		}
	}

	m.selected = newSelected

	return true
}

func (m *SelectModel) MoveTo(offsetX, offsetY int) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		return m.moveTo(offsetX, offsetY)
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *SelectModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}

func (m *SelectModel) Selected() map[utils.Int2]Selected {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return maps.Clone(m.selected)
}

func (m *SelectModel) SetSelected(selected map[utils.Int2]Selected) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.selected = maps.Clone(selected)

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}
