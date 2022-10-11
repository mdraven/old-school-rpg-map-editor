package map_widget

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"old-school-rpg-map-editor/configuration"
	"old-school-rpg-map-editor/models/mode_model"
	"old-school-rpg-map-editor/models/notes_model"
	"old-school-rpg-map-editor/models/rot_map_model"
	"old-school-rpg-map-editor/models/rot_select_model"
	"old-school-rpg-map-editor/models/rotate_model"
	"old-school-rpg-map-editor/utils"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/disintegration/imaging"
)

type Mode int

const (
	SetMode    Mode = 0
	SelectMode Mode = 1 // выделять область мышкой, а не scroll'ить
	MoveMode   Mode = 2 // перемещаем элементы из selected
)

type modeData interface {
	getMode() Mode
}

type setModeData struct{}

func (*setModeData) getMode() Mode {
	return SetMode
}

type selectModeData struct {
	selectionArea *utils.VectorInt // область(рамка) выделения, если nil, то рамки нет
}

func (*selectModeData) getMode() Mode {
	return SelectMode
}

type moveModeData struct {
	begin *utils.Int2
}

func (*moveModeData) getMode() Mode {
	return MoveMode
}

type draggedSecondary struct {
	begin *utils.Float2
}

type MapWidget struct {
	widget.BaseWidget
	mutex                  sync.Mutex
	origFloorImage         image.Image
	floorImage             image.Image // resized
	origWallImage          image.Image
	wallImage              image.Image // resized
	wallImage90            image.Image // resized
	origFloorSelectedImage image.Image
	floorSelectedImage     image.Image // resized
	origWallSelectedImage  image.Image
	wallSelectedImage      image.Image // resized
	wallSelectedImage90    image.Image // resized
	imageConfig            configuration.ImageConfig
	mapModel               *rot_map_model.RotMapModel
	disconnectMapModel     utils.Signal0
	selectModel            *rot_select_model.RotSelectModel
	disconnectSelectModel  func()
	modeModel              *mode_model.ModeModel
	disconnectModeModel    func()
	rotateModel            *rotate_model.RotateModel
	disconnectRotateModel  utils.Signal0
	notesModel             *notes_model.NotesModel
	disconnectNotesModel   utils.Signal0
	clickFloor             func(x, y int)
	clickWall              func(x, y int, isRight bool /*or bottom*/)
	moveSelectedTo         func(offsetX, offsetY int, startGrab bool)
	selectArea             func(floors []utils.Int2, rightWall []utils.Int2, bottomWall []utils.Int2)
	unselectAll            func()
	isClickFloor           bool // обрабатывать click на floor или wall
	modeData               modeData
	offset                 utils.Float2
	scale                  float32
	draggedSecondary       draggedSecondary
}

func NewMapWidget(floorImage image.Image, wallImage image.Image, floorSelectedImage image.Image, wallSelectedImage image.Image, imageConfig configuration.ImageConfig, rotateModel *rotate_model.RotateModel, mapModel *rot_map_model.RotMapModel, selectModel *rot_select_model.RotSelectModel, modeModel *mode_model.ModeModel, notesModel *notes_model.NotesModel, clickFloor func(x, y int), clickWall func(x, y int, isRight bool), moveSelectedTo func(offsetX, offsetY int, startGrab bool), selectArea func(floors []utils.Int2, rightWall []utils.Int2, bottomWall []utils.Int2), unselectAll func()) *MapWidget {
	w := &MapWidget{
		origFloorImage:         floorImage,
		floorImage:             floorImage,
		origWallImage:          wallImage,
		wallImage:              wallImage,
		wallImage90:            imaging.Rotate270(wallImage),
		origFloorSelectedImage: floorSelectedImage,
		floorSelectedImage:     floorSelectedImage,
		origWallSelectedImage:  wallSelectedImage,
		wallSelectedImage:      wallSelectedImage,
		wallSelectedImage90:    imaging.Rotate270(wallSelectedImage),
		imageConfig:            imageConfig,
		clickFloor:             clickFloor,
		clickWall:              clickWall,
		moveSelectedTo:         moveSelectedTo,
		selectArea:             selectArea,
		unselectAll:            unselectAll,
		modeData:               &setModeData{},
		scale:                  1.,
	}

	w.SetRotateModel(rotateModel)
	w.SetMapModel(mapModel)
	w.SetSelectModel(selectModel)
	w.SetModeModel(modeModel)
	w.SetNotesModel(notesModel)

	w.ExtendBaseWidget(w)
	return w
}

