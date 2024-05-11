// Copyright (c) 2019-2024 Duke Leto and The Hush developers
// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the GPLv3 software license
// Package frontend implements the gRPC handlers called by the wallets.
package frontend

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/btcsuite/btcd/rpcclient"
	"git.hush.is/hush/lightwalletd/common"
	"git.hush.is/hush/lightwalletd/parser"
	"git.hush.is/hush/lightwalletd/walletrpc"
)

type lwdStreamer struct {
	cache      *common.BlockCache
	chainName  string
	pingEnable bool
	mutex      sync.Mutex
	walletrpc.UnimplementedCompactTxStreamerServer
	log    *logrus.Entry
	client *rpcclient.Client
}

// NewLwdStreamer constructs a gRPC context.
func NewLwdStreamer(cache *common.BlockCache, chainName string, enablePing bool) (walletrpc.CompactTxStreamerServer, error) {
	return &lwdStreamer{cache: cache, chainName: chainName, pingEnable: enablePing}, nil
}

var (
	ErrUnspecified = errors.New("request for unspecified identifier")
)

// Test to make sure Address is a single transparent address
func checkTaddress(taddr string) error {
	match, err := regexp.Match("\\AR[a-zA-Z0-9]{33}\\z", []byte(taddr))
	if err != nil || !match {
		return errors.New("invalid address")
	}
	return nil
}

// GetLatestBlock returns the height of the best chain, according to hushd
func (s *lwdStreamer) GetLatestBlock(ctx context.Context, placeholder *walletrpc.ChainSpec) (*walletrpc.BlockID, error) {
	// Lock to ensure we return consistent height and hash
	s.mutex.Lock()
	defer s.mutex.Unlock()
	blockChainInfo, err := common.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}
	bestBlockHash, err := hex.DecodeString(blockChainInfo.BestBlockHash)
	if err != nil {
		return nil, err
	}
	return &walletrpc.BlockID{Height: uint64(blockChainInfo.Blocks), Hash: []byte(bestBlockHash)}, nil
}

// GetTaddressTxids is a streaming RPC that returns transaction IDs that have
// the given transparent address (taddr) as either an input or output.
func (s *lwdStreamer) GetTaddressTxids(addressBlockFilter *walletrpc.TransparentAddressBlockFilter, resp walletrpc.CompactTxStreamer_GetTaddressTxidsServer) error {
	if err := checkTaddress(addressBlockFilter.Address); err != nil {
		return err
	}

	if addressBlockFilter.Range == nil {
		return errors.New("must specify block range")
	}
	if addressBlockFilter.Range.Start == nil {
		return errors.New("must specify a start block height")
	}
	if addressBlockFilter.Range.End == nil {
		return errors.New("must specify an end block height")
	}
	params := make([]json.RawMessage, 1)
	request := &common.HushdRpcRequestGetaddresstxids{
		Addresses: []string{addressBlockFilter.Address},
		Start:     addressBlockFilter.Range.Start.Height,
		End:       addressBlockFilter.Range.End.Height,
	}
	param, err := json.Marshal(request)
	if err != nil {
		return err
	}
	params[0] = param
	result, rpcErr := common.CallRpcWithRetries("getaddresstxids", params)

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		return rpcErr
	}

	var txids []string
	err = json.Unmarshal(result, &txids)
	if err != nil {
		return err
	}

	timeout, cancel := context.WithTimeout(resp.Context(), 30*time.Second)
	defer cancel()

	for _, txidstr := range txids {
		txid, _ := hex.DecodeString(txidstr)
		// Txid is read as a string, which is in big-endian order. But when converting
		// to bytes, it should be little-endian
		tx, err := s.GetTransaction(timeout, &walletrpc.TxFilter{Hash: parser.Reverse(txid)})
		if err != nil {
			return err
		}
		if err = resp.Send(tx); err != nil {
			return err
		}
	}
	return nil
}

