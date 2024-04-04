package httpclient

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/nlnwa/whatwg-url/url"
)

func ptr[T any](v T) *T {
	return &v
}

type HttpCodes int

const (
	Ok                          HttpCodes = 200
	MultipleChoices             HttpCodes = 300
	MovedPermanently            HttpCodes = 301
	ResourceMoved               HttpCodes = 302
	SeeOther                    HttpCodes = 303
	NotModified                 HttpCodes = 304
	UseProxy                    HttpCodes = 305
	SwitchProxy                 HttpCodes = 306
	TemporaryRedirect           HttpCodes = 307
	PermanentRedirect           HttpCodes = 308
	BadRequest                  HttpCodes = 400
	Unauthorized                HttpCodes = 401
	PaymentRequired             HttpCodes = 402
	Forbidden                   HttpCodes = 403
	NotFound                    HttpCodes = 404
	MethodNotAllowed            HttpCodes = 405
	NotAcceptable               HttpCodes = 406
	ProxyAuthenticationRequired HttpCodes = 407
	RequestTimeout              HttpCodes = 408
	Conflict                    HttpCodes = 409
	Gone                        HttpCodes = 410
	TooManyRequests             HttpCodes = 429
	InternalServerError         HttpCodes = 500
	NotImplemented              HttpCodes = 501
	BadGateway                  HttpCodes = 502
	ServiceUnavailable          HttpCodes = 503
	GatewayTimeout              HttpCodes = 504
)

type Headers string

const (
	Accept      Headers = "accept"
	ContentType Headers = "content-type"
)

type MediaTypes string

const (
	ApplicationJson MediaTypes = "application/json"
)

func GetProxyUrl(serverUrl string) (string, error) {
	serverUrlParsed, err := url.Parse(serverUrl)
	if err != nil {
		return "", err
	}
	if checkBypass(serverUrlParsed) {
		return "", nil
	}

	var proxyVar string
	if serverUrlParsed.Protocol() == "https" {
		proxyVar = os.Getenv("https_proxy")
		if proxyVar == "" {
			proxyVar = os.Getenv("HTTPS_PROXY")
		}
	} else {
		proxyVar = os.Getenv("http_proxy")
		if proxyVar == "" {
			proxyVar = os.Getenv("HTTP_PROXY")
		}
	}
	if proxyVar == "" {
		return "", nil
	} else {
		if proxyVarUrl, err := url.Parse(proxyVar); err == nil {
			return proxyVarUrl.Href(false), nil
		} else if !strings.HasPrefix(proxyVar, "http:") && !strings.HasPrefix(proxyVar, "https:") {
			return "http://" + proxyVar, nil
		} else {
			return "", nil
		}
	}
}

func checkBypass(url *url.Url) bool {
	if url.Hostname() == "" {
		return false
	}
	if isLoopbackAddress(url) {
		return true
	}
	
	noProxy := os.Getenv("no_proxy")
	if noProxy != "" {
		noProxy = os.Getenv("NO_PROXY")
	}
	if noProxy == "" {
		return false
	}

	var port *int
	if url.Port() != "" {
		port = ptr(url.DecodedPort())
	} else if url.Protocol() == "http" {
		port = ptr(80)
	} else if url.Protocol() == "https" {
		port = ptr(443)
	}
	
	upperHosts := []string{strings.ToUpper(url.Hostname())}
	if port != nil {
		additionalHost := upperHosts[0] + ":" + string(*port)
		upperHosts = append(upperHosts, additionalHost)
	}

	noProxyList := strings.Split(noProxy, ",")
	for i, v := range noProxyList {
		noProxyList[i] = strings.ToUpper(strings.TrimSpace(v))
	}
	noProxyList2 := []string{}
	for _, v := range noProxyList {
		if v != "" {
			noProxyList2 = append(noProxyList2, v)
		}
	}

	for _, noProxy := range noProxyList2 {
		if noProxy == "*" {
			return true
		}
		for _, host := range upperHosts {
			if host == noProxy {
				return true
			}
			if strings.HasSuffix(host, "." + noProxy) {
				return true
			}
			if strings.HasPrefix(noProxy, ".") && strings.HasSuffix(host, noProxy) {
				return true
			}
		}
	}
	return false
}

