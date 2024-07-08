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
	"bufio"
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
	"t73f.de/r/zsc/sexp"
	"t73f.de/r/zsc/sz"
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
// using the Zettelstore. The original behaviour can be restored by settinh
// "allow" to false.
func (c *Client) AllowRedirect(allow bool) {
	if allow {
		c.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		c.client.CheckRedirect = nil
	}
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

// NewURLBuilder creates a new URL builder for the client with the given key.
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
func (c *Client) RefreshToken(ctx context.Context) error {
	req, err := c.newRequest(ctx, http.MethodPut, c.NewURLBuilder('a'), nil)
	if err != nil {
		return err
	}
	return c.executeAuthRequest(req)
}

// CreateZettel creates a new zettel and returns its URL.
func (c *Client) CreateZettel(ctx context.Context, data []byte) (api.ZettelID, error) {
	ub := c.NewURLBuilder('z')
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
	if _, err := sx.Print(&buf, sexp.EncodeZettel(data)); err != nil {
		return api.InvalidZID, err
	}
	ub := c.NewURLBuilder('z').AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodPost, ub, &buf, nil)
	if err != nil {
		return api.InvalidZID, err
	}
	defer resp.Body.Close()
	rdr := sxreader.MakeReader(resp.Body)
	obj, err := rdr.Read()
	if resp.StatusCode != http.StatusCreated {
		return api.InvalidZID, statusToError(resp)
	}
	if err != nil {
		return api.InvalidZID, err
	}
	return makeZettelID(obj)
}

var bsLF = []byte{'\n'}

// QueryZettel returns a list of all Zettel.
func (c *Client) QueryZettel(ctx context.Context, query string) ([][]byte, error) {
	ub := c.NewURLBuilder('z').AppendQuery(query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, statusToError(resp)
	}
	if err != nil {
		return nil, err
	}
	lines := bytes.Split(data, bsLF)
	if len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}

// QueryZettelData returns a list of zettel metadata.
func (c *Client) QueryZettelData(ctx context.Context, query string) (string, string, []api.ZidMetaRights, error) {
	ub := c.NewURLBuilder('z').AppendKVQuery(api.QueryKeyEncoding, api.EncodingData).AppendQuery(query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return "", "", nil, err
	}
	defer resp.Body.Close()
	rdr := sxreader.MakeReader(resp.Body)
	obj, err := rdr.Read()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
		return "", "", nil, nil
	default:
		return "", "", nil, statusToError(resp)
	}
	if err != nil {
		return "", "", nil, err
	}
	vals, err := sexp.ParseList(obj, "yppp")
	if err != nil {
		return "", "", nil, err
	}
	qVals, err := sexp.ParseList(vals[1], "ys")
	if err != nil {
		return "", "", nil, err
	}
	hVals, err := sexp.ParseList(vals[2], "ys")
	if err != nil {
		return "", "", nil, err
	}
	metaList, err := parseMetaList(vals[3].(*sx.Pair))
	return sz.GoValue(qVals[1]), sz.GoValue(hVals[1]), metaList, err
}

func parseMetaList(metaPair *sx.Pair) ([]api.ZidMetaRights, error) {
	if metaPair == nil {
		return nil, fmt.Errorf("no zettel list")
	}
	if errSym := sexp.CheckSymbol(metaPair.Car(), "list"); errSym != nil {
		return nil, errSym
	}
	var result []api.ZidMetaRights
	for node := metaPair.Cdr(); !sx.IsNil(node); {
		elem, isPair := sx.GetPair(node)
		if !isPair {
			return nil, fmt.Errorf("meta-list not a proper list: %v", metaPair.String())
		}
		node = elem.Cdr()
		vals, err := sexp.ParseList(elem.Car(), "yppp")
		if err != nil {
			return nil, err
		}

		if errSym := sexp.CheckSymbol(vals[0], "zettel"); errSym != nil {
			return nil, errSym
		}

		idVals, err := sexp.ParseList(vals[1], "yi")
		if err != nil {
			return nil, err
		}
		if errSym := sexp.CheckSymbol(idVals[0], "id"); errSym != nil {
			return nil, errSym
		}
		zid, err := makeZettelID(idVals[1])
		if err != nil {
			return nil, err
		}

		meta, err := sexp.ParseMeta(vals[2].(*sx.Pair))
		if err != nil {
			return nil, err
		}

		rights, err := sexp.ParseRights(vals[3])
		if err != nil {
			return nil, err
		}

		result = append(result, api.ZidMetaRights{
			ID:     zid,
			Meta:   meta,
			Rights: rights,
		})
	}
	return result, nil
}
func makeZettelID(obj sx.Object) (api.ZettelID, error) {
	val, isInt64 := obj.(sx.Int64)
	if !isInt64 || val <= 0 {
		return api.InvalidZID, fmt.Errorf("invalid zettel ID: %v", val)
	}
	sVal := strconv.FormatInt(int64(val), 10)
	if len(sVal) < 14 {
		sVal = "00000000000000"[0:14-len(sVal)] + sVal
	}
	zid := api.ZettelID(sVal)
	if !zid.IsValid() {
		return api.InvalidZID, fmt.Errorf("invalid zettel ID: %v", val)
	}
	return zid, nil
}

