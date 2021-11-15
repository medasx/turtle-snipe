package pancake_swap

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

func EthToWei(val *big.Float) *big.Int {
	wei := new(big.Float).Mul(big.NewFloat(params.Ether), val)
	result := new(big.Int)
	wei.Int(result)
	return result
}

func WeiToEth(val *big.Int) *big.Float {
	eth := new(big.Float).Quo(new(big.Float).SetInt(val), big.NewFloat(params.Ether))
	return eth
}
