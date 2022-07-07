//-----------------------------------------------------------------------------
// Copyright (c) 2021-2022 Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

// Package client provides a client for accessing the Zettelstore via its API.
package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"codeberg.org/t73fde/sxpf"
	"zettelstore.de/c/api"
	"zettelstore.de/c/sexpr"
	"zettelstore.de/c/zjson"
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

// NewClient create a new client.
func NewClient(u *url.URL) *Client {
	myURL := *u
	myURL.User = nil
	myURL.ForceQuery = false
	myURL.RawQuery = ""
	myURL.Fragment = ""
	myURL.RawFragment = ""
	base := myURL.String()
	if !strings.HasSuffix(base, "/") {
		base += "/"
	}
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

// Error encapsulates the possible client call errors.
type Error struct {
	StatusCode int
	Message    string
	Body       []byte
}

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

func (c *Client) newURLBuilder(key byte) *api.URLBuilder {
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
			resp.Body.Close()
		}
		return nil, err
	}
	return resp, err
}

func (c *Client) buildAndExecuteRequest(
	ctx context.Context, method string, ub *api.URLBuilder, body io.Reader, h http.Header) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, ub, body)
	if err != nil {
		return nil, err
	}
	err = c.updateToken(ctx)
	if err != nil {
		return nil, err
	}
	for key, val := range h {
		req.Header[key] = append(req.Header[key], val...)
	}
	return c.executeRequest(req)
}

// SetAuth sets authentication data.
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var tinfo api.AuthJSON
	err = dec.Decode(&tinfo)
	if err != nil {
		return err
	}
	c.token = tinfo.Token
	c.tokenType = tinfo.Type
	c.expires = time.Now().Add(time.Duration(tinfo.Expires*10/9) * time.Second)
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
func (c *Client) Authenticate(ctx context.Context) error {
	authData := url.Values{"username": {c.username}, "password": {c.password}}
	req, err := c.newRequest(ctx, http.MethodPost, c.newURLBuilder('a'), strings.NewReader(authData.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.executeAuthRequest(req)
}

// RefreshToken updates the access token
func (c *Client) RefreshToken(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodPut, c.newURLBuilder('a'), nil)
	if err != nil {
		return err
	}
	return c.executeAuthRequest(req)
}

// CreateZettel creates a new zettel and returns its URL.
func (c *Client) CreateZettel(ctx context.Context, data []byte) (api.ZettelID, error) {
	ub := c.newURLBuilder('z')
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPost, ub, bytes.NewBuffer(data), nil)
	if err != nil {
		return api.InvalidZID, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return api.InvalidZID, statusToError(resp)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.InvalidZID, err
	}
	if zid := api.ZettelID(b); zid.IsValid() {
		return zid, nil
	}
	return api.InvalidZID, err
}

// CreateZettelJSON creates a new zettel and returns its URL.
func (c *Client) CreateZettelJSON(ctx context.Context, data *api.ZettelDataJSON) (api.ZettelID, error) {
	var buf bytes.Buffer
	if err := encodeZettelData(&buf, data); err != nil {
		return api.InvalidZID, err
	}
	ub := c.newURLBuilder('j')
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPost, ub, &buf, nil)
	if err != nil {
		return api.InvalidZID, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return api.InvalidZID, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var newZid api.ZidJSON
	err = dec.Decode(&newZid)
	if err != nil {
		return api.InvalidZID, err
	}
	if zid := newZid.ID; zid.IsValid() {
		return zid, nil
	}
	return api.InvalidZID, err
}

func encodeZettelData(buf *bytes.Buffer, data *api.ZettelDataJSON) error {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	return enc.Encode(&data)
}

var bsLF = []byte{'\n'}

