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

package client

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"t73f.de/r/sx"
	"t73f.de/r/sx/sxreader"
	"t73f.de/r/zsc/api"
	"t73f.de/r/zsc/domain/id"
	"t73f.de/r/zsc/sexp"
	"t73f.de/r/zsx"
)

var bsLF = []byte{'\n'}

// QueryZettel returns a list of all Zettel based on the given query.
//
// query is a search expression, as described in [Query the list of all zettel].
//
// The functions returns a slice of bytes slices, where each byte slice contains the
// zettel identifier within its first 14 bytes. The next byte is a space character,
// followed by the title of the zettel.
//
// [Query the list of all zettel]: https://zettelstore.de/manual/h/00001012051400
func (c *Client) QueryZettel(ctx context.Context, query string) ([][]byte, error) {
	ub := c.NewURLBuilder('z').AppendQuery(query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
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
//
// query is a search expression, as described in [Query the list of all zettel].
//
// The functions returns the normalized query and its human-readable representation as
// its first two result values.
//
// [Query the list of all zettel]: https://zettelstore.de/manual/h/00001012051400
func (c *Client) QueryZettelData(ctx context.Context, query string) (string, string, []api.ZidMetaRights, error) {
	ub := c.NewURLBuilder('z').AppendKVQuery(api.QueryKeyEncoding, api.EncodingData).AppendQuery(query)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return "", "", nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	rdr := sxreader.MakeReader(resp.Body).SetListLimit(0) // No limit b/c number of zettel may be more than 100000. We must trust the server
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
	vals, err := sexp.ParseList(obj, "yppr")
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
	return zsx.GoValue(qVals[1]), zsx.GoValue(hVals[1]), metaList, err
}

func parseMetaList(metaPair *sx.Pair) ([]api.ZidMetaRights, error) {
	var result []api.ZidMetaRights
	for node := metaPair; !sx.IsNil(node); {
		elem, isPair := sx.GetPair(node)
		if !isPair {
			return nil, fmt.Errorf("meta-list not a proper list: %v", metaPair.String())
		}
		node = elem.Tail()
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

// QueryAggregate returns a aggregate as a result of a query.
// It is most often used in a query with an action, where the action is either
// a metadata key of type Word or of type TagSet.
//
// query is a search expression, as described in [Query the list of all zettel].
// It must contain an aggregate action.
//
// [Query the list of all zettel]: https://zettelstore.de/manual/h/00001012051400
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
				if zid, zidErr := id.Parse(string(field)); zidErr == nil {
					agg[key] = append(agg[key], zid)
				}
			}
		}
	}
	return agg, nil
}

// TagZettel returns the identifier of the tag zettel for a given tag.
//
// This method only works if c.AllowRedirect(true) was called.
func (c *Client) TagZettel(ctx context.Context, tag string) (id.Zid, error) {
	return c.fetchTagOrRoleZettel(ctx, api.QueryKeyTag, tag)
}

// RoleZettel returns the identifier of the tag zettel for a given role.
//
// This method only works if c.AllowRedirect(true) was called.
func (c *Client) RoleZettel(ctx context.Context, role string) (id.Zid, error) {
	return c.fetchTagOrRoleZettel(ctx, api.QueryKeyRole, role)
}

func (c *Client) fetchTagOrRoleZettel(ctx context.Context, key, val string) (id.Zid, error) {
	if c.client.CheckRedirect == nil {
		panic("client does not allow to track redirect")
	}
	ub := c.NewURLBuilder('z').AppendKVQuery(key, val)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return id.Invalid, err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return id.Invalid, err
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return id.Invalid, nil
	case http.StatusFound:
		return id.Parse(string(data))
	default:
		return id.Invalid, statusToError(resp)
	}
}

