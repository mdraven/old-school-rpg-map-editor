package center_model

import (
	"old-school-rpg-map-editor/utils"
	"sync"
)

type CenterModel struct {
	mutex sync.Mutex
	pos   utils.Int2

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewCenterModel(pos utils.Int2) *CenterModel {
	m := &CenterModel{pos: pos, listeners: utils.NewSignal0()}
	return m
}

func (m *CenterModel) Get() utils.Int2 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.pos
}

func (m *CenterModel) Set(pos utils.Int2) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if m.pos == pos {
			return false
		}

		m.pos = pos

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *CenterModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
