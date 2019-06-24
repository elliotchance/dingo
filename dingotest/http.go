package dingotest

import (
	"net/http"
	"time"
)

type HTTPSignerClient struct {
	CreateSigner func(req *http.Request) *Signer
}

func (c *HTTPSignerClient) Do(req *http.Request) (*http.Response, error) {
	signer := c.CreateSigner(req)
	req.Header.Set("Authorization", signer.Auth())

	return http.DefaultClient.Do(req)
}

type Signer struct {
	req *http.Request
}

func NewSigner(req *http.Request) *Signer {
	return &Signer{req: req}
}

// Produces something like "Mon Jan 2 15:04:05 2006 POST"
func (signer *Signer) Auth() string {
	return time.Now().Format(time.ANSIC) + " " + signer.req.Method
}
