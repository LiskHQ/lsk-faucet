# lsk-faucet

[![Report](https://goreportcard.com/badge/github.com/liskhq/lsk-faucet)](https://goreportcard.com/report/github.com/liskhq/lsk-faucet)
[![Go](https://img.shields.io/github/go-mod/go-version/liskhq/lsk-faucet)](https://go.dev/)
[![License](https://img.shields.io/github/license/LiskHQ/lsk-faucet)](https://github.com/liskhq/lsk-faucet/blob/main/LICENSE)
![GitHub repo size](https://img.shields.io/github/repo-size/liskhq/lsk-faucet)
![GitHub issues](https://img.shields.io/github/issues-raw/liskhq/lsk-faucet)
![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/liskhq/lsk-faucet)

LSK faucet is a web application to get Lisk (LSK) tokens on the Lisk Sepolia Testnet. The tokens can be used to test and troubleshoot your decentralized application or protocol before going live on the Lisk Mainnet.

## Features

* Allow to configure the funding account via private key or keystore
* Asynchronous processing Txs to achieve parallel execution of user requests
* Rate limiting by ETH address and IP address as a precaution against spam
* Prevent X-Forwarded-For spoofing by specifying the count of reverse proxies

## Get started

### Prerequisites

* Go (1.22 or later)
* Node.js

### Installation

1. Clone the repository and navigate to the app’s directory
```bash
git clone https://github.com/LiskHQ/lsk-faucet.git
cd lsk-faucet
```

2. Bundle Frontend web with Vite
```bash
go generate
```

3. Build Go project 
```bash
go build -o lsk-faucet
```

## Usage

**Use private key to fund users**

```bash
./lsk-faucet -httpport 8080 -wallet.provider http://localhost:8545 -wallet.privkey privkey
```

**Use keystore to fund users**

```bash
./lsk-faucet -httpport 8080 -wallet.provider http://localhost:8545 -wallet.keyjson keystore -wallet.keypass password.txt
```

### Configuration
Below is a list of environment variables that can be configured.

- `WEB3_PROVIDER`: Endpoint for Lisk JSON-RPC connection.
- `PRIVATE_KEY`: Private key hex to fund user requests with.
- `KEYSTORE`: Keystore file to fund user requests with.
- `HCAPTCHA_SITEKEY`: hCaptcha sitekey.
- `HCAPTCHA_SECRET`: hCaptcha secret.
- `LSK_TOKEN_ADDRESS`: Contract address of LSK token on the Lisk L2.

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
./lsk-faucet -httpport 8080
```

**Optional Flags**

The following are the available command-line flags(excluding above wallet flags):

| Flag              | Description                                      | Default Value                              |
| ----------------- | ------------------------------------------------ | ------------------------------------------ |
| -httpport         | Listener port to serve HTTP connection           | 8080                                       |
| -proxycount       | Count of reverse proxies in front of the server  | 0                                          |
| -token-address    | Token contract address                           | 0x8a21CF9Ba08Ae709D64Cb25AfAA951183EC9FF6D |
| -faucet.amount    | Number of LSK to transfer per user request       | 1                                          |
| -faucet.minutes   | Number of minutes to wait between funding rounds | 10080 (1 week)                             |
| -faucet.name      | Network name to display on the frontend          | sepolia                                    |
| -faucet.symbol    | Token symbol to display on the frontend          | LSK                                        |
| -hcaptcha.sitekey | hCaptcha sitekey                                 |                                            |
| -hcaptcha.secret  | hCaptcha secret                                  |                                            |

### Docker deployment
#### Build docker image
Run the following command to build docker image:
```bash
docker build -t liskhq/lsk-faucet .
```


#### Run faucet
Run the following command to start the application:

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=<rpc-endpoint> -e PRIVATE_KEY=<hex-private-key> liskhq/lsk-faucet

```
**NOTE**: Please replace `<rpc-endpoint>` and `<hex-private-key>` with appropriate values.

or

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=<rpc-endpoint> -e KEYSTORE=<keystore-path> -v `pwd`/keystore:/app/keystore -v `pwd`/password.txt:/app/password.txt liskhq/lsk-faucet
```

**NOTE**: Please replace `<rpc-endpoint>` and `<keystore-path>` with appropriate values.

## License

Distributed under the MIT License. See LICENSE for more information.
