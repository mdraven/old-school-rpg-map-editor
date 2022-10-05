package palette_widget

import (
	"image"
	"image/color"
	"image/draw"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.CanvasObject = &PaletteWidget{}
var _ fyne.Widget = &PaletteWidget{}

type PaletteWidget struct {
	widget.BaseWidget
	mutex        sync.Mutex
	img          image.Image
	elemWidth    uint
	elemHeight   uint
	widthPadding uint
	selected     int
}

func NewPaletteWidget(img image.Image, elemWidth uint, elemHeight uint, widthPadding uint) *PaletteWidget {
	w := &PaletteWidget{img: img, elemWidth: elemWidth, elemHeight: elemHeight, widthPadding: widthPadding}
	w.ExtendBaseWidget(w)
	return w
}

func (w *PaletteWidget) Selected() int {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.selected
}

func (w *PaletteWidget) Tapped(ev *fyne.PointEvent) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	x := int(ev.Position.X)
	y := int(ev.Position.Y)

	step := int(w.elemWidth) + 2*int(w.widthPadding) + 1

	rect := image.Rect(1, 1, step, int(w.elemHeight)+1)
	for i := 0; i != w.img.Bounds().Dx()/int(w.elemWidth); i++ {
		if x >= rect.Min.X && y >= rect.Min.Y && x <= rect.Max.X && y <= rect.Max.Y {
			w.selected = i
			w.Refresh()
			return
		}

		rect = rect.Add(image.Pt(step, 0))
		if rect.Max.X > int(w.Size().Width) {
			rect.Min.X = 1
			rect.Max.X = step
			rect = rect.Add(image.Pt(0, int(w.elemHeight)+1))
		}
	}
}

func (w *PaletteWidget) CreateRenderer() fyne.WidgetRenderer {
	return newPaletteWidgetRenderer(w)
}

var _ fyne.WidgetRenderer = &paletteWidgetRenderer{}

type paletteWidgetRenderer struct {
	widget *PaletteWidget
	raster *canvas.Raster
}

func newPaletteWidgetRenderer(w *PaletteWidget) *paletteWidgetRenderer {
	backgroundUniform := image.NewUniform(color.RGBA{0xff, 0xff, 0xff, 0xff})
	selectUniform := image.NewUniform(color.RGBA{0xaa, 0xaa, 0xff, 0xff})

	var img *image.RGBA

	raster := canvas.NewRaster(func(width, height int) image.Image {
		w.mutex.Lock()
		defer w.mutex.Unlock()

		if img == nil || img.Bounds() != image.Rect(0, 0, width, height) {
			img = image.NewRGBA(image.Rect(0, 0, width, height))
		}

		draw.Draw(img, image.Rect(0, 0, width, height), backgroundUniform, image.Point{}, draw.Src)

		step := int(w.elemWidth) + 2*int(w.widthPadding) + 1

		rect := image.Rect(1, 1, step, int(w.elemHeight)+1)
		for i := 0; i != w.img.Bounds().Dx()/int(w.elemWidth); i++ {
			imgRect := image.Rect(
				rect.Min.X+int(w.widthPadding), rect.Min.Y,
				rect.Max.X-int(w.widthPadding), rect.Max.Y,
			)
			draw.Draw(img, imgRect, w.img, image.Pt(i*int(w.elemWidth), 0), draw.Src)

			if i == w.selected {
				selectRect := image.Rect(rect.Min.X-1, rect.Min.Y-1, rect.Min.X, rect.Max.Y+1)
				draw.Draw(img, selectRect, selectUniform, image.Point{}, draw.Src)

				selectRect = image.Rect(rect.Max.X, rect.Min.Y, rect.Max.X+1, rect.Max.Y+1)
				draw.Draw(img, selectRect, selectUniform, image.Point{}, draw.Src)

				selectRect = image.Rect(rect.Min.X-1, rect.Min.Y-1, rect.Max.X+1, rect.Min.Y)
				draw.Draw(img, selectRect, selectUniform, image.Point{}, draw.Src)

				selectRect = image.Rect(rect.Min.X, rect.Max.Y, rect.Max.X+1, rect.Max.Y+1)
				draw.Draw(img, selectRect, selectUniform, image.Point{}, draw.Src)
			}

			rect = rect.Add(image.Pt(step, 0))
			if rect.Max.X > width {
				rect.Min.X = 1
				rect.Max.X = step
				rect = rect.Add(image.Pt(0, int(w.elemHeight)+1))
			}
		}

		return img
	})

	return &paletteWidgetRenderer{widget: w, raster: raster}
}

func (r *paletteWidgetRenderer) Destroy() {}

func (r *paletteWidgetRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}

func (r *paletteWidgetRenderer) MinSize() fyne.Size {
	r.widget.mutex.Lock()
	defer r.widget.mutex.Unlock()

	step := int(r.widget.elemWidth) + 2*int(r.widget.widthPadding) + 1

	w := -1
	h := 1
	for i := 0; i != r.widget.img.Bounds().Dx()/int(r.widget.elemWidth); i++ {
		w += step
		if w > int(r.raster.Size().Width) {
			w = 0
			h++
		}
	}

	return fyne.NewSize(0, float32(r.widget.elemHeight*uint(h)+2))
}

func (r *paletteWidgetRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func (r *paletteWidgetRenderer) Refresh() {
	r.raster.Refresh()
}
