package credential

import "github.com/docker/docker-credential-helpers/credentials"

func Set(lbl, url, user, secret string) error {
	cr := &credentials.Credentials{
		ServerURL: url,
		Username:  user,
		Secret:    secret,
	}

	credentials.SetCredsLabel(lbl)
	return ns.Add(cr)
}

func Get(lbl, url string) (string, string, error) {
	credentials.SetCredsLabel(lbl)
	return ns.Get(url)
}

func Del(lbl, url string) error {
	credentials.SetCredsLabel(lbl)
	return ns.Delete(url)
}