// ListZettel returns a list of all Zettel.
func (c *Client) ListZettel(ctx context.Context, query url.Values) ([][]byte, error) {
	ub := c.newQueryURLBuilder('z', query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, bsLF)
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// ListZettelJSON returns a list of zettel.
func (c *Client) ListZettelJSON(ctx context.Context, query url.Values) (string, []api.ZidMetaJSON, error) {
	ub := c.newQueryURLBuilder('j', query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var zl api.ZettelListJSON
	err = dec.Decode(&zl)
	if err != nil {
		return "", nil, err
	}
	return zl.Query, zl.List, nil
}

// GetZettel returns a zettel as a string.
func (c *Client) GetZettel(ctx context.Context, zid api.ZettelID, part string) ([]byte, error) {
	ub := c.newURLBuilder('z').SetZid(zid)
	if part != "" && part != api.PartContent {
		ub.AppendQuery(api.QueryKeyPart, part)
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	return io.ReadAll(resp.Body)
}

// GetZettelJSON returns a zettel as a JSON struct.
func (c *Client) GetZettelJSON(ctx context.Context, zid api.ZettelID) (*api.ZettelDataJSON, error) {
	ub := c.newURLBuilder('j').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var out api.ZettelDataJSON
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetParsedZettel return a parsed zettel in a defined encoding.
func (c *Client) GetParsedZettel(ctx context.Context, zid api.ZettelID, enc api.EncodingEnum) ([]byte, error) {
	return c.getZettelString(ctx, 'p', zid, enc)
}

// GetEvaluatedZettel return an evaluated zettel in a defined encoding.
func (c *Client) GetEvaluatedZettel(ctx context.Context, zid api.ZettelID, enc api.EncodingEnum) ([]byte, error) {
	return c.getZettelString(ctx, 'v', zid, enc)
}

func (c *Client) getZettelString(ctx context.Context, key byte, zid api.ZettelID, enc api.EncodingEnum) ([]byte, error) {
	ub := c.newURLBuilder(key).SetZid(zid)
	ub.AppendQuery(api.QueryKeyEncoding, enc.String())
	ub.AppendQuery(api.QueryKeyPart, api.PartContent)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	return io.ReadAll(resp.Body)
}

// GetParsedZettelZJSON returns an parsed zettel as a JSON-decoded data structure.
func (c *Client) GetParsedSexpr(ctx context.Context, zid api.ZettelID, part string) (sxpf.Value, error) {
	return c.getSexpr(ctx, 'p', zid, part)
}

// GetEvaluatedZettelZJSON returns an evaluated zettel as a JSON-decoded data structure.
func (c *Client) GetEvaluatedSexpr(ctx context.Context, zid api.ZettelID, part string) (sxpf.Value, error) {
	return c.getSexpr(ctx, 'v', zid, part)
}

func (c *Client) getSexpr(ctx context.Context, key byte, zid api.ZettelID, part string) (sxpf.Value, error) {
	ub := c.newURLBuilder(key).SetZid(zid)
	ub.AppendQuery(api.QueryKeyEncoding, api.EncodingSexpr)
	if part != "" {
		ub.AppendQuery(api.QueryKeyPart, part)
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}

	return sxpf.ParseValue(sexpr.Smk, bufio.NewReaderSize(resp.Body, 8))
}

// GetParsedZettelZJSON returns an parsed zettel as a JSON-decoded data structure.
func (c *Client) GetParsedZJSON(ctx context.Context, zid api.ZettelID, part string) (zjson.Value, error) {
	return c.getZJSON(ctx, 'p', zid, part)
}

// GetEvaluatedZettelZJSON returns an evaluated zettel as a JSON-decoded data structure.
func (c *Client) GetEvaluatedZJSON(ctx context.Context, zid api.ZettelID, part string) (zjson.Value, error) {
	return c.getZJSON(ctx, 'v', zid, part)
}

func (c *Client) getZJSON(ctx context.Context, key byte, zid api.ZettelID, part string) (zjson.Value, error) {
	ub := c.newURLBuilder(key).SetZid(zid)
	ub.AppendQuery(api.QueryKeyEncoding, api.EncodingZJSON)
	if part != "" {
		ub.AppendQuery(api.QueryKeyPart, part)
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	return zjson.Decode(resp.Body)
}

// GetMeta returns the metadata of a zettel.
func (c *Client) GetMeta(ctx context.Context, zid api.ZettelID) (api.ZettelMeta, error) {
	ub := c.newURLBuilder('m').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var out api.MetaJSON
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out.Meta, nil
}

// GetZettelOrder returns metadata of the given zettel and, more important,
// metadata of zettel that are referenced in a list within the first zettel.
func (c *Client) GetZettelOrder(ctx context.Context, zid api.ZettelID) (*api.ZidMetaRelatedList, error) {
	ub := c.newURLBuilder('o').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var out api.ZidMetaRelatedList
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// ContextDirection specifies how the context should be calculated.
type ContextDirection uint8

// Allowed values for ContextDirection
const (
	_ ContextDirection = iota
	DirBoth
	DirBackward
	DirForward
)

// GetZettelContext returns metadata of the given zettel and, more important,
// metadata of zettel that for the context of the first zettel.
func (c *Client) GetZettelContext(
	ctx context.Context, zid api.ZettelID, dir ContextDirection, depth, limit int) (
	*api.ZidMetaRelatedList, error,
) {
	ub := c.newURLBuilder('x').SetZid(zid)
	switch dir {
	case DirBackward:
		ub.AppendQuery(api.QueryKeyDir, api.DirBackward)
	case DirForward:
		ub.AppendQuery(api.QueryKeyDir, api.DirForward)
	}
	if depth > 0 {
		ub.AppendQuery(api.QueryKeyDepth, strconv.Itoa(depth))
	}
	if limit > 0 {
		ub.AppendQuery(api.QueryKeyLimit, strconv.Itoa(limit))
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var out api.ZidMetaRelatedList
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUnlinkedReferences returns connections to other zettel, embedded material, externals URLs.
func (c *Client) GetUnlinkedReferences(
	ctx context.Context, zid api.ZettelID, query url.Values) (*api.ZidMetaRelatedList, error) {
	ub := c.newQueryURLBuilder('u', query).SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var out api.ZidMetaRelatedList
	err = dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateZettel updates an existing zettel.
func (c *Client) UpdateZettel(ctx context.Context, zid api.ZettelID, data []byte) error {
	ub := c.newURLBuilder('z').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPut, ub, bytes.NewBuffer(data), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

// UpdateZettelJSON updates an existing zettel.
func (c *Client) UpdateZettelJSON(ctx context.Context, zid api.ZettelID, data *api.ZettelDataJSON) error {
	var buf bytes.Buffer
	if err := encodeZettelData(&buf, data); err != nil {
		return err
	}
	ub := c.newURLBuilder('j').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPut, ub, &buf, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

// RenameZettel renames a zettel.
func (c *Client) RenameZettel(ctx context.Context, oldZid, newZid api.ZettelID) error {
	ub := c.newURLBuilder('z').SetZid(oldZid)
	h := http.Header{
		api.HeaderDestination: {c.newURLBuilder('z').SetZid(newZid).String()},
	}
	resp, err := c.buildAndExecuteRequest(ctx, api.MethodMove, ub, nil, h)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

// DeleteZettel deletes a zettel with the given identifier.
func (c *Client) DeleteZettel(ctx context.Context, zid api.ZettelID) error {
	ub := c.newURLBuilder('z').SetZid(zid)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodDelete, ub, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

// ExecuteCommand will execute a given command at the Zettelstore.
func (c *Client) ExecuteCommand(ctx context.Context, command api.Command) error {
	ub := c.newURLBuilder('x').AppendQuery(api.QueryKeyCommand, string(command))
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPost, ub, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return statusToError(resp)
	}
	return nil
}

func (c *Client) newQueryURLBuilder(key byte, query url.Values) *api.URLBuilder {
	ub := c.newURLBuilder(key)
	for key, values := range query {
		if key == api.QueryKeyEncoding {
			continue
		}
		for _, val := range values {
			ub.AppendQuery(key, val)
		}
	}
	return ub
}

// ListMapMeta returns a map of all metadata values with the given key to the
// list of zettel IDs containing this value.
func (c *Client) ListMapMeta(ctx context.Context, key string) (api.MapMeta, error) {
	err := c.updateToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodGet, c.newURLBuilder('m').AppendQuery(api.QueryKeyKey, key), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.executeRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var mlj api.MapListJSON
	err = dec.Decode(&mlj)
	if err != nil {
		return nil, err
	}
	return mlj.Map, nil
}

// GetVersionJSON returns version information..
func (c *Client) GetVersionJSON(ctx context.Context) (api.VersionJSON, error) {
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, c.newURLBuilder('x'), nil, nil)
	if err != nil {
		return api.VersionJSON{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return api.VersionJSON{}, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var version api.VersionJSON
	err = dec.Decode(&version)
	if err != nil {
		return api.VersionJSON{}, err
	}
	return version, nil
}
