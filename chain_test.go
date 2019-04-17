// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factom_test

import (
	"testing"

	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/FactomProject/factom"
)

func TestNewChain(t *testing.T) {
	ent := new(Entry)
	ent.ChainID = ""
	ent.Content = []byte("This is a test Entry.")
	ent.ExtIDs = append(ent.ExtIDs, []byte("This is the first extid."))
	ent.ExtIDs = append(ent.ExtIDs, []byte("This is the second extid."))

	newChain := NewChain(ent)
	expectedID := "5a402200c5cf278e47905ce52d7d64529a0291829a7bd230072c5468be709069"

	if newChain.ChainID != expectedID {
		t.Errorf("expected:%s\nrecieved:%s", expectedID, newChain.ChainID)
	}
	t.Log(newChain.ChainID)

	cfb := NewChainFromBytes(ent.Content, ent.ExtIDs...)
	if cfb.ChainID != expectedID {
		t.Errorf("expected:%s\nrecieved:%s", expectedID, cfb.ChainID)
	}
	t.Log(cfb.ChainID)

	cfs := NewChainFromStrings(
		"This is a test Entry.",
		"This is the first extid.",
		"This is the second extid.",
	)
	if cfs.ChainID != expectedID {
		t.Errorf("expected:%s\nrecieved:%s", expectedID, cfs.ChainID)
	}
	t.Log(cfs.ChainID)
}

func TestIfExists(t *testing.T) {
	simlatedFactomdResponse := `{
  "jsonrpc": "2.0",
  "id": 0,
  "result": {
    "ChainHead": "f65f67774139fa78344dcdd302631a0d646db0c2be4d58e3e48b2a188c1b856c"
  }
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, simlatedFactomdResponse)
	}))
	defer ts.Close()

	url := ts.URL[7:]
	SetFactomdServer(url)

	expectedID := "f65f67774139fa78344dcdd302631a0d646db0c2be4d58e3e48b2a188c1b856c"
	//fmt.Println(ChainExists(expectedID))
	if ChainExists(expectedID) != true {
		t.Errorf("chain %s does not exist", expectedID)
	}
}

func TestIfNotExists(t *testing.T) {
	simlatedFactomdResponse := `{"jsonrpc":"2.0","id":0,"error":{"code":-32009,"message":"Missing Chain Head"}}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, simlatedFactomdResponse)
	}))
	defer ts.Close()

	url := ts.URL[7:]
	SetFactomdServer(url)
	unexpectedID := "5a402200c5cf278e47905ce52d7d64529a0291829a7bd230072c5468be709069"

	if ChainExists(unexpectedID) != false {
		t.Errorf("chain %s shouldn't exist", unexpectedID)
	}
}

func TestComposeChainCommit(t *testing.T) {
	type response struct {
		Message string `json:"message"`
	}
	ecAddr, err := GetECAddress("Es2Rf7iM6PdsqfYCo3D1tnAR65SkLENyWJG1deUzpRMQmbh9F3eG")
	if err != nil {
		t.Error(err)
	}

	ent := new(Entry)
	ent.ChainID = "954d5a49fd70d9b8bcdb35d252267829957f7ef7fa6c74f88419bdc5e82209f4"
	ent.Content = []byte("test!")
	ent.ExtIDs = append(ent.ExtIDs, []byte("test"))
	newChain := NewChain(ent)

	cCommit, err := ComposeChainCommit(newChain, ecAddr)
	if err != nil {
		t.Error(err)
	}

	r := new(response)
	json.Unmarshal(cCommit.Params, r)
	binCommit, _ := hex.DecodeString(r.Message)
	t.Logf("%x", binCommit)

	//the commit has a timestamp which is updated new for each time it is called.  This means it is different after each call.
	//we will check the non-changing parts

	if len(binCommit) != 200 {
		t.Error("expected commit to be 200 bytes long, instead got", len(binCommit))
	}
	result := binCommit[0:1]
	expected := []byte{0x00}
	if !bytes.Equal(result, expected) {
		t.Errorf("expected:%s\nrecieved:%s", expected, result)
	}
	//skip the 6 bytes of the timestamp
	result = binCommit[7:136]
	expected, err = hex.DecodeString("516870d4c0e1ee2d5f0d415e51fc10ae6b8d895561e9314afdc33048194d76f07cc61c8a81aea23d76ff6447689757dc1e36af66e300ce3e06b8d816c79acfd2285ed45081d5b8819a678d13c7c2d04f704b34c74e8aaecd9bd34609bee047200b3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29")
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(result, expected) {
		t.Errorf("expected:%s\nrecieved:%s", expected, result)
	}
}

