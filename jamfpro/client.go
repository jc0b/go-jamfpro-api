package jamfpro

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/pkg/errors"
)

const (
	uriAuthToken  = "/api/v1/auth/token"
	uriOAuthToken = "/api/oauth/token"
)

// Client ... stores an object to talk with Jamf API
type Client struct {
	username, password, clientId, clientSecret string
	token                                      *string
	tokenExpiration                            *time.Time

	instanceUrl *url.URL

	// The Http Client that is used to make requests
	client           *http.Client
	HttpRetryTimeout time.Duration

	Buildings  BuildingsService
	Categories CategoriesService

	// Option to specify extra headers like User-Agent
	ExtraHeader map[string]string
}

// Response is a Zentral response. This wraps the standard http.Response returned from Zentral.
type Response struct {
	*http.Response
}

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string `json:"message"`
}

type responseAuthToken struct {
	Token   *string    `json:"token,omitempty"`
	Expires *time.Time `json:"expires,omitempty"`
}

type responseOAuthToken struct {
	AccessToken *string `json:"access_token,omitempty"`
	Scope       *string `json:"scope,omitempty"`
	TokenType   *string `json:"token_type,omitempty"`
	ExpiresIn   *int64  `json:"expires_in,omitempty"`
}

type FormOptions struct {
	ClientId     string `url:"client_id"`
	ClientSecret string `url:"client_secret"`
	GrantType    string `url:"grant_type"`
}

// NewClient ... returns a new jamf.Client which can be used to access the API using the new bearer tokens
func NewClient(clientId, clientSecret, instance string) (*Client, error) {
	instanceUrl, err := url.Parse(instance)

	if err != nil {
		return nil, err
	}
	c := &Client{
		clientId:         clientId,
		clientSecret:     clientSecret,
		instanceUrl:      instanceUrl,
		token:            nil,
		client:           http.DefaultClient,
		HttpRetryTimeout: 60 * time.Second,
		ExtraHeader:      make(map[string]string),
	}

	c.Buildings = &BuildingsServiceOp{client: c}
	c.Categories = &CategoriesServiceOp{client: c}

	if err := c.refreshAuthToken(); err != nil {
		return c, errors.Wrap(err, "Error getting bearer auth token")
	}

	return c, nil
}

func (c *Client) refreshAuthToken() error {
	if c.tokenExpiration != nil {
		if c.tokenExpiration.After(time.Now()) {
			return nil
		}
	}

	c.token = nil

	var out *responseOAuthToken
	data := url.Values{}
	data.Set("client_id", c.clientId)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "client_credentials")

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, c.instanceUrl.String()+uriOAuthToken, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	decodeErr := json.NewDecoder(resp.Body).Decode(&out)
	if decodeErr != nil {
		return nil
	}
	c.token = out.AccessToken
	expiration := time.Now().Add(time.Duration(*out.ExpiresIn) * time.Second)
	c.tokenExpiration = &expiration

	return nil
}

func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.instanceUrl.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var request *http.Request
	var mediaType string
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		request, err = http.NewRequest(method, u.String(), nil)
		if err != nil {
			return nil, err
		}

	default:
		buf := new(bytes.Buffer)
		if body != nil {
			switch {
			case strings.Contains(urlStr, "JSSResource"):
				err := xml.NewEncoder(buf).Encode(body)
				if err != nil {
					return nil, err
				}
				mediaType = "application/xml"
			case strings.Contains(urlStr, uriOAuthToken):
				b, err := query.Values(body)

				if err != nil {
					return nil, err
				}

				buf = bytes.NewBufferString(b.Encode())
				mediaType = "application/x-www-form-urlencoded"
			default:
				err = json.NewEncoder(buf).Encode(body)
				if err != nil {
					return nil, err
				}
				mediaType = "application/json"
			}
		}

		request, err = http.NewRequest(method, u.String(), buf)
		if err != nil {
			return nil, err
		}
		request.Header.Set("Content-Type", mediaType)
	}
	request.Header.Set("Authorization", "Bearer "+*c.token)
	return request, nil
}

// newResponse creates a new Response for the provided http.Response
func newResponse(r *http.Response) *Response {
	response := Response{Response: r}

	return &response
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*Response, error) {
	resp, err := DoRequestWithClient(ctx, c.client, req)
	if err != nil {
		return nil, err
	}

	defer func() {
		// Ensure the response body is fully read and closed
		// before we reconnect, so that we reuse the same TCPConnection.
		// Close the previous response's body. But read at least some of
		// the body so if it's small the underlying TCP connection will be
		// re-used. No need to check for errors: if it fails, the Transport
		// won't reuse it anyway.
		const maxBodySlurpSize = 2 << 10
		if resp.ContentLength == -1 || resp.ContentLength <= maxBodySlurpSize {
			io.CopyN(io.Discard, resp.Body, maxBodySlurpSize)
		}

		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	response := newResponse(resp)

	err = CheckResponse(resp)
	if err != nil {
		return response, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				println("Blah")
				return nil, err
			}
		}
	}

	return response, err
}

// DoRequestWithClient submits an HTTP request using the specified client.
func DoRequestWithClient(
	ctx context.Context,
	client *http.Client,
	req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}

// CheckResponse checks the API response for errors, and returns them if present. A response is considered an
// error if it has a status code outside the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other response body will be silently ignored.
// If the API error response does not include the request ID in its body, the one from its header will be used.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		errorResponse.Message = string(data)
	}

	return errorResponse
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}
