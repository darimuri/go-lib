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
	LoginBeforeSuccessCheckHandler func(pt *PageTemplate) (bool, error)
	LoginBeforeSubmitHandler func(pt *PageTemplate) error
	LoginPostSubmitHandler       func(pt *PageTemplate) error
	LoginPostSuccessCheckHandler func(pt *PageTemplate) (bool, error)
}

type BrowserTemplate struct {
	*rod.Browser
}

func (b *BrowserTemplate) Login(h LoginHandler) (*PageTemplate, error) {
	var pt *PageTemplate

	log.Println("go to login gate", h.LoginGateURL)

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

	if h.LoginSuccessSelector != "" && pt.Has(h.LoginSuccessSelector) {
		return pt, nil
	} else if h.LoginBeforeSuccessCheckHandler != nil {
		succ, errHandle := h.LoginBeforeSuccessCheckHandler(pt)
		if succ && errHandle == nil {
			return pt, nil
		} else if errHandle != nil {
			log.Printf("failed to check login succes for error %s", errHandle.Error())
		}
	}


	if h.LoginURL != "" {
		log.Println("go to login page", h.LoginURL)
		if err = pt.Navigate(h.LoginURL); err != nil {
			return nil, err
		}
	} else if h.LoginLinkHandler != nil {
		log.Println("go to login page", "with LoginLinkHandler")
		if err = h.LoginLinkHandler(pt); err != nil {
			return nil, err
		}
	} else {
		log.Println("go to login page", "with LoginLinkSelector")
		pt.ClickWhenAvailable(h.LoginLinkSelector)
	}

	login := &Login{PageTemplate: pt, Handler: h}

	log.Println("validate login")
	if err = login.Validate(); err != nil {
		return nil, err
	}

	log.Println("submit login")
	if err = login.Submit(b.Browser); err != nil {
		return nil, err
	}

	return pt, nil
}

func NewBrowserTemplate(b *rod.Browser) *BrowserTemplate {
	return &BrowserTemplate{Browser: b}
}
