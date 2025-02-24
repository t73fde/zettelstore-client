//-----------------------------------------------------------------------------
// Copyright (c) 2021-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2021-present Detlef Stern
//-----------------------------------------------------------------------------

// Package client provides a client for accessing the Zettelstore via its API.
package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxreader"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsc/sexp"
)

// Client contains all data to execute requests.
type Client struct {
	base      string
	username  string
	password  string
	token     string
	tokenType string
	expires   time.Time
	client    http.Client
}

// Base returns the base part of the URLs that are used to communicate with a Zettelstore.
func (c *Client) Base() string { return c.base }

// NewClient creates a new client with a given base URL to a Zettelstore.
func NewClient(u *url.URL) *Client {
	myURL := *u
	myURL.User = nil
	myURL.ForceQuery = false
	myURL.RawQuery = ""
	myURL.Fragment = ""
	myURL.RawFragment = ""
	base := myURL.String()
	c := Client{
		base: base,
		client: http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second, // TCP connect timeout
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
			},
		},
	}
	return &c
}

// AllowRedirect will modify the client to not follow redirect status code when
// using the Zettelstore. The original behaviour can be restored by setting
// allow to false.
func (c *Client) AllowRedirect(allow bool) {
	if allow {
		c.client.CheckRedirect = func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		c.client.CheckRedirect = nil
	}
}

// Error encapsulates the possible client call errors.
//
//   - StatusCode is the HTTP status code, e.g. 200
//   - Message is the HTTP message, e.g. "OK"
//   - Body is the HTTP body returned by a request.
type Error struct {
	StatusCode int
	Message    string
	Body       []byte
}

// Error returns the error as a string.
func (err *Error) Error() string {
	var body string
	if err.Body == nil {
		body = "nil"
	} else if bl := len(err.Body); bl == 0 {
		body = "empty"
	} else {
		const maxBodyLen = 79
		b := bytes.ToValidUTF8(err.Body, nil)
		if len(b) > maxBodyLen {
			if len(b)-3 > maxBodyLen {
				b = append(b[:maxBodyLen-3], "..."...)
			} else {
				b = b[:maxBodyLen]
			}
			b = bytes.ToValidUTF8(b, nil)
		}
		body = string(b) + " (" + strconv.Itoa(bl) + ")"
	}
	return strconv.Itoa(err.StatusCode) + " " + err.Message + ", body: " + body
}

func statusToError(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = nil
	}
	return &Error{
		StatusCode: resp.StatusCode,
		Message:    resp.Status[4:],
		Body:       body,
	}
}

// NewURLBuilder creates a new URL builder for the client with the given key.
//
// key is one of the defined lower case letters to specify an endpoint.
// See [Endpoints used by the API] for details.
//
// [Endpoints used by the API]: https://zettelstore.de/manual/h/00001012920000
func (c *Client) NewURLBuilder(key byte) *api.URLBuilder {
	return api.NewURLBuilder(c.base, key)
}
func (*Client) newRequest(ctx context.Context, method string, ub *api.URLBuilder, body io.Reader) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, ub.String(), body)
}

func (c *Client) executeRequest(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Add("Authorization", c.tokenType+" "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		return nil, err
	}
	return resp, err
}

func (c *Client) buildAndExecuteRequest(
	ctx context.Context,
	method string,
	ub *api.URLBuilder,
	body io.Reader,
) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, ub, body)
	if err != nil {
		return nil, err
	}
	err = c.updateToken(ctx)
	if err != nil {
		return nil, err
	}
	return c.executeRequest(req)
}

// SetAuth sets authentication data.
//
// username and password are the same values that are used to authenticate via the Web-UI.
func (c *Client) SetAuth(username, password string) {
	c.username = username
	c.password = password
	c.token = ""
	c.tokenType = ""
	c.expires = time.Time{}
}

func (c *Client) executeAuthRequest(req *http.Request) error {
	resp, err := c.executeRequest(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return statusToError(resp)
	}
	rd := sxreader.MakeReader(resp.Body)
	obj, err := rd.Read()
	if err != nil {
		return err
	}
	vals, err := sexp.ParseList(obj, "ssi")
	if err != nil {
		return err
	}
	token := vals[1].(sx.String).GetValue()
	if len(token) < 4 {
		return fmt.Errorf("no valid token found: %q", token)
	}
	c.token = token
	c.tokenType = vals[0].(sx.String).GetValue()
	c.expires = time.Now().Add(time.Duration(vals[2].(sx.Int64)*9/10) * time.Second)
	return nil
}

