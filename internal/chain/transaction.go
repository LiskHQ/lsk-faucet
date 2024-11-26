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
	supportsEIP1559  bool
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
	supportsEIP1559, err := checkEIP1559Support(client)
	if err != nil {
		return nil, err
	}

	txBuilder := &TxBuild{
		client:           client,
		privateKey:       privateKey,
		signer:           types.NewLondonSigner(chainID),
		fromAddress:      crypto.PubkeyToAddress(privateKey.PublicKey),
		supportsEIP1559:  supportsEIP1559,
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
	toAddress := common.HexToAddress(to)
	nonce := b.getAndIncrementNonce()

	var err error
	var unsignedTx *types.Transaction

	if b.supportsEIP1559 {
		unsignedTx, err = b.buildEIP1559Tx(ctx, &toAddress, value, gasLimit, nonce)
	} else {
		unsignedTx, err = b.buildLegacyTx(ctx, &toAddress, value, gasLimit, nonce)
	}

	if err != nil {
		return common.Hash{}, err
	}

	signedTx, err := types.SignTx(unsignedTx, b.signer, b.privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err = b.client.SendTransaction(ctx, signedTx); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "nonce") {
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
func (b *TxBuild) buildEIP1559Tx(ctx context.Context, to *common.Address, value *big.Int, gasLimit uint64, nonce uint64) (*types.Transaction, error) {
	header, err := b.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	gasTipCap, err := b.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}

	// gasFeeCap = baseFee * 2 + gasTipCap
	gasFeeCap := new(big.Int).Mul(header.BaseFee, big.NewInt(2))
	gasFeeCap = new(big.Int).Add(gasFeeCap, gasTipCap)

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   b.signer.ChainID(),
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        to,
		Value:     value,
	}), nil
}

func (b *TxBuild) buildLegacyTx(ctx context.Context, to *common.Address, value *big.Int, gasLimit uint64, nonce uint64) (*types.Transaction, error) {
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       to,
		Value:    value,
	}), nil
}

func (b *TxBuild) getAndIncrementNonce() uint64 {
	return atomic.AddUint64(&b.nonce, 1) - 1
}

func (b *TxBuild) refreshNonce(ctx context.Context) {
	nonce, err := b.client.PendingNonceAt(ctx, b.Sender())
	if err != nil {
		log.WithFields(log.Fields{
			"address": b.Sender(),
			"error":   err,
		}).Error("failed to refresh account nonce")
		return
	}

	atomic.StoreUint64(&b.nonce, nonce)
}

func checkEIP1559Support(client *ethclient.Client) (bool, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return false, err
	}

	return header.BaseFee != nil && header.BaseFee.Cmp(big.NewInt(0)) > 0, nil
}
