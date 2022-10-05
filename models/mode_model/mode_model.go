package mode_model

import (
	"old-school-rpg-map-editor/utils"
	"sync"
)

type Mode int

const (
	SetMode    Mode = 0
	SelectMode Mode = 1
	MoveMode   Mode = 2
)

type ModeModel struct {
	mutex     sync.Mutex
	mode      Mode
	listeners utils.Signal0
}

func NewModeModel() *ModeModel {
	return &ModeModel{listeners: utils.NewSignal0()}
}

func (m *ModeModel) Mode() Mode {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.mode
}

func (m *ModeModel) SetMode(value Mode) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if m.mode == value {
			return false
		}

		m.mode = value

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *ModeModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
