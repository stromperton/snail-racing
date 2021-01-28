package main

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"

	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client/models"
	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	"github.com/MinterTeam/minter-go-sdk/v2/wallet"
	"github.com/buger/jsonparser"
)

var (
	nodeUrl         = "https://api.minter.one/v2"
	testnodeUrl     = "https://node-api.testnet.minter.network/v2"
	minterClient, _ = http_client.NewConcise(nodeUrl)
)

func SendCoin(flyt float64, fromAddress string, address string, privateKey string) (*models.SendTransactionResponse, error) {

	minGasPrice, _ := minterClient.MinGasPrice()
	gasPrice := minGasPrice.MinGasPrice
	nonce, _ := minterClient.Nonce(fromAddress)

	value, ok := new(big.Int).SetString(fmt.Sprintf("%.0f", flyt*1000000000000000000), 10)
	if !ok {
		fmt.Println("SetString: error")
		return nil, nil
	}

	data, _ := transaction.NewSendData().SetCoin(0).SetValue(value).SetTo(address)
	transactionsBuilder := transaction.NewBuilder(transaction.MainNetChainID)
	tx, _ := transactionsBuilder.NewTransaction(data)
	sign, _ := tx.SetNonce(nonce).SetGasPrice(uint8(gasPrice)).Sign(privateKey)
	encode, _ := sign.Encode()

	res, err := minterClient.SendTransaction(encode)

	return res, err
}

func GetBalance(address string) float64 {

	response, err := minterClient.Address(address)

	if err != nil {
		fmt.Println(err)
	}

	num, err := strconv.ParseFloat(response.BipValue, 64)
	return num / 1000000000000000000
}

func CreateWallet() (string, string) {
	walletData, _ := wallet.New()
	return walletData.Address, walletData.PrivateKey
}

func GetPrivateKeyFromMnemonic(mnemonic string) string {
	seed, _ := wallet.Seed(mnemonic)
	prKey, _ := wallet.PrivateKeyBySeed(seed)
	return prKey
}

func GetBipPrice() float64 {
	resp, err := http.Get("https://api.coingecko.com/api/v3/simple/price?ids=bip&vs_currencies=usd")
	if err != nil {
		fmt.Println(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)

	}

	price, err := jsonparser.GetFloat(body, "bip", "usd")
	if err != nil {
		fmt.Println(err)
	}

	return price

}
