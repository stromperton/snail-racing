package main

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"

	"github.com/MinterTeam/minter-go-sdk/api"
	"github.com/MinterTeam/minter-go-sdk/transaction"
	"github.com/MinterTeam/minter-go-sdk/wallet"
	"github.com/buger/jsonparser"
	"github.com/go-resty/resty/v2"
)

var (
	nodeUrl = "https://test.mnt.funfasy.dev/v0/"
	restyC  = resty.New().SetHeaders(map[string]string{
		"Content-Type":     "application/json",
		"X-Project-Id":     os.Getenv("FUNFASY_ID"),
		"X-Project-Secret": os.Getenv("FUNFASY_SECRET"),
	})

	minterClient = api.NewApiWithClient(nodeUrl, restyC)
)

func SendCoin(num int, fromAddress string, address string, privateKey string) (*api.SendTransactionResult, error) {
	snum := strconv.Itoa(num * 1000000000000000000)
	value, ok := new(big.Int).SetString(snum, 10)
	if !ok {
		fmt.Println("SetString: error")
		return nil, nil
	}
	data, _ := transaction.NewSendData().SetCoin("MNT").SetValue(value).SetTo(address)
	minGasPrice, _ := minterClient.MinGasPrice()
	gasPrice, _ := strconv.ParseUint(minGasPrice, 10, 8)
	nonce, _ := minterClient.Nonce(fromAddress)

	tx, _ := transaction.NewBuilder(transaction.TestNetChainID).NewTransaction(data)
	tx.SetNonce(nonce).SetGasPrice(uint8(gasPrice)).SetGasCoin("MNT")
	signedTx, _ := tx.Sign(privateKey)
	result, err := minterClient.SendTransaction(signedTx)

	return result, err
}

func GetBalance(address string) float64 {

	response, err := minterClient.Balance(address)

	if err != nil {
		fmt.Println(err)
	}

	num, err := strconv.ParseFloat(response["MNT"], 64)
	return num / 1000000000000000000
}

func CreateWallet() (string, string) {
	walletData, _ := wallet.Create()
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
