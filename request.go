package wiremock

import (
	"encoding/json"
)

// A Request is the part of StubRule describing the matching of the http request
type Request struct {
	urlMatcher           URLMatcherInterface
	method               string
	headers              map[string]MultiParamMatcherInterface
	queryParams          map[string]MultiParamMatcherInterface
	cookies              map[string]ParamMatcherInterface
	bodyPatterns         []ParamMatcher
	multipartPatterns    []*MultipartPattern
	basicAuthCredentials *struct {
		username string
		password string
	}
}

// NewRequest constructs minimum possible Request
func NewRequest(method string, urlMatcher URLMatcherInterface) *Request {
	return &Request{
		method:     method,
		urlMatcher: urlMatcher,
	}
}

// WithMethod is fluent-setter for http verb
func (r *Request) WithMethod(method string) *Request {
	r.method = method
	return r
}

// WithURLMatched is fluent-setter url matcher
func (r *Request) WithURLMatched(urlMatcher URLMatcherInterface) *Request {
	r.urlMatcher = urlMatcher
	return r
}

// WithBodyPattern adds body pattern to list
func (r *Request) WithBodyPattern(matcher ParamMatcher) *Request {
	r.bodyPatterns = append(r.bodyPatterns, matcher)
	return r
}

// WithMultipartPattern adds multipart pattern to list
func (r *Request) WithMultipartPattern(pattern *MultipartPattern) *Request {
	r.multipartPatterns = append(r.multipartPatterns, pattern)
	return r
}

// WithBasicAuth adds basic auth credentials to Request
func (r *Request) WithBasicAuth(username, password string) *Request {
	r.basicAuthCredentials = &struct {
		username string
		password string
	}{
		username: username,
		password: password,
	}
	return r
}

// WithQueryParam add param to query param list
func (r *Request) WithQueryParam(param string, matcher ParamMatcherInterface) *Request {
	if r.queryParams == nil {
		r.queryParams = map[string]MultiParamMatcherInterface{}
	}

	r.queryParams[param] = ToMultiParamMatcher(matcher)
	return r
}

// WithQueryParams add param to query param list
func (r *Request) WithQueryParams(param string, matcher MultiParamMatcherInterface) *Request {
	if r.queryParams == nil {
		r.queryParams = map[string]MultiParamMatcherInterface{}
	}

	r.queryParams[param] = matcher
	return r
}

// WithHeader add header to header list
func (r *Request) WithHeader(header string, matcher ParamMatcherInterface) *Request {
	if r.headers == nil {
		r.headers = map[string]MultiParamMatcherInterface{}
	}

	r.headers[header] = ToMultiParamMatcher(matcher)
	return r
}

// WithHeaders add header to header list
func (r *Request) WithHeaders(header string, matcher MultiParamMatcherInterface) *Request {
	if r.headers == nil {
		r.headers = map[string]MultiParamMatcherInterface{}
	}

	r.headers[header] = matcher
	return r
}

// WithCookie is fluent-setter for cookie
func (r *Request) WithCookie(cookie string, matcher ParamMatcherInterface) *Request {
	if r.cookies == nil {
		r.cookies = map[string]ParamMatcherInterface{}
	}

	r.cookies[cookie] = matcher
	return r
}

// MarshalJSON gives valid JSON or error.
func (r *Request) MarshalJSON() ([]byte, error) {
	request := map[string]interface{}{
		"method":                        r.method,
		string(r.urlMatcher.Strategy()): r.urlMatcher.Value(),
	}
	if len(r.bodyPatterns) > 0 {
		bodyPatterns := make([]map[string]interface{}, len(r.bodyPatterns))
		for i, bodyPattern := range r.bodyPatterns {
			bodyPatterns[i] = map[string]interface{}{
				string(bodyPattern.Strategy()): bodyPattern.Value(),
			}

			for flag, value := range bodyPattern.flags {
				bodyPatterns[i][flag] = value
			}
		}
		request["bodyPatterns"] = bodyPatterns
	}
	if len(r.multipartPatterns) > 0 {
		request["multipartPatterns"] = r.multipartPatterns
	}
	if len(r.headers) > 0 {
		headers := make(map[string]map[string]interface{}, len(r.headers))
		for key, header := range r.headers {
			if header.IsSingleParam() {
				headers[key] = map[string]interface{}{
					string(header.Strategy()): header.FirstValue(),
				}
			} else {
				headers[key] = map[string]interface{}{
					string(header.Strategy()): make([]interface{}, 0, header.Length()),
				}

				subKey := headers[key][string(header.Strategy())].([]interface{})
				for _, v := range header.Values() {
					subKey = append(subKey, map[string]string{
						string(v.Strategy()): v.Value(),
					})
				}
				headers[key][string(header.Strategy())] = subKey
			}

			for flag, value := range header.Flags() {
				headers[key][flag] = value
			}
		}
		request["headers"] = headers
	}
	if len(r.cookies) > 0 {
		cookies := make(map[string]map[string]interface{}, len(r.cookies))
		for key, cookie := range r.cookies {
			cookies[key] = map[string]interface{}{
				string(cookie.Strategy()): cookie.Value(),
			}

			for flag, value := range cookie.Flags() {
				cookies[key][flag] = value
			}
		}
		request["cookies"] = cookies
	}
	if len(r.queryParams) > 0 {
		params := make(map[string]map[string]interface{}, len(r.queryParams))
		for key, param := range r.queryParams {
			if param.IsSingleParam() {
				params[key] = map[string]interface{}{
					string(param.Strategy()): param.FirstValue(),
				}
			} else {
				params[key] = map[string]interface{}{
					string(param.Strategy()): make([]map[string]string, 0, param.Length()),
				}

				subKey := params[key][string(param.Strategy())].([]map[string]string)
				for _, v := range param.Values() {
					subKey = append(subKey, map[string]string{
						string(v.Strategy()): v.Value(),
					})
				}
				params[key][string(param.Strategy())] = subKey
			}

			for flag, value := range param.Flags() {
				params[key][flag] = value
			}
		}
		request["queryParameters"] = params
	}

	if r.basicAuthCredentials != nil {
		request["basicAuthCredentials"] = map[string]string{
			"password": r.basicAuthCredentials.password,
			"username": r.basicAuthCredentials.username,
		}
	}

	return json.Marshal(request)
}
