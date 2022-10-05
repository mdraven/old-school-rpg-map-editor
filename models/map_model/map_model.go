package map_model

import (
	"encoding/json"
	"math"
	"old-school-rpg-map-editor/utils"
	"sync"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Location struct {
	Floor      uint32 `json:"floor,omitempty"`
	RightWall  uint32 `json:"right_wall,omitempty"`
	BottomWall uint32 `json:"bottom_wall,omitempty"`
	NoteId     string `json:"note_id,omitempty"` // id заметки для этой клетки
}

func (l *Location) isEmptyLocation() bool {
	return l.Floor == 0 && l.RightWall == 0 && l.BottomWall == 0 && len(l.NoteId) == 0
}

type LayerInfo struct {
	Uuid    uuid.UUID `json:"uuid"`
	Name    string    `json:"name"`
	Visible bool      `json:"visible"`
	System  bool      `json:"system"` // слой создан для целей приложения(например для select)
}

type Layer struct {
	LayerInfo
	locations map[utils.Int2]Location
}

func newLayer(uuid uuid.UUID) *Layer {
	return &Layer{LayerInfo: LayerInfo{Uuid: uuid, Visible: true}, locations: make(map[utils.Int2]Location)}
}

func (l *Layer) MarshalJSON() ([]byte, error) {
	t := struct {
		Info      *LayerInfo              `json:"info"`
		Locations map[utils.Int2]Location `json:"locations"`
	}{Info: &l.LayerInfo, Locations: l.locations}

	return json.Marshal(t)
}

func (l *Layer) UnmarshalJSON(d []byte) error {
	var t struct {
		Info      *LayerInfo              `json:"info"`
		Locations map[utils.Int2]Location `json:"locations"`
	}

	err := json.Unmarshal(d, &t)
	if err != nil {
		return err
	}

	l.LayerInfo = *t.Info
	l.locations = t.Locations

	return nil
}

func (l *Layer) Clone() *Layer {
	return &Layer{
		LayerInfo: l.LayerInfo,
		locations: maps.Clone(l.locations),
	}
}

type MapModel struct {
	mutex  sync.Mutex
	layers []*Layer

	listeners utils.Signal0 // listener'ы на изменение списка

	beforeDeleteLayerListeners utils.Signal0
	afterDeleteLayerListeners  utils.Signal0

	beforeMoveLayerListeners utils.Signal0
	afterMoveLayerListeners  utils.Signal0
}

func NewMapModel() *MapModel {
	return &MapModel{}
}

func (m *MapModel) MarshalJSON() ([]byte, error) {
	t := struct {
		Layers []*Layer `json:"layers"`
	}{Layers: m.layers}

	return json.Marshal(t)
}

func (m *MapModel) UnmarshalJSON(d []byte) error {
	var t struct {
		Layers []*Layer `json:"layers"`
	}

	err := json.Unmarshal(d, &t)
	if err != nil {
		return err
	}

	m.layers = t.Layers

	return nil
}

func (m *MapModel) LayerIndexById(uuid uuid.UUID) int32 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	layerIndex := slices.IndexFunc(m.layers, func(l *Layer) bool {
		return l.Uuid == uuid
	})

	return int32(layerIndex)
}

func (m *MapModel) LayerIndexByName(name string, system bool) int32 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	layerIndex := slices.IndexFunc(m.layers, func(l *Layer) bool {
		return l.Name == name && l.System == system
	})

	return int32(layerIndex)
}

func (m *MapModel) AddLayerWithId(uuid uuid.UUID) (layerIndex int32) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.layers = append(m.layers, newLayer(uuid))

		layerIndex = int32(len(m.layers) - 1)

		return true
	}()

	if send {
		m.listeners.Emit()
	}

	return
}

func (m *MapModel) AddLayer(layer *Layer) (layerIndex int32) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.layers = append(m.layers, layer.Clone())

		layerIndex = int32(len(m.layers) - 1)

		return true
	}()

	if send {
		m.listeners.Emit()
	}

	return
}

func (m *MapModel) DeleteLayer(layerIndex int32) {
	m.beforeDeleteLayerListeners.Emit()

	func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.layers = slices.Delete(m.layers, int(layerIndex), int(layerIndex)+1)
	}()

	m.afterDeleteLayerListeners.Emit()
	m.listeners.Emit()
}

