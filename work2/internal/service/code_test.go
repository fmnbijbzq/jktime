package service

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestGenerate(t *testing.T) {
	t.Log(fmt.Sprintf("%06d", rand.Intn(1000000)))

}
