package rodtemplate

import (
	"log"

	"github.com/go-rod/rod"
)

type LoginHandler struct {
	LoginGateURL          string
	LoginAfterURL         string
	LoginLinkSelector     string
	LoginInputSelector    string
	PasswordInputSelector string

	LoginURL string

	LoginSuccessSelector string

	ID       string
	Password string
	EnvID    string

	EnvPassword string

	CaptchaHandler           func(pt *PageTemplate) error
	LoginLinkHandler         func(pt *PageTemplate) error
	LoginBeforeSubmitHandler func(pt *PageTemplate) error
	LoginPostSubmitHandler   func(pt *PageTemplate) error
	LoginSuccessCheckHandler func(pt *PageTemplate) (bool, error)
}

type BrowserTemplate struct {
	*rod.Browser
}

func (b *BrowserTemplate) Login(h LoginHandler) (*PageTemplate, error) {
	var pt *PageTemplate

	page := b.MustPage(h.LoginGateURL)
	page.MustWaitRequestIdle()

	pt = &PageTemplate{P: page}
	pt.MaximizeToWindowBounds()

	pages, err := b.Browser.Pages()
	if err != nil {
		return nil, err
	}

	for _, p := range pages {
		if p.FrameID == pt.P.FrameID {
			continue
		}
		p.MustClose()
	}

	pt.WaitRequestIdle()

	if h.LoginSuccessSelector != "" && pt.Has(h.LoginSuccessSelector) {
		return pt, nil
	} else if h.LoginSuccessCheckHandler != nil {
		succ, errHandle := h.LoginSuccessCheckHandler(pt)
		if succ && errHandle == nil {
			return pt, nil
		} else if errHandle != nil {
			log.Printf("failed to check login succes for error %s", errHandle.Error())
		}
	}

	if h.LoginURL != "" {
		if err = pt.Navigate(h.LoginURL); err != nil {
			return nil, err
		}
	} else if h.LoginLinkHandler != nil {
		if err = h.LoginLinkHandler(pt); err != nil {
			return nil, err
		}
	} else {
		pt.ClickWhenAvailable(h.LoginLinkSelector)
	}

	login := &Login{PageTemplate: pt, Handler: h}

	if err = login.Validate(); err != nil {
		return nil, err
	}

	if err = login.Submit(b.Browser); err != nil {
		return nil, err
	}

	return pt, nil
}

func NewBrowserTemplate(b *rod.Browser) *BrowserTemplate {
	return &BrowserTemplate{Browser: b}
}