func (w *MapWidget) Destroy() {
	w.SetRotateModel(nil)
	w.SetMapModel(nil)
	w.SetSelectModel(nil)
	w.SetModeModel(nil)
	w.SetNotesModel(nil)
}

func (w *MapWidget) CreateRenderer() fyne.WidgetRenderer {
	return newMapWidgetRenderer(w)
}

// x, y - координаты на экране в пикселях
func (w *MapWidget) isFloorSelected(model *rot_select_model.RotSelectModel, x, y uint) bool {
	fFloorSize := float32(w.imageConfig.FloorSize)
	scaledFloorWbSize := int((fFloorSize + 1) * w.scale) // With Border

	mapX, mapY, _, _ := w.screenPixelToFloorCoords(x, y, uint(scaledFloorWbSize))
	return model.IsFloorSelected(mapX, mapY)
}

// x, y - координаты на экране в пикселях
func (w *MapWidget) isWallSelected(model *rot_select_model.RotSelectModel, x, y uint) bool {
	fFloorSize := float32(w.imageConfig.FloorSize)
	scaledFloorWbSize := int((fFloorSize + 1) * w.scale) // With Border
	halfScaledWallWidth := uint(float32(w.imageConfig.WallWidth) * w.scale / 2)

	mapX, mapY, imgX, imgY := w.screenPixelToFloorCoords(x, y, uint(scaledFloorWbSize))

	if imgX < halfScaledWallWidth {
		return model.IsWallSelected(mapX-1, mapY, true)
	} else if imgX > uint(fFloorSize)-halfScaledWallWidth {
		return model.IsWallSelected(mapX, mapY, true)
	}

	if imgY < halfScaledWallWidth {
		return model.IsWallSelected(mapX, mapY-1, false)
	} else if imgY > uint(fFloorSize)-halfScaledWallWidth {
		return model.IsWallSelected(mapX, mapY, false)
	}

	return false
}

func (w *MapWidget) Tapped(ev *fyne.PointEvent) {
	w.mutex.Lock()
	once := sync.Once{}
	defer once.Do(w.mutex.Unlock)

	switch w.modeData.(type) {
	case *setModeData:
		fFloorSize := float32(w.imageConfig.FloorSize)
		fWallWidth := float32(w.imageConfig.WallWidth)

		if w.isClickFloor {
			mapX, mapY, _, _ := w.screenPixelToFloorCoords(uint(ev.Position.X), uint(ev.Position.Y), uint((fFloorSize+1)*w.scale))
			once.Do(w.mutex.Unlock)
			w.clickFloor(mapX, mapY)
		} else {
			// Так как половина стены выползает за floor, то подвинем координаты на wallWidth/2
			pos := ev.Position.Subtract(fyne.NewDelta(fWallWidth*w.scale/2, fWallWidth*w.scale/2))

			mapX, mapY, imgX, imgY := w.screenPixelToFloorCoords(uint(pos.X), uint(pos.Y), uint((fFloorSize+1)*w.scale))

			floorWithoutWall := (fFloorSize + 1 - fWallWidth) * w.scale

			if float32(imgX) > 0 && float32(imgX) < floorWithoutWall && float32(imgY) > floorWithoutWall {
				w.clickWall(mapX, mapY, false)
			}
			if float32(imgY) > 0 && float32(imgY) < floorWithoutWall && float32(imgX) > floorWithoutWall {
				w.clickWall(mapX, mapY, true)
			}
		}
		//case *selectModeData:
	}
}

