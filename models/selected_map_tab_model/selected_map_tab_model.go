package selected_map_tab_model

import (
	"old-school-rpg-map-editor/utils"
	"sync"

	"github.com/google/uuid"
)

type SelectedMapTabModel struct {
	mutex    sync.Mutex
	selected uuid.UUID

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewSelectedLayerModel() *SelectedMapTabModel {
	return &SelectedMapTabModel{}
}

func (m *SelectedMapTabModel) SetSelected(value uuid.UUID) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if m.selected == value {
			return false
		}

		m.selected = value

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *SelectedMapTabModel) Selected() uuid.UUID {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.selected
}

func (m *SelectedMapTabModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