func (c *Client) updateToken(ctx context.Context) error {
	if c.username == "" {
		return nil
	}
	if time.Now().After(c.expires) {
		return c.Authenticate(ctx)
	}
	return c.RefreshToken(ctx)
}

// Authenticate sets a new token by sending user name and password.
//
// [Client.SetAuth] should be called before.
func (c *Client) Authenticate(ctx context.Context) error {
	authData := url.Values{"username": {c.username}, "password": {c.password}}
	req, err := c.newRequest(ctx, http.MethodPost, c.NewURLBuilder('a'), strings.NewReader(authData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.executeAuthRequest(req)
}

// RefreshToken updates the access token
//
// [Client.SetAuth] should be called before.
func (c *Client) RefreshToken(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodPut, c.NewURLBuilder('a'), nil)
	if err != nil {
		return err
	}
	return c.executeAuthRequest(req)
}

// CreateZettel creates a new zettel and returns its URL.
//
// data contains the zettel metadata and content, as it is stored in a file in a zettel box,
// or as returned by [Client.GetZettel].
// Metadata is separated from zettel content by an empty line.
func (c *Client) CreateZettel(ctx context.Context, data []byte) (id.Zid, error) {
	ub := c.NewURLBuilder('z')
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPost, ub, bytes.NewBuffer(data))
	if err != nil {
		return id.Invalid, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		return id.Invalid, statusToError(resp)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return id.Invalid, err
	}
	return id.Parse(string(b))
}

// CreateZettelData creates a new zettel and returns its URL.
//
// data contains the zettel date, encoded as explicit struct.
func (c *Client) CreateZettelData(ctx context.Context, data api.ZettelData) (id.Zid, error) {
	var buf bytes.Buffer
	if _, err := sx.Print(&buf, sexp.EncodeZettel(data)); err != nil {
		return id.Invalid, err
	}
	ub := c.NewURLBuilder('z').AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPost, ub, &buf)
	if err != nil {
		return id.Invalid, err
	}
	defer func() { _ = resp.Body.Close() }()
	rdr := sxreader.MakeReader(resp.Body)
	obj, err := rdr.Read()
	if resp.StatusCode != http.StatusCreated {
		return id.Invalid, statusToError(resp)
	}
	if err != nil {
		return id.Invalid, err
	}
	return makeZettelID(obj)
}

func makeZettelID(obj sx.Object) (id.Zid, error) {
	val, isInt64 := obj.(sx.Int64)
	if !isInt64 || val <= 0 {
		return id.Invalid, fmt.Errorf("invalid zettel ID: %v", val)
	}
	sVal := strconv.FormatInt(int64(val), 10)
	if len(sVal) < 14 {
		sVal = "00000000000000"[0:14-len(sVal)] + sVal
	}
	zid, err := id.Parse(sVal)
	if err != nil {
		return id.Invalid, fmt.Errorf("invalid zettel ID %v: %w", val, err)
	}
	return zid, nil
}

// UpdateZettel updates an existing zettel, specified by its zettel identifier.
//
// data contains the zettel metadata and content, as it is stored in a file in a zettel box,
// or as returned by [Client.GetZettel].
// Metadata is separated from zettel content by an empty line.
func (c *Client) UpdateZettel(ctx context.Context, zid id.Zid, data []byte) error {
	ub := c.NewURLBuilder('z').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPut, ub, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

// UpdateZettelData updates an existing zettel, specified by its zettel identifier.
func (c *Client) UpdateZettelData(ctx context.Context, zid id.Zid, data api.ZettelData) error {
	var buf bytes.Buffer
	if _, err := sx.Print(&buf, sexp.EncodeZettel(data)); err != nil {
		return err
	}
	ub := c.NewURLBuilder('z').SetZid(zid).AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPut, ub, &buf)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

// DeleteZettel deletes a zettel with the given identifier.
func (c *Client) DeleteZettel(ctx context.Context, zid id.Zid) error {
	ub := c.NewURLBuilder('z').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodDelete, ub, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

// ExecuteCommand will execute a given command at the Zettelstore.
//
// See [API commands] for a list of valid commands.
//
// [API commands]: https://zettelstore.de/manual/h/00001012080100
func (c *Client) ExecuteCommand(ctx context.Context, command api.Command) error {
	ub := c.NewURLBuilder('x').AppendKVQuery(api.QueryKeyCommand, string(command))
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPost, ub, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}