// QueryAggregate returns a aggregate as a result of a query.
// It is most often used in a query with an action, where the action is either
// a metadata key of type Word or of type TagSet.
func (c *Client) QueryAggregate(ctx context.Context, query string) (api.Aggregate, error) {
	lines, err := c.QueryZettel(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, nil
	}
	agg := make(api.Aggregate, len(lines))
	for _, line := range lines {
		if fields := bytes.Fields(line); len(fields) > 1 {
			key := string(fields[0])
			for _, field := range fields[1:] {
				if zid := api.ZettelID(string(field)); zid.IsValid() {
					agg[key] = append(agg[key], zid)
				}
			}
		}
	}
	return agg, nil
}

// TagZettel returns the tag zettel of a given tag.
//
// This method only works if c.AllowRedirect(true) was called.
func (c *Client) TagZettel(ctx context.Context, tag string) (api.ZettelID, error) {
	return c.fetchTagOrRoleZettel(ctx, api.QueryKeyTag, tag)
}

// RoleZettel returns the tag zettel of a given tag.
//
// This method only works if c.AllowRedirect(true) was called.
func (c *Client) RoleZettel(ctx context.Context, role string) (api.ZettelID, error) {
	return c.fetchTagOrRoleZettel(ctx, api.QueryKeyRole, role)
}

func (c *Client) fetchTagOrRoleZettel(ctx context.Context, key, val string) (api.ZettelID, error) {
	if c.client.CheckRedirect == nil {
		panic("client does not allow to track redirect")
	}
	ub := c.NewURLBuilder('z').AppendKVQuery(key, val)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return api.InvalidZID, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return api.InvalidZID, err
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return "", nil
	case http.StatusFound:
		zid := api.ZettelID(data)
		if zid.IsValid() {
			return zid, nil
		}
		return api.InvalidZID, nil
	default:
		return api.InvalidZID, statusToError(resp)
	}
}

// GetZettel returns a zettel as a string.
func (c *Client) GetZettel(ctx context.Context, zid api.ZettelID, part string) ([]byte, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	if part != "" && part != api.PartContent {
		ub.AppendKVQuery(api.QueryKeyPart, part)
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, statusToError(resp)
	}
	return data, err
}

// GetZettelData returns a zettel as a struct of its parts.
func (c *Client) GetZettelData(ctx context.Context, zid api.ZettelID) (api.ZettelData, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	ub.AppendKVQuery(api.QueryKeyPart, api.PartZettel)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return api.ZettelData{}, statusToError(resp)
		}
		rdr := sxreader.MakeReader(resp.Body)
		obj, err2 := rdr.Read()
		if err2 == nil {
			return sexp.ParseZettel(obj)
		}
	}
	return api.ZettelData{}, err
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
	ub := c.NewURLBuilder('z').SetZid(zid)
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
func (c *Client) GetParsedSz(ctx context.Context, zid api.ZettelID, part string) (sx.Object, error) {
	return c.getSz(ctx, zid, part, true)
}

// GetEvaluatedSz returns an evaluated zettel as a Sexpr-decoded data structure.
func (c *Client) GetEvaluatedSz(ctx context.Context, zid api.ZettelID, part string) (sx.Object, error) {
	return c.getSz(ctx, zid, part, false)
}

