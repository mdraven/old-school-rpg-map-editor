package notes_model

import (
	"encoding/json"
	"image"
	"log"
	"old-school-rpg-map-editor/utils"
	"regexp"
	"strings"
	"sync"

	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"
	"golang.org/x/exp/slices"
	"golang.org/x/image/font"
)

var re = regexp.MustCompile(`(?m:^-\s*(\S{1,5})\b)`)

func textWidthAndHeight(text string, fnt *truetype.Font, fontSize float64) (width int, height int, descent int) {
	opts := truetype.Options{}
	opts.Size = fontSize
	opts.Hinting = font.HintingNone
	face := truetype.NewFace(fnt, &opts)
	defer face.Close()

	width = 0
	height = 0
	descent = 0

	if len(text) > 0 {
		runes := []rune(text)

		b, a, _ := face.GlyphBounds(runes[0])
		width += a.Round()
		height = utils.Max(height, b.Max.Y.Round()-b.Min.Y.Round())
		descent = utils.Max(descent, b.Max.Y.Round())

		for i := 1; i < len(runes); i++ {
			b, a, _ := face.GlyphBounds(runes[i])
			width += a.Round()
			height = utils.Max(height, b.Max.Y.Round()-b.Min.Y.Round())
			descent = utils.Max(descent, b.Max.Y.Round())

			a = face.Kern(runes[i-1], runes[i])
			width += a.Round()
		}
	}

	return width, height, descent
}

func textImage(text string, fnt *truetype.Font, fontSize float64) image.Image {
	if fnt == nil {
		return nil
	}

	width, height, descent := textWidthAndHeight(text, fnt, fontSize)
	textImg := image.NewRGBA(image.Rect(0, 0, width, height))

	c := freetype.NewContext()
	c.SetDst(textImg)
	c.SetSrc(image.Black)
	c.SetFont(fnt)
	c.SetFontSize(fontSize)
	//c.SetDPI(72)
	c.SetClip(textImg.Bounds())
	c.SetHinting(font.HintingNone)

	pt := freetype.Pt(0, height-descent)
	_, err := c.DrawString(text, pt)
	if err != nil {
		log.Fatal(err)
	}

	return textImg
}

type Note struct {
	noteId   string
	noteImg  image.Image
	position int // line in text
}

type NotesModel struct {
	mutex    sync.Mutex
	fontSize float64
	font     *truetype.Font
	text     string
	notes    []Note

	listeners utils.Signal0 // listener'ы на изменение списка
}

func NewNotesModel(fontSize float64, font *truetype.Font) *NotesModel {
	return &NotesModel{fontSize: fontSize, font: font}
}

func (m *NotesModel) MarshalJSON() ([]byte, error) {
	t := struct {
		Text string `json:"text"`
	}{Text: m.text}

	return json.Marshal(t)
}

func (m *NotesModel) UnmarshalJSON(d []byte) error {
	var t struct {
		Text string `json:"text"`
	}

	err := json.Unmarshal(d, &t)
	if err != nil {
		return err
	}

	m.Update(t.Text)

	return nil
}

func (m *NotesModel) Update(text string) {
	send := false
	func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		deleteNoteIds := m.getNoteIds()

		entryNodeIds := make(map[string]int)
		for i, s := range strings.Split(text, "\n") {
			sub := re.FindStringSubmatch(s)
			if len(sub) == 2 {
				entryNodeIds[sub[1]] = i
			}
		}

		for noteId, position := range entryNodeIds {
			index := slices.Index(deleteNoteIds, noteId)
			if index != -1 {
				deleteNoteIds = slices.Delete(deleteNoteIds, index, index+1)
			} else {
				result := m.add(noteId, position)
				send = send || result
			}
		}

		for _, n := range deleteNoteIds {
			result := m.delete(n)
			send = send || result
		}

		m.text = text
	}()

	if send {
		m.listeners.Emit()
	}
}

func (m *NotesModel) SetFont(fontSize float64, font *truetype.Font) {
	m.mutex.Lock()
	if m.font == font {
		return
	}
	m.mutex.Unlock()

	func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		m.fontSize = fontSize

		for i := range m.notes {
			m.notes[i].noteImg = textImage(m.notes[i].noteId, m.font, m.fontSize)
		}
	}()

	m.listeners.Emit()
}

func (m *NotesModel) add(noteId string, position int) bool {
	index := slices.IndexFunc(m.notes, func(n Note) bool { return n.noteId == noteId })
	if index != -1 {
		return false
	}

	m.notes = append(m.notes, Note{noteId: noteId, noteImg: textImage(noteId, m.font, m.fontSize), position: position})

	return true
}

func (m *NotesModel) delete(noteId string) bool {
	index := slices.IndexFunc(m.notes, func(n Note) bool { return n.noteId == noteId })
	if index == -1 {
		return false
	}

	m.notes = slices.Delete(m.notes, index, index+1)

	return true
}

func (m *NotesModel) GetNoteImage(noteId string) image.Image {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	index := slices.IndexFunc(m.notes, func(n Note) bool { return n.noteId == noteId })
	if index == -1 {
		return nil
	}

	return m.notes[index].noteImg
}

func (m *NotesModel) GetNotePosition(noteId string) int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	index := slices.IndexFunc(m.notes, func(n Note) bool { return n.noteId == noteId })
	if index == -1 {
		return -1
	}

	return m.notes[index].position
}

func (m *NotesModel) Len() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return len(m.notes)
}

func (m *NotesModel) GetNoteIdByIndex(index int) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.notes[index].noteId
}

func (m *NotesModel) getNoteIds() []string {
	nodeIds := make([]string, len(m.notes))
	for i, n := range m.notes {
		nodeIds[i] = n.noteId
	}

	return nodeIds
}

func (m *NotesModel) GetNoteIds() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.getNoteIds()
}

func (m *NotesModel) AddDataChangeListener(listener func()) func() {
	return m.listeners.AddSlot(listener)
}
