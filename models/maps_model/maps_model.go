package maps_model

import (
	"encoding/json"
	"old-school-rpg-map-editor/common/load_save"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/mode_model"
	"old-school-rpg-map-editor/models/notes_model"
	"old-school-rpg-map-editor/models/rot_map_model"
	"old-school-rpg-map-editor/models/rot_select_model"
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/models/select_model"
	"old-school-rpg-map-editor/models/selected_layer_model"
	"old-school-rpg-map-editor/undo_redo"
	"old-school-rpg-map-editor/utils"
	"os"
	"sync"

	"github.com/goki/freetype/truetype"
	"github.com/google/uuid"
)

type MapElem struct {
	Model                 *map_model.MapModel
	SelectModel           *select_model.SelectModel
	ModeModel             *mode_model.ModeModel
	RotateModel           *rotate_model.RotateModel
	RotMapModel           *rot_map_model.RotMapModel
	RotSelectModel        *rot_select_model.RotSelectModel
	NotesModel            *notes_model.NotesModel
	UndoRedoQueue         *undo_redo.UndoRedoQueue
	SelectedLayerModel    *selected_layer_model.SelectedLayerModel
	MapId                 uuid.UUID
	FilePath              string // если пустой, то файла нет
	ChangeGeneration      uint64 // используется чтобы определять был изменён файл или нет
	SavedChangeGeneration uint64
	ExternalData          any
}

type MapsModel struct {
	mutex     sync.Mutex
	fontSize  float64
	fnt       *truetype.Font
	maps      map[uuid.UUID]MapElem
	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewMapsModel(fontSize float64, fnt *truetype.Font) *MapsModel {
	return &MapsModel{
		fontSize:  fontSize,
		fnt:       fnt,
		maps:      make(map[uuid.UUID]MapElem),
		listeners: utils.NewSignal0(),
	}
}

func (m *MapsModel) MarshalJSON() ([]byte, error) {
	t := struct {
		OpenFiles []string `json:"open_files"`
	}{OpenFiles: nil}

	for _, e := range m.maps {
		t.OpenFiles = append(t.OpenFiles, e.FilePath)
	}

	return json.Marshal(t)
}

func (m *MapsModel) UnmarshalJSON(d []byte) error {
	var t struct {
		OpenFiles []string `json:"open_files"`
	}

	err := json.Unmarshal(d, &t)
	if err != nil {
		return err
	}

	for _, file := range t.OpenFiles {
		f, err := os.Open(file)
		if err != nil {
			continue
		}

		mapModel, notesModel, err := load_save.LoadMapFile(f)
		if err != nil {
			return err
		}

		notesModel.SetFont(m.fontSize, m.fnt)

		selectedLayerModel := selected_layer_model.NewSelectedLayerModel()
		rotateModel := rotate_model.NewRotateModel(0)
		selectModel := select_model.NewSelectModel(mapModel, selectedLayerModel)
		rotMapModel := rot_map_model.NewRotMapMode(mapModel, rotateModel)
		rotSelectModel := rot_select_model.NewRotSelectModel(selectModel, rotateModel)
		undoRedoQueue := undo_redo.NewUndoRedoQueue(100 /*TODO*/)

		m.Add(mapModel, selectModel, mode_model.NewModeModel(), rotateModel, rotMapModel, rotSelectModel, notesModel, undoRedoQueue, selectedLayerModel, file)
	}

	return nil
}

func (m *MapsModel) Add(model *map_model.MapModel, selectModel *select_model.SelectModel, modeModel *mode_model.ModeModel, rotateModel *rotate_model.RotateModel, rotMapModel *rot_map_model.RotMapModel, rotSelectModel *rot_select_model.RotSelectModel, notesModel *notes_model.NotesModel, undoRedoQueue *undo_redo.UndoRedoQueue, selectedLayerModel *selected_layer_model.SelectedLayerModel, filePath string) (mapId uuid.UUID) {
	listeners := func() utils.Signal0 {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		mapId = uuid.New()

		m.maps[mapId] = MapElem{
			Model:              model,
			SelectModel:        selectModel,
			ModeModel:          modeModel,
			RotateModel:        rotateModel,
			RotMapModel:        rotMapModel,
			RotSelectModel:     rotSelectModel,
			NotesModel:         notesModel,
			UndoRedoQueue:      undoRedoQueue,
			SelectedLayerModel: selectedLayerModel,
			MapId:              mapId,
			FilePath:           filePath,
		}

		return m.listeners.Clone()
	}()

	listeners.Emit()

	return
}

func (m *MapsModel) Delete(mapId uuid.UUID) {
	listeners := func() utils.Signal0 {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		delete(m.maps, mapId)

		return m.listeners.Clone()
	}()

	listeners.Emit()
}

func (m *MapsModel) GetById(mapId uuid.UUID) MapElem {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.maps[mapId]
}

func (m *MapsModel) GetByExternalData(externalData any) MapElem {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, el := range m.maps {
		if externalData == el.ExternalData {
			return el
		}
	}

	return MapElem{}
}

type IdAndExternalData struct {
	MapId        uuid.UUID
	ExternalData any
}

func (m *MapsModel) GetIdAndExternalData() []IdAndExternalData {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	res := make([]IdAndExternalData, 0, len(m.maps))

	for id, elem := range m.maps {
		res = append(res, IdAndExternalData{MapId: id, ExternalData: elem.ExternalData})
	}

	return res
}

func (m *MapsModel) SetExternalData(mapId uuid.UUID, externalData any) {
	listeners := func() utils.Signal0 {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		mp, exists := m.maps[mapId]
		if exists {
			mp.ExternalData = externalData
			m.maps[mapId] = mp
			return m.listeners.Clone()
		}

		return utils.Signal0{}
	}()

	listeners.Emit()
}

func (m *MapsModel) SetFilePath(mapId uuid.UUID, filePath string) {
	listeners := func() utils.Signal0 {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		mp, exists := m.maps[mapId]
		if exists {
			mp.FilePath = filePath
			m.maps[mapId] = mp
			return m.listeners.Clone()
		}

		return utils.Signal0{}
	}()

	listeners.Emit()
}

func (m *MapsModel) SetChangeGeneration(mapId uuid.UUID, changeGeneration uint64) {
	listeners := func() utils.Signal0 {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		mp, exists := m.maps[mapId]
		if exists {
			mp.ChangeGeneration = changeGeneration
			m.maps[mapId] = mp
			return m.listeners.Clone()
		}

		return utils.Signal0{}
	}()

	listeners.Emit()
}

func (m *MapsModel) UpdateSaveChangeGeneration(mapId uuid.UUID) {
	listeners := func() utils.Signal0 {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		mp, exists := m.maps[mapId]
		if exists {
			mp.SavedChangeGeneration = mp.ChangeGeneration
			m.maps[mapId] = mp
			return m.listeners.Clone()
		}

		return utils.Signal0{}
	}()

	listeners.Emit()
}

func (m *MapsModel) Length() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return len(m.maps)
}

func (m *MapsModel) AddDataChangeListener(listener func()) func() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.listeners.AddSlot(listener)
}
