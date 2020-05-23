package main

import (
	"fmt"
	"math/big"
	"strconv"

	"github.com/MinterTeam/minter-go-sdk/api"
	"github.com/MinterTeam/minter-go-sdk/transaction"
	"github.com/MinterTeam/minter-go-sdk/wallet"
	"github.com/go-resty/resty/v2"
)

var (
	nodeUrl = "https://texas.mnt.funfasy.dev/v0/" //"https://api.minter.one/ https://texasnet.node-api.minter.network/"
	restyC  = resty.New().SetHeaders(map[string]string{
		"Content-Type":     "application/json",
		"X-Project-Id":     "5311bad1-99e0-462e-943e-6c6e77714d26",
		"X-Project-Secret": "67b63b2937087ddb1d266ec33b26787f",
	})

	minterClient = api.NewApiWithClient(nodeUrl, restyC)
)

func SendCoin(num string, address string, privateKey string) {
	value, ok := new(big.Int).SetString(num, 10)
	if !ok {
		fmt.Println("SetString: error")
		return
	}
	data, _ := transaction.NewSendData().SetCoin("MNT").SetValue(value).SetTo(address)
	minGasPrice, _ := minterClient.MinGasPrice()
	gasPrice, _ := strconv.ParseUint(minGasPrice, 10, 8)
	nonce, _ := minterClient.Nonce(address)

	tx, _ := transaction.NewBuilder(transaction.TestNetChainID).NewTransaction(data)
	tx.SetNonce(nonce).SetGasPrice(uint8(gasPrice)).SetGasCoin("MNT")
	signedTx, _ := tx.Sign(privateKey)
	minterClient.SendTransaction(signedTx)
}

func GetBalance(address string) string {

	response, err := minterClient.Balance(address)

	if err != nil {
		fmt.Println(err)
	}

	return response["MNT"]
}

func CreateWallet() (string, string) {
	walletData, _ := wallet.Create()
	return walletData.Address, walletData.PrivateKey
}