func (w *MapWidget) MouseDown(ev *desktop.MouseEvent) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if ev.Button == desktop.MouseButtonSecondary {
		float2 := utils.NewFloat2(ev.Position.X, ev.Position.Y)
		w.draggedSecondary.begin = &float2
	}
}

func (w *MapWidget) MouseUp(ev *desktop.MouseEvent) {
	w.mutex.Lock()
	once := sync.Once{}
	defer once.Do(w.mutex.Unlock)

	if ev.Button == desktop.MouseButtonSecondary {
		// TODO: если нужен DraggedSecondaryEnd, то делаем его тут
		//if *w.draggedSecondary.begin != utils.NewFloat2(ev.Position.X, ev.Position.Y) {
		//	once.Do(w.mutex.Unlock)
		//	w.DraggedSecondaryEnd(XXX)
		//}
		w.draggedSecondary.begin = nil
	}
}

func (w *MapWidget) MouseMoved(ev *desktop.MouseEvent) {
	w.mutex.Lock()
	once := sync.Once{}
	defer once.Do(w.mutex.Unlock)

	if w.draggedSecondary.begin != nil {
		oldBegin := *w.draggedSecondary.begin
		w.draggedSecondary.begin.X = ev.Position.X
		w.draggedSecondary.begin.Y = ev.Position.Y

		once.Do(w.mutex.Unlock)
		w.DraggedSecondary(ev.Position.X-oldBegin.X, ev.Position.Y-oldBegin.Y)
		return
	}
}

func (w *MapWidget) MouseIn(ev *desktop.MouseEvent) {}

func (w *MapWidget) MouseOut() {}

func (w *MapWidget) DraggedSecondary(dX float32, dY float32) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	offset := w.offset
	offset.X -= dX
	offset.Y -= dY
	w.offset = offset

	w.Refresh()
}

func (w *MapWidget) Dragged(ev *fyne.DragEvent) {
	w.mutex.Lock()
	once := sync.Once{}
	defer once.Do(w.mutex.Unlock)

	switch mode := w.modeData.(type) {
	case *setModeData:
	case *selectModeData:
		if mode.selectionArea == nil {
			mode.selectionArea = utils.ToPtr(utils.NewVectorInt(utils.NewInt2(int(ev.Position.X), int(ev.Position.Y)), utils.Int2{}))

			func() {
				w.mutex.Unlock()
				defer w.mutex.Lock()
				w.unselectAll()
			}()
		}
		mode.selectionArea.End.X = int(ev.Position.X)
		mode.selectionArea.End.Y = int(ev.Position.Y)

		offset := w.offset

		// скролинг у краёв по-горизонтали
		if ev.Position.X > w.Size().Width {
			diff := utils.Max(w.Size().Width-ev.Position.X, -10)
			offset.X -= diff
			mode.selectionArea.Begin.X += int(diff)
		} else if ev.Position.X < 0 {
			diff := utils.Max(ev.Position.X, -10)
			offset.X += diff
			mode.selectionArea.Begin.X -= int(diff)
		}

		// скролинг у краёв по-вертикали
		if ev.Position.Y > w.Size().Height {
			diff := utils.Max(w.Size().Height-ev.Position.Y, -10)
			offset.Y -= diff
			mode.selectionArea.Begin.Y += int(diff)
		} else if ev.Position.Y < 0 {
			diff := utils.Max(ev.Position.Y, -10)
			offset.Y += diff
			mode.selectionArea.Begin.Y -= int(diff)
		}

		if offset != w.offset {
			w.offset = offset
		}

		w.Refresh()
	case *moveModeData:
		fFloorSize := float32(w.imageConfig.FloorSize)
		scaledFloorWbSize := int((fFloorSize + 1) * w.scale) // With Border

		if mode.begin == nil {
			if w.isFloorSelected(w.selectModel, uint(ev.Position.X), uint(ev.Position.Y)) ||
				w.isWallSelected(w.selectModel, uint(ev.Position.X), uint(ev.Position.Y)) {

				mapX, mapY, _, _ := w.screenPixelToFloorCoords(uint(ev.Position.X), uint(ev.Position.Y), uint(scaledFloorWbSize))

				int2 := utils.NewInt2(mapX, mapY)
				mode.begin = &int2

				once.Do(w.mutex.Unlock)
				w.moveSelectedTo(0, 0, true)
			}

			return
		}

		mapX, mapY, _, _ := w.screenPixelToFloorCoords(uint(ev.Position.X), uint(ev.Position.Y), uint(scaledFloorWbSize))

		offsetX := mapX - mode.begin.X
		offsetY := mapY - mode.begin.Y

		if offsetX != 0 || offsetY != 0 {
			mode.begin.X = mapX
			mode.begin.Y = mapY

			once.Do(w.mutex.Unlock)
			w.moveSelectedTo(offsetX, offsetY, false)
		}
	default:
		panic(fmt.Sprintf("unknown map widget mode: %v", mode))
	}
}

