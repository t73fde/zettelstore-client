//-----------------------------------------------------------------------------
// Copyright (c) 2021-present Detlef Stern
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
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"zettelstore.de/client.fossil/api"
	"zettelstore.de/client.fossil/sx"
	"zettelstore.de/sx.fossil/sxpf"
	"zettelstore.de/sx.fossil/sxpf/reader"
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
	rd := reader.MakeReader(resp.Body)
	obj, err := rd.Read()
	if err != nil {
		return err
	}
	vals, err := sx.ParseObject(obj, "ssi")
	if err != nil {
		return err
	}
	token := vals[1].(sxpf.String).String()
	if len(token) < 4 {
		return fmt.Errorf("no valid token found: %q", token)
	}
	c.token = token
	c.tokenType = vals[0].(sxpf.String).String()
	c.expires = time.Now().Add(time.Duration(vals[2].(sxpf.Int64)*9/10) * time.Second)
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

// CreateZettelData creates a new zettel and returns its URL.
func (c *Client) CreateZettelData(ctx context.Context, data api.ZettelData) (api.ZettelID, error) {
	var buf bytes.Buffer
	if err := encodeZettelData(&buf, &data); err != nil {
		return api.InvalidZID, err
	}
	ub := c.newURLBuilder('z').AppendKVQuery(api.QueryKeyEncoding, api.EncodingJson)
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

func encodeZettelData(buf *bytes.Buffer, data *api.ZettelData) error {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	return enc.Encode(&data)
}

var bsLF = []byte{'\n'}

// ListZettel returns a list of all Zettel.
func (c *Client) ListZettel(ctx context.Context, query string) ([][]byte, error) {
	ub := c.newURLBuilder('z').AppendQuery(query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
	default:
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
func (c *Client) ListZettelJSON(ctx context.Context, query string) (string, string, []api.ZidMetaJSON, error) {
	ub := c.newURLBuilder('z').AppendKVQuery(api.QueryKeyEncoding, api.EncodingJson).AppendQuery(query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", nil, statusToError(resp)
	}
	dec := json.NewDecoder(resp.Body)
	var zl api.ZettelListJSON
	err = dec.Decode(&zl)
	if err != nil {
		return "", "", nil, err
	}
	return zl.Query, zl.Human, zl.List, nil
}

// GetZettel returns a zettel as a string.
func (c *Client) GetZettel(ctx context.Context, zid api.ZettelID, part string) ([]byte, error) {
	ub := c.newURLBuilder('z').SetZid(zid)
	if part != "" && part != api.PartContent {
		ub.AppendKVQuery(api.QueryKeyPart, part)
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
	default:
		return nil, statusToError(resp)
	}
	return io.ReadAll(resp.Body)
}

// GetZettelData returns a zettel as a struct of its parts.
func (c *Client) GetZettelData(ctx context.Context, zid api.ZettelID) (api.ZettelData, error) {
	ub := c.newURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	ub.AppendKVQuery(api.QueryKeyPart, api.PartZettel)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return api.ZettelData{}, statusToError(resp)
		}
		rdr := reader.MakeReader(resp.Body)
		obj, err2 := rdr.Read()
		if err2 == nil {
			var data api.ZettelData
			err2 = parseZettelSxToStruct(obj, &data)
			return data, err2
		}
	}
	return api.ZettelData{}, err
}

func parseZettelSxToStruct(obj sxpf.Object, data *api.ZettelData) error {
	vals, err := sx.ParseObject(obj, "yppppp")
	if err != nil {
		return err
	}
	if errSym := checkSymbol(vals[0], "zettel"); errSym != nil {
		return errSym
	}

	// Ignore vals[1] (id "12345678901234"), we don't need it in ZettelData

	meta, err := parseMetaSxToMap(vals[2].(*sxpf.Pair))
	if err != nil {
		return err
	}

	// Ignore vals[3] (rights 4), we don't need the rights in ZettelData

	encVals, err := sx.ParseObject(vals[4], "ys")
	if err != nil {
		return err
	}
	if errSym := checkSymbol(encVals[0], "encoding"); errSym != nil {
		return errSym
	}

	contentVals, err := sx.ParseObject(vals[5], "ys")
	if err != nil {
		return err
	}
	if errSym := checkSymbol(contentVals[0], "content"); errSym != nil {
		return errSym
	}

	data.Meta = meta
	data.Encoding = encVals[1].(sxpf.String).String()
	data.Content = contentVals[1].(sxpf.String).String()
	return nil
}
func checkSymbol(obj sxpf.Object, exp string) error {
	if got := obj.(*sxpf.Symbol).Name(); got != exp {
		return fmt.Errorf("symbol %q expected, but got: %q", exp, got)
	}
	return nil
}
func parseMetaSxToMap(pair *sxpf.Pair) (api.ZettelMeta, error) {
	if err := checkSymbol(pair.Car(), "meta"); err != nil {
		return nil, err
	}
	res := api.ZettelMeta{}
	for node := pair.Tail(); node != nil; node = node.Tail() {
		mVals, err := sx.ParseObject(node.Car(), "ys")
		if err != nil {
			return nil, err
		}
		res[mVals[0].(*sxpf.Symbol).Name()] = mVals[1].(sxpf.String).String()
	}
	return res, nil
}

// GetParsedZettel return a parsed zettel in a defined encoding.
func (c *Client) GetParsedZettel(ctx context.Context, zid api.ZettelID, enc api.EncodingEnum) ([]byte, error) {
	return c.getZettelString(ctx, zid, enc, true)
}

// GetEvaluatedZettel return an evaluated zettel in a defined encoding.
func (c *Client) GetEvaluatedZettel(ctx context.Context, zid api.ZettelID, enc api.EncodingEnum) ([]byte, error) {
	return c.getZettelString(ctx, zid, enc, false)
}

func (c *Client) getZettelString(ctx context.Context, zid api.ZettelID, enc api.EncodingEnum, parseOnly bool) ([]byte, error) {
	ub := c.newURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, enc.String())
	ub.AppendKVQuery(api.QueryKeyPart, api.PartContent)
	if parseOnly {
		ub.AppendKVQuery(api.QueryKeyParseOnly, "")
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
	default:
		return nil, statusToError(resp)
	}
	return io.ReadAll(resp.Body)
}

