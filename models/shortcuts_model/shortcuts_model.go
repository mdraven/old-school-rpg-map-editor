package shortcuts_model

import (
	"encoding/json"
	"old-school-rpg-map-editor/utils"
	"strconv"
	"strings"
	"sync"

	"github.com/elliotchance/pie/v2"
	"golang.org/x/exp/slices"
)

type ShortcutType string

const (
	ScrollMapLeft  ShortcutType = "Scroll the map west"
	ScrollMapRight ShortcutType = "Scroll the map east"
	ScrollMapUp    ShortcutType = "Scroll the map north"
	ScrollMapDown  ShortcutType = "Scroll the map south"

	RotateMapClockwise        ShortcutType = "Rotate the map clockwise"
	RotateMapCounterClockwise ShortcutType = "Rotate the map counterclockwise"
)

var modifiers = map[string]string{
	"LeftControl":  "Control",
	"RightControl": "Control",
	"LeftAlt":      "Alt",
	"RightAlt":     "Alt",
	"LeftShift":    "Shift",
	"RightShift":   "Shift",
	"LeftSuper":    "Super",
	"RightSuper":   "Super",
}

type Shortcut []string

func NewShortcut(keys []string) Shortcut {
	result := pie.Unique(keys)

	for ind, from := range result {
		if to, has := modifiers[from]; has {
			result[ind] = to
		}
	}

	return result
}

func ShortcutFromMap(keys map[string]struct{}) Shortcut {
	return NewShortcut(pie.Keys(keys))
}

func ShortcutFromString(keys string) Shortcut {
	withQuotes := strings.Split(keys, "+")
	for i := range withQuotes {
		s, err := strconv.Unquote(withQuotes[i])
		if err == nil {
			withQuotes[i] = s
		}
	}
	return NewShortcut(withQuotes)
}

type ShortcutsModel struct {
	mutex sync.Mutex

	shortcuts map[ShortcutType]Shortcut

	beforeChangeListeners utils.Signal0
	afterChangeListeners  utils.Signal0
}

func NewShortcutsModel() *ShortcutsModel {
	return &ShortcutsModel{
		shortcuts: map[ShortcutType]Shortcut{
			ScrollMapLeft:  {"A"},
			ScrollMapRight: {"D"},
			ScrollMapUp:    {"W"},
			ScrollMapDown:  {"S"},

			RotateMapClockwise:        {"E"},
			RotateMapCounterClockwise: {"Q"},
		},
	}
}

func (m *ShortcutsModel) MarshalJSON() ([]byte, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	t := struct {
		Shortcuts map[ShortcutType]string `json:"shortcuts"`
	}{Shortcuts: make(map[ShortcutType]string)}

	for k, v := range m.shortcuts {
		var withQuotes []string
		for _, s := range v {
			withQuotes = append(withQuotes, strconv.Quote(s))
		}

		t.Shortcuts[k] = strings.Join(withQuotes, "+")
	}

	return json.Marshal(t)
}

func (m *ShortcutsModel) UnmarshalJSON(d []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var t struct {
		Shortcuts map[ShortcutType]string `json:"shortcuts"`
	}

	err := json.Unmarshal(d, &t)
	if err != nil {
		return err
	}

	for k, v := range t.Shortcuts {
		m.shortcuts[k] = ShortcutFromString(v)
	}

	return nil
}

func (m *ShortcutsModel) Set(t ShortcutType, sc Shortcut) {
	m.beforeChangeListeners.Emit()

	m.mutex.Lock()
	m.shortcuts[t] = sc
	m.mutex.Unlock()

	m.afterChangeListeners.Emit()
}

func (m *ShortcutsModel) Get(sc Shortcut) []ShortcutType {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var result []ShortcutType

loop:
	for ty, sh := range m.shortcuts {
		for _, k := range sh {
			if !slices.Contains(sc, k) {
				continue loop
			}
		}
		result = append(result, ty)
	}

	return result
}

func (m *ShortcutsModel) AddBeforeChangeListener(listener func()) func() {
	return m.beforeChangeListeners.AddSlot(listener)
}

func (m *ShortcutsModel) AddAfterChangeListener(listener func()) func() {
	return m.afterChangeListeners.AddSlot(listener)
}