// this is the old name of GetTaddressTxids to provide backcompat to old clients
func (s *lwdStreamer) GetAddressTxids(addressBlockFilter *walletrpc.TransparentAddressBlockFilter, resp walletrpc.CompactTxStreamer_GetAddressTxidsServer) error {
    return s.GetTaddressTxids(addressBlockFilter, resp)
}

// GetBlock returns the compact block at the requested height. Requesting a
// block by hash is not yet supported.
func (s *lwdStreamer) GetBlock(ctx context.Context, id *walletrpc.BlockID) (*walletrpc.CompactBlock, error) {
	if id.Height == 0 && id.Hash == nil {
		return nil, errors.New("request for unspecified identifier")
	}

	// Precedence: a hash is more specific than a height. If we have it, use it first.
	if id.Hash != nil {
		// TODO: Get block by hash
		return nil, errors.New("gRPC GetBlock by Hash is not yet implemented")
	}
	cBlock, err := common.GetBlock(s.cache, int(id.Height))

	if err != nil {
		return nil, err
	}

	return cBlock, err
}

// GetBlockRange is a streaming RPC that returns blocks, in compact form,
// (as also returned by GetBlock) from the block height 'start' to height
// 'end' inclusively.
func (s *lwdStreamer) GetBlockRange(span *walletrpc.BlockRange, resp walletrpc.CompactTxStreamer_GetBlockRangeServer) error {
	blockChan := make(chan *walletrpc.CompactBlock)
	if span.Start == nil || span.End == nil {
		return errors.New("must specify start and end heights")
	}
	errChan := make(chan error)
	go common.GetBlockRange(s.cache, blockChan, errChan, int(span.Start.Height), int(span.End.Height))

	for {
		select {
		case err := <-errChan:
			// this will also catch context.DeadlineExceeded from the timeout
			return err
		case cBlock := <-blockChan:
			err := resp.Send(cBlock)
			if err != nil {
				return err
			}
		}
	}
}

func (s *lwdStreamer) GetLatestTreeState(ctx context.Context, in *walletrpc.Empty) (*walletrpc.TreeState, error) {
	blockChainInfo, err := common.GetBlockChainInfo()
	if err != nil {
		return nil, err
	}
	latestHeight := blockChainInfo.Blocks
	return s.GetTreeState(ctx, &walletrpc.BlockID{Height: uint64(latestHeight)})
}

// GetTransaction returns the raw transaction bytes that are returned by the 'getrawtransaction' RPC
func (s *lwdStreamer) GetTransaction(ctx context.Context, txf *walletrpc.TxFilter) (*walletrpc.RawTransaction, error) {
	if txf.Hash != nil {
		if len(txf.Hash) != 32 {
			return nil, errors.New("transaction ID has invalid length")
		}
		leHashStringJSON, err := json.Marshal(hex.EncodeToString(parser.Reverse(txf.Hash)))
		if err != nil {
			return nil, err
		}
		params := []json.RawMessage{
			leHashStringJSON,
			json.RawMessage("1"),
		}
		result, rpcErr := common.CallRpcWithRetries("getrawtransaction", params)

		// For some reason, the error responses are not JSON
		if rpcErr != nil {
			return nil, rpcErr
		}
		// Many other fields are returned, but we need only these two.
		var txinfo common.HushdRpcReplyGetrawtransaction
		err = json.Unmarshal(result, &txinfo)
		if err != nil {
			return nil, err
		}
		txBytes, err := hex.DecodeString(txinfo.Hex)
		if err != nil {
			return nil, err
		}
		return &walletrpc.RawTransaction{
			Data:   txBytes,
			Height: uint64(txinfo.Height),
		}, nil
	}

	if txf.Block != nil && txf.Block.Hash != nil {
		return nil, errors.New("can't GetTransaction with a blockhash+num, please call GetTransaction with txid")
	}
	return nil, errors.New("please call GetTransaction with txid")
}