func (w *MapWidget) DragEnd() {
	w.mutex.Lock()
	once := sync.Once{}
	defer once.Do(w.mutex.Unlock)

	if modeData, ok := w.modeData.(*selectModeData); ok {
		if modeData.selectionArea != nil {
			fFloorSize := float32(w.imageConfig.FloorSize)
			scaledFloorWbSize := int((fFloorSize + 1) * w.scale) // With Border
			halfScaledWallWidth := uint(float32(w.imageConfig.WallWidth) * w.scale / 2)

			rect := utils.VectorInt{
				Begin: utils.Int2{
					X: utils.Min(modeData.selectionArea.Begin.X, modeData.selectionArea.End.X),
					Y: utils.Min(modeData.selectionArea.Begin.Y, modeData.selectionArea.End.Y),
				},
				End: utils.Int2{
					X: utils.Max(modeData.selectionArea.Begin.X, modeData.selectionArea.End.X),
					Y: utils.Max(modeData.selectionArea.Begin.Y, modeData.selectionArea.End.Y),
				},
			}

			mapLeft, mapTop, imgLeft, imgTop := w.screenPixelToFloorCoords(uint(rect.Begin.X), uint(rect.Begin.Y), uint(scaledFloorWbSize))
			mapRight, mapBottom, imgRight, imgBottom := w.screenPixelToFloorCoords(uint(rect.End.X), uint(rect.End.Y), uint(scaledFloorWbSize))
			mapRight++
			mapBottom++

			var floors []utils.Int2
			var rightWalls []utils.Int2
			var bottomWalls []utils.Int2

			// внутренняя часть рамки
			for y := mapTop; y < mapBottom; y++ {
				for x := mapLeft; x < mapRight; x++ {
					floors = append(floors, utils.NewInt2(x, y))

					if x < mapRight-1 {
						rightWalls = append(rightWalls, utils.NewInt2(x, y))
					}
					if y < mapBottom-1 {
						bottomWalls = append(bottomWalls, utils.NewInt2(x, y))
					}
				}
			}

			// края рамки
			if imgLeft < halfScaledWallWidth {
				for i := mapTop; i < mapBottom; i++ {
					rightWalls = append(rightWalls, utils.NewInt2(mapLeft-1, i))
				}
			} else if imgRight > uint(fFloorSize)-halfScaledWallWidth {
				for i := mapTop; i < mapBottom; i++ {
					rightWalls = append(rightWalls, utils.NewInt2(mapRight-1, i))
				}
			}

			// края рамки
			if imgTop < halfScaledWallWidth {
				for i := mapLeft; i < mapRight; i++ {
					bottomWalls = append(bottomWalls, utils.NewInt2(mapTop-1, i))
				}
			} else if imgBottom > uint(fFloorSize)-halfScaledWallWidth {
				for i := mapLeft; i < mapRight; i++ {
					bottomWalls = append(bottomWalls, utils.NewInt2(mapBottom-1, i))
				}
			}

			// рамка больше не нужна
			modeData.selectionArea = nil

			once.Do(w.mutex.Unlock)
			w.selectArea(floors, rightWalls, bottomWalls)
		}
	} else if modeData, ok := w.modeData.(*moveModeData); ok {
		//once.Do(w.mutex.Unlock)
		modeData.begin = nil
		//w.moveSelectedTo(0, 0, true)
	}
}

