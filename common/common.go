// Copyright (c) 2019-2023 Duke Leto and The Hush developers
// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the GPLv3 software license
package common

import (
    "fmt"
    "encoding/hex"
    "encoding/json"
    "strconv"
    "strings"
    "time"
    "git.hush.is/hush/lightwalletd/parser"
    "git.hush.is/hush/lightwalletd/walletrpc"
    "github.com/pkg/errors"
    "github.com/sirupsen/logrus"
)

// TODO: 'make build' will overwrite this string with the output of git-describe (tag)
var (
	Version   = "v0.1.3"
	GitCommit = ""
	Branch    = ""
	BuildDate = ""
	BuildUser = ""
)

type Options struct {
	GRPCBindAddr        string `json:"grpc_bind_address,omitempty"`
	GRPCLogging         bool   `json:"grpc_logging_insecure,omitempty"`
	HTTPBindAddr        string `json:"http_bind_address,omitempty"`
	TLSCertPath         string `json:"tls_cert_path,omitempty"`
	TLSKeyPath          string `json:"tls_cert_key,omitempty"`
	LogLevel            uint64 `json:"log_level,omitempty"`
	LogFile             string `json:"log_file,omitempty"`
	HushConfPath        string `json:"hush_conf,omitempty"`
	RPCUser             string `json:"rpcuser"`
	RPCPassword         string `json:"rpcpassword"`
	RPCHost             string `json:"rpchost"`
	RPCPort             string `json:"rpcport"`
	NoTLS               bool   `json:"no_tls,omitempty"`
	GenCertVeryInsecure bool   `json:"gen_cert_very_insecure,omitempty"`
	Redownload          bool   `json:"redownload"`
	SyncFromHeight      int    `json:"sync_from_height"`
	DataDir             string `json:"data_dir"`
	PingEnable          bool   `json:"ping_enable"`
	Darkside            bool   `json:"darkside"`
	DarksideTimeout     uint64 `json:"darkside_timeout"`
}

// RawRequest points to the function to send a an RPC request to hushd;
// in production, it points to btcsuite/btcd/rpcclient/rawrequest.go:RawRequest();
// in unit tests it points to a function to mock RPCs to hushd.
var RawRequest func(method string, params []json.RawMessage) (json.RawMessage, error)

// Time allows time-related functions to be mocked for testing,
// so that tests can be deterministic and so they don't require
// real time to elapse. In production, these point to the standard
// library `time` functions; in unit tests they point to mock
// functions (set by the specific test as required).
// More functions can be added later.
var Time struct {
	Sleep func(d time.Duration)
	Now   func() time.Time
}


// Log as a global variable simplifies logging
var Log *logrus.Entry

// The following are JSON hushd rpc requests and replies.
type (
	// hushd rpc "getblockchaininfo"
	Upgradeinfo struct {
		// unneeded fields can be omitted
		ActivationHeight int
		Status           string // "active"
	}
	ConsensusInfo struct { // consensus branch IDs
		Nextblock string // example: "e9ff75a6" (canopy)
		Chaintip  string // example: "e9ff75a6" (canopy)
	}
	HushdRpcReplyGetblockchaininfo struct {
		Chain           string
		Upgrades        map[string]Upgradeinfo
		Blocks          int
		BestBlockHash   string
		Consensus       ConsensusInfo
		EstimatedHeight int
	}

	// hushd rpc "getinfo"
	HushdRpcReplyGetinfo struct {
		Build      string
		Subversion string
	}

	// hushd rpc "getaddresstxids"
	HushdRpcRequestGetaddresstxids struct {
		Addresses []string `json:"addresses"`
		Start     uint64   `json:"start"`
		End       uint64   `json:"end"`
	}

	// TODO: hushd rpc "z_gettreestate"
	HushdRpcReplyGettreestate struct {
		Height  int
		Hash    string
		Time    uint32
		Sapling struct {
			Commitments struct {
				FinalState string
			}
			SkipHash string
		}
		Orchard struct {
			Commitments struct {
				FinalState string
			}
			SkipHash string
		}
	}

	// hushd rpc "getrawtransaction txid 1" (1 means verbose), there are
	// many more fields but these are the only ones we current need.
	HushdRpcReplyGetrawtransaction struct {
		Hex    string
		Height int
	}

	// hushd rpc "getaddressbalance"
	HushdRpcRequestGetaddressbalance struct {
		Addresses []string `json:"addresses"`
	}
	HushdRpcReplyGetaddressbalance struct {
		Balance int64
	}

	// hushd rpc "getaddressutxos"
	HushdRpcRequestGetaddressutxos struct {
		Addresses []string `json:"addresses"`
	}
	HushdRpcReplyGetaddressutxos struct {
		Address     string
		Txid        string
		OutputIndex int64
		Script      string
		Satoshis    uint64
		Height      int
	}

	// reply to getblock verbose=1 (json includes txid list)
	HushdRpcReplyGetblock1 struct {
		Hash string
		Tx   []string
	}
)

