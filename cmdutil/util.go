package cmdutil

import (
	"fmt"
	"strings"

	"github.com/docker/docker-credential-helpers/credentials"
	"github.com/manifoldco/promptui"

	"github.com/darimuri/go-lib/credential"
)

func PromptInput(prompt string) (string, error) {
	p := promptui.Prompt{
		Label: prompt,
	}

	result, err := p.Run()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

func PromptPassword(prompt string) (string, error) {
	p := promptui.Prompt{
		Label: prompt,
		Mask:  '*',
	}

	result, err := p.Run()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(result), nil
}

func PromptYN(prompt string) string {
	p := promptui.Prompt{
		Label:     prompt,
		IsConfirm: true,
	}

	result, _ := p.Run()

	return strings.TrimSpace(result)
}

func ResetCredential(credLabel string, credURL string) {
	if err := credential.Del(credLabel, credURL); err != nil {
		panic(err)
	}
}

func GetCredential(credLabel string, credURL string) (string, string) {
	id, pass, err := credential.Get(credLabel, credURL)
	if err != nil && false == credentials.IsErrCredentialsNotFound(err) {
		handleCredentialError(err)
	}

	if id == "" || pass == "" {
		fmt.Println("ID, Password is required for", credLabel, credURL)

		if id == "" {
			id, err = PromptInput("ID")
			if err != nil {
				panic(err)
			}
		}

		if pass == "" {
			pass, err = PromptPassword("Password")
			if err != nil {
				panic(err)
			}
		}

		yn := PromptYN("Store ID, Password?")

		if strings.ToLower(yn) == "y" {
			if err := credential.Set(credLabel, credURL, id, pass); err != nil {
				handleCredentialError(err)
			}
		}
	}

	return id, pass
}

func handleCredentialError(err error) {
	errorString := strings.TrimSpace(err.Error())
	if strings.HasSuffix(errorString, `executable file not found in $PATH:`) {
		panic(`
pass should be installed.
Refer Download section in https://www.passwordstore.org/ to install pass.`)
	} else if strings.HasSuffix(errorString, `Error: password store is empty. Try "pass init".`) {
		panic(`
pass is not initialized.
Init pass on command line as follows.
---
gpg --full-generate-key
...
public and secret key created and signed.

pub   rsa4096 2020-12-25 [SC]
      XXXXXXXXXXXXXXXX <- gpg-id
uid                      Your Name (Comment) <Email Address>
sub   rsa4096 2020-12-25 [E]
...
pass init gpg-id
---`)
	}
	panic(err)
}
