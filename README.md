# lsk-faucet

[![Report](https://goreportcard.com/badge/github.com/liskhq/lsk-faucet)](https://goreportcard.com/report/github.com/liskhq/lsk-faucet)
[![Go](https://img.shields.io/github/go-mod/go-version/liskhq/lsk-faucet)](https://go.dev/)
[![License](https://img.shields.io/github/license/LiskHQ/lsk-faucet)](https://github.com/liskhq/lsk-faucet/blob/main/LICENSE)
![GitHub repo size](https://img.shields.io/github/repo-size/liskhq/lsk-faucet)
![GitHub issues](https://img.shields.io/github/issues-raw/liskhq/lsk-faucet)
![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/liskhq/lsk-faucet)

LSK faucet is a web application that can be configured and deployed to get ETH and custom ERC20 tokens on any test network. The tokens can be used to test and troubleshoot your decentralized application or protocol before going live on the Mainnet.

## Features

* Allow to configure the funding account via private key or keystore
* Asynchronous processing Txs to achieve parallel execution of user requests
* Rate limiting by ETH address and IP address as a precaution against spam
* Prevent X-Forwarded-For spoofing by specifying the count of reverse proxies

## Get started

### Prerequisites

* Go (1.22 or later)
* Node.js
* Solidity compiler (solc)
* Abigen

### Installation

1. Clone the repository and navigate to the app’s directory
```bash
git clone https://github.com/LiskHQ/lsk-faucet.git
cd lsk-faucet
```

2. Bundle Frontend web with Vite

**NOTE**: Please make sure to update the token icon under `web/public/` with the specific ERC20 token icon. The file must be named `token.png`. We recommend the image dimensions to be 128px x 128px.

```bash
make build-frontend
```
3. Generate Go binding for ERC20 token smart contract
```bash
make generate-binding ERC20_CONTRACT_FILE_PATH=<erc20-contract-file-path>
```

**NOTE**: Please make sure to generate the Go binding corresponding to your contract before building the project.

4. Build Go project 
```bash
make build-backend
```

## Usage

**Use private key to fund users**

```bash
make run FLAGS="-httpport 8080 -wallet.provider http://localhost:8545 -wallet.privkey privkey"
```

**Use keystore to fund users**

```bash
make run FLAGS="-httpport 8080 -wallet.provider http://localhost:8545 -wallet.keyjson keystore -wallet.keypass password.txt"
```

### Configuration
Below is a list of environment variables that can be configured.

- `WEB3_PROVIDER`: RPC Endpoint to connect with the network node.
- `PRIVATE_KEY`: Private key hex to fund user requests with.
- `KEYSTORE`: Keystore file to fund user requests with.
- `HCAPTCHA_SITEKEY`: hCaptcha sitekey.
- `HCAPTCHA_SECRET`: hCaptcha secret.
- `ERC20_TOKEN_ADDRESS`: Contract address of the ERC20 token on the configured network, defaults to contract address for Lisk ERC20 tokens on Lisk Sepolia.

You can configure the funder by setting any of the following environment variable instead of command-line flags:
```bash
export PRIVATE_KEY=<hex-private-key>
```

or

```bash
export KEYSTORE=<keystore-path>
echo "your keystore password" > `pwd`/password.txt
```

Then run the faucet application without the wallet command-line flags:
```bash
make run FLAGS="-httpport 8080"
```

**Optional Flags**

The following are the available command-line flags(excluding above wallet flags):

| Flag              | Description                                         | Default Value                              |
| ----------------- | --------------------------------------------------- | ------------------------------------------ |
| -httpport         | Listener port to serve HTTP connection              | 8080                                       |
| -proxycount       | Count of reverse proxies in front of the server     | 0                                          |
| -token.address    | Token contract address                              | 0x8a21CF9Ba08Ae709D64Cb25AfAA951183EC9FF6D |
| -faucet.amount    | Number of ERC20 tokens to transfer per user request | 1                                          |
| -faucet.minutes   | Number of minutes to wait between funding rounds    | 10080 (1 week)                             |
| -faucet.name      | Network name to display on the frontend             | sepolia                                    |
| -faucet.symbol    | Token symbol to display on the frontend             | LSK                                        |
| -explorer.url     | Block Explorer URL                                  | https://sepolia-blockscout.lisk.com        |
| -explorer.tx.path | Block explorer transaction path                     | tx                                         |
| -hcaptcha.sitekey | hCaptcha sitekey                                    |                                            |
| -hcaptcha.secret  | hCaptcha secret                                     |                                            |

### Docker deployment
#### Build docker image
Run the following command to build docker image:
```bash
make build-image
```

#### Run faucet
Set the appropriate configuration in `.env` file and run the following command to start the application:

```bash
make docker-start

```

## License

Distributed under the MIT License. See LICENSE for more information.