//TODO: this function is not currently used, but some of it's code
// needs to be implemented elsewhere
func GetSaplingInfo() (int, int, string, string, int, int, int, error) {
	result, rpcErr := CallRpcWithRetries("getblockchaininfo", []json.RawMessage{})

	var err error
	var errCode int64

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		errParts := strings.SplitN(rpcErr.Error(), ":", 2)
		errCode, err = strconv.ParseInt(errParts[0], 10, 32)
		//Check to see if we are requesting a height the hushd doesn't have yet
		if err == nil && errCode == -8 {
			return -1, -1, "", "", -1, -1, -1, nil
		}
		return -1, -1, "", "", -1, -1, -1, errors.Wrap(rpcErr, "error requesting block")
	}

	var f interface{}
	err = json.Unmarshal(result, &f)
	if err != nil {
		return -1, -1, "", "", -1, -1, -1, errors.Wrap(err, "error reading JSON response")
	}

	chainName := f.(map[string]interface{})["chain"].(string)

	upgradeJSON := f.(map[string]interface{})["upgrades"]
	saplingJSON := upgradeJSON.(map[string]interface{})["76b809bb"] // Sapling ID
	saplingHeight := saplingJSON.(map[string]interface{})["activationheight"].(float64)

	blockHeight := f.(map[string]interface{})["headers"].(float64)
	difficulty := f.(map[string]interface{})["difficulty"].(float64)
	longestchain := f.(map[string]interface{})["longestchain"].(float64)
	notarized := f.(map[string]interface{})["notarized"].(float64)

	consensus := f.(map[string]interface{})["consensus"]
	branchID := consensus.(map[string]interface{})["nextblock"].(string)

	return int(saplingHeight), int(blockHeight), chainName, branchID, int(difficulty), int(longestchain), int(notarized), nil
}

func GetLightdInfo() (*walletrpc.LightdInfo, error) {
    params := []json.RawMessage{}
    result, rpcErr := CallRpcWithRetries("getinfo", params)
	if rpcErr != nil {
		return nil, rpcErr
	}
	var getinfoReply HushdRpcReplyGetinfo
	err := json.Unmarshal(result, &getinfoReply)
	if err != nil {
		return nil, rpcErr
	}

    params = []json.RawMessage{}
    result, rpcErr = CallRpcWithRetries("getblockchaininfo", params)
	if rpcErr != nil {
		return nil, rpcErr
	}
	var getblockchaininfoReply HushdRpcReplyGetblockchaininfo
	err = json.Unmarshal(result, &getblockchaininfoReply)
	if err != nil {
		return nil, rpcErr
	}
	// If the sapling consensus branch doesn't exist, it must be regtest
	var saplingHeight int
	if saplingJSON, ok := getblockchaininfoReply.Upgrades["76b809bb"]; ok { // Sapling ID
		saplingHeight = saplingJSON.ActivationHeight
	}

	vendor := "Hush lightwalletd"
	return &walletrpc.LightdInfo{
		Version:                 Version,
		Vendor:                  vendor,
		TaddrSupport:            true,
		ChainName:               getblockchaininfoReply.Chain,
		SaplingActivationHeight: uint64(saplingHeight),
		ConsensusBranchId:       getblockchaininfoReply.Consensus.Chaintip,
		BlockHeight:             uint64(getblockchaininfoReply.Blocks),
        //TODO: set these via LDFLAGS
		//GitCommit:               GitCommit,
		//Branch:                  Branch,
		//BuildDate:               BuildDate,
		//BuildUser:               BuildUser,
		//EstimatedHeight:         uint64(getblockchaininfoReply.EstimatedHeight),
		//HushdBuild:              getinfoReply.Build,
		//HushSubversion:          getinfoReply.Subversion,
        // TODO: get notarized key
        // TODO: longestchain
		Notarized:               uint64(0),
	}, nil
}

