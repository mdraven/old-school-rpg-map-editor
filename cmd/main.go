package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"log"
	"old-school-rpg-map-editor/common"
	"old-school-rpg-map-editor/configuration"
	"old-school-rpg-map-editor/models/copy_model"
	"old-school-rpg-map-editor/models/maps_model"
	"old-school-rpg-map-editor/models/mode_model"
	"old-school-rpg-map-editor/models/selected_map_tab_model"
	"old-school-rpg-map-editor/models/shortcuts_model"
	"old-school-rpg-map-editor/undo_redo"
	"old-school-rpg-map-editor/widgets/doc_tabs_widget"
	"old-school-rpg-map-editor/widgets/layer_buttons_widget"
	"old-school-rpg-map-editor/widgets/layers_widget"
	"old-school-rpg-map-editor/widgets/notes_widget"
	"old-school-rpg-map-editor/widgets/palette_widget"
	"old-school-rpg-map-editor/widgets/toolbar_widget"
	"os"
	"reflect"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"github.com/goki/freetype/truetype"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"golang.org/x/exp/maps"
)

func restoreMainWindowSettings(c *configuration.Config, w fyne.Window) {
	mainWindowSize := fyne.Size{}
	mainWindowSize.Height = float32(c.MainWindowHeight)
	mainWindowSize.Width = float32(c.MainWindowWidth)
	w.Resize(mainWindowSize)

	mainWindowPos := fyne.Position{}
	mainWindowPos.X = float32(c.MainWindowX)
	mainWindowPos.Y = float32(c.MainWindowY)
	w.Content().Move(mainWindowPos)
}

func saveMainWindowSettings(f *os.File, c *configuration.Config, w fyne.Window) {
	mainWindowSize := w.Canvas().Size()
	c.MainWindowHeight = int(mainWindowSize.Height)
	c.MainWindowWidth = int(mainWindowSize.Width)

	mainWindowPos := w.Content().Position()
	c.MainWindowX = int(mainWindowPos.X)
	c.MainWindowY = int(mainWindowPos.Y)
}

func restoreContentAndToolsSettings(c *configuration.Config, content, tools *container.Split) {
	content.SetOffset(float64(c.ContentOffset))
	tools.SetOffset(float64(c.ToolsOffset))
}

func saveContentAndToolsSettings(f *os.File, c *configuration.Config, content, tools *container.Split) {
	c.ContentOffset = float32(content.Offset)
	c.ToolsOffset = float32(tools.Offset)
}

func loadImage(file string) (image.Image, error) {
	reader, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrap(err, "cannot load images")
	}
	defer reader.Close()

	image, _, err := image.Decode(reader)
	if err != nil {
		return nil, errors.Wrap(err, "cannot load images")
	}

	return image, nil
}

func makeSelectedButtonIcon(border image.Image, res fyne.Resource) (fyne.Resource, error) {
	resImg, err := png.Decode(bytes.NewReader(res.Content()))
	if err != nil {
		return nil, err
	}

	selectedImg := image.NewRGBA(resImg.Bounds())
	draw.Draw(selectedImg, selectedImg.Bounds(), resImg, image.Point{}, draw.Over)
	draw.Draw(selectedImg, selectedImg.Bounds(), border, image.Point{}, draw.Over)

	var buf bytes.Buffer

	err = png.Encode(&buf, selectedImg)
	if err != nil {
		log.Fatal(err)
	}

	return fyne.NewStaticResource(fmt.Sprintf("%s.selected.png", res.Name()), buf.Bytes()), nil
}

func processKeymaps(keyMap map[string]struct{}, shortcutsModel *shortcuts_model.ShortcutsModel, mapsModel *maps_model.MapsModel, selectedMapTabModel *selected_map_tab_model.SelectedMapTabModel) {
	for _, st := range shortcutsModel.Get(shortcuts_model.ShortcutFromMap(keyMap)) {
		switch st {
		case shortcuts_model.RotateMapClockwise, shortcuts_model.RotateMapCounterClockwise:
			{
				mapElem := mapsModel.GetById(selectedMapTabModel.Selected())

				var action undo_redo.UndoRedoAction
				if st == shortcuts_model.RotateMapClockwise {
					action = undo_redo.NewRotateClockwiseAction()
				}
				if st == shortcuts_model.RotateMapCounterClockwise {
					action = undo_redo.NewRotateCounterclockwiseAction()
				}

				err := common.MakeAction(action, mapsModel, mapElem.MapId, reflect.TypeOf((*undo_redo.RotateMapContainer)(nil)))
				if err != nil {
					// TODO
					fmt.Println(err)
					return
				}
			}
		case shortcuts_model.ScrollMapLeft, shortcuts_model.ScrollMapRight, shortcuts_model.ScrollMapUp, shortcuts_model.ScrollMapDown:
			{
				mapElem := mapsModel.GetById(selectedMapTabModel.Selected())

				center := mapElem.CenterModel.Get()

				if st == shortcuts_model.ScrollMapLeft {
					center.X -= 50
				}
				if st == shortcuts_model.ScrollMapRight {
					center.X += 50
				}
				if st == shortcuts_model.ScrollMapUp {
					center.Y -= 50
				}
				if st == shortcuts_model.ScrollMapDown {
					center.Y += 50
				}

				err := common.MakeAction(undo_redo.NewSetCenterAction(center), mapsModel, mapElem.MapId, reflect.TypeOf((*undo_redo.SetCenterContainer)(nil)))
				if err != nil {
					// TODO
					fmt.Println(err)
					return
				}
			}
		}
	}
}

