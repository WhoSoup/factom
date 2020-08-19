// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factom_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	. "github.com/FactomProject/factom"

	"testing"
)

type walletcall struct {
	Jsonrpc string `json:"jsonrps"`
	Id      int    `json:"id"`
	Result  struct {
		FactoidAccountBalances struct {
			Ack   int64 `json:"ack"`
			Saved int64 `json:"saved"`
		} `json:"fctaccountbalances"`
		EntryCreditAccountBalances struct {
			Ack   int64 `json:"ack"`
			Saved int64 `json:"saved"`
		} `json:"ecaccountbalances"`
	} `json:"result"`
}

func helper(t *testing.T, addr []string) (*walletcall, string) {
	for _, k := range addr {
		if _, _, err := ImportAddresses(k); err != nil {
			return nil, "failed"
		}
	}

	url := "http://localhost:8089/v2"
	jsonStrEC := []byte(`{"jsonrpc": "2.0", "id": 0, "method": "wallet-balances"}`)
	reqEC, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStrEC))
	reqEC.Header.Set("content-type", "text/plain;")

	clientEC := &http.Client{}
	callRespEC, err := clientEC.Do(reqEC)
	if err != nil {
		t.Error(err)
	}

	defer callRespEC.Body.Close()
	bodyEC, _ := ioutil.ReadAll(callRespEC.Body)
	fmt.Println("BODY: ", string(bodyEC))

	respEC := new(walletcall)
	errEC := json.Unmarshal([]byte(bodyEC), &respEC)
	if errEC != nil {
		t.Error(errEC)
	}
	return respEC, ""
}

// helper functions for testing

func populateTestWallet() error {
	//FA3T1gTkuKGG2MWpAkskSoTnfjxZDKVaAYwziNTC1pAYH5B9A1rh
	//Fs2TCa7Mo4XGy9FQSoZS8JPnDfv7SjwUSGqrjMWvc1RJ9sKbJeXA
	//
	//FA3oaS2D2GkrZJuWuiDohnLruxV3AWbrM3PmG3HSSE7DHzPWio36
	//Fs1os7xg2mN9fTuJmaYZLk6EXz51x2wmmHr2365UAuPMJW3aNr25
	//
	//EC2CyGKaNddLFxrjkFgiaRZnk77b8iQia3Zj6h5fxFReAcDwCo3i
	//Es4KmwK65t9HCsibYzVDFrijvkgTFZKdEaEAgfMtYTPSVtM3NDSx
	//
	//EC2R4bPDj9WQ8eWA4X3K8NYfTkBh4HFvCopLBq48FyrNXNumSK6w
	//Es355qB6tWo1ZZRTK8cXpHjxGECXaPGw98AFCRJ6kxZ3J6vp1M2i

	_, _, err := ImportAddresses(
		"Fs2TCa7Mo4XGy9FQSoZS8JPnDfv7SjwUSGqrjMWvc1RJ9sKbJeXA",
		"Fs1os7xg2mN9fTuJmaYZLk6EXz51x2wmmHr2365UAuPMJW3aNr25",
		"Es4KmwK65t9HCsibYzVDFrijvkgTFZKdEaEAgfMtYTPSVtM3NDSx",
		"Es355qB6tWo1ZZRTK8cXpHjxGECXaPGw98AFCRJ6kxZ3J6vp1M2i",
	)
	if err != nil {
		return err
	}

	return nil
}
