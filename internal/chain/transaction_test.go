package chain

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

func TestTxBuilder_TransferETH(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA("976f9f7772781ff6d1c93941129d417c49a209c674056a3cf5e27e225ee55fa8")
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	simBackend := simulated.NewBackend(
		types.GenesisAlloc{
			fromAddress: {Balance: big.NewInt(10000000000000000)},
		})
	defer simBackend.Close()
	var s *backends.SimulatedBackend
	patches := gomonkey.ApplyMethod(reflect.TypeOf(s), "SuggestGasPrice", func(_ *backends.SimulatedBackend, _ context.Context) (*big.Int, error) {
		return big.NewInt(875000000), nil
	})
	defer patches.Reset()

	txBuilder := &TxBuild{
		client:       simBackend.Client(),
		privateKey:   privateKey,
		signer:       types.NewEIP155Signer(big.NewInt(1337)),
		fromAddress:  crypto.PubkeyToAddress(privateKey.PublicKey),
		chainID:      big.NewInt(1337),
		tokenAddress: "0xbb5801a7D398351b8bE11C439e05C5B3259aeux0",
	}
	bgCtx := context.Background()
	toAddress := common.HexToAddress("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B")
	value := big.NewInt(1000)
	txHash, err := txBuilder.TransferETH(bgCtx, toAddress.Hex(), value)
	if err != nil {
		t.Errorf("could not add tx to pending block: %v", err)
	}
	simBackend.Commit()

	block, err := simBackend.Client().BlockByNumber(bgCtx, big.NewInt(1))
	if err != nil {
		t.Errorf("could not get block at height 1: %v", err)
	}
	if txHash != block.Transactions()[0].Hash() {
		t.Errorf("did not commit sent transaction. expected hash %v got hash %v", block.Transactions()[0].Hash(), txHash)
	}

	bal, err := simBackend.Client().BalanceAt(bgCtx, toAddress, nil)
	if err != nil {
		t.Error(err)
	}
	if bal.Cmp(value) != 0 {
		t.Errorf("expected balance for to address not received. expected: %v actual: %v", value, bal)
	}
}

func TestTxBuilder_TransferERC20(t *testing.T) {
	privateKey, _ := crypto.HexToECDSA("976f9f7772781ff6d1c93941129d417c49a209c674056a3cf5e27e225ee55fa8")
	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	simBackend := simulated.NewBackend(
		types.GenesisAlloc{
			fromAddress: {Balance: big.NewInt(10000000000000000)},
		})

	defer simBackend.Close()
	var s *backends.SimulatedBackend
	patches := gomonkey.ApplyMethod(reflect.TypeOf(s), "SuggestGasPrice", func(_ *backends.SimulatedBackend, _ context.Context) (*big.Int, error) {
		return big.NewInt(875000000), nil
	})
	defer patches.Reset()

	txBuilder := &TxBuild{
		client:       simBackend.Client(),
		privateKey:   privateKey,
		signer:       types.NewEIP155Signer(big.NewInt(1337)),
		fromAddress:  crypto.PubkeyToAddress(privateKey.PublicKey),
		chainID:      big.NewInt(1337),
		tokenAddress: "0xbb5801a7D398351b8bE11C439e05C5B3259aeux0",
	}
	bgCtx := context.Background()
	toAddress := common.HexToAddress("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B")
	value := big.NewInt(1000)
	txHash, err := txBuilder.TransferERC20(bgCtx, toAddress.Hex(), value)
	if err != nil {
		t.Errorf("could not add tx to pending block: %v", err)
	}
	simBackend.Commit()

	block, err := simBackend.Client().BlockByNumber(bgCtx, big.NewInt(1))
	if err != nil {
		t.Errorf("could not get block at height 1: %v", err)
	}
	if txHash != block.Transactions()[0].Hash() {
		t.Errorf("did not commit sent transaction. expected hash %v got hash %v", block.Transactions()[0].Hash(), txHash)
	}

	// TODO: Verify ERC20 token balance of recipient if possible in future with the simulator
}