func (w *MapWidget) Scrolled(ev *fyne.ScrollEvent) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	oldOffset := w.offset

	coef := float32(0)

	if ev.Scrolled.DY < 0 {
		coef = 1 / (-ev.Scrolled.DY * 0.15)
	} else {
		coef = ev.Scrolled.DY * 0.15
	}

	fFloorSize := float32(w.imageConfig.FloorSize)
	mapX, mapY, _, _ := w.screenPixelToFloorCoords(uint(w.Size().Width/2), uint(w.Size().Height/2), uint((fFloorSize+1)*w.scale))

	w.offset = utils.Float2{}

	pX, pY := w.floorCoordsToScreenPixel(mapX, mapY, 0, 0, uint((fFloorSize+1)*w.scale*coef))
	w.offset = utils.NewFloat2(float32(pX)-w.Size().Width/2, float32(pY)-w.Size().Height/2)

	if w.setScale(w.scale * coef) {
		w.Refresh()
		return
	}

	w.offset = oldOffset
}

func (w *MapWidget) SetCenter(pos utils.Int2) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	fFloorSize := float32(w.imageConfig.FloorSize)

	pX, pY := w.floorCoordsToScreenPixel(pos.X, pos.Y, 0, 0, uint((fFloorSize+1)*w.scale))
	w.offset = utils.NewFloat2(float32(pX)-w.Size().Width/2, float32(pY)-w.Size().Height/2)
}

func (w *MapWidget) Center() utils.Int2 {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	fFloorSize := float32(w.imageConfig.FloorSize)

	mapX, mapY, _, _ := w.screenPixelToFloorCoords(uint(w.Size().Width/2), uint(w.Size().Height/2), uint((fFloorSize+1)*w.scale))

	return utils.NewInt2(mapX, mapY)
}

func (w *MapWidget) Scale() float32 {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.scale
}

func (w *MapWidget) setScale(v float32) bool {
	if v < 0.5 || v > 6. {
		return false
	}

	floorBounds := w.origFloorImage.Bounds()
	wallBounds := w.origWallImage.Bounds()

	w.scale = v

	w.floorImage = imaging.Resize(w.origFloorImage, int(float32(floorBounds.Dx())*w.scale), int(float32(floorBounds.Dy())*w.scale), imaging.Lanczos)

	w.wallImage = imaging.Resize(w.origWallImage, int(float32(wallBounds.Dx())*w.scale), int(float32(wallBounds.Dy())*w.scale), imaging.Lanczos)
	w.wallImage90 = imaging.Rotate270(w.wallImage)

	w.floorSelectedImage = imaging.Resize(w.origFloorSelectedImage, int(float32(floorBounds.Dx())*w.scale), int(float32(floorBounds.Dy())*w.scale), imaging.Lanczos)

	w.wallSelectedImage = imaging.Resize(w.origWallSelectedImage, int(float32(wallBounds.Dx())*w.scale), int(float32(wallBounds.Dy())*w.scale), imaging.Lanczos)
	w.wallSelectedImage90 = imaging.Rotate270(w.wallSelectedImage)

	return true
}

func (w *MapWidget) SetScale(v float32) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	result := w.setScale(v)

	w.Refresh()

	return result
}

func (w *MapWidget) IsClickFloor() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.isClickFloor
}

func (w *MapWidget) SetIsClickFloor(v bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.isClickFloor = v
}

func (w *MapWidget) Mode() Mode {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.modeData.getMode()
}

func (w *MapWidget) setSetMode() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.modeData = &setModeData{}
}

func (w *MapWidget) setSelectMode() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.modeData = &selectModeData{}
}

func (w *MapWidget) setMoveMode() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if _, ok := w.modeData.(*selectModeData); ok {
		w.modeData = &moveModeData{}
	}
}

