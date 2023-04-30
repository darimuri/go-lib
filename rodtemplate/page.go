package rodtemplate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/cdp"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

type ScreenShotOption struct {
	Format      proto.PageCaptureScreenshotFormat
	Quality     int
	XDelta      float64
	YDelta      float64
	WidthDelta  float64
	HeightDelta float64
}

var _ ElementSelector = (*PageTemplate)(nil)

type PageTemplate struct {
	P *rod.Page
}

func (p *PageTemplate) Navigate(url string) error {
	if p.P == nil {
		return errors.New("page is nil")
	}

	p.P.MustNavigate(url)

	p.P.MustWaitRequestIdle()

	return nil
}

func (p *PageTemplate) ClickElement(selector string) {
	p.P.MustWaitIdle()

	el := p.P.MustElement(selector)
	p.MoveMouseTo(el)

	el.MustClick()
}

func (p *PageTemplate) ClickWhenAvailable(selector string) bool {
	for i := 0; i < 1000; i++ {
		if p.Has(selector) {
			el := p.El(selector)
			if el.MustVisible() {
				el.MustFocus()
				el.MustScrollIntoView()
				el.MustClick()
				return true
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
	return false
}

func (p *PageTemplate) FocusWhenAvailable(selector string) bool {
	for i := 0; i < 1000; i++ {
		if p.Has(selector) {
			el := p.El(selector)
			el.MustFocus()
			return true
		}
		time.Sleep(time.Millisecond * 100)
	}
	return false
}

func (p *PageTemplate) MoveMouseTo(el *rod.Element) {
	shape, err := el.Shape()
	if err == nil {
		point := shape.OnePointInside()
		p.P.Mouse.MustMove(point.X, point.Y)
	} else {
		if cErr, ok := err.(*cdp.Error); ok {
			log.Println("failed to get element shape", cErr)
		} else {
			panic(err)
		}
	}
}

func (p *PageTemplate) URL() string {
	return p.P.MustInfo().URL
}

func (p *PageTemplate) Input(selector string, value string) {
	for i := 0; i < 100; i++ {
		if p.P.MustHas(selector) {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if !p.P.MustHas(selector) {
		log.Fatalf("failed to find input having selector %s\n", selector)
	}

	el := p.P.MustElement(selector)
	el.MustClick().MustSelectAllText().MustInput(value)
}

func (p *PageTemplate) PressKey(key input.Key) {
	p.Type(key)
}

func (p *PageTemplate) Type(key ...input.Key) {
	p.P.Keyboard.MustType(key...)
}

func (p *PageTemplate) WaitRequestIdle(excludes ...string) {
	wait := p.P.MustWaitRequestIdle(excludes...)
	wait()
}

func (p *PageTemplate) Has(selector string) bool {
	has, _, err := p.P.Has(selector)
	if err != nil {
		log.Printf("failed to find element using selector %s\n", selector)
		return false
	}

	return has
}

func (p PageTemplate) GetVisibleHeight(selector string) float64 {
	if p.Has(selector) {
		el := p.El(selector)
		if el.MustVisible() {
			return el.Height()
		}
	}

	return 0.0
}

func (p *PageTemplate) El(selector string) *ElementTemplate {
	return &ElementTemplate{Element: p.P.MustElement(selector)}
}

func (p *PageTemplate) Els(selector string) ElementsTemplate {
	return toElementsTemplate(p.P.MustElements(selector))
}

func (p *PageTemplate) Reload() {
	p.P.MustReload()
}

func (p *PageTemplate) FrameID() proto.PageFrameID {
	return p.P.FrameID
}

func (p *PageTemplate) WaitRepaint() {
	if err := p.P.WaitRepaint(); err != nil {
		log.Println("failed to wait", err)
	}
}

func (p *PageTemplate) ScrollTop() {
	p.Type(input.Home)
}

func (p *PageTemplate) ScrollBottom() {
	p.Type(input.End)
}

func (p *PageTemplate) ScrollBottomHuman() {
	metrics, err := proto.PageGetLayoutMetrics{}.Call(p.P)
	if err != nil {
		panic(err)
	}

	width := int(metrics.ContentSize.Width)
	height := int(metrics.ContentSize.Height)

	if err = p.P.Mouse.Scroll(float64(width), float64(height), height/128); err != nil {
		panic(fmt.Errorf("failed to scroll for %s", err.Error()))
	}
}

func (p *PageTemplate) ScrollTo(e *ElementTemplate) {
	quad := e.MustShape().Quads[0]
	bottom := quad[7]
	if err := p.P.Mouse.Scroll(0.0, bottom, 1); err != nil {
		log.Println("failed to scroll mouse", err)
	}
}

func (p *PageTemplate) Body() string {
	return p.El("body").MustHTML()
}

func (p *PageTemplate) HTML() string {
	return p.El("html").MustHTML()
}

func (p *PageTemplate) Event() <-chan *rod.Message {
	return p.P.Event()
}

func (p *PageTemplate) MaximizeToWindowBounds() {
	bounds := p.P.MustGetWindow()
	p.SetViewport(*bounds.Width, *bounds.Height)
}

func (p *PageTemplate) SetViewport(width, height int) {
	p.P.MustSetViewport(width, height, 0, false)
}

func (p *PageTemplate) ScreenShotFull(dumpPath string) []byte {
	opt := ScreenShotOption{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: 95,
	}

	return p.ScreenShotFullWithOption(dumpPath, opt)
}

func (p *PageTemplate) ScreenShotFullWithOption(dumpPath string, opt ScreenShotOption) []byte {
	metrics, err := proto.PageGetLayoutMetrics{}.Call(p.P)
	if err != nil {
		panic(err)
	}

	oldView := proto.EmulationSetDeviceMetricsOverride{}
	set := p.P.LoadState(&oldView)
	view := oldView
	view.Width = int(metrics.ContentSize.Width)
	view.Height = int(metrics.ContentSize.Height)

	err = p.P.SetViewport(&view)
	if err != nil {
		panic(err)
	}

	defer func() { // try to recover the viewport
		if !set {
			_ = proto.EmulationClearDeviceMetricsOverride{}.Call(p.P)
			return
		}

		_ = p.P.SetViewport(&oldView)
	}()

	return p.ScreenShotWithOption(p.El("html"), dumpPath, opt)
}

func (p *PageTemplate) ScreenShot(el *ElementTemplate, dumpPath string, yDelta float64) []byte {
	opt := ScreenShotOption{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: 95,
		YDelta:  yDelta,
	}

	return p.ScreenShotWithOption(el, dumpPath, opt)
}

func (p *PageTemplate) ScreenShotWithOption(el *ElementTemplate, dumpPath string, opt ScreenShotOption) []byte {
	err := el.ScrollIntoView()
	if err != nil {
		panic(err)
	}

	quad := el.MustShape().Quads[0]

	width := quad[2] - quad[0] + opt.WidthDelta
	height := quad[7] - quad[1] + opt.HeightDelta

	req := &proto.PageCaptureScreenshot{
		Format:  opt.Format,
		Quality: &opt.Quality,
		Clip: &proto.PageViewport{
			X:      quad[0] + opt.XDelta,
			Y:      quad[1] + opt.YDelta,
			Width:  width,
			Height: height,
			Scale:  1,
		},
	}

	byteArr, errScreenShot := p.P.Screenshot(false, req)
	if errScreenShot != nil {
		panic(errScreenShot)
	}

	errWrite := ioutil.WriteFile(dumpPath, byteArr, 0644)
	if errWrite != nil {
		panic(errWrite)
	}

	return byteArr
}

func (p *PageTemplate) SelectOrPanic(selector string) *ElementTemplate {
	if !p.Has(selector) {
		panic(fmt.Errorf("%s block is missing", selector))
	}

	return p.El(selector)
}

//NewPageTemplate ...
func NewPageTemplate(p *rod.Page) *PageTemplate {
	return &PageTemplate{P: p}
}
