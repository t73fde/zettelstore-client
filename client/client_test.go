//-----------------------------------------------------------------------------
// Copyright (c) 2022-present Detlef Stern
//
// This file is part of zettelstore-client.
//
// Zettelstore client is licensed under the latest version of the EUPL
// (European Union Public License). Please see file LICENSE.txt for your rights
// and obligations under this license.
//-----------------------------------------------------------------------------

package client_test

import (
	"context"
	"flag"
	"net/http"
	"net/url"
	"testing"

	"zettelstore.de/c/api"
	"zettelstore.de/c/client"
	"zettelstore.de/c/sz"
	"zettelstore.de/sx.fossil/sxpf"
)

func TestZettelList(t *testing.T) {
	c := getClient()
	_, err := c.ListZettel(context.Background(), "")
	if err != nil {
		t.Error(err)
		return
	}
}

func TestGetProtectedZettel(t *testing.T) {
	c := getClient()
	_, err := c.GetZettel(context.Background(), api.ZidStartupConfiguration, api.PartZettel)
	if err != nil {
		if cErr, ok := err.(*client.Error); ok && cErr.StatusCode == http.StatusForbidden {
			return
		} else {
			t.Error(err)
		}
		return
	}
}

func TestGetSzZettel(t *testing.T) {
	c := getClient()
	sf := sxpf.MakeMappedFactory()
	var zetSyms sz.ZettelSymbols
	zetSyms.InitializeZettelSymbols(sf)
	value, err := c.GetEvaluatedSz(context.Background(), api.ZidDefaultHome, api.PartContent, sf)
	if err != nil {
		t.Error(err)
		return
	}
	if value.IsNil() {
		t.Error("No data")
	}
}

var baseURL string

func init() {
	flag.StringVar(&baseURL, "base-url", "http://localhost:23123/", "Base URL")
}

func getClient() *client.Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	return client.NewClient(u)
}

func TestBase(t *testing.T) {
	exp := baseURL
	got := getClient().Base()
	if exp != got {
		t.Errorf("Base: expected=%q, but got=%q", exp, got)
	}
}

// TestMain controls whether client API tests should run or not.
func TestMain(m *testing.M) {
	flag.Parse()
	if baseURL != "" {
		m.Run()
	}
}
