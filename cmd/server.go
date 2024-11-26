package cmd

import (
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/LiskHQ/lsk-faucet/internal/chain"
	"github.com/LiskHQ/lsk-faucet/internal/server"
)

var (
	appVersion = "v1.2.0"
	chainIDMap = map[string]int{"sepolia": 11155111, "holesky": 17000, "lisk_sepolia": 4202}

	tokenAddress      = flag.String("token.address", os.Getenv("ERC20_TOKEN_ADDRESS"), "Contract address of ERC20 token")
	tokenDecimalsFlag = flag.Int("token.decimals", 18, "Token decimals")

	httpPortFlag = flag.Int("httpport", 8080, "Listener port to serve HTTP connection")
	proxyCntFlag = flag.Int("proxycount", 0, "Count of reverse proxies in front of the server")
	versionFlag  = flag.Bool("version", false, "Print version number")

	payoutFlag   = flag.Float64("faucet.amount", 0.1, "Number of ERC20 tokens to transfer per user request")
	intervalFlag = flag.Int("faucet.minutes", 10080, "Number of minutes to wait between funding rounds")
	netnameFlag  = flag.String("faucet.name", "lisk_sepolia", "Network name to display on the frontend")
	symbolFlag   = flag.String("faucet.symbol", "LSK", "Token symbol to display on the frontend")

	explorerURL    = flag.String("explorer.url", "https://sepolia-blockscout.lisk.com", "Block explorer URL")
	explorerTxPath = flag.String("explorer.tx.path", "tx", "Block explorer transaction path fragment")

	keyJSONFlag  = flag.String("wallet.keyjson", os.Getenv("KEYSTORE"), "Keystore file to fund user requests with")
	keyPassFlag  = flag.String("wallet.keypass", "password.txt", "Passphrase text file to decrypt keystore")
	privKeyFlag  = flag.String("wallet.privkey", os.Getenv("PRIVATE_KEY"), "Private key hex to fund user requests with")
	providerFlag = flag.String("wallet.provider", os.Getenv("WEB3_PROVIDER"), "Endpoint for Lisk JSON-RPC connection")

	hcaptchaSiteKeyFlag = flag.String("hcaptcha.sitekey", os.Getenv("HCAPTCHA_SITEKEY"), "hCaptcha sitekey")
	hcaptchaSecretFlag  = flag.String("hcaptcha.secret", os.Getenv("HCAPTCHA_SECRET"), "hCaptcha secret")
)

func init() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}
}

func Execute() {
	privateKey, err := getPrivateKeyFromFlags()
	if err != nil {
		panic(fmt.Errorf("failed to read private key: %w", err))
	}
	var chainID *big.Int
	if value, ok := chainIDMap[strings.ToLower(*netnameFlag)]; ok {
		chainID = big.NewInt(int64(value))
	}

	txBuilder, err := chain.NewTxBuilder(*providerFlag, privateKey, *tokenAddress, chainID)
	if err != nil {
		panic(fmt.Errorf("cannot connect to web3 provider: %w", err))
	}
	config := server.NewConfig(*netnameFlag, *symbolFlag, *payoutFlag, *tokenDecimalsFlag, *httpPortFlag, *intervalFlag, *proxyCntFlag, *hcaptchaSiteKeyFlag, *hcaptchaSecretFlag, *explorerURL, *explorerTxPath)
	go server.NewServer(txBuilder, config).Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func getPrivateKeyFromFlags() (*ecdsa.PrivateKey, error) {
	if *privKeyFlag != "" {
		hexkey := *privKeyFlag
		if chain.Has0xPrefix(hexkey) {
			hexkey = hexkey[2:]
		}
		return crypto.HexToECDSA(hexkey)
	} else if *keyJSONFlag == "" {
		return nil, errors.New("missing private key or keystore")
	}

	keyfile, err := chain.ResolveKeyfilePath(*keyJSONFlag)
	if err != nil {
		return nil, err
	}
	password, err := os.ReadFile(*keyPassFlag)
	if err != nil {
		return nil, err
	}

	return chain.DecryptKeyfile(keyfile, strings.TrimRight(string(password), "\r\n"))
}
