package Random

import (
	"crypto/rand"
	"math/big"
)

func GetRandomInt(maxValue int) (int, error) {
	nBig, err := rand.Int(rand.Reader, big.NewInt(int64(maxValue)))
	if err != nil {
		return 0, err
	}
	result := int(nBig.Int64())
	return result, nil
}
