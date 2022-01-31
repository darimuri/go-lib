package rodtemplate

import (
	"fmt"
	"log"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-rod/rod"
)

type TimeoutError struct {
	Timout  time.Duration
	Started time.Time
	Message string
}

func (e TimeoutError) Error() string {
	return e.Message
}

func (e TimeoutError) Timeout() bool {
	return true
}

func WaitFor(targetName string, timeout, retryDuration time.Duration, checkFunc func() bool, retryFunc func()) error {
	started := time.Now()
	lastRetry := time.Now()

	for {
		if checkFunc() {
			return nil
		}

		elapsed := time.Since(started)
		if timeout < elapsed {
			break
		} else {
			sleepDuration := retryDuration - time.Since(lastRetry)
			log.Println("retry after sleep", sleepDuration, "for retryDuration", retryDuration, "waiting for", targetName)
			if sleepDuration < 0 {
				sleepDuration = retryDuration
			}
			time.Sleep(sleepDuration)
			retryFunc()
			lastRetry = time.Now()
		}
	}

	message := fmt.Sprintf("timeout %s exceeded after %s waiting for %s", timeout, started, targetName)
	log.Println(message)

	return &TimeoutError{Timout: timeout, Started: started, Message: message}
}

var errNotFound = &rod.ErrObjectNotFound{}

func IsObjectNotFoundError(err error) bool {
	return reflect.TypeOf(errNotFound) == reflect.TypeOf(err)
}

var errCovered = &rod.ErrCovered{}

func IsCoveredError(err error) bool {
	return reflect.TypeOf(errCovered) == reflect.TypeOf(err)
}

func IsSameDomainUrl(url1, url2 string) (bool, error) {
	pUrl1, errParseUrl1 := url.Parse(url1)
	if errParseUrl1 != nil {
		return false, errParseUrl1
	}

	pUrl2, errParseUrl2 := url.Parse(url2)
	if errParseUrl2 != nil {
		return false, errParseUrl2
	}

	host1 := pUrl1.Hostname()
	host2 := pUrl2.Hostname()

	if !govalidator.IsDNSName(host1) || !govalidator.IsDNSName(host2) {
		return host1 == host2, nil
	}

	host1Split := strings.Split(host1, ".")
	host2Split := strings.Split(host2, ".")

	host1Domain := strings.Join(host1Split[len(host1Split)-2:], ".")
	host2Domain := strings.Join(host2Split[len(host2Split)-2:], ".")

	return host1Domain == host2Domain, nil
}