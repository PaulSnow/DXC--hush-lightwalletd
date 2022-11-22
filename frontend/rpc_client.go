// Copyright (c) 2019-2022 Duke Leto and The Hush developers
// Copyright (c) 2019-2020 The Zcash developers
// Distributed under the GPLv3 software license
package frontend

import (
	"net"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/pkg/errors"
	"git.hush.is/hush/lightwalletd/common"
	ini "gopkg.in/ini.v1"
)

func NewZRPCFromConf(confPath string) (*rpcclient.Client, error) {
	cfg, err := ini.Load(confPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	rpcaddr := cfg.Section("").Key("rpcbind").String()
	rpcport := cfg.Section("").Key("rpcport").String()
	username := cfg.Section("").Key("rpcuser").String()
	password := cfg.Section("").Key("rpcpassword").String()

	return NewZRPCFromCreds(net.JoinHostPort(rpcaddr, rpcport), username, password)
}

// NewZRPCFromFlags gets zcashd rpc connection information from provided flags.
func NewZRPCFromFlags(opts *common.Options) (*rpcclient.Client, error) {
	// Connect to local Zcash RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         net.JoinHostPort(opts.RPCHost, opts.RPCPort),
		User:         opts.RPCUser,
		Pass:         opts.RPCPassword,
		HTTPPostMode: true, // Zcash only supports HTTP POST mode
		DisableTLS:   true, // Zcash does not provide TLS by default
	}
	return rpcclient.New(connCfg, nil)
}

func NewZRPCFromCreds(addr, username, password string) (*rpcclient.Client, error) {
	// Connect to local hush RPC server using HTTP POST mode.
	connCfg := &rpcclient.ConnConfig{
		Host:         addr,
		User:         username,
		Pass:         password,
		HTTPPostMode: true, // Hush only supports HTTP POST mode
		DisableTLS:   true, // Hush does not provide TLS by default
	}
	// Notice the notification parameter is nil since notifications are
	// not supported in HTTP POST mode.
	return rpcclient.New(connCfg, nil)
}