func (w *MapWidget) SetMapModel(mapModel *rot_map_model.RotMapModel) {
	if w.mapModel == mapModel {
		return
	}

	if w.mapModel != nil {
		w.disconnectMapModel.Emit()
		w.disconnectMapModel.Clear()
	}

	w.mapModel = mapModel

	if mapModel != nil {
		w.disconnectMapModel.AddSlot(mapModel.AddDataChangeListener(w.Refresh))
	}

	w.Refresh()
}

func (w *MapWidget) SetRotateModel(rotateModel *rotate_model.RotateModel) {
	if w.rotateModel == rotateModel {
		return
	}

	if w.rotateModel != nil {
		w.disconnectRotateModel.Emit()
		w.disconnectRotateModel.Clear()
	}

	w.rotateModel = rotateModel

	if rotateModel != nil {
		w.disconnectRotateModel.AddSlot(rotateModel.AddDataChangeListener(w.Refresh))

		var offsetBeforeRotate utils.Int2
		w.disconnectRotateModel.AddSlot(rotateModel.AddBeforeRotateListener(func() {
			w.mutex.Lock()
			offsetBeforeRotate = utils.NewInt2(int(w.offset.X), int(w.offset.Y))
			w.mutex.Unlock()

			offsetBeforeRotate.X += int(w.Size().Width / 2)
			offsetBeforeRotate.Y += int(w.Size().Height / 2)

			offsetBeforeRotate.X, offsetBeforeRotate.Y = w.rotateModel.TransformToRot(offsetBeforeRotate.X, offsetBeforeRotate.Y)
		}))
		w.disconnectRotateModel.AddSlot(rotateModel.AddAfterRotateListener(func() {
			offsetBeforeRotate.X, offsetBeforeRotate.Y = w.rotateModel.TransformFromRot(offsetBeforeRotate.X, offsetBeforeRotate.Y)

			w.mutex.Lock()
			w.offset.X = float32(offsetBeforeRotate.X) - w.Size().Width/2.
			w.offset.Y = float32(offsetBeforeRotate.Y) - w.Size().Height/2.
			w.mutex.Unlock()
		}))
	}

	w.Refresh()
}

func (w *MapWidget) SetSelectModel(selectModel *rot_select_model.RotSelectModel) {
	if w.selectModel == selectModel {
		return
	}

	if w.selectModel != nil {
		w.disconnectSelectModel()
	}

	w.selectModel = selectModel

	if selectModel != nil {
		w.disconnectSelectModel = selectModel.AddDataChangeListener(w.Refresh)
	}

	w.Refresh()
}

func (w *MapWidget) SetModeModel(modeModel *mode_model.ModeModel) {
	if w.modeModel == modeModel {
		return
	}

	if w.modeModel != nil {
		w.disconnectModeModel()
	}

	w.modeModel = modeModel

	if modeModel != nil {
		w.disconnectModeModel = modeModel.AddDataChangeListener(func() {
			if modeModel.Mode() == mode_model.SetMode {
				w.setSetMode()
			} else if modeModel.Mode() == mode_model.MoveMode {
				w.setMoveMode()
			} else if modeModel.Mode() == mode_model.SelectMode {
				w.setSelectMode()
			}
		})
	}

	w.Refresh()
}

func (w *MapWidget) SetNotesModel(notesModel *notes_model.NotesModel) {
	if w.notesModel == notesModel {
		return
	}

	if w.notesModel != nil {
		w.disconnectNotesModel.Emit()
		w.disconnectNotesModel.Clear()
	}

	w.notesModel = notesModel

	if notesModel != nil {
		w.disconnectNotesModel.AddSlot(notesModel.AddDataChangeListener(w.Refresh))
	}

	w.Refresh()
}