// FirstRPC tests that we can successfully reach hushd through the RPC
// interface. The specific RPC used here is not important.
func FirstRPC() {
	retryCount := 0
	for {
		_, err := GetBlockChainInfo()
		if err == nil {
			if retryCount > 0 {
				Log.Warn("getblockchaininfo RPC successful")
			}
			break
		}
		retryCount++
		if retryCount > 20 {
			Log.WithFields(logrus.Fields{
				"timeouts": retryCount,
			}).Fatal("unable to issue getblockchaininfo RPC call to hushd node")
		}
		Log.WithFields(logrus.Fields{
			"error": err.Error(),
			"retry": retryCount,
		}).Warn("error with getblockchaininfo rpc, retrying...")
		Time.Sleep(time.Duration(10+retryCount*5) * time.Second) // backoff
	}
}

func CallRpcWithRetries(method string, params []json.RawMessage) (json.RawMessage, error) {
    retryCount := 0
    maxRetries := 50
    for {
        // params := []json.RawMessage{}
        result, err := RawRequest(method, params)
        if err == nil {
            if retryCount > 0 {
                Log.Warn(fmt.Sprintf("%s RPC successful", method))
            }
            return result, err
            break
        }
        retryCount++
        if retryCount > maxRetries && method != "sendrawtransaction" {
            Log.WithFields(logrus.Fields{
                "timeouts": retryCount,
            }).Fatal(fmt.Sprintf("unable to issue %s RPC call to hushd node", method))
        }
        if retryCount > maxRetries && method == "sendrawtransaction" {
            // TODO: it would be better to only give up if the error is an expired tx
            Log.WithFields(logrus.Fields{
                "error": err.Error(),
                "retry": retryCount,
            }).Warn(fmt.Sprintf("giving up on %s rpc", method))
            // TODO: return a better error
            return nil, nil
        }
        Log.WithFields(logrus.Fields{
            "error": err.Error(),
            "retry": retryCount,
        }).Warn(fmt.Sprintf("error with %s rpc, retrying...", method))
        Time.Sleep(time.Duration(10+retryCount*5) * time.Second) // backoff
    }
    return nil, nil
}

func GetBlockChainInfo() (*HushdRpcReplyGetblockchaininfo, error) {
    // we don't use CallRpcWithRetries here because the calling code does it already
	result, rpcErr := RawRequest("getblockchaininfo", []json.RawMessage{})
	if rpcErr != nil {
		return nil, rpcErr
	}
	var getblockchaininfoReply HushdRpcReplyGetblockchaininfo
	err := json.Unmarshal(result, &getblockchaininfoReply)
	if err != nil {
		return nil, err
	}
	return &getblockchaininfoReply, nil
}

