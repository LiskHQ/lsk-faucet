package chain

import (
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func Has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func IsValidAddress(address string, checksummed bool) bool {
	if !common.IsHexAddress(address) {
		return false
	}
	return !checksummed || common.HexToAddress(address).Hex() == address
}

func addLeftPadding(input []byte) []byte {
	return common.LeftPadBytes(input, 32)
}

func TokenToWei(amount float64, decimals int) *big.Int {
	ethDecimals := 18
	oneEthToWei := math.Pow10(ethDecimals)

	amountInt := int64(amount * oneEthToWei)

	oneTokenInWei := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	return new(big.Int).Div(new(big.Int).Mul(big.NewInt(amountInt), oneTokenInWei), big.NewInt(int64(oneEthToWei)))
}