// GetLightdInfo gets the lightwalletd (this server) info, and includes information
// it gets from its backend hushd
func (s *lwdStreamer) GetLightdInfo(ctx context.Context, in *walletrpc.Empty) (*walletrpc.LightdInfo, error) {
	return common.GetLightdInfo()
}


// GetCoinsupply gets the Coinsupply  info
func (s *lwdStreamer) GetCoinsupply(ctx context.Context, in *walletrpc.Empty) (*walletrpc.Coinsupply, error) {
	result, coin, height, supply, zfunds, total, err := common.GetCoinsupply()

	if err != nil {
		s.log.WithFields(logrus.Fields{
			"error": err,
		}).Warn("Unable to get Coinsupply")
		return nil, err
	}

	// TODO these are called Error but they aren't at the moment.
	// A success will return code 0 and message txhash.
	return &walletrpc.Coinsupply{
		Result: result,
		Coin:   coin,
		Height: uint64(height),
		Supply: uint64(supply),
		Zfunds: uint64(zfunds),
		Total:  uint64(total),
	}, nil
}

// SendTransaction forwards raw transaction bytes to a full node over JSON-RPC
func (s *lwdStreamer) SendTransaction(ctx context.Context, rawtx *walletrpc.RawTransaction) (*walletrpc.SendResponse, error) {
	// sendrawtransaction "hexstring" ( allowhighfees )
	//
	// Submits raw transaction (binary) to local node and network.
	//
	// Result:
	// "hex"             (string) The transaction hash in hex

	// Verify rawtx
	if rawtx == nil || rawtx.Data == nil {
		return nil, errors.New("bad transaction data")
	}

	// Construct raw JSON-RPC params
	params := make([]json.RawMessage, 1)
	txJSON, err := json.Marshal(hex.EncodeToString(rawtx.Data))
	if err != nil {
		return &walletrpc.SendResponse{}, err
	}
	params[0] = txJSON
	result, rpcErr := common.CallRpcWithRetries("sendrawtransaction", params)

	var errCode int64
	var errMsg string

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		errParts := strings.SplitN(rpcErr.Error(), ":", 2)
		if len(errParts) < 2 {
			return nil, errors.New("sendTransaction couldn't parse error code")
		}
		errMsg = strings.TrimSpace(errParts[1])
		errCode, err = strconv.ParseInt(errParts[0], 10, 32)
		if err != nil {
			// This should never happen. We can't panic here, but it's that class of error.
			// This is why we need integration testing to work better than regtest currently does. TODO.
			return nil, errors.New("sendTransaction couldn't parse error code")
		}
	} else {
		errMsg = string(result)
	}

	// TODO these are called Error but they aren't at the moment.
	// A success will return code 0 and message txhash.
	return &walletrpc.SendResponse{
		ErrorCode:    int32(errCode),
		ErrorMessage: errMsg,
	}, nil
}

func getTaddressBalanceHushdRpc(addressList []string) (*walletrpc.Balance, error) {
	for _, addr := range addressList {
		if err := checkTaddress(addr); err != nil {
			return &walletrpc.Balance{}, err
		}
	}
	params := make([]json.RawMessage, 1)
	addrList := &common.HushdRpcRequestGetaddressbalance{
		Addresses: addressList,
	}
	param, err := json.Marshal(addrList)
	if err != nil {
		return &walletrpc.Balance{}, err
	}
	params[0] = param

	result, rpcErr := common.CallRpcWithRetries("getaddressbalance", params)
	if rpcErr != nil {
		return &walletrpc.Balance{}, rpcErr
	}
	var balanceReply common.HushdRpcReplyGetaddressbalance
	err = json.Unmarshal(result, &balanceReply)
	if err != nil {
		return &walletrpc.Balance{}, err
	}
	return &walletrpc.Balance{ValueZat: balanceReply.Balance}, nil
}