func GetCoinsupply() (string, string, int, int, int, int, error) {
    params := []json.RawMessage{}
	result1, rpcErr := CallRpcWithRetries("coinsupply", params)

	var err error
	var errCode int64

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		errParts := strings.SplitN(rpcErr.Error(), ":", 2)
		errCode, err = strconv.ParseInt(errParts[0], 10, 32)
		//Check to see if we are requesting a height the hushd doesn't have yet
		if err == nil && errCode == -8 {
			return "", "", -1, -1, -1, -1, nil
		}
		return "", "", -1, -1, -1, -1, errors.Wrap(rpcErr, "error requesting coinsupply")
	}

	var f interface{}
	err = json.Unmarshal(result1, &f)
	if err != nil {
		return "", "", -1, -1, -1, -1, errors.Wrap(err, "error reading JSON response")
	}

	result := f.(map[string]interface{})["result"].(string)
	coin := f.(map[string]interface{})["coin"].(string)
	height := f.(map[string]interface{})["height"].(float64)
	supply := f.(map[string]interface{})["supply"].(float64)
	zfunds := f.(map[string]interface{})["zfunds"].(float64)
	total := f.(map[string]interface{})["total"].(float64)

	return result, coin, int(height), int(supply), int(zfunds), int(total), nil
}

func getBlockFromRPC(height int) (*walletrpc.CompactBlock, error) {
	// `block.ParseFromSlice` correctly parses blocks containing v5
	// transactions, but incorrectly computes the IDs of the v5 transactions.
	// We temporarily paper over this bug by fetching the correct txids via a
	// verbose getblock RPC call, which returns the txids.
	//
	// Unfortunately, this RPC doesn't return the raw hex for the block,
	// so a second getblock RPC (non-verbose) is needed (below).
	// https://github.com/zcash/lightwalletd/issues/392

	params := make([]json.RawMessage, 2)
	heightJSON, err := json.Marshal(strconv.Itoa(height))
	if err != nil {
		Log.Fatal("getBlockFromRPC bad height argument", height, err)
	}
	params[0] = heightJSON
	// Fetch the block using the verbose option ("1") because it provides
	// both the list of txids and the block hash (block ID), which
	// we need to fetch the raw data format of the same block. Don't fetch
	// by height in case a reorg occurs between the two getblock calls;
	// using block hash ensures that we're fetching the same block.
	params[1] = json.RawMessage("1")
	result, rpcErr := CallRpcWithRetries("getblock", params)
	if rpcErr != nil {
		// Check to see if we are requesting a height the hushd doesn't have yet
		if (strings.Split(rpcErr.Error(), ":"))[0] == "-8" {
			return nil, nil
		}
		return nil, errors.Wrap(rpcErr, "error requesting verbose block")
	}
	var block1 HushdRpcReplyGetblock1
	err = json.Unmarshal(result, &block1)
	if err != nil {
		return nil, err
	}
	blockHash, err := json.Marshal(block1.Hash)
	if err != nil {
		Log.Fatal("getBlockFromRPC bad block hash", block1.Hash)
	}
	params[0] = blockHash
	params[1] = json.RawMessage("0") // non-verbose (raw hex)
	result, rpcErr = CallRpcWithRetries("getblock", params)

	// For some reason, the error responses are not JSON
	if rpcErr != nil {
		return nil, errors.Wrap(rpcErr, "error requesting block")
	}

	var blockDataHex string
	err = json.Unmarshal(result, &blockDataHex)
	if err != nil {
		return nil, errors.Wrap(err, "error reading JSON response")
	}

	blockData, err := hex.DecodeString(blockDataHex)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding getblock output")
	}

	block := parser.NewBlock()
	rest, err := block.ParseFromSlice(blockData)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing block")
	}
	if len(rest) != 0 {
		return nil, errors.New("received overlong message")
	}
	if block.GetHeight() != height {
		return nil, errors.New("received unexpected height block")
	}
	for i, t := range block.Transactions() {
		txid, err := hex.DecodeString(block1.Tx[i])
		if err != nil {
			return nil, errors.Wrap(err, "error decoding getblock txid")
		}
		// convert from big-endian
		t.SetTxID(parser.Reverse(txid))
	}
	return block.ToCompact(), nil
}

var (
	ingestorRunning  bool
	stopIngestorChan = make(chan struct{})
)

