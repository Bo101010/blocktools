package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"strings"
	"sync"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
)

func completeVout(vout *btcjson.Vout) {
	if len(vout.ScriptPubKey.Addresses) != 0 {
		return
	}
	script, err := hex.DecodeString(vout.ScriptPubKey.Hex)
	if err != nil {
		panic(err.Error())
	}
	_, addressObjs, _, err := txscript.ExtractPkScriptAddrs(script, &chaincfg.MainNetParams)
	if err != nil {
		panic(err.Error())
	}
	vout.ScriptPubKey.Addresses = []string{}
	for _, addr := range addressObjs {
		vout.ScriptPubKey.Addresses = append(vout.ScriptPubKey.Addresses, addr.EncodeAddress())
	}
}

type txTask struct {
	id     int
	tx     btcjson.TxRawResult
	result chan string
}

func newTxTask(tx btcjson.TxRawResult, id int) *txTask {
	return &txTask{
		id:     id,
		tx:     tx,
		result: make(chan string, 1),
	}
}

type BlockAnalyser struct {
	client  *rpcclient.Client
	mtx     sync.RWMutex
	txCache map[string]*btcjson.TxRawResult

	n        int
	taskChan chan *txTask
	tasks    []*txTask
}

func NewBlockAnalyser(client *rpcclient.Client, workerN int, blockHash string) *BlockAnalyser {
	hash, err := chainhash.NewHashFromStr(blockHash)
	if err != nil {
		panic(err.Error())
	}
	block, err := client.GetBlockVerboseTx(hash)
	if err != nil {
		panic(err.Error())
	}
	taskChan := make(chan *txTask, len(block.Tx))
	tasks := make([]*txTask, 0, len(block.Tx))
	for i, tx := range block.Tx {
		task := newTxTask(tx, i)
		taskChan <- task
		tasks = append(tasks, task)
	}
	return &BlockAnalyser{
		client:   client,
		txCache:  make(map[string]*btcjson.TxRawResult),
		n:        workerN,
		taskChan: taskChan,
		tasks:    tasks,
	}
}

func (analyser *BlockAnalyser) Run() {
	for i := 0; i < analyser.n; i++ {
		go analyser.work()
	}
	for _, task := range analyser.tasks {
		fmt.Println(<-task.result)
	}
}

func (analyser *BlockAnalyser) work() {
	for task := range analyser.taskChan {
		analyser.analyseTx(task)
	}
}

func (analyser *BlockAnalyser) analyseTx(task *txTask) {
	client := analyser.client
	tx := task.tx
	id := task.id
	lines := []string{}

	lines = append(lines, fmt.Sprintf("Tx:%d %s", id, tx.Txid))
	lines = append(lines, "\tinputs:")
	var inputValue float64
	for _, input := range tx.Vin {
		if input.IsCoinBase() {
			lines = append(lines, fmt.Sprintf("\t\tcoinbase"))
			continue
		}

		var preTx *btcjson.TxRawResult
		analyser.mtx.RLock()
		preTx, _ = analyser.txCache[input.Txid]
		analyser.mtx.RUnlock()
		if preTx == nil {
			preTxHash, err := chainhash.NewHashFromStr(input.Txid)
			if err != nil {
				panic(err.Error())
			}
			preTx, err = client.GetRawTransactionVerbose(preTxHash)
			if err != nil {
				panic(err.Error())
			}
			analyser.mtx.Lock()
			analyser.txCache[input.Txid] = preTx
			analyser.mtx.Unlock()
		}

		preOut := preTx.Vout[input.Vout]
		completeVout(&preOut)
		inputValue += preOut.Value
		lines = append(lines, fmt.Sprintf("\t\t[%s] %f", strings.Join(preOut.ScriptPubKey.Addresses, ","), preOut.Value))
	}
	lines = append(lines, "\toutputs:")
	var outputValue float64
	for _, output := range tx.Vout {
		completeVout(&output)
		lines = append(lines, fmt.Sprintf("\t\t[%s] %f", strings.Join(output.ScriptPubKey.Addresses, ","), output.Value))
		outputValue += output.Value
	}
	if id != 0 {
		lines = append(lines, fmt.Sprintf("\ttxFee:%f", inputValue-outputValue))
	}

	task.result <- strings.Join(lines, "\n")
}

var httpsHost = flag.String("httpsHost", "morning-old-diamond.btc.discover.quiknode.pro/f9864dae2c85ff8be7000e49b753e89a8cc1396c", "https host")
var user = flag.String("user", "placeholder", "rpc server username, cant be empty string")
var pass = flag.String("pass", "placeholder", "rpc server password, cant be empty string")
var blockId = flag.String("blockId", "0000000000000000000453c1bdd26714aaa5dcc00708d7b07cdd3f7dd1ab34f6", "block hash")
var worker = flag.Int("worker", 4, "worker count")

func main() {
	flag.Parse()
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:         *httpsHost,
		User:         *user,
		Pass:         *pass,
		HTTPPostMode: true,
		DisableTLS:   false,
	}, nil)
	if err != nil {
		panic(err.Error())
	}
	defer client.Shutdown()

	NewBlockAnalyser(client, *worker, *blockId).Run()
}
