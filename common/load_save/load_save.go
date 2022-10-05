package load_save

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"old-school-rpg-map-editor/models/map_model"
	"old-school-rpg-map-editor/models/notes_model"
)

func LoadMapFile(reader io.ReadCloser) (*map_model.MapModel, *notes_model.NotesModel, error) {
	greader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, nil, err
	}

	d, err := io.ReadAll(greader)
	if err != nil {
		return nil, nil, err
	}

	var t struct {
		Version    int                     `json:"version"`
		MapModel   *map_model.MapModel     `json:"map"`
		NotesModel *notes_model.NotesModel `json:"notes"`
	}

	err = json.Unmarshal(d, &t)
	if err != nil {
		return nil, nil, err
	}

	if t.Version != 1 {
		return nil, nil, errors.New("unsupported version")
	}

	return t.MapModel, t.NotesModel, nil
}