func (m *MapModel) ClearLayer(layerIndex int32) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.layers[layerIndex].locations = make(map[utils.Int2]Location)

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) checkMoveUp(layerIndex int32, offset int32) bool {
	if offset < 0 {
		panic("offset < 0")
	}

	offset = utils.Min(layerIndex, offset)

	if offset == 0 {
		return false
	}

	if len(m.layers) == 0 || layerIndex >= int32(len(m.layers)) {
		return false
	}

	return true
}

func (m *MapModel) moveUp(layerIndex int32, offset int32) bool {
	if !m.checkMoveUp(layerIndex, offset) {
		return false
	}

	offset = utils.Min(layerIndex, offset)

	layer := m.layers[layerIndex]

	for i := layerIndex; i != layerIndex-offset; i-- {
		m.layers[i] = m.layers[i-1]
	}

	m.layers[layerIndex-offset] = layer

	return true
}

func (m *MapModel) MoveUp(layerIndex int32, offset int32) {
	ok := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		return m.checkMoveUp(layerIndex, offset)
	}()

	if !ok {
		return
	}

	m.beforeMoveLayerListeners.Emit()

	func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.moveUp(layerIndex, offset)
	}()

	m.afterMoveLayerListeners.Emit()
	m.listeners.Emit()
}

func (m *MapModel) checkMoveDown(layerIndex int32, offset int32) bool {
	if offset < 0 {
		panic("offset < 0")
	}

	if len(m.layers) == 0 || layerIndex >= int32(len(m.layers)) {
		return false
	}

	offset = utils.Min(int32(len(m.layers)-int(layerIndex)-1), offset)

	return offset != 0
}

func (m *MapModel) moveDown(layerIndex int32, offset int32) bool {
	if !m.checkMoveDown(layerIndex, offset) {
		return false
	}

	offset = utils.Min(int32(len(m.layers)-int(layerIndex)-1), offset)

	layer := m.layers[layerIndex]

	for i := layerIndex; i != layerIndex+offset; i++ {
		m.layers[i] = m.layers[i+1]
	}

	m.layers[layerIndex+offset] = layer

	return true
}

func (m *MapModel) MoveDown(layerIndex int32, offset int32) {
	ok := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		return m.checkMoveDown(layerIndex, offset)
	}()

	if !ok {
		return
	}

	m.beforeMoveLayerListeners.Emit()

	func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.moveDown(layerIndex, offset)
	}()

	m.afterMoveLayerListeners.Emit()
	m.listeners.Emit()
}

func (m *MapModel) NumLayers() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return len(m.layers)
}

func (m *MapModel) LayerInfos() []LayerInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]LayerInfo, len(m.layers))

	for i, l := range m.layers {
		result[i] = l.LayerInfo
	}

	return result
}

func (m *MapModel) LayerInfo(layerIndex int32) LayerInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if layerIndex < 0 || int(layerIndex) >= len(m.layers) {
		return LayerInfo{}
	}

	return m.layers[layerIndex].LayerInfo
}

func (m *MapModel) Layer(layerIndex int32) *Layer {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.layers[layerIndex].Clone()
}

func (m *MapModel) Locations(layerIndex int32) map[utils.Int2]Location {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return maps.Clone(m.layers[layerIndex].locations)
}

func (m *MapModel) SetVisible(layerIndex int32, value bool) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if m.layers[layerIndex].Visible == value {
			return false
		}

		m.layers[layerIndex].Visible = value

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) SetSystem(layerIndex int32, value bool) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if m.layers[layerIndex].System == value {
			return false
		}

		m.layers[layerIndex].System = value

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) SetName(layerIndex int32, value string) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		if m.layers[layerIndex].Name == value {
			return false
		}

		m.layers[layerIndex].Name = value

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) SetLocations(layerIndex int32, value map[utils.Int2]Location) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.layers[layerIndex].locations = value

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) VisibleFloor(x, y int) (layerUuid uuid.UUID, floor uint32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.layers) == 0 {
		return uuid.UUID{}, 0
	}

	for _, l := range m.layers {
		if l.Visible {
			v, exists := l.locations[utils.NewInt2(x, y)]
			if exists && v.Floor > 0 {
				return l.Uuid, v.Floor
			}
		}
	}

	return m.layers[0].Uuid, 0
}

func (m *MapModel) floor(x, y int, layerIndex int32) uint32 {
	v, exists := m.layers[layerIndex].locations[utils.NewInt2(x, y)]
	if !exists {
		return 0
	}
	return v.Floor
}

func (m *MapModel) Floor(x, y int, layerIndex int32) uint32 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.floor(x, y, layerIndex)
}