// mapX, mapY - координаты у w.getFloor(); imgX, imgY - координата у floor в mapX, mapY.
func (w *MapWidget) screenPixelToFloorCoords(x, y uint, floorSize uint) (mapX, mapY int, imgX, imgY uint) {
	// Делим с округлением в меньшую сторону. Т.е.: `1 / 2 == 0` и `-1 / 2 == -1`
	floorDiv := func(a int, b int) int {
		res := a / b
		if a < 0 && a%b != 0 {
			res--
		}
		return res
	}
	mapX = floorDiv(int(x)+int(w.offset.X), int(floorSize))
	mapY = floorDiv(int(y)+int(w.offset.Y), int(floorSize))
	// Если value < 0, то оно завернётся с другой стороны interval
	wrapValue := func(value int, interval uint) uint {
		v := value % int(interval)
		if v < 0 {
			return uint(int(interval) + v)
		}
		return uint(v)
	}
	imgX = (x + wrapValue(int(w.offset.X), floorSize)) % floorSize
	imgY = (y + wrapValue(int(w.offset.Y), floorSize)) % floorSize
	return
}

func (w *MapWidget) floorCoordsToScreenPixel(mapX, mapY int, imgX, imgY uint, floorSize uint) (x, y int) {
	x = mapX*int(floorSize) - int(w.offset.X) + int(imgX)
	y = mapY*int(floorSize) - int(w.offset.Y) + int(imgY)
	return
}

type mapWidgetRenderer struct {
	widget *MapWidget
	raster *canvas.Raster
}

