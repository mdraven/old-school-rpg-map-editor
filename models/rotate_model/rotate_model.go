package rotate_model

import (
	"old-school-rpg-map-editor/utils"
	"sync"
)

type RotateModel struct {
	mutex sync.Mutex
	angle int

	listeners utils.Signal0 // listener'ы на изменение списка

	beforeRotateListeners utils.Signal0
	afterRotateListeners  utils.Signal0
}

func NewRotateModel(angle int) *RotateModel {
	m := &RotateModel{angle: angle, listeners: utils.NewSignal0(), beforeRotateListeners: utils.NewSignal0(), afterRotateListeners: utils.NewSignal0()}
	return m
}

func (m *RotateModel) transformToRot(x, y int) (tX, tY int) {
	if m.angle == 0 {
		return x, y
	} else if m.angle == 90 {
		return y, -x
	} else if m.angle == 180 {
		return -x, -y
	} else if m.angle == 270 {
		return -y, x
	}

	panic("incorrect angle")
}

func (m *RotateModel) translateWallToRot(x, y int, isRight bool) (tX, tY int, tIsRight bool) {
	if m.angle == 90 {
		if isRight {
			isRight = false
			y--
		} else {
			isRight = true
		}
	} else if m.angle == 180 {
		if isRight {
			x--
		} else {
			y--
		}
	} else if m.angle == 270 {
		if isRight {
			isRight = false
		} else {
			isRight = true
			x--
		}
	}

	return x, y, isRight
}

func (m *RotateModel) transformFromRot(x, y int) (tX, tY int) {
	if m.angle == 0 {
		return x, y
	} else if m.angle == 90 {
		return -y, x
	} else if m.angle == 180 {
		return -x, -y
	} else if m.angle == 270 {
		return y, -x
	}

	panic("incorrect angle")
}

func (m *RotateModel) TransformToRot(x, y int) (tX, tY int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.transformToRot(x, y)
}

func (m *RotateModel) TranslateWallToRot(x, y int, isRight bool) (tX, tY int, tIsRight bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.translateWallToRot(x, y, isRight)
}

func (m *RotateModel) TransformFromRot(x, y int) (tX, tY int) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.transformFromRot(x, y)
}

func (m *RotateModel) RotateClockwise() {
	m.beforeRotateListeners.Emit()

	func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.angle = (m.angle + 90) % 360

		return true
	}()

	m.listeners.Emit()
	m.afterRotateListeners.Emit()
}

func (m *RotateModel) RotateCounterclockwise() {
	m.beforeRotateListeners.Emit()

	func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.angle = (360 + m.angle - 90) % 360

		return true
	}()

	m.listeners.Emit()
	m.afterRotateListeners.Emit()
}

func (m *RotateModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}

func (m *RotateModel) AddBeforeRotateListener(listener func()) func() {
	return m.beforeRotateListeners.AddSlot(listener)
}

func (m *RotateModel) AddAfterRotateListener(listener func()) func() {
	return m.afterRotateListeners.AddSlot(listener)
}
