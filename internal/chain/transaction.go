package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"sync/atomic"

	"github.com/LiskHQ/lsk-faucet/internal/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/sha3"
)

var (
	gasLimitInitializedAccount    uint64 = 35000
	gasLimitNonInitializedAccount uint64 = 52000
)

type TxBuilder interface {
	Sender() common.Address
	GetContractInstance() *bindings.Token
	TransferETH(ctx context.Context, to string, value *big.Int) (common.Hash, error)
	TransferERC20(ctx context.Context, to string, value *big.Int, balance *big.Int) (common.Hash, error)
}

type TxBuild struct {
	client           bind.ContractBackend
	privateKey       *ecdsa.PrivateKey
	signer           types.Signer
	fromAddress      common.Address
	nonce            uint64
	chainID          *big.Int
	tokenAddress     string
	contractInstance *bindings.Token
}

func NewTxBuilder(provider string, privateKey *ecdsa.PrivateKey, tokenAddress string, chainID *big.Int) (TxBuilder, error) {
	client, err := ethclient.Dial(provider)
	if err != nil {
		return nil, err
	}

	if chainID == nil {
		chainID, err = client.ChainID(context.Background())
		if err != nil {
			return nil, err
		}
	}

	contractInstance, err := bindings.NewToken(common.HexToAddress(tokenAddress), client)
	if err != nil {
		return nil, err
	}

	txBuilder := &TxBuild{
		client:           client,
		privateKey:       privateKey,
		signer:           types.NewEIP155Signer(chainID),
		fromAddress:      crypto.PubkeyToAddress(privateKey.PublicKey),
		chainID:          chainID,
		tokenAddress:     tokenAddress,
		contractInstance: contractInstance,
	}
	txBuilder.refreshNonce(context.Background())

	return txBuilder, nil
}

func (b *TxBuild) Sender() common.Address {
	return b.fromAddress
}

func (b *TxBuild) GetContractInstance() *bindings.Token {
	return b.contractInstance
}

func (b *TxBuild) TransferETH(ctx context.Context, to string, value *big.Int) (common.Hash, error) {
	gasLimit := uint64(21000)
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return common.Hash{}, err
	}

	toAddress := common.HexToAddress(to)
	unsignedTx := types.NewTx(&types.LegacyTx{
		Nonce:    b.getAndIncrementNonce(),
		To:       &toAddress,
		Value:    value,
		Gas:      gasLimit,
		GasPrice: gasPrice,
	})

	signedTx, err := types.SignTx(unsignedTx, b.signer, b.privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err = b.client.SendTransaction(ctx, signedTx); err != nil {
		log.Error("failed to send tx", "tx hash", signedTx.Hash().String(), "err", err)
		if strings.Contains(err.Error(), "nonce") {
			b.refreshNonce(context.Background())
		}
		return common.Hash{}, err
	}

	return signedTx.Hash(), nil
}

func (b *TxBuild) TransferERC20(ctx context.Context, to string, value *big.Int, balance *big.Int) (common.Hash, error) {
	emptyHash := common.Hash{}
	publicKey := b.privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return emptyHash, fmt.Errorf("invalid type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := b.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return emptyHash, err
	}

	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return emptyHash, err
	}

	toAddress := common.HexToAddress(to)
	tokenAddress := common.HexToAddress(b.tokenAddress)

	transferFnSignature := []byte("transfer(address,uint256)")
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	methodID := hash.Sum(nil)[:4]

	paddedAddress := addLeftPadding(toAddress.Bytes())
	paddedAmount := addLeftPadding(value.Bytes())

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	var gasLimit uint64
	if balance.BitLen() == 0 {
		gasLimit = gasLimitNonInitializedAccount
	} else {
		gasLimit = gasLimitInitializedAccount
	}

	tx := types.NewTransaction(nonce, tokenAddress, big.NewInt(0), gasLimit, gasPrice, data)

	signedTx, err := types.SignTx(tx, b.signer, b.privateKey)
	if err != nil {
		return emptyHash, err
	}

	if err = b.client.SendTransaction(ctx, signedTx); err != nil {
		log.Error("failed to send tx", "tx hash", signedTx.Hash().String(), "err", err)
		if strings.Contains(err.Error(), "nonce") {
			b.refreshNonce(ctx)
		}
		return emptyHash, err
	}

	return signedTx.Hash(), nil
}

func (b *TxBuild) getAndIncrementNonce() uint64 {
	return atomic.AddUint64(&b.nonce, 1) - 1
}

func (b *TxBuild) refreshNonce(ctx context.Context) {
	nonce, err := b.client.PendingNonceAt(ctx, b.Sender())
	if err != nil {
		log.Error("failed to refresh nonce", "address", b.Sender(), "err", err)
		return
	}

	b.nonce = nonce
}
