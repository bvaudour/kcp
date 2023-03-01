package database

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bvaudour/kcp/common"
)

type ctx struct {
	username string
	password string
	accept   string
}

func execRequest(url string, opt ctx, args ...any) (io.Reader, error) {
	request, err := http.NewRequest("GET", fmt.Sprintf(url, args...), nil)
	if err != nil {
		return nil, err
	}
	if opt.username != "" && opt.password != "" {
		request.SetBasicAuth(opt.username, opt.password)
	}
	if opt.accept != "" {
		request.Header.Set("Accept", opt.accept)
	}
	response, err := new(http.Client).Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	b, err := io.ReadAll(response.Body)
	if err == nil {
		return bytes.NewBuffer(b), nil
	}
	return bytes.NewBuffer([]byte{}), nil
}

func sliceToSet(sl []string) map[string]bool {
	out := make(map[string]bool)
	for _, e := range sl {
		out[e] = true
	}

	return out
}

// Counter is a counter of updated packages.
type Counter struct {
	Updated int
	Deleted int
	Added   int
}

// String returns the string representation of the counter
func (c Counter) String() string {
	var out []string
	if c.Added > 0 {
		out = append(out, common.Tr(msgAdded, c.Added))
	}
	if c.Deleted > 0 {
		out = append(out, common.Tr(msgDeleted, c.Deleted))
	}
	if c.Updated > 0 {
		out = append(out, common.Tr(msgUpdated, c.Updated))
	}

	return strings.Join(out, "\n")
}