func isLoopbackAddress(url *url.Url) bool {
	hostLower := strings.ToLower(url.Host())
	return hostLower == "localhost" || strings.HasPrefix(hostLower, "127.") || strings.HasPrefix(hostLower, "[::1]") || strings.HasPrefix(hostLower, "[0:0:0:0:0:0:0:1]")
}

type HttpClientError struct {
	Message string
	Name string
	StatusCode int
	result any
}

func NewHttpClientError(message string, statusCode int) *HttpClientError {
	return &HttpClientError{
		Message: message,
		Name: "HttpClientError",
		StatusCode: statusCode,
	}
}

func (e *HttpClientError) Error() string {
	return e.Name + ": " + e.Message
}

type HttpClientResponse struct {
	Message *http.Request
}

func NewHttpClientResponse(message *http.Request) *HttpClientResponse {
	return &HttpClientResponse{
		Message: message,
	}
}

func (r *HttpClientResponse) ReadBody() (string, error) {
	bytes, err := io.ReadAll(r.Message.Body)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (r *HttpClientResponse) ReadBodyBuffer() ([]byte, error) {
	return io.ReadAll(r.Message.Body)
}

func (r *HttpClientResponse) IsHttps() bool {
	return r.Message.URL.Scheme == "https"
}

type HttpClient struct {
	UserAgent *string
	Handlers []RequestHandler
	RequestOptions *RequestOptions
	ignoreSslError bool
	socketTimeout *int
	allowRedirects bool
	allowRedirectDowngrade bool
	maxRedirects int
	allowRetries bool
	maxRetries int
	agent any
	proxyAgent any
	proxyAgentDispatcher any
	keepAlive bool
	disposed bool
}

func NewHttpClient(userAgent *string, handlers []RequestHandler, requestOptions *RequestOptions) *HttpClient {
	handlers2 := []RequestHandler{}
	if handlers != nil {
		handlers2 = handlers
	}
	self := &HttpClient{
		UserAgent: userAgent,
		Handlers: handlers2,
		RequestOptions: requestOptions,
		ignoreSslError: false,
		socketTimeout: nil,
		allowRedirects: true,
		allowRedirectDowngrade: false,
		maxRedirects: 50,
		allowRetries: false,
		maxRetries: 1,
		agent: nil,
		proxyAgent: nil,
		proxyAgentDispatcher: nil,
		keepAlive: false,
		disposed: false,
	}
	if requestOptions != nil {
		if requestOptions.IgnoreSslError != nil {
			self.ignoreSslError = *requestOptions.IgnoreSslError
		}
		self.socketTimeout = requestOptions.SocketTimeout
		if requestOptions.AllowRedirects != nil {
			self.allowRedirects = *requestOptions.AllowRedirects
		}
		if requestOptions.AllowRedirectDowngrade != nil {
			self.allowRedirectDowngrade = *requestOptions.AllowRedirectDowngrade
		}
		if requestOptions.MaxRedirects != nil {
			self.maxRedirects = max(0, *requestOptions.MaxRedirects)
		}
		if requestOptions.KeepAlive != nil {
			self.keepAlive = *requestOptions.KeepAlive
		}
		if requestOptions.AllowRetries != nil {
			self.allowRetries = *requestOptions.AllowRetries
		}
		if requestOptions.MaxRetries != nil {
			self.maxRetries = max(0, *requestOptions.MaxRetries)
		}
	}
	return self
}

func (c *HttpClient) Options(requestUrl string, additionalHeaders map[string]string) (*HttpClientResponse, error) {
	additionalHeaders2 := map[string]string{}
	if additionalHeaders != nil {
		additionalHeaders2 = additionalHeaders
	}
	return c.Request("OPTIONS", requestUrl, nil, additionalHeaders2)
}

func (c *HttpClient) Get(requestUrl string, additionalHeaders map[string]string) (*HttpClientResponse, error) {
	additionalHeaders2 := map[string]string{}
	if additionalHeaders != nil {
		additionalHeaders2 = additionalHeaders
	}
	return c.Request("GET", requestUrl, nil, additionalHeaders2)
}

func (c *HttpClient) Del(requestUrl string, additionalHeaders map[string]string) (*HttpClientResponse, error) {
	additionalHeaders2 := map[string]string{}
	if additionalHeaders != nil {
		additionalHeaders2 = additionalHeaders
	}
	return c.Request("DELETE", requestUrl, nil, additionalHeaders2)
}

func (c *HttpClient) Post(requestUrl string, data any, additionalHeaders map[string]string) (*HttpClientResponse, error) {
	additionalHeaders2 := map[string]string{}
	if additionalHeaders != nil {
		additionalHeaders2 = additionalHeaders
	}
	return c.Request("POST", requestUrl, data, additionalHeaders2)
}

func (c *HttpClient) Patch(requestUrl string, data any, additionalHeaders map[string]string) (*HttpClientResponse, error) {
	additionalHeaders2 := map[string]string{}
	if additionalHeaders != nil {
		additionalHeaders2 = additionalHeaders
	}
	return c.Request("PATCH", requestUrl, data, additionalHeaders2)
}

func (c *HttpClient) Put(requestUrl string, data any, additionalHeaders map[string]string) (*HttpClientResponse, error) {
	additionalHeaders2 := map[string]string{}
	if additionalHeaders != nil {
		additionalHeaders2 = additionalHeaders
	}
	return c.Request("PUT", requestUrl, data, additionalHeaders2)
}

func (c *HttpClient) Head(requestUrl string, additionalHeaders map[string]string) (*HttpClientResponse, error) {
	additionalHeaders2 := map[string]string{}
	if additionalHeaders != nil {
		additionalHeaders2 = additionalHeaders
	}
	return c.Request("HEAD", requestUrl, nil, additionalHeaders2)
}

func (c *HttpClient) SendStream(verb string, requestUrl string, stream io.Reader) (*HttpClientResponse, error) {
	return c.Request(verb, requestUrl, stream, nil)
}

func (c *HttpClient) GetJson(requestUrl string, additionalHeaders map[string]string) (any, error) {
	res, err := c.Get(requestUrl, additionalHeaders)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *HttpClient) PostJson(requestUrl string, data any, additionalHeaders map[string]string) (any, error) {
	res, err := c.Post(requestUrl, data, additionalHeaders)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *HttpClient) PutJson(requestUrl string, data any, additionalHeaders map[string]string) (any, error) {
	res, err := c.Put(requestUrl, data, additionalHeaders)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *HttpClient) PatchJson(requestUrl string, data any, additionalHeaders map[string]string) (any, error) {
	res, err := c.Patch(requestUrl, data, additionalHeaders)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *HttpClient) Request(verb string, requestUrl string, data any, headers map[string]string) (*HttpClientResponse, error) {
	if c.disposed {
		return nil, errors.New("client has already been disposed")
	}

	parsedUrl, err := url.Parse(requestUrl)
	if err != nil {
		return nil, err
	}
	info := map[string]string{}

	maxTries := 1
	if c.allowRetries && slices.Contains([]string{"OPTIONS", "GET", "DELETE", "HEAD"}, verb) {
		maxTries = c.maxRetries + 1
	}

	var response *HttpClientResponse
	return response, nil
}

func (c *HttpClient) Dispose() {
	c.disposed = true
}

func (c *HttpClient) RequestRaw(info any, data any) (*HttpClientResponse, error) {
	return nil, nil
}

func (c *HttpClient) RequestRawWithCallback(info any, data any, callback func(error, *HttpClientResponse)) {
}

func (c *HttpClient) GetAgent(serverUrl string) any {
	return nil
}

func (c *HttpClient) GetAgentDispatcher(serverUrl string) any {
	return nil
}

type RequestHandler interface {
	PrepareRequest(options any)
	CanHandleAuthentication(response *HttpClientResponse) bool
	HandleAuthentication(httpClient *HttpClient, requestInfo any, data any) (*HttpClientResponse, error)
}

type RequestOptions struct {
	Headers map[string]string
	SocketTimeout *int
	IgnoreSslError *bool
	AllowRedirects *bool
	AllowRedirectDowngrade *bool
	MaxRedirects *int
	MaxSockets *int
	KeepAlive *bool
	// DeserializeDates *bool
	AllowRetries *bool
	MaxRetries *int
}