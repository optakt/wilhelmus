package b

import (
	"encoding/hex"
	"math/big"
)

func FromHex(s string) *big.Int {
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	b := big.NewInt(0).SetBytes(bytes)
	return b
}
