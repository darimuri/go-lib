package rodtemplate

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

type Login struct {
	*PageTemplate
	Handler LoginHandler
}

func (l *Login) Validate() error {
	if l.Handler.ID == "" {
		l.Handler.ID = os.Getenv(l.Handler.EnvID)
	}

	if l.Handler.Password == "" {
		l.Handler.Password = os.Getenv(l.Handler.EnvPassword)
	}

	if l.Handler.ID == "" || l.Handler.Password == "" {
		return fmt.Errorf("id, password is required as parameter or os environment variables with names(%s, %s)", l.Handler.EnvID, l.Handler.EnvPassword)
	}

	return nil
}

func (l *Login) Submit(b *rod.Browser) error {
	h := l.Handler
	pt := l.PageTemplate

	var loginPt *PageTemplate

	pt.WaitIdle()

	log.Println("find login page")

	for i := 0; i < 10; i++ {
		//find login input selector in iframes
		if pt.Has("iframe") {
			for _, e := range pt.Els("iframe") {
				iFrame, err := e.Frame()
				if err != nil {
					continue
				}

				//to prevent nil pointer reference
				if srcAttr, errAttr := e.Attribute("src"); errAttr != nil {
					log.Println("failed to get iframe src attribute", errAttr)
				} else if srcAttr == nil {
					continue
				} else if strings.HasPrefix(*srcAttr, "http") {
					sameUrl, errSameUrl := IsSameDomainUrl(pt.URL(), *srcAttr)
					if errSameUrl != nil {
						log.Println("failed to parse url", pt.URL(), "error", errSameUrl.Error())
					}
					if !sameUrl {
						log.Println("prevent nil pointer reference on iframe url(", *srcAttr, ")outside owner url domain", pt.URL())
						continue
					}
				}

				if has, _, errHas := iFrame.Has("body"); !has || errHas != nil {
					continue
				}

				_, err = iFrame.Element("body")
				if err != nil {
					errMessage := err.Error()
					if strings.Contains(errMessage, "Frame with the given id was not found.") {
						continue
					} else {
						return err
					}
				}

				myPt := &PageTemplate{P: iFrame}
				//myPt.WaitRequestIdle()

				if myPt.Has(h.LoginInputSelector) {
					loginPt = myPt
					log.Println("found LoginInputSelector", h.LoginLinkSelector, "in iframes")
					break
				}
			}
		}

		//find login input selector in another windows
		if loginPt == nil {
			for _, p := range b.MustPages() {
				if p.FrameID == pt.FrameID() {
					continue
				}

				myPt := &PageTemplate{P: p}

				if myPt.Has(h.LoginInputSelector) {
					loginPt = myPt
					log.Println("found LoginInputSelector", h.LoginLinkSelector, "in pages")
					break
				}
			}
		}

		//find login input selector in page
		if loginPt == nil {
			if pt.Has(h.LoginInputSelector) {
				loginPt = pt
				log.Println("found LoginInputSelector", h.LoginLinkSelector, "in current page")
				break
			}

		}

		if loginPt != nil {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	if loginPt == nil {
		return errors.New("failed to find login input selector " + h.LoginInputSelector)
	}

	loginPageURL := pt.URL()
	if pt != loginPt {
		loginPageURL = loginPt.URL()
	}

	if h.CaptchaHandler != nil {
		errCaptcha := h.CaptchaHandler(loginPt)
		if errCaptcha != nil {
			return errCaptcha
		}
	}

	if l.Handler.LoginBeforeSubmitHandler != nil {
		errBeforeSubmit := l.Handler.LoginBeforeSubmitHandler(loginPt)
		if errBeforeSubmit != nil {
			return errBeforeSubmit
		}
	}

	loginPt.Input(h.LoginInputSelector, h.ID)
	loginPt.Input(h.PasswordInputSelector, h.Password)
	loginPt.PressKey(input.Enter)

	if h.LoginPostSubmitHandler != nil {
		errPostSubmit := h.LoginPostSubmitHandler(pt)
		if errPostSubmit != nil {
			return errPostSubmit
		}
	}

	if h.LoginPostSuccessCheckHandler != nil {
		success, errSuccessCheck := h.LoginPostSuccessCheckHandler(pt)
		if errSuccessCheck != nil {
			return errSuccessCheck
		}

		if !success {
			return fmt.Errorf("login failed for LoginSuccessCheckHandler returned success failed")
		}

		return nil
	}

	for i := 0; i < 100; i++ {
		if !pt.Has(h.LoginSuccessSelector) {
			time.Sleep(time.Millisecond * 100)
			continue
		}

		if str, err := pt.El(h.LoginSuccessSelector).Text(); err != nil {
			if IsObjectNotFoundError(err) {
				continue
			} else {
				return err
			}
		} else if str != "" {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	currentPageURL := pt.URL()

	if loginPageURL == currentPageURL {
		screenshotPath := os.Getenv("PATH_SCREENSHOT")
		screenshotLoginFailed := os.Getenv("SCREENSHOT_LOGIN_FAILED")
		if screenshotLoginFailed == "1" && screenshotPath != "" {
			if _, err := os.Stat(screenshotPath); err != nil {
				if os.IsNotExist(err) {
					if err = os.MkdirAll(screenshotPath, 0755); err != nil {
						log.Println(err)
						return errors.New("failed to login")
					}
				} else {
					log.Println(err)
					return errors.New("failed to login")
				}
			}

			screenshotFile := fmt.Sprintf("loginfailed.%s.png", time.Now().Format("20060102150405"))
			pt.ScreenShot(pt.El("html"), path.Join(screenshotPath, screenshotFile), 0)
		}
		return errors.New("failed to login")
	}

	if h.LoginAfterURL != "" {
		if err := pt.Navigate(h.LoginAfterURL); err != nil {
			return err
		}
	}

	return nil
}
