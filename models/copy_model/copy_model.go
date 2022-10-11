package copy_model

import (
	"math"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/utils"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

type CopyResultLocations struct {
	Locations map[utils.Int2]map_model.Location
}

type CopyResult struct {
	Layers map[uuid.UUID]CopyResultLocations
}

func (m *CopyResult) Bounds() (leftTop, rightBottom utils.Int2) {
	leftTop = utils.NewInt2(math.MaxInt32, math.MaxInt32)
	rightBottom = utils.NewInt2(math.MinInt32, math.MinInt32)

	for _, layer := range m.Layers {
		for k, v := range layer.Locations {
			if !v.IsEmptyLocation() {
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
	}

	// В случае, если или layer.Locations пустой, или в нём нет ни одного выделенного элемента
	if leftTop.X == math.MaxInt32 {
		leftTop = utils.Int2{}
		rightBottom = utils.Int2{}
	} else {
		rightBottom.X++
		rightBottom.Y++
	}

	return
}

func (r *CopyResult) Clone() CopyResult {
	result := CopyResult{}
	result.Layers = make(map[uuid.UUID]CopyResultLocations)

	for k, v := range r.Layers {
		result.Layers[k] = CopyResultLocations{Locations: maps.Clone(v.Locations)}
	}

	return result
}

type CopyModel struct {
	mutex      sync.Mutex
	copyResult *CopyResult

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewCopyModel() *CopyModel {
	return &CopyModel{}
}

func (m *CopyModel) SetCopyResult(copyResult CopyResult) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		t := copyResult.Clone()
		m.copyResult = &t

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *CopyModel) CopyResult() CopyResult {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.copyResult.Clone()
}

func (m *CopyModel) IsEmpty() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.copyResult == nil
}

func (m *CopyModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
