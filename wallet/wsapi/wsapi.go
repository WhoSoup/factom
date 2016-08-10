// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factom/wallet"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/web"
	
)

const APIVersion string = "2.0"

var (
	webServer *web.Server
	fctWallet *wallet.Wallet
)

func Start(w *wallet.Wallet, net string) {
	webServer = web.NewServer()
	fctWallet = w

	webServer.Post("/v2", handleV2)
	webServer.Get("/v2", handleV2)
	webServer.Run(net)
}

func Stop() {
	fctWallet.Close()
	webServer.Close()
}

func handleV2(ctx *web.Context) {
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		handleV2Error(ctx, nil, newInvalidRequestError())
		return
	}

	j, err := factom.ParseJSON2Request(string(body))
	if err != nil {
		handleV2Error(ctx, nil, newInvalidRequestError())
		return
	}

	jsonResp, jsonError := handleV2Request(j)

	if jsonError != nil {
		handleV2Error(ctx, j, jsonError)
		return
	}

	ctx.Write([]byte(jsonResp.String()))
}

func handleV2Request(j *factom.JSON2Request) (*factom.JSON2Response, *factom.JSONError) {
	var resp interface{}
	var jsonError *factom.JSONError
	params := []byte(j.Params)

	switch j.Method {
	case "address":
		resp, jsonError = handleAddress(params)
	case "all-addresses":
		resp, jsonError = handleAllAddresses(params)
	case "generate-ec-address":
		resp, jsonError = handleGenerateECAddress(params)
	case "generate-factoid-address":
		resp, jsonError = handleGenerateFactoidAddress(params)
	case "import-addresses":
		resp, jsonError = handleImportAddresses(params)
	case "import-mnemonic":
		resp, jsonError = handleImportMnemonic(params)
	case "wallet-backup":
		resp, jsonError = handleWalletBackup(params)
	case "all-transactions":
		resp, jsonError = handleAllTransactions(params)
	case "new-transaction":
		resp, jsonError = handleNewTransaction(params)
	case "delete-transaction":
		resp, jsonError = handleDeleteTransaction(params)
	case "transactions":
		resp, jsonError = handleTransactions(params)
	case "transaction-hash":
		resp, jsonError = handleTransactionHash(params)
	case "add-input":
		resp, jsonError = handleAddInput(params)
	case "add-output":
		resp, jsonError = handleAddOutput(params)
	case "add-ec-output":
		resp, jsonError = handleAddECOutput(params)
	case "add-fee":
		resp, jsonError = handleAddFee(params)
	case "sub-fee":
		resp, jsonError = handleSubFee(params)
	case "sign-transaction":
		resp, jsonError = handleSignTransaction(params)
	case "compose-transaction":
		resp, jsonError = handleComposeTransaction(params)
	case "properties":
		resp, jsonError = handleProperties(params)
	default:
		jsonError = newMethodNotFoundError()
	}
	if jsonError != nil {
		return nil, jsonError
	}

	jsonResp := factom.NewJSON2Response()
	jsonResp.ID = j.ID
	if b, err := json.Marshal(resp); err != nil {
		return nil, newCustomInternalError(err.Error())
	} else {
		jsonResp.Result = b
	}

	return jsonResp, nil
}

func handleAddress(params []byte) (interface{}, *factom.JSONError) {
	req := new(addressRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	resp := new(addressResponse)
	switch factom.AddressStringType(req.Address) {
	case factom.ECPub:
		e, err := fctWallet.GetECAddress(req.Address)
		if err != nil {
			return nil, newCustomInternalError(err.Error())
		}
		resp = mkAddressResponse(e)
	case factom.FactoidPub:
		f, err := fctWallet.GetFCTAddress(req.Address)
		if err != nil {
			return nil, newCustomInternalError(err.Error())
		}
		resp = mkAddressResponse(f)
	default:
		return nil, newCustomInternalError("Invalid address type")
	}

	return resp, nil
}

func handleAllAddresses(params []byte) (interface{}, *factom.JSONError) {
	resp := new(multiAddressResponse)

	fs, es, err := fctWallet.GetAllAddresses()
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	for _, f := range fs {
		a := mkAddressResponse(f)
		resp.Addresses = append(resp.Addresses, a)
	}
	for _, e := range es {
		a := mkAddressResponse(e)
		resp.Addresses = append(resp.Addresses, a)
	}

	return resp, nil
}

func handleGenerateFactoidAddress(params []byte) (interface{}, *factom.JSONError) {
	a, err := fctWallet.GenerateFCTAddress()
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}

	resp := mkAddressResponse(a)

	return resp, nil
}

