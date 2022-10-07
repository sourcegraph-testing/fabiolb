package route

import (
	gkm "github.com/go-kit/kit/metrics"
	"net/http"
	"net/url"
	"strings"
)

type Target struct {
	// Service is the name of the service the targetURL points to
	Service string

	// Tags are the list of tags for this target
	Tags []string

	// Opts is the raw options for the target.
	Opts map[string]string

	// StripPath will be removed from the front of the outgoing
	// request path
	StripPath string

	// PrependPath will be added to the front of the outgoing
	// request path (after StripPath has been removed)
	PrependPath string

	// TLSSkipVerify disables certificate validation for upstream
	// TLS connections.
	TLSSkipVerify bool

	// Host signifies what the proxy will set the Host header to.
	// The proxy does not modify the Host header by default.
	// When Host is set to 'dst' the proxy will use the host name
	// of the target host for the outgoing request.
	Host string

	// URL is the endpoint the service instance listens on
	URL *url.URL

	// RedirectCode is the HTTP status code used for redirects.
	// When set to a value > 0 the client is redirected to the target url.
	RedirectCode int

	// RedirectURL is the redirect target based on the request.
	// This is cached here to prevent multiple generations per request.
	RedirectURL *url.URL

	// FixedWeight is the weight assigned to this target.
	// If the value is 0 the targets weight is dynamic.
	FixedWeight float64

	// Weight is the actual weight for this service in percent.
	Weight float64

	// Histogram measures throughput and latency of this target
	Timer gkm.Histogram

	// Counters for rx and tx
	RxCounter gkm.Counter
	TxCounter gkm.Counter

	// accessRules is map of access information for the target.
	accessRules map[string][]any

	// name of the auth handler for this target
	AuthScheme string

	// ProxyProto enables PROXY Protocol on upstream connection
	ProxyProto bool

	// Transport allows for different types of transports
	Transport *http.Transport
}

func (t *Target) BuildRedirectURL(requestURL *url.URL) {
	t.RedirectURL = &url.URL{
		Scheme:   t.URL.Scheme,
		Host:     t.URL.Host,
		Path:     t.URL.Path,
		RawPath:  t.URL.Path,
		RawQuery: t.URL.RawQuery,
	}
	// treat case of $path not separated with a / from host
	if strings.HasSuffix(t.RedirectURL.Host, "$path") {
		t.RedirectURL.Host = t.RedirectURL.Host[:len(t.RedirectURL.Host)-len("$path")]
		t.RedirectURL.Path = "$path"
	}
	// remove / before $path in redirect url
	if strings.Contains(t.RedirectURL.Path, "/$path") {
		t.RedirectURL.Path = strings.Replace(t.RedirectURL.Path, "/$path", "$path", 1)
		t.RedirectURL.RawPath = strings.Replace(t.RedirectURL.RawPath, "/$path", "$path", 1)
	}
	// remove strip path, insert passed request path, set query
	if strings.Contains(t.RedirectURL.Path, "$path") {
		// set replacement paths
		replacePath := requestURL.Path
		var replaceRawPath string
		if requestURL.RawPath == "" {
			replaceRawPath = requestURL.Path
		} else {
			replaceRawPath = requestURL.RawPath
		}
		// strip path before replacement
		if t.StripPath != "" {
			if strings.HasPrefix(replacePath, t.StripPath) {
				replacePath = replacePath[len(t.StripPath):]
			}
			if strings.HasPrefix(replaceRawPath, t.StripPath) {
				replaceRawPath = replaceRawPath[len(t.StripPath):]
			}
		}
		// add prepend path
		if t.PrependPath != "" {
			replacePath = t.PrependPath + replacePath
			replaceRawPath = t.PrependPath + replaceRawPath
		}
		// do path replacement
		t.RedirectURL.Path = strings.Replace(t.RedirectURL.Path, "$path", replacePath, 1)
		t.RedirectURL.RawPath = strings.Replace(t.RedirectURL.RawPath, "$path", replaceRawPath, 1)
		// set query
		if t.RedirectURL.RawQuery == "" && requestURL.RawQuery != "" {
			t.RedirectURL.RawQuery = requestURL.RawQuery
		}
	}
	if t.RedirectURL.Path == "" {
		t.RedirectURL.Path = "/"
	}
	if strings.Contains(t.RedirectURL.Host, "$host") {
		t.RedirectURL.Host = strings.Replace(t.RedirectURL.Host, "$host", requestURL.Host, 1)
	}
}