func (c *Client) getSz(ctx context.Context, zid api.ZettelID, part string, parseOnly bool) (sx.Object, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
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
	return sxreader.MakeReader(bufio.NewReaderSize(resp.Body, 8)).Read()
}

// GetMetaData returns the metadata of a zettel.
func (c *Client) GetMetaData(ctx context.Context, zid api.ZettelID) (api.MetaRights, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	ub.AppendKVQuery(api.QueryKeyPart, api.PartMeta)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return api.MetaRights{}, err
	}
	defer resp.Body.Close()
	rdr := sxreader.MakeReader(resp.Body)
	obj, err := rdr.Read()
	if resp.StatusCode != http.StatusOK {
		return api.MetaRights{}, statusToError(resp)
	}
	if err != nil {
		return api.MetaRights{}, err
	}
	vals, err := sexp.ParseList(obj, "ypp")
	if err != nil {
		return api.MetaRights{}, err
	}
	if errSym := sexp.CheckSymbol(vals[0], "list"); errSym != nil {
		return api.MetaRights{}, err
	}

	meta, err := sexp.ParseMeta(vals[1].(*sx.Pair))
	if err != nil {
		return api.MetaRights{}, err
	}

	rights, err := sexp.ParseRights(vals[2])
	if err != nil {
		return api.MetaRights{}, err
	}

	return api.MetaRights{
		Meta:   meta,
		Rights: rights,
	}, nil
}

// UpdateZettel updates an existing zettel.
func (c *Client) UpdateZettel(ctx context.Context, zid api.ZettelID, data []byte) error {
	ub := c.NewURLBuilder('z').SetZid(zid)
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
	if _, err := sx.Print(&buf, sexp.EncodeZettel(data)); err != nil {
		return err
	}
	ub := c.NewURLBuilder('z').SetZid(zid).AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
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
//
// This function is deprecated and will be removed in v0.19 (or later).
func (c *Client) RenameZettel(ctx context.Context, oldZid, newZid api.ZettelID) error {
	ub := c.NewURLBuilder('z').SetZid(oldZid)
	h := http.Header{
		api.HeaderDestination: {c.NewURLBuilder('z').SetZid(newZid).String()},
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
	ub := c.NewURLBuilder('z').SetZid(zid)
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
	ub := c.NewURLBuilder('x').AppendKVQuery(api.QueryKeyCommand, string(command))
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

// GetVersionInfo returns version information.
func (c *Client) GetVersionInfo(ctx context.Context) (VersionInfo, error) {
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, c.NewURLBuilder('x'), nil, nil)
	if err != nil {
		return VersionInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return VersionInfo{}, statusToError(resp)
	}
	rdr := sxreader.MakeReader(resp.Body)
	obj, err := rdr.Read()
	if err == nil {
		if vals, errVals := sexp.ParseList(obj, "iiiss"); errVals == nil {
			return VersionInfo{
				Major: int(vals[0].(sx.Int64)),
				Minor: int(vals[1].(sx.Int64)),
				Patch: int(vals[2].(sx.Int64)),
				Info:  vals[3].(sx.String).GetValue(),
				Hash:  vals[4].(sx.String).GetValue(),
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

// GetApplicationZid returns the zettel identifier used to configure client
// application with the given name.
func (c *Client) GetApplicationZid(ctx context.Context, appname string) (api.ZettelID, error) {
	mr, err := c.GetMetaData(ctx, api.ZidAppDirectory)
	if err != nil {
		return api.InvalidZID, err
	}
	key := appname + "-zid"
	val, found := mr.Meta[key]
	if !found {
		return api.InvalidZID, fmt.Errorf("no application registered: %v", appname)
	}
	if zid := api.ZettelID(val); zid.IsValid() {
		return zid, nil
	}
	return api.InvalidZID, fmt.Errorf("invalid identifier for application %v: %v", appname, val)
}

// Get executes a GET request to the given URL and returns the read data.
func (c *Client) Get(ctx context.Context, ub *api.URLBuilder) ([]byte, error) {
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
		return nil, nil
	default:
		return nil, statusToError(resp)
	}
	data, err := io.ReadAll(resp.Body)
	return data, err
}
