package selected_layer_model

import (
	"old-school-rpg-map-editor/utils"
	"sync"
)

type SelectedLayerModel struct {
	mutex    sync.Mutex
	selected int32

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewSelectedLayerModel() *SelectedLayerModel {
	return &SelectedLayerModel{}
}

func (m *SelectedLayerModel) SetSelected(value int32) {
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

func (m *SelectedLayerModel) Selected() int32 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.selected
}

func (m *SelectedLayerModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
