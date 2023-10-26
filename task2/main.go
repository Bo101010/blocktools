package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/Bo101010/blocktools/erc20/gen"
)

var url = flag.String("url", "https://rpc.ankr.com/eth_goerli", "eth node rpc endpoint")
var hexPrivateKey = flag.String("private", "", "hex of privateKey, WARNING: DONT USE PRIVATEKEY IN PRODUCTION ENV")
var sendToAddr = flag.String("to", "", "account address receiving token, prefix with 0x")
var contractAddr = flag.String("contract", "0x5c486db7559adAC22516Ae7676750f5105A1F3d1", "contract address")
var amount = flag.String("amt", "", "token account * 10^decimals")

func main() {
	flag.Parse()
	sendAmt, ok := big.NewInt(0).SetString(*amount, 10)
	if !ok {
		panic("invalid amount:" + *amount)
	}

	initCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	rpcClient, err := rpc.DialContext(initCtx, *url)
	if err != nil {
		panic(err.Error())
	}
	client := ethclient.NewClient(rpcClient)
	chainID, err := client.ChainID(initCtx)
	if err != nil {
		panic(err.Error())
	}
	privaeKey, err := crypto.HexToECDSA(*hexPrivateKey)
	if err != nil {
		panic(err.Error())
	}
	opts, err := bind.NewKeyedTransactorWithChainID(privaeKey, chainID)
	if err != nil {
		panic(err.Error())
	}

	token, err := gen.NewERC20(common.HexToAddress(*contractAddr), client)
	if err != nil {
		panic(err.Error())
	}
	tx, err := token.Transfer(opts, common.HexToAddress(*sendToAddr), sendAmt)
	if err != nil {
		panic("send tx failed " + err.Error())
	}
	fmt.Println("txId: " + tx.Hash().Hex())
}