func newMapWidgetRenderer(w *MapWidget) *mapWidgetRenderer {
	backgroundUniform := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	wallUniform := image.NewUniform(color.RGBA{0xaa, 0xaa, 0xaa, 0xff})
	selectUniform := image.NewUniform(color.RGBA{0xaa, 0xaa, 0xff, 0xff})

	var img *image.RGBA

	raster := canvas.NewRaster(func(width, height int) image.Image {
		w.mutex.Lock()
		defer w.mutex.Unlock()

		if img == nil || img.Bounds() != image.Rect(0, 0, width, height) {
			img = image.NewRGBA(image.Rect(0, 0, width, height))
		}

		draw.Draw(img, image.Rect(0, 0, width, height), backgroundUniform, image.Point{}, draw.Src)

		if !w.mapModel.Model().HasVisible() {
			return img
		}

		fFloorSize := float32(w.imageConfig.FloorSize)

		scaledFloorWbSize := int((fFloorSize + 1) * w.scale) // With Border
		scaledFloorWobSize := int(fFloorSize * w.scale)      // Without Border
		scaledWallWidth := int(float32(w.imageConfig.WallWidth) * w.scale)
		halfScaledWallWidth := int(float32(w.imageConfig.WallWidth) * w.scale / 2)

		mapLeft, mapTop, _ /* imgX */, _ /* imgY */ := w.screenPixelToFloorCoords(uint(0), uint(0), uint(scaledFloorWbSize))
		mapRight, mapBottom, _, _ := w.screenPixelToFloorCoords(uint(width), uint(height), uint(scaledFloorWbSize))
		mapRight++
		mapBottom++

		floorRect := func(x, y int) image.Rectangle {
			pX, pY := w.floorCoordsToScreenPixel(x, y, 0, 0, uint(scaledFloorWbSize))
			return image.Rect(pX, pY, int(pX)+scaledFloorWobSize, int(pY)+scaledFloorWobSize)
		}

		for y := mapTop; y < mapBottom; y++ {
			for x := mapLeft; x < mapRight; x++ {
				_, index := w.mapModel.VisibleFloor(x, y)
				if index > 0 {
					rect := floorRect(x, y)
					draw.Draw(img, rect, w.floorImage, image.Pt(int(index)*int(fFloorSize*w.scale), 0), draw.Src)
				}
			}
		}

		for y := mapTop; y < mapBottom; y++ {
			for x := mapLeft; x < mapRight; x++ {
				_, rightIndex := w.mapModel.VisibleWall(x, y, true)
				_, bottomIndex := w.mapModel.VisibleWall(x, y, false)
				rect := floorRect(x, y)

				if rightIndex > 0 {
					wallRect := image.Rect(rect.Max.X-halfScaledWallWidth, rect.Min.Y, rect.Max.X+halfScaledWallWidth, rect.Max.Y)
					draw.Draw(img, wallRect, w.wallImage, image.Pt(int(rightIndex)*scaledWallWidth, 0), draw.Over)
				} else {
					wallRect := image.Rect(rect.Max.X, rect.Min.Y, rect.Max.X+int(w.scale), rect.Max.Y)
					draw.Draw(img, wallRect, wallUniform, image.Point{}, draw.Src)
				}

				if bottomIndex > 0 {
					wallRect := image.Rect(rect.Min.X, rect.Max.Y-halfScaledWallWidth, rect.Max.X, rect.Max.Y+halfScaledWallWidth)
					draw.Draw(img, wallRect, w.wallImage90, image.Pt(0, int(bottomIndex)*scaledWallWidth), draw.Over)
				} else {
					wallRect := image.Rect(rect.Min.X, rect.Max.Y, rect.Max.X, rect.Max.Y+int(w.scale))
					draw.Draw(img, wallRect, wallUniform, image.Point{}, draw.Src)
				}
			}
		}

		{
			for y := mapTop; y < mapBottom; y++ {
				for x := mapLeft; x < mapRight; x++ {
					if w.selectModel.IsFloorSelected(x, y) {
						rect := floorRect(x, y)
						draw.Draw(img, rect, w.floorSelectedImage, image.Pt(0, 0), draw.Over)
					}
				}
			}

			for y := mapTop; y < mapBottom; y++ {
				for x := mapLeft; x < mapRight; x++ {
					rect := floorRect(x, y)

					if w.selectModel.IsWallSelected(x, y, true) {
						wallRect := image.Rect(rect.Max.X-halfScaledWallWidth, rect.Min.Y, rect.Max.X+halfScaledWallWidth, rect.Max.Y)
						draw.Draw(img, wallRect, w.wallSelectedImage, image.Pt(0, 0), draw.Over)
					}

					if w.selectModel.IsWallSelected(x, y, false) {
						wallRect := image.Rect(rect.Min.X, rect.Max.Y-halfScaledWallWidth, rect.Max.X, rect.Max.Y+halfScaledWallWidth)
						draw.Draw(img, wallRect, w.wallSelectedImage90, image.Pt(0, 0), draw.Over)
					}
				}
			}
		}

		for y := mapTop; y < mapBottom; y++ {
			for x := mapLeft; x < mapRight; x++ {
				_, value := w.mapModel.VisibleNoteId(x, y)
				if len(value) > 0 {
					rect := floorRect(x, y)
					noteImg := w.notesModel.GetNoteImage(value)
					draw.Draw(img, rect, noteImg, image.Point{}, draw.Src)
				}
			}
		}

		if modeData, ok := w.modeData.(*selectModeData); ok {
			if modeData.selectionArea != nil {
				selectionArea := modeData.selectionArea
				rect := image.Rect(selectionArea.Begin.X, selectionArea.Begin.Y, selectionArea.End.X, selectionArea.Begin.Y+1)
				draw.Draw(img, rect, selectUniform, image.Point{}, draw.Src)

				rect = image.Rect(selectionArea.End.X, selectionArea.Begin.Y, selectionArea.End.X+1, selectionArea.End.Y)
				draw.Draw(img, rect, selectUniform, image.Point{}, draw.Src)

				rect = image.Rect(selectionArea.Begin.X, selectionArea.End.Y, selectionArea.End.X, selectionArea.End.Y+1)
				draw.Draw(img, rect, selectUniform, image.Point{}, draw.Src)

				rect = image.Rect(selectionArea.Begin.X, selectionArea.Begin.Y, selectionArea.Begin.X+1, selectionArea.End.Y)
				draw.Draw(img, rect, selectUniform, image.Point{}, draw.Src)
			}
		}

		return img
	})

	return &mapWidgetRenderer{widget: w, raster: raster}
}

func (r *mapWidgetRenderer) Destroy() {
}

func (r *mapWidgetRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}

func (r *mapWidgetRenderer) MinSize() fyne.Size {
	return fyne.Size{}
}

func (r *mapWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func (r *mapWidgetRenderer) Refresh() {
	r.raster.Refresh()
}