func handleGenerateECAddress(params []byte) (interface{}, *factom.JSONError) {
	a, err := fctWallet.GenerateECAddress()
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}

	resp := mkAddressResponse(a)

	return resp, nil
}

func handleImportAddresses(params []byte) (interface{}, *factom.JSONError) {
	req := new(importRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	resp := new(multiAddressResponse)
	for _, v := range req.Addresses {
		switch factom.AddressStringType(v.Secret) {
		case factom.FactoidSec:
			f, err := factom.GetFactoidAddress(v.Secret)
			if err != nil {
				return nil, newCustomInternalError(err.Error())
			}
			if err := fctWallet.InsertFCTAddress(f); err != nil {
				return nil, newCustomInternalError(err.Error())
			}
			a := mkAddressResponse(f)
			resp.Addresses = append(resp.Addresses, a)
		case factom.ECSec:
			e, err := factom.GetECAddress(v.Secret)
			if err != nil {
				return nil, newCustomInternalError(err.Error())
			}
			if err := fctWallet.InsertECAddress(e); err != nil {
				return nil, newCustomInternalError(err.Error())
			}
			a := mkAddressResponse(e)
			resp.Addresses = append(resp.Addresses, a)
		default:
			return nil, newCustomInternalError("address could not be imported")
		}
	}
	return resp, nil
}

func handleImportMnemonic(params []byte) (interface{}, *factom.JSONError) {
	req := new(importMnemonicRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	f, err := factom.MakeFactoidAddressFromMnemonic(req.Words)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	if err := fctWallet.InsertFCTAddress(f); err != nil {
		return nil, newCustomInternalError(err.Error())
	}

	return mkAddressResponse(f), nil
}

func handleWalletBackup(params []byte) (interface{}, *factom.JSONError) {
	resp := new(walletBackupResponse)

	if seed, err := fctWallet.GetSeed(); err != nil {
		return nil, newCustomInternalError(err.Error())
	} else {
		resp.Seed = seed
	}

	fs, es, err := fctWallet.GetAllAddresses()
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	for _, f := range fs {
		a := mkAddressResponse(f)
		resp.Addresses = append(resp.Addresses, a)
	}
	for _, e := range es {
		a := mkAddressResponse(e)
		resp.Addresses = append(resp.Addresses, a)
	}

	return resp, nil
}

func handleAllTransactions(params []byte) (interface{}, *factom.JSONError) {
	type transactionList struct {
		Transactions []json.RawMessage `json:"transactions"`
	}

	if fctWallet.TXDB() == nil {
		return nil, newCustomInternalError(
			"Wallet does not have a transaction database")
	}
	
	resp := new(transactionList)
	txs, err := fctWallet.TXDB().GetAllTXs()
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	for _, tx := range txs {
		p, err := tx.JSONByte()
		if err != nil {
			return nil, newCustomInternalError(err.Error())
		}
		resp.Transactions = append(resp.Transactions, p)
	}
	
	return resp, nil
}

// transaction handlers

func handleNewTransaction(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	if err := fctWallet.NewTransaction(req.Name); err != nil {
		return nil, newCustomInternalError(err.Error())
	}

	t := fctWallet.GetTransactions()[req.Name]
	resp, err := mkTransactionResponse(t)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp.Name = req.Name

	return resp, nil
}

func handleDeleteTransaction(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	if err := fctWallet.DeleteTransaction(req.Name); err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp := transactionResponse{Name: req.Name}
	return resp, nil
}

func handleTransactions(params []byte) (interface{}, *factom.JSONError) {
	resp := new(multiTransactionResponse)
	txs := fctWallet.GetTransactions()

	for name, tx := range txs {
		r := transactionResponse{Name: name}
		r.TxID = hex.EncodeToString(tx.GetSigHash().Bytes())
		if i, err := tx.TotalInputs(); err != nil {
			return nil, newCustomInternalError(err.Error())
		} else {
			r.TotalInputs = i
		}
		if i, err := tx.TotalOutputs(); err != nil {
			return nil, newCustomInternalError(err.Error())
		} else {
			r.TotalOutputs = i
		}
		if i, err := tx.TotalECs(); err != nil {
			return nil, newCustomInternalError(err.Error())
		} else {
			r.TotalECOutputs = i
		}
		if t, err := tx.MarshalBinary(); err != nil {
			return nil, newCustomInternalError(err.Error())
		} else {
			r.RawTransaction = hex.EncodeToString(t)
		}

		resp.Transactions = append(resp.Transactions, r)
	}

	return resp, nil
}

func handleTransactionHash(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	resp := new(transactionResponse)
	txs := fctWallet.GetTransactions()

	for name, tx := range txs {
		if name == req.Name {
			resp.Name = name
			resp.TxID = tx.GetSigHash().String()
			return resp, nil
		}
	}

	return nil, newCustomInternalError("Transaction not found")
}

func handleAddInput(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionValueRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	if err := fctWallet.AddInput(req.Name, req.Address, req.Amount); err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	t := fctWallet.GetTransactions()[req.Name]
	resp, err := mkTransactionResponse(t)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp.Name = req.Name

	return resp, nil
}

func handleAddOutput(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionValueRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	if err := fctWallet.AddOutput(req.Name, req.Address, req.Amount); err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	t := fctWallet.GetTransactions()[req.Name]
	resp, err := mkTransactionResponse(t)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp.Name = req.Name

	return resp, nil
}

func handleAddECOutput(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionValueRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	if err := fctWallet.AddECOutput(req.Name, req.Address, req.Amount); err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	t := fctWallet.GetTransactions()[req.Name]
	resp, err := mkTransactionResponse(t)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp.Name = req.Name

	return resp, nil
}

func handleAddFee(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionAddressRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	rate, err := factom.GetRate()
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	if err := fctWallet.AddFee(req.Name, req.Address, rate); err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	t := fctWallet.GetTransactions()[req.Name]
	resp, err := mkTransactionResponse(t)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp.Name = req.Name

	return resp, nil
}

func handleSubFee(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionAddressRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	rate, err := factom.GetRate()
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	if err := fctWallet.SubFee(req.Name, req.Address, rate); err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	t := fctWallet.GetTransactions()[req.Name]
	resp, err := mkTransactionResponse(t)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp.Name = req.Name

	return resp, nil
}

func handleSignTransaction(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	if err := fctWallet.SignTransaction(req.Name); err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	t := fctWallet.GetTransactions()[req.Name]
	resp, err := mkTransactionResponse(t)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	resp.Name = req.Name

	return resp, nil
}

func handleComposeTransaction(params []byte) (interface{}, *factom.JSONError) {
	req := new(transactionRequest)
	if err := json.Unmarshal(params, req); err != nil {
		return nil, newInvalidParamsError()
	}

	t, err := fctWallet.ComposeTransaction(req.Name)
	if err != nil {
		return nil, newCustomInternalError(err.Error())
	}
	return t, nil
}

func handleProperties(params []byte) (interface{}, *factom.JSONError) {
	props := new(propertiesResponse)
	props.WalletVersion = fctWallet.GetProperties()
	return props, nil
}

// utility functions

type addressResponder interface {
	String() string
	SecString() string
}

func mkAddressResponse(a addressResponder) *addressResponse {
	r := new(addressResponse)
	r.Public = a.String()
	r.Secret = a.SecString()
	return r
}

func mkTransactionResponse(t *factoid.Transaction) (*transactionResponse, error) {
	r := new(transactionResponse)
	r.TxID = hex.EncodeToString(t.GetSigHash().Bytes())

	if i, err := t.TotalInputs(); err != nil {
		return nil, err
	} else {
		r.TotalInputs = i
	}

	if i, err := t.TotalOutputs(); err != nil {
		return nil, err
	} else {
		r.TotalOutputs = i
	}

	if i, err := t.TotalECs(); err != nil {
		return nil, err
	} else {
		r.TotalECOutputs = i
	}

	return r, nil
}