func main() {
	a := app.NewWithID("old-school-rpg-map-editor-4130b499-2e11-4f95-86c4-d2ff537d8bea")

	floorImage, err := loadImage("images/floor.png")
	if err != nil {
		log.Fatal(err)
	}

	wallImage, err := loadImage("images/wall.png")
	if err != nil {
		log.Fatal(err)
	}

	floorSelectedImage, err := loadImage("images/floor_selected.png")
	if err != nil {
		log.Fatal(err)
	}

	wallSelectedImage, err := loadImage("images/wall_selected.png")
	if err != nil {
		log.Fatal(err)
	}

	borderImage, err := loadImage("images/selected.png")
	if err != nil {
		log.Fatal(err)
	}

	rotateLeftIcon, err := fyne.LoadResourceFromPath("images/rotate_left.png")
	if err != nil {
		log.Fatal(err)
	}

	rotateRightIcon, err := fyne.LoadResourceFromPath("images/rotate_right.png")
	if err != nil {
		log.Fatal(err)
	}

	setModeIcon, err := fyne.LoadResourceFromPath("images/set_mode.png")
	if err != nil {
		log.Fatal(err)
	}

	setModeSelectedIcon, err := makeSelectedButtonIcon(borderImage, setModeIcon)
	if err != nil {
		log.Fatal(err)
	}

	selectModeIcon, err := fyne.LoadResourceFromPath("images/select_mode.png")
	if err != nil {
		log.Fatal(err)
	}

	selectModeSelectedIcon, err := makeSelectedButtonIcon(borderImage, selectModeIcon)
	if err != nil {
		log.Fatal(err)
	}

	moveModeIcon, err := fyne.LoadResourceFromPath("images/move_mode.png")
	if err != nil {
		log.Fatal(err)
	}

	moveModeSelectedIcon, err := makeSelectedButtonIcon(borderImage, moveModeIcon)
	if err != nil {
		log.Fatal(err)
	}

	searchNoteIcon, err := fyne.LoadResourceFromPath("images/search_note.png")
	if err != nil {
		log.Fatal(err)
	}

	searchNoteSelectedIcon, err := makeSelectedButtonIcon(borderImage, searchNoteIcon)
	if err != nil {
		log.Fatal(err)
	}

	fnt, err := truetype.Parse(fyne.CurrentApp().Settings().Theme().Font(fyne.TextStyle{Monospace: true}).Content())
	if err != nil {
		log.Fatal(err)
	}

	configFile, err := configuration.GetConfigFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	defer configFile.Close()

	config, err := configuration.LoadConfig(configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer configuration.SaveConfig(configFile, config)

	imageConfig := configuration.ImageConfig{
		FloorSize: uint(floorImage.Bounds().Dy()),
		WallWidth: 16,
	}

	w := a.NewWindow("Title")

	restoreMainWindowSettings(config, w)
	defer saveMainWindowSettings(configFile, config, w)

	w.SetMaster()

	mapsModel := maps_model.NewMapsModel(8, fnt)
	copyModel := copy_model.NewCopyModel()

	floorPaletteWidget := palette_widget.NewPaletteWidget(floorImage, imageConfig.FloorSize, imageConfig.FloorSize, 0)
	wallPaletteWidget := palette_widget.NewPaletteWidget(wallImage, imageConfig.WallWidth, imageConfig.FloorSize, (imageConfig.FloorSize-imageConfig.WallWidth)/2)

	selectedMapTabModel := selected_map_tab_model.NewSelectedLayerModel()

	shortcutsModel := shortcuts_model.NewShortcutsModel()

	if deskCanvas, ok := w.Canvas().(desktop.Canvas); ok {
		mutex := sync.Mutex{}
		keyMap := make(map[string]struct{})

		deskCanvas.SetOnKeyDown(func(ev *fyne.KeyEvent) {
			mutex.Lock()
			keyMap[string(ev.Name)] = struct{}{}
			mutex.Unlock()
		})
		deskCanvas.SetOnKeyUp(func(ev *fyne.KeyEvent) {
			mutex.Lock()
			delete(keyMap, string(ev.Name))
			mutex.Unlock()
		})

		go func() {
			for range time.Tick(time.Millisecond * 150) {
				mutex.Lock()
				keyMap := maps.Clone(keyMap)
				mutex.Unlock()

				processKeymaps(keyMap, shortcutsModel, mapsModel, selectedMapTabModel)
			}
		}()
	}

	keyS := desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(&keyS, func(shortcut fyne.Shortcut) {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		if mapElem.ModeModel.Mode() == mode_model.SetMode {
			toolbar_widget.SetMode(mapsModel, mapElem.MapId, mode_model.SelectMode)
		} else if mapElem.ModeModel.Mode() == mode_model.SelectMode {
			toolbar_widget.SetMode(mapsModel, mapElem.MapId, mode_model.SetMode)
		}
	})

	keyM := desktop.CustomShortcut{KeyName: fyne.KeyM, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(&keyM, func(shortcut fyne.Shortcut) {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		toolbar_widget.SetMode(mapsModel, mapElem.MapId, mode_model.MoveMode)
	})

	keyZ := desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(&keyZ, func(shortcut fyne.Shortcut) {
		toolbar_widget.Undo(selectedMapTabModel, mapsModel)
	})

	keyY := desktop.CustomShortcut{KeyName: fyne.KeyY, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(&keyY, func(shortcut fyne.Shortcut) {
		toolbar_widget.Redo(selectedMapTabModel, mapsModel)
	})

	layersWidget := layers_widget.NewLayersWidget(theme.VisibilityIcon(), theme.VisibilityOffIcon())

	layerButtons := layer_buttons_widget.NewLayerButtonsWidget(w, mapsModel, selectedMapTabModel)

	notesWidget := notes_widget.NewNotesWidget(nil, borderImage, searchNoteIcon, searchNoteSelectedIcon)
	paletteTabFloors := container.NewTabItem("Floors", container.NewVScroll(floorPaletteWidget))
	paletteTabWalls := container.NewTabItem("Walls", container.NewVScroll(wallPaletteWidget))
	paletteTabNotes := container.NewTabItem("Notes", notesWidget.Container())
	paletteTabs := container.NewAppTabs(
		paletteTabFloors,
		paletteTabWalls,
		paletteTabNotes,
	)

	isFloorTabSelected := func() bool {
		selectedTab := paletteTabs.Selected()
		if selectedTab == nil {
			return false
		}

		return selectedTab == paletteTabFloors || selectedTab == paletteTabNotes
	}

	paletteTabs.OnSelected = func(ti *container.TabItem) {
		mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
		if (mapElem.MapId != uuid.UUID{}) {
			mapWidget := doc_tabs_widget.GetMapWidget(mapElem.ExternalData)
			if mapWidget == nil {
				return
			}
			mapWidget.SetIsClickFloor(isFloorTabSelected())
		}
	}

	mapTabs := doc_tabs_widget.NewDocTabsWidget(mapsModel, selectedMapTabModel, floorPaletteWidget, wallPaletteWidget, notesWidget, paletteTabFloors, paletteTabNotes, paletteTabs, layersWidget, floorImage, wallImage, floorSelectedImage, wallSelectedImage, imageConfig)
	mapTabs.IsFloorTabSelected = isFloorTabSelected

	tools := container.NewVSplit(paletteTabs, container.NewBorder(nil, layerButtons.Container(), nil, nil, layersWidget))
	content := container.NewHSplit(mapTabs.Container(), tools)
	content.SetOffset(0.7)

	restoreContentAndToolsSettings(config, content, tools)
	defer saveContentAndToolsSettings(configFile, config, content, tools)

	toolbar := toolbar_widget.NewToolbar(w, fnt, mapsModel, selectedMapTabModel, copyModel, rotateLeftIcon, rotateRightIcon, setModeIcon, setModeSelectedIcon, selectModeIcon, selectModeSelectedIcon, moveModeIcon, moveModeSelectedIcon)

	w.SetContent(container.NewBorder(toolbar, nil, nil, nil, content))

	disconnectDataChangeSelectedMapTabModel := selectedMapTabModel.AddDataChangeListener(func() {
		if mapsModel.Length() == 0 {
			layersWidget.SetMapModel(nil)
			layersWidget.SetSelectedLayerModel(nil)
			notesWidget.SetNotesModel(nil)
		} else {
			mapElem := mapsModel.GetById(selectedMapTabModel.Selected())
			if (mapElem.MapId != uuid.UUID{}) {
				layersWidget.SetMapModel(mapElem.Model)
				layersWidget.SetSelectedLayerModel(mapElem.SelectedLayerModel)
				notesWidget.SetNotesModel(mapElem.NotesModel)
			}
		}
	})
	defer disconnectDataChangeSelectedMapTabModel()

	w.Show()
	a.Run()
}
