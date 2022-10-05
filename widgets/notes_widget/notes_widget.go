package notes_widget

import (
	"bytes"
	"image"
	"image/draw"
	"image/png"
	"log"
	"old-school-rpg-map-editor/models/notes_model"
	"old-school-rpg-map-editor/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func makeSetNoteIcon(buttonSize fyne.Size, border, note image.Image, selected bool) fyne.Resource {
	dX := border.Bounds().Dx() //utils.Min(border.Bounds().Dx(), int(buttonSize.Width))
	dY := border.Bounds().Dy() //utils.Min(border.Bounds().Dy(), int(buttonSize.Height))

	dX = utils.Max(dX, note.Bounds().Dx())
	dY = utils.Max(dY, note.Bounds().Dy())

	setNoteImg := image.NewRGBA(image.Rect(0, 0, dX, dY))

	bounds := image.Rectangle{}
	bounds.Min.X = (dX - note.Bounds().Dx()) / 2
	bounds.Min.Y = (dY - note.Bounds().Dy()) / 2
	bounds.Max.X = bounds.Min.X + note.Bounds().Dx()
	bounds.Max.Y = bounds.Min.Y + note.Bounds().Dy()

	draw.Draw(setNoteImg, bounds, note, image.Point{}, draw.Over)
	if selected {
		draw.Draw(setNoteImg, setNoteImg.Bounds(), border, image.Point{}, draw.Over)
	}

	var buf bytes.Buffer

	err := png.Encode(&buf, setNoteImg)
	if err != nil {
		log.Fatal(err)
	}

	return fyne.NewStaticResource("notes.png", buf.Bytes())
}

type NotesWidget struct {
	container  *fyne.Container
	selected   string
	model      *notes_model.NotesModel
	disconnect utils.Signal0

	OnSelected func(value string)
}

func NewNotesWidget(model *notes_model.NotesModel, border image.Image, searchIcon, searchSelectedIcon fyne.Resource) *NotesWidget {
	setNoteOnFloor := widget.NewButtonWithIcon("", nil, func() {})
	notesTools := container.NewHBox(widget.NewButtonWithIcon("", searchIcon, func() {}), setNoteOnFloor)

	notesEntry := widget.NewEntry()
	notesEntry.MultiLine = true

	w := &NotesWidget{}

	w.SetNotesModel(model)

	notesEntry.OnChanged = func(s string) {
		if w.model != nil {
			w.model.Update(s)
		}
	}

	sel := widget.NewSelect(nil, func(s string) {})

	sel.OnChanged = func(s string) {
		w.selected = s
		if w.OnSelected != nil {
			w.OnSelected(s)
		}

		var image image.Image
		if w.model != nil {
			image = w.model.GetNoteImage(s)
		}

		if image != nil {
			setNoteOnFloor.Icon = makeSetNoteIcon(setNoteOnFloor.Size(), border, image, true)
			setNoteOnFloor.Refresh()
		} else {
			// TODO: ставим в setNoteOnFloor placeholder
		}

		if position := w.model.GetNotePosition(s); notesEntry.CursorRow != position {
			notesEntry.CursorRow = position
			notesEntry.Refresh()
		}
	}

	w.container = container.NewBorder(container.NewBorder(nil, nil, notesTools, nil, sel), nil, nil, nil, notesEntry)

	return w
}

func (w *NotesWidget) SetNotesModel(model *notes_model.NotesModel) {
	if w.model == model {
		return
	}

	if w.model != nil {
		w.disconnect.Emit()
		w.disconnect.Clear()
	}

	w.model = model

	if model != nil {
		sel := w.container.Objects[1].(*fyne.Container).Objects[0].(*widget.Select)
		w.disconnect.AddSlot(model.AddDataChangeListener(func() {
			sel.Options = w.model.GetNoteIds()
			sel.Refresh()
		}))
	}

	w.container.Refresh()
}

func (w *NotesWidget) Container() *fyne.Container {
	return w.container
}

func (w *NotesWidget) Selected() string {
	return w.selected
}
