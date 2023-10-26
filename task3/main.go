package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

var url = flag.String("url", "https://rpc.ankr.com/eth_goerli", "eth node rpc endpoint")
var block = flag.Int64("block", 9933867, "block number")

var eventSeletor = common.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")

func isERC20Contract(addr common.Address) bool {
	// TODO: check verified contract ABI
	return true
}

func parseAddress(hash common.Hash) (addr common.Address, ok bool) {
	for i := 0; i < 12; i++ {
		if hash[i] != 0 {
			return
		}
	}
	addr = common.BytesToAddress(hash[:])
	ok = true
	return
}

func parseErc20TransferLog(log *types.Log, dataParser abi.Arguments) (from common.Address, to common.Address, amount *big.Int, ok bool) {
	if len(log.Topics) != 3 {
		return
	}
	if log.Topics[0] != eventSeletor {
		return
	}
	from, ok = parseAddress(log.Topics[1])
	if !ok {
		return
	}
	to, ok = parseAddress(log.Topics[2])
	if !ok {
		return
	}
	params, err := dataParser.Unpack(log.Data)
	if err != nil {
		return
	}
	if len(params) != 1 {
		return
	}
	amount, ok = params[0].(*big.Int)
	if !ok {
		return
	}
	ok = isERC20Contract(log.Address)
	return
}

func ABIParser(s string) (abi.Arguments, error) {
	selector, err := abi.ParseSelector(fmt.Sprintf("noname(%s)", s))
	if err != nil {
		return nil, err
	}
	args := abi.Arguments{}
	for _, input := range selector.Inputs {
		aType, err := abi.NewType(input.Type, input.InternalType, input.Components)
		if err != nil {
			return nil, err
		}
		args = append(args, abi.Argument{
			Name: input.Name,
			Type: aType,
		})
	}
	return args, nil
}

func main() {
	flag.Parse()

	dataParser, _ := ABIParser("uint256")

	initCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	rpcClient, err := rpc.DialContext(initCtx, *url)
	if err != nil {
		panic(err.Error())
	}
	client := ethclient.NewClient(rpcClient)

	logs, err := client.FilterLogs(initCtx, ethereum.FilterQuery{
		FromBlock: big.NewInt(*block),
		Topics: [][]common.Hash{
			{eventSeletor},
		},
	})
	if err != nil {
		panic(err.Error())
	}
	for _, log := range logs {
		from, to, amt, ok := parseErc20TransferLog(&log, dataParser)
		if !ok {
			continue
		}
		fmt.Printf("tx:%s erc20Token:%s transferFrom %s to %s amount:%s\n", log.TxHash, log.Address, from, to, amt)
	}
}
