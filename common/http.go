package common

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type BasicAuth struct {
	Username string
	Password string
}

type Context struct {
	*BasicAuth
	Method string
	Header http.Header
	Query  url.Values
	Body   io.Reader
}

func GetAuthParameters() []string {
	if Token != "" {
		return []string{Token}
	} else if User != "" && Password != "" {
		return []string{User, Password}
	}
	return nil
}

func Request(requestUrl string, context ...Context) (responseBody io.Reader, responseHeader http.Header, err error) {
	var ctx Context
	if len(context) > 0 {
		ctx = context[0]
	}

	method := http.MethodGet
	if ctx.Method != "" {
		method = ctx.Method
	}

	if ctx.Query != nil && len(ctx.Query) > 0 {
		if p, e := url.Parse(requestUrl); e == nil {
			query := p.Query()
			for k, vv := range ctx.Query {
				for _, v := range vv {
					query.Add(k, v)
				}
			}
			p.RawQuery = query.Encode()
			requestUrl = p.String()
		}
	}

	var request *http.Request
	if request, err = http.NewRequest(method, requestUrl, ctx.Body); err != nil {
		return
	}

	if ctx.BasicAuth != nil {
		request.SetBasicAuth(ctx.BasicAuth.Username, ctx.BasicAuth.Password)
	}

	if ctx.Header != nil {
		for k := range ctx.Header {
			request.Header.Set(k, ctx.Header.Get(k))
		}
	}

	var response *http.Response
	if response, err = new(http.Client).Do(request); err != nil {
		return
	}
	defer response.Body.Close()

	var rb *bytes.Buffer
	if b, e := io.ReadAll(response.Body); e == nil {
		rb = bytes.NewBuffer(b)
	} else {
		rb = bytes.NewBuffer([]byte{})
	}

	responseBody, responseHeader = rb, response.Header

	return
}
