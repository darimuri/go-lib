package credential

import (
	"log"
	"testing"

	"github.com/docker/docker-credential-helpers/credentials"
)

// Uses 자격 증명 관리자 on windows
func TestSetGet(t *testing.T) {
	url := "github.com/dormael/naver-shop-click"
	lbl := "naver-shop-click"

	user, secret, err := Get(lbl, url)
	if err != nil && false == credentials.IsErrCredentialsNotFound(err) {
		t.Errorf("Expecting CredentialNotFound error, got %v", err)
	}

	if user != "" {
		t.Errorf("Expecting empey user, got %s", user)
	}

	if secret != "" {
		t.Errorf("Expecting empey secret, got %s", secret)
	}

	err = Set(lbl, url, "user", "password")
	if err != nil {
		t.Errorf("Expecting empty error, got %v", err)
	}

	user, secret, err = Get(lbl, url)
	if err == nil {
		if user != "user" {
			t.Errorf("Expecting user, got %s", user)
		}

		if secret != "password" {
			t.Errorf("Expecting password, got %s", secret)
		}
	} else {
		log.Println("got error:", err)
	}

	err = Del(lbl, url)
	if err != nil {
		t.Errorf("Expecting empty error, got %v", err)
	}
}