func startIngestor(c *BlockCache) {
	if !ingestorRunning {
		ingestorRunning = true
		go BlockIngestor(c, 0)
	}
}
func stopIngestor() {
	if ingestorRunning {
		ingestorRunning = false
		stopIngestorChan <- struct{}{}
	}
}

// BlockIngestor runs as a goroutine and polls hushd for new blocks, adding them
// to the cache. The repetition count, rep, is nonzero only for unit-testing.
func BlockIngestor(c *BlockCache, rep int) {
	lastLog := Time.Now()
	lastHeightLogged := 0

	// Start listening for new blocks
	for i := 0; rep == 0 || i < rep; i++ {
		// stop if requested
		select {
		case <-stopIngestorChan:
			return
		default:
		}

        params := []json.RawMessage{}
		result, err := CallRpcWithRetries("getbestblockhash", params)
		if err != nil {
			Log.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("error hushd getbestblockhash rpc")
		}
		var hashHex string
		err = json.Unmarshal(result, &hashHex)
		if err != nil {
			Log.Fatal("bad getbestblockhash return:", err, result)
		}
		lastBestBlockHash, err := hex.DecodeString(hashHex)
		if err != nil {
			Log.Fatal("error decoding getbestblockhash", err, hashHex)
		}

		height := c.GetNextHeight()
		if string(lastBestBlockHash) == string(parser.Reverse(c.GetLatestHash())) {
			// Synced
			c.Sync()
			if lastHeightLogged != height-1 {
				lastHeightLogged = height - 1
				Log.Info("Waiting for block: ", height)
			}
			Time.Sleep(2 * time.Second)
			lastLog = Time.Now()
			continue
		}
		var block *walletrpc.CompactBlock
		block, err = getBlockFromRPC(height)
		if err != nil {
			Log.Info("getblock ", height, " failed, will retry: ", err)
			Time.Sleep(8 * time.Second)
			continue
		}
		if block != nil && c.HashMatch(block.PrevHash) {
			if err = c.Add(height, block); err != nil {
				Log.Fatal("Cache add failed:", err)
			}
			// Don't log these too often.
			if /* DarksideEnabled || */ Time.Now().Sub(lastLog).Seconds() >= 4 {
				lastLog = Time.Now()
				Log.Info("Adding block to cache ", height, " ", displayHash(block.Hash))
			}
			continue
		}
		if height == c.GetFirstHeight() {
			c.Sync()
			Log.Info("Waiting for hushd height to reach Sapling activation height ",
				"(", c.GetFirstHeight(), ")...")
			Time.Sleep(20 * time.Second)
			return
		}
		Log.Info("REORG: dropping block ", height-1, " ", displayHash(c.GetLatestHash()))
		c.Reorg(height - 1)
	}
}


// GetBlock returns the compact block at the requested height, first by querying
// the cache, then, if not found, will request the block from hushd. It returns
// nil if no block exists at this height.
func GetBlock(cache *BlockCache, height int) (*walletrpc.CompactBlock, error) {
	// First, check the cache to see if we have the block
	block := cache.Get(height)
	if block != nil {
		return block, nil
	}

	// Not in the cache
	block, err := getBlockFromRPC(height)
	if err != nil {
		return nil, err
	}
	if block == nil {
		// Block height is too large
		return nil, errors.New("block requested is newer than latest block")
	}
	return block, nil
}

// GetBlockRange returns a sequence of consecutive blocks in the given range.
func GetBlockRange(cache *BlockCache, blockOut chan<- *walletrpc.CompactBlock, errOut chan<- error, start, end int) {
	// Go over [start, end] inclusive
	low := start
	high := end
	if start > end {
		// reverse the order
		low, high = end, start
	}
	for i := low; i <= high; i++ {
		j := i
		if start > end {
			// reverse the order
			j = high - (i - low)
		}
		block, err := GetBlock(cache, j)
		if err != nil {
			errOut <- err
			return
		}
		blockOut <- block
	}
	errOut <- nil
}


func displayHash(hash []byte) string {
	return hex.EncodeToString(parser.Reverse(hash))
}
