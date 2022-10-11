package b

import (
	"encoding/hex"
	"math/big"
)

func FromHex(v interface{}) *big.Int {
	s, ok := v.(string)
	if !ok {
		return big.NewInt(0)
	}
	bytes, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	b := big.NewInt(0).SetBytes(bytes)
	return b
}
