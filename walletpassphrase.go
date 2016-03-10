// Copyright (c) 2016 Josh Rickmar <jrick@devio.us>
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/btcsuite/btcrpcclient"
	"github.com/btcsuite/btcutil"
)

var (
	defaultAppData  = btcutil.AppDataDir("btcwallet", false)
	defaultCertFile = filepath.Join(defaultAppData, "rpc.cert")
)

const (
	defaultNetworkAddress = "localhost:8332"
	defaultSeconds        = 60
)

var (
	host     = flag.String("c", defaultNetworkAddress, "network address (host:port) of wallet RPC server")
	rpcUser  = flag.String("u", "", "RPC username")
	certFile = flag.String("cert", defaultCertFile, "certificate file for RPC TLS")
	seconds  = flag.Int64("s", defaultSeconds, "seconds to keep wallet unlocked")
)

func main() {
	flag.Parse()

	err := unlock()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func unlock() error {
	if *rpcUser == "" {
		return errors.New("No RPC username (use -u to set)")
	}
	if *seconds < 0 {
		return errors.New("Negative seconds option (use -s to set)")
	}
	if *seconds > 60*60 {
		return errors.New("Insane seconds value exceeds 1hr (use -s to set)")
	}
	certExists, err := fileExists(*certFile)
	if err != nil {
		return err
	}
	if !certExists {
		return fmt.Errorf("TLS certificate file `%s` not found (use -cert to set)", *certFile)
	}

	rpcPass, err := promptSecret("RPC password")
	if err != nil {
		return err
	}

	certs, err := ioutil.ReadFile(*certFile)
	if err != nil {
		return err
	}
	client, err := btcrpcclient.New(&btcrpcclient.ConnConfig{
		Host:         *host,
		User:         *rpcUser,
		Pass:         rpcPass,
		Certificates: certs,
		HTTPPostMode: true,
	}, nil)
	if err != nil {
		return err
	}

	pass, err := promptSecret("Wallet passphrase")
	if err != nil {
		return err
	}

	err = client.WalletPassphrase(pass, *seconds)
	if err != nil {
		return err
	}

	fmt.Printf("Wallet unlocked for %d seconds.\n", *seconds)
	return nil
}

func fileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func promptSecret(what string) (string, error) {
	fmt.Printf("%s: ", what)
	fd := int(os.Stdin.Fd())
	input, err := terminal.ReadPassword(fd)
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(input), nil
}
