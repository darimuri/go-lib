package rodtemplate

import (
	"fmt"
	"log"
	"reflect"
	"time"

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
		if true == checkFunc() {
			return nil
		}

		elapsed := time.Now().Sub(started)
		if timeout < elapsed {
			break
		} else {
			sleepDuration := retryDuration - time.Now().Sub(lastRetry)
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