// GetParsedSz returns an parsed zettel as a Sexpr-decoded data structure.
func (c *Client) GetParsedSz(ctx context.Context, zid api.ZettelID, part string, sf sxpf.SymbolFactory) (sxpf.Object, error) {
	return c.getSz(ctx, zid, part, true, sf)
}

// GetEvaluatedSz returns an evaluated zettel as a Sexpr-decoded data structure.
func (c *Client) GetEvaluatedSz(ctx context.Context, zid api.ZettelID, part string, sf sxpf.SymbolFactory) (sxpf.Object, error) {
	return c.getSz(ctx, zid, part, false, sf)
}

func (c *Client) getSz(ctx context.Context, zid api.ZettelID, part string, parseOnly bool, sf sxpf.SymbolFactory) (sxpf.Object, error) {
	ub := c.newURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingSz)
	if part != "" {
		ub.AppendKVQuery(api.QueryKeyPart, part)
	}
	if parseOnly {
		ub.AppendKVQuery(api.QueryKeyParseOnly, "")
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	return reader.MakeReader(bufio.NewReaderSize(resp.Body, 8), reader.WithSymbolFactory(sf)).Read()
}

// GetMeta returns the metadata of a zettel.
func (c *Client) GetMeta(ctx context.Context, zid api.ZettelID) (api.ZettelMeta, error) {
	ub := c.newURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingJson)
	ub.AppendKVQuery(api.QueryKeyPart, api.PartMeta)
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

// UpdateZettelData updates an existing zettel.
func (c *Client) UpdateZettelData(ctx context.Context, zid api.ZettelID, data api.ZettelData) error {
	var buf bytes.Buffer
	if err := encodeZettelData(&buf, &data); err != nil {
		return err
	}
	ub := c.newURLBuilder('z').SetZid(zid).AppendKVQuery(api.QueryKeyEncoding, api.EncodingJson)
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
	ub := c.newURLBuilder('x').AppendKVQuery(api.QueryKeyCommand, string(command))
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
			ub.AppendKVQuery(key, val)
		}
	}
	return ub
}

// QueryMapMeta returns a map of all metadata values with the given query action to the
// list of zettel IDs containing this value.
func (c *Client) QueryMapMeta(ctx context.Context, query string) (api.MapMeta, error) {
	err := c.updateToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest(ctx, http.MethodGet, c.newURLBuilder('z').AppendKVQuery(api.QueryKeyEncoding, api.EncodingJson).AppendQuery(query), nil)
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

// GetVersionInfo returns version information..
func (c *Client) GetVersionInfo(ctx context.Context) (VersionInfo, error) {
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, c.newURLBuilder('x'), nil, nil)
	if err != nil {
		return VersionInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return VersionInfo{}, statusToError(resp)
	}
	rdr := reader.MakeReader(resp.Body)
	obj, err := rdr.Read()
	if err == nil {
		if vals, errVals := sx.ParseObject(obj, "iiiss"); errVals == nil {
			return VersionInfo{
				Major: int(vals[0].(sxpf.Int64)),
				Minor: int(vals[1].(sxpf.Int64)),
				Patch: int(vals[2].(sxpf.Int64)),
				Info:  vals[3].(sxpf.String).String(),
				Hash:  vals[4].(sxpf.String).String(),
			}, nil
		}
	}
	return VersionInfo{}, err
}

// VersionInfo contains version information.
type VersionInfo struct {
	Major int
	Minor int
	Patch int
	Info  string
	Hash  string
}