func (m *MapModel) setFloor(x, y int, layerIndex int32, value uint32) {
	pos := utils.NewInt2(x, y)

	f, exists := m.layers[layerIndex].locations[pos]
	if exists || value > 0 {
		f.Floor = value
		m.layers[layerIndex].locations[pos] = f
	}

	if f.isEmptyLocation() {
		delete(m.layers[layerIndex].locations, pos)
	}
}

func (m *MapModel) SetFloor(x, y int, layerIndex int32, value uint32) {
	if layerIndex < 0 {
		return
	}

	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.setFloor(x, y, layerIndex, value)

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) VisibleWall(x, y int, isRight bool) (layerUuid uuid.UUID, wall uint32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.layers) == 0 {
		return uuid.UUID{}, 0
	}

	for _, l := range m.layers {
		if l.Visible {
			v, exists := l.locations[utils.NewInt2(x, y)]
			if isRight && exists && v.RightWall > 0 {
				return l.Uuid, v.RightWall
			}
			if !isRight && exists && v.BottomWall > 0 {
				return l.Uuid, v.BottomWall
			}
		}
	}

	return m.layers[0].Uuid, 0
}

func (m *MapModel) wall(x, y int, layerIndex int32, isRight bool) (wall uint32) {
	v, exists := m.layers[layerIndex].locations[utils.NewInt2(x, y)]
	if isRight && exists {
		return v.RightWall
	}
	if !isRight && exists {
		return v.BottomWall
	}
	return 0
}

func (m *MapModel) Wall(x, y int, layerIndex int32, isRight bool) (wall uint32) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.wall(x, y, layerIndex, isRight)
}

func (m *MapModel) setWall(x, y int, layerIndex int32, isRight bool, value uint32) {
	pos := utils.NewInt2(x, y)

	f, exists := m.layers[layerIndex].locations[pos]
	if exists || value > 0 {
		if isRight {
			f.RightWall = value
		} else {
			f.BottomWall = value
		}
		m.layers[layerIndex].locations[pos] = f
	}

	if f.isEmptyLocation() {
		delete(m.layers[layerIndex].locations, pos)
	}
}

func (m *MapModel) SetWall(x, y int, layerIndex int32, isRight bool, value uint32) {
	if layerIndex < 0 {
		return
	}

	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.setWall(x, y, layerIndex, isRight, value)

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) VisibleNoteId(x, y int) (layerUuid uuid.UUID, noteId string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.layers) == 0 {
		return uuid.UUID{}, ""
	}

	for _, l := range m.layers {
		if l.Visible {
			v, exists := l.locations[utils.NewInt2(x, y)]
			if exists && len(v.NoteId) > 0 {
				return l.Uuid, v.NoteId
			}
		}
	}

	return m.layers[0].Uuid, ""
}

func (m *MapModel) noteId(x, y int, layerIndex int32) string {
	v, exists := m.layers[layerIndex].locations[utils.NewInt2(x, y)]
	if !exists {
		return ""
	}
	return v.NoteId
}

func (m *MapModel) NoteId(x, y int, layerIndex int32) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.noteId(x, y, layerIndex)
}

func (m *MapModel) setNoteId(x, y int, layerIndex int32, value string) {
	pos := utils.NewInt2(x, y)

	f, exists := m.layers[layerIndex].locations[pos]
	if exists || len(value) > 0 {
		f.NoteId = value
		m.layers[layerIndex].locations[pos] = f
	}

	if f.isEmptyLocation() {
		delete(m.layers[layerIndex].locations, pos)
	}
}