func TestComposeChainReveal(t *testing.T) {
	ent := new(Entry)
	ent.ChainID = "954d5a49fd70d9b8bcdb35d252267829957f7ef7fa6c74f88419bdc5e82209f4"
	ent.Content = []byte("test!")
	ent.ExtIDs = append(ent.ExtIDs, []byte("test"))
	newChain := NewChain(ent)

	cReveal, err := ComposeChainReveal(newChain)
	if err != nil {
		t.Error(err)
	}

	expectedResponse := `{"entry":"00954d5a49fd70d9b8bcdb35d252267829957f7ef7fa6c74f88419bdc5e82209f400060004746573747465737421"}`
	if expectedResponse != string(cReveal.Params) {
		t.Errorf("expected:%s\nrecieved:%s", expectedResponse, cReveal.Params)
	}
}

func TestCommitChain(t *testing.T) {
	simlatedFactomdResponse := `{
   "jsonrpc":"2.0",
   "id":0,
   "result":{
      "message":"Chain Commit Success",
      "txid":"76e123d133a841fe3e08c5e3f3d392f8431f2d7668890c03f003f541efa8fc61"
   }
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, simlatedFactomdResponse)
	}))
	defer ts.Close()

	url := ts.URL[7:]
	SetFactomdServer(url)

	ent := new(Entry)
	ent.ChainID = "954d5a49fd70d9b8bcdb35d252267829957f7ef7fa6c74f88419bdc5e82209f4"
	ent.Content = []byte("test!")
	ent.ExtIDs = append(ent.ExtIDs, []byte("test"))
	newChain := NewChain(ent)
	ecAddr, err := GetECAddress("Es2Rf7iM6PdsqfYCo3D1tnAR65SkLENyWJG1deUzpRMQmbh9F3eG")
	if err != nil {
		t.Error(err)
	}

	expectedResponse := "76e123d133a841fe3e08c5e3f3d392f8431f2d7668890c03f003f541efa8fc61"
	response, _ := CommitChain(newChain, ecAddr)

	if expectedResponse != response {
		t.Errorf("expected:%s\nrecieved:%s", expectedResponse, response)
	}
	t.Log(response)
}

func TestRevealChain(t *testing.T) {
	simlatedFactomdResponse := `{
  "jsonrpc": "2.0",
  "id": 0,
  "result": {
    "message": "Entry Reveal Success",
    "entryhash": "f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d60669854734"
  }
}`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, simlatedFactomdResponse)
	}))
	defer ts.Close()

	url := ts.URL[7:]
	SetFactomdServer(url)

	ent := new(Entry)
	ent.ChainID = "954d5a49fd70d9b8bcdb35d252267829957f7ef7fa6c74f88419bdc5e82209f4"
	ent.Content = []byte("test!")
	ent.ExtIDs = append(ent.ExtIDs, []byte("test"))
	newChain := NewChain(ent)

	expectedResponse := "f5c956749fc3eba4acc60fd485fb100e601070a44fcce54ff358d60669854734"
	response, err := RevealChain(newChain)
	if err != nil {
		t.Error(err)
	}

	if expectedResponse != response {
		t.Errorf("expected:%s\nrecieved:%s", expectedResponse, response)
	}
	t.Log(response)
}
