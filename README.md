# lsk-faucet

[![Report](https://goreportcard.com/badge/github.com/liskhq/lsk-faucet)](https://goreportcard.com/report/github.com/liskhq/lsk-faucet)
[![Go](https://img.shields.io/github/go-mod/go-version/liskhq/lsk-faucet)](https://go.dev/)
[![License](https://img.shields.io/github/license/LiskHQ/lsk-faucet)](https://github.com/liskhq/lsk-faucet/blob/main/LICENSE)
![GitHub repo size](https://img.shields.io/github/repo-size/liskhq/lsk-faucet)
![GitHub issues](https://img.shields.io/github/issues-raw/liskhq/lsk-faucet)
![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/liskhq/lsk-faucet)

LSK faucet is a web application to get sepolia Lisk (LSK) in order to test and troubleshoot your decentralized application or protocol before going live on Lisk mainnet.
This faucet is designed to provide LSK test tokens to the developers who need to test their smart contracts and interact with the blockchain.

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

You can configure the funder by using environment variables instead of command-line flags as follows:
```bash
export WEB3_PROVIDER=rpc endpoint
export PRIVATE_KEY=hex private key
```

or

```bash
export WEB3_PROVIDER=rpc endpoint
export KEYSTORE=keystore path
echo "your keystore password" > `pwd`/password.txt
```

Then run the faucet application without the wallet command-line flags:
```bash
./lsk-faucet -httpport 8080
```

**Optional Flags**

The following are the available command-line flags(excluding above wallet flags):

| Flag              | Description                                      | Default Value |
| ----------------- | ------------------------------------------------ | ------------- |
| -httpport         | Listener port to serve HTTP connection           | 8080          |
| -proxycount       | Count of reverse proxies in front of the server  | 0             |
| -token-address    | Token contract address                           |               |
| -faucet.amount    | Number of LSK to transfer per user request       | 1             |
| -faucet.minutes   | Number of minutes to wait between funding rounds | 1440          |
| -faucet.name      | Network name to display on the frontend          | sepolia       |
| -faucet.symbol    | Token symbol to display on the frontend          | LSK           |
| -hcaptcha.sitekey | hCaptcha sitekey                                 |               |
| -hcaptcha.secret  | hCaptcha secret                                  |               |

### Docker deployment
#### Build docker image
Run the following command to build docker image:
```bash
docker build -t liskhq/lsk-faucet .
```


#### Run faucet
Run the following command to start the application:

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc endpoint -e PRIVATE_KEY=hex private key liskhq/lsk-faucet
```

or

```bash
docker run -d -p 8080:8080 -e WEB3_PROVIDER=rpc endpoint -e KEYSTORE=keystore path -v `pwd`/keystore:/app/keystore -v `pwd`/password.txt:/app/password.txt liskhq/lsk-faucet
```

## License

Distributed under the MIT License. See LICENSE for more information.