func (m *MapModel) SetNoteId(x, y int, layerIndex int32, value string) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.setNoteId(x, y, layerIndex, value)

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) bounds(layerIndex int32) (leftTop, rightBottom utils.Int2) {
	leftTop = utils.NewInt2(math.MaxInt32, math.MaxInt32)
	rightBottom = utils.NewInt2(math.MinInt32, math.MinInt32)

	for k, v := range m.layers[layerIndex].locations {
		if !v.isEmptyLocation() {
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

	// В случае, если или m.layers[layerIndex].locations пустой, или в нём нет ни одного выделенного элемента
	if leftTop.X == math.MaxInt32 {
		leftTop = utils.Int2{}
		rightBottom = utils.Int2{}
	} else {
		rightBottom.X++
		rightBottom.Y++
	}

	return
}

func (m *MapModel) Bounds(layerIndex int32) (leftTop, rightBottom utils.Int2) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.bounds(layerIndex)
}

func (m *MapModel) MoveTo(layerIndex int32, offsetX, offsetY int) {
	send := func() bool {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		leftTop, rightBottom := m.bounds(layerIndex)
		if leftTop == rightBottom {
			return false
		}

		newLocations := make(map[utils.Int2]Location)

		for y := leftTop.Y; y < rightBottom.Y; y++ {
			for x := leftTop.X; x < rightBottom.X; x++ {
				pos := utils.NewInt2(x, y)
				offsetPos := utils.NewInt2(x+offsetX, y+offsetY)
				newLocations[offsetPos] = m.layers[layerIndex].locations[pos]
			}
		}

		m.layers[layerIndex].locations = newLocations

		return true
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *MapModel) HasVisible() bool {
	for _, l := range m.LayerInfos() {
		if !l.System && l.Visible {
			return true
		}
	}

	return false
}

func (m *MapModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}

func (m *MapModel) AddBeforeDeleteLayerListener(listener func()) func() {
	return m.beforeDeleteLayerListeners.AddSlot(listener)
}

func (m *MapModel) AddAfterDeleteLayerListener(listener func()) func() {
	return m.afterDeleteLayerListeners.AddSlot(listener)
}

func (m *MapModel) AddBeforeMoveLayerListener(listener func()) func() {
	return m.beforeMoveLayerListeners.AddSlot(listener)
}

func (m *MapModel) AddAfterMoveLayerListener(listener func()) func() {
	return m.afterMoveLayerListeners.AddSlot(listener)
}

func LayerIndexWithoutSystemToWithSystem(model *MapModel, layerIndex int32) int32 {
	if layerIndex < 0 {
		panic("layerIndex < 0")
	}

	model.mutex.Lock()
	defer model.mutex.Unlock()

	for i := 0; i < len(model.layers); i++ {
		layerInfo := model.layers[i].LayerInfo
		if !layerInfo.System {
			if layerIndex == 0 {
				return int32(i)
			}
			layerIndex--
		}
	}

	return -1
}

func LayerIndexWithSystemToWithoutSystem(model *MapModel, layerIndex int32) int32 {
	if layerIndex < 0 {
		panic("layerIndex < 0")
	}

	model.mutex.Lock()
	defer model.mutex.Unlock()

	if int(layerIndex) >= len(model.layers) {
		return -1
	}

	if model.layers[int(layerIndex)].LayerInfo.System {
		return -1
	}

	counter := int32(0)

	for i := 0; i <= int(layerIndex); i++ {
		layerInfo := model.layers[i].LayerInfo
		if !layerInfo.System {
			counter++
		}
	}

	return counter - 1
}

func LengthWithoutSystem(model *MapModel) int {
	model.mutex.Lock()
	defer model.mutex.Unlock()

	counter := 0
	for i := 0; i < len(model.layers); i++ {
		layerInfo := model.layers[i].LayerInfo
		if !layerInfo.System {
			counter++
		}
	}

	return counter
}

func MoveUpWithoutSystem(model *MapModel, layerIndex int32, offset int32) {
	ok, from := func() (bool, int32) {
		model.mutex.Lock()
		defer model.mutex.Unlock()

		from := layerIndex - 1
		for ; from >= 0; from-- {
			layerInfo := model.layers[from].LayerInfo
			if !layerInfo.System {
				offset--
				if offset == 0 {
					break
				}
			}
		}

		return model.checkMoveUp(layerIndex, layerIndex-from), from
	}()

	if !ok {
		return
	}

	model.beforeMoveLayerListeners.Emit()

	func() {
		model.mutex.Lock()
		defer model.mutex.Unlock()

		model.moveUp(layerIndex, layerIndex-from)
	}()

	model.afterMoveLayerListeners.Emit()
	model.listeners.Emit()
}

func MoveDownWithoutSystem(model *MapModel, layerIndex int32, offset int32) {
	ok, from := func() (bool, int32) {
		model.mutex.Lock()
		defer model.mutex.Unlock()

		from := layerIndex + 1
		for ; from < int32(len(model.layers)); from++ {
			layerInfo := model.layers[from].LayerInfo
			if !layerInfo.System {
				offset--
				if offset == 0 {
					break
				}
			}
		}

		return model.checkMoveDown(layerIndex, from-layerIndex), from
	}()

	if !ok {
		return
	}

	model.beforeMoveLayerListeners.Emit()

	func() {
		model.mutex.Lock()
		defer model.mutex.Unlock()

		model.moveDown(layerIndex, from-layerIndex)
	}()

	model.afterMoveLayerListeners.Emit()
	model.listeners.Emit()
}