func getAddressUtxos(arg *walletrpc.GetAddressUtxosArg, f func(*walletrpc.GetAddressUtxosReply) error) error {
	for _, a := range arg.Addresses {
		if err := checkTaddress(a); err != nil {
			return err
		}
	}
	params := make([]json.RawMessage, 1)
	addrList := &common.HushdRpcRequestGetaddressutxos{
		Addresses: arg.Addresses,
	}
	param, err := json.Marshal(addrList)
	if err != nil {
		return err
	}
	params[0] = param
	result, rpcErr := common.CallRpcWithRetries("getaddressutxos", params)
	if rpcErr != nil {
		return rpcErr
	}
	var utxosReply []common.HushdRpcReplyGetaddressutxos
	err = json.Unmarshal(result, &utxosReply)
	if err != nil {
		return err
	}
	n := 0
	for _, utxo := range utxosReply {
		if uint64(utxo.Height) < arg.StartHeight {
			continue
		}
		n++
		if arg.MaxEntries > 0 && uint32(n) > arg.MaxEntries {
			break
		}
		txidBytes, err := hex.DecodeString(utxo.Txid)
		if err != nil {
			return err
		}
		scriptBytes, err := hex.DecodeString(utxo.Script)
		if err != nil {
			return err
		}
		err = f(&walletrpc.GetAddressUtxosReply{
			Address:  utxo.Address,
			Txid:     parser.Reverse(txidBytes),
			Index:    int32(utxo.OutputIndex),
			Script:   scriptBytes,
			ValueZat: int64(utxo.Satoshis),
			Height:   uint64(utxo.Height),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// GetTaddressBalance returns the total balance for a list of taddrs
func (s *lwdStreamer) GetTaddressBalance(ctx context.Context, addresses *walletrpc.AddressList) (*walletrpc.Balance, error) {
	return getTaddressBalanceHushdRpc(addresses.Addresses)
}

// GetTaddressBalanceStream returns the total balance for a list of taddrs
func (s *lwdStreamer) GetTaddressBalanceStream(addresses walletrpc.CompactTxStreamer_GetTaddressBalanceStreamServer) error {
	addressList := make([]string, 0)
	for {
		addr, err := addresses.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		addressList = append(addressList, addr.Address)
	}
	balance, err := getTaddressBalanceHushdRpc(addressList)
	if err != nil {
		return err
	}
	addresses.SendAndClose(balance)
	return nil
}

func (s *lwdStreamer) GetMempoolStream(_empty *walletrpc.Empty, resp walletrpc.CompactTxStreamer_GetMempoolStreamServer) error {
	err := common.GetMempool(func(tx *walletrpc.RawTransaction) error {
		return resp.Send(tx)
	})
	return err
}

// Key is 32-byte txid (as a 64-character string), data is pointer to compact tx.
var mempoolMap *map[string]*walletrpc.CompactTx
var mempoolList []string

// Last time we pulled a copy of the mempool from hushd
var lastMempool time.Time

func (s *lwdStreamer) GetMempoolTx(exclude *walletrpc.Exclude, resp walletrpc.CompactTxStreamer_GetMempoolTxServer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if time.Since(lastMempool).Seconds() >= 2 {
		lastMempool = time.Now()
		// Refresh our copy of the mempool.
		params := make([]json.RawMessage, 0)
		result, rpcErr := common.CallRpcWithRetries("getrawmempool", params)
		if rpcErr != nil {
			return rpcErr
		}
		err := json.Unmarshal(result, &mempoolList)
		if err != nil {
			return err
		}
		newmempoolMap := make(map[string]*walletrpc.CompactTx)
		if mempoolMap == nil {
			mempoolMap = &newmempoolMap
		}
		for _, txidstr := range mempoolList {
			if ctx, ok := (*mempoolMap)[txidstr]; ok {
				// This ctx has already been fetched, copy pointer to it.
				newmempoolMap[txidstr] = ctx
				continue
			}
			txidJSON, err := json.Marshal(txidstr)
			if err != nil {
				return err
			}
			// The "0" is because we only need the raw hex, which is returned as
			// just a hex string, and not even a json string (with quotes).
            params := []json.RawMessage{txidJSON, json.RawMessage("0")}
            result, rpcErr := common.CallRpcWithRetries("getrawtransaction", params)
			if rpcErr != nil {
				// Not an error; mempool transactions can disappear
				continue
			}
			// strip the quotes
			var txStr string
			err = json.Unmarshal(result, &txStr)
			if err != nil {
				return err
			}

			// conver to binary
			txBytes, err := hex.DecodeString(txStr)
			if err != nil {
				return err
			}
			tx := parser.NewTransaction()
			txdata, err := tx.ParseFromSlice(txBytes)
			if err != nil {
				return err
			}
			if len(txdata) > 0 {
				return errors.New("extra data deserializing transaction")
			}
			newmempoolMap[txidstr] = &walletrpc.CompactTx{}
			if tx.HasShieldedElements() {
				txidBytes, err := hex.DecodeString(txidstr)
				if err != nil {
					return err
				}
				tx.SetTxID(txidBytes)
				newmempoolMap[txidstr] = tx.ToCompact( /* height */ 0)
			}
		}
		mempoolMap = &newmempoolMap
	}
	excludeHex := make([]string, len(exclude.Txid))
	for i := 0; i < len(exclude.Txid); i++ {
		excludeHex[i] = hex.EncodeToString(parser.Reverse(exclude.Txid[i]))
	}
	for _, txid := range MempoolFilter(mempoolList, excludeHex) {
		tx := (*mempoolMap)[txid]
		if len(tx.Hash) > 0 {
			err := resp.Send(tx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Return the subset of items that aren't excluded, but
// if more than one item matches an exclude entry, return
// all those items.
func MempoolFilter(items, exclude []string) []string {
	sort.Slice(items, func(i, j int) bool {
		return items[i] < items[j]
	})
	sort.Slice(exclude, func(i, j int) bool {
		return exclude[i] < exclude[j]
	})
	// Determine how many items match each exclude item.
	nmatches := make([]int, len(exclude))
	// is the exclude string less than the item string?
	lessthan := func(e, i string) bool {
		l := len(e)
		if l > len(i) {
			l = len(i)
		}
		return e < i[0:l]
	}
	ei := 0
	for _, item := range items {
		for ei < len(exclude) && lessthan(exclude[ei], item) {
			ei++
		}
		match := ei < len(exclude) && strings.HasPrefix(item, exclude[ei])
		if match {
			nmatches[ei]++
		}
	}

	// Add each item that isn't uniquely excluded to the results.
	tosend := make([]string, 0)
	ei = 0
	for _, item := range items {
		for ei < len(exclude) && lessthan(exclude[ei], item) {
			ei++
		}
		match := ei < len(exclude) && strings.HasPrefix(item, exclude[ei])
		if !match || nmatches[ei] > 1 {
			tosend = append(tosend, item)
		}
	}
	return tosend
}



func (s *lwdStreamer) GetAddressUtxos(ctx context.Context, arg *walletrpc.GetAddressUtxosArg) (*walletrpc.GetAddressUtxosReplyList, error) {
	addressUtxos := make([]*walletrpc.GetAddressUtxosReply, 0)
	err := getAddressUtxos(arg, func(utxo *walletrpc.GetAddressUtxosReply) error {
		addressUtxos = append(addressUtxos, utxo)
		return nil
	})
	if err != nil {
		return &walletrpc.GetAddressUtxosReplyList{}, err
	}
	return &walletrpc.GetAddressUtxosReplyList{AddressUtxos: addressUtxos}, nil
}

func (s *lwdStreamer) GetAddressUtxosStream(arg *walletrpc.GetAddressUtxosArg, resp walletrpc.CompactTxStreamer_GetAddressUtxosStreamServer) error {
	err := getAddressUtxos(arg, func(utxo *walletrpc.GetAddressUtxosReply) error {
		return resp.Send(utxo)
	})
	if err != nil {
		return err
	}
	return nil
}