// GetZettel returns a zettel as a byte slice.
//
// part must be one of "meta", "content", or "zettel".
//
// The format of the byte slice is described in [Layout of a zettel].
//
// [Layout of a zettel]: https://zettelstore.de/manual/h/00001006000000
func (c *Client) GetZettel(ctx context.Context, zid id.Zid, part string) ([]byte, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	if part != "" && part != api.PartContent {
		ub.AppendKVQuery(api.QueryKeyPart, part)
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
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
func (c *Client) GetZettelData(ctx context.Context, zid id.Zid) (api.ZettelData, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	ub.AppendKVQuery(api.QueryKeyPart, api.PartZettel)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err == nil {
		defer func() { _ = resp.Body.Close() }()
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

// GetParsedZettel return a parsed zettel in a specified text-based encoding.
//
// A parsed zettel is just read from its box and is not processed any further.
//
// Valid encoding values are given as constants. They are described in more
// detail in [Encodings available via the API].
//
// [Encodings available via the API]: https://zettelstore.de/manual/h/00001012920500
func (c *Client) GetParsedZettel(ctx context.Context, zid id.Zid, enc api.EncodingEnum) ([]byte, error) {
	return c.getZettelString(ctx, zid, enc, true)
}

// GetEvaluatedZettel return an evaluated zettel in a specified text-based encoding.
//
// An evaluated zettel was parsed, and any transclusions etc. are resolved.
// This is the zettel representation you typically see on the Web UI.
//
// Valid encoding values are given as constants. They are described in more
// detail in [Encodings available via the API].
//
// [Encodings available via the API]: https://zettelstore.de/manual/h/00001012920500
func (c *Client) GetEvaluatedZettel(ctx context.Context, zid id.Zid, enc api.EncodingEnum) ([]byte, error) {
	return c.getZettelString(ctx, zid, enc, false)
}

func (c *Client) getZettelString(ctx context.Context, zid id.Zid, enc api.EncodingEnum, parseOnly bool) ([]byte, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, enc.String())
	ub.AppendKVQuery(api.QueryKeyPart, api.PartContent)
	if parseOnly {
		ub.AppendKVQuery(api.QueryKeyParseOnly, "")
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNoContent:
	default:
		return nil, statusToError(resp)
	}
	return io.ReadAll(resp.Body)
}

// GetParsedSz returns a part of an parsed zettel as a Sexpr-decoded data structure.
//
// A parsed zettel is just read from its box and is not processed any further.
//
// part must be one of "meta", "content", or "zettel".
//
// Basically, this function returns the sz encoding of a part of a zettel.
func (c *Client) GetParsedSz(ctx context.Context, zid id.Zid, part string) (sx.Object, error) {
	return c.getSz(ctx, zid, part, true)
}

// GetEvaluatedSz returns an evaluated zettel as a Sexpr-decoded data structure.
//
// An evaluated zettel was parsed, and any transclusions etc. are resolved.
// This is the zettel representation you typically see on the Web UI.
//
// part must be one of "meta", "content", or "zettel".
//
// Basically, this function returns the sz encoding of a part of a zettel.
func (c *Client) GetEvaluatedSz(ctx context.Context, zid id.Zid, part string) (sx.Object, error) {
	return c.getSz(ctx, zid, part, false)
}

func (c *Client) getSz(ctx context.Context, zid id.Zid, part string, parseOnly bool) (sx.Object, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingSz)
	if part != "" {
		ub.AppendKVQuery(api.QueryKeyPart, part)
	}
	if parseOnly {
		ub.AppendKVQuery(api.QueryKeyParseOnly, "")
	}
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, statusToError(resp)
	}
	return sxreader.MakeReader(bufio.NewReaderSize(resp.Body, 8)).Read()
}

// GetMetaData returns the metadata of a zettel.
func (c *Client) GetMetaData(ctx context.Context, zid id.Zid) (api.MetaRights, error) {
	ub := c.NewURLBuilder('z').SetZid(zid)
	ub.AppendKVQuery(api.QueryKeyEncoding, api.EncodingData)
	ub.AppendKVQuery(api.QueryKeyPart, api.PartMeta)
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return api.MetaRights{}, err
	}
	defer func() { _ = resp.Body.Close() }()
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

// GetVersionInfo returns version information of the Zettelstore that is used.
func (c *Client) GetVersionInfo(ctx context.Context) (VersionInfo, error) {
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, c.NewURLBuilder('x'), nil)
	if err != nil {
		return VersionInfo{}, err
	}
	defer func() { _ = resp.Body.Close() }()
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

// VersionInfo contains version information of the associated Zettelstore.
//
//   - Major is an integer containing the major software version of Zettelstore.
//     If its value is greater than zero, different major versions are not compatible.
//   - Minor is an integer specifying the minor software version for the given major version.
//     If the major version is greater than zero, minor versions are backward compatible.
//   - Patch is an integer that specifies a change within a minor version.
//     A version that have equal major and minor versions and differ in patch version are
//     always compatible, even if the major version equals zero.
//   - Info contains some optional text, i.e. it may be the empty string. Typically, Info
//     specifies a developer version by containing the string "dev".
//   - Hash contains the value of the source code version stored in the Zettelstore repository.
//     You can use it to reproduce bugs that occured, when source code was changed since
//     its introduction.
type VersionInfo struct {
	Major int
	Minor int
	Patch int
	Info  string
	Hash  string
}

// GetApplicationZid returns the zettel identifier used to configure a client
// application with the given name.
func (c *Client) GetApplicationZid(ctx context.Context, appname string) (id.Zid, error) {
	mr, err := c.GetMetaData(ctx, id.ZidAppDirectory)
	if err != nil {
		return id.Invalid, err
	}
	key := appname + "-zid"
	val, found := mr.Meta[key]
	if !found {
		return id.Invalid, fmt.Errorf("no application registered: %v", appname)
	}
	zid, err := id.Parse(val)
	if err == nil {
		return zid, nil
	}
	return id.Invalid, fmt.Errorf("invalid identifier for application %v: %v", appname, val)
}

// Get executes a GET request to the given URL and returns the read data.
func (c *Client) Get(ctx context.Context, ub *api.URLBuilder) ([]byte, error) {
	resp, err := c.buildAndExecuteRequest(ctx, http.MethodGet, ub, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
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
