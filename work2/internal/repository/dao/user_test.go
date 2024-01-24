package dao_test

import (
	"fmt"
	"testing"
)

type Deom struct {
}

func (d *Deom) Hello() {
	fmt.Println("---------")

}

type DeomPrint interface {
	Hello()
}

func TestUser(t *testing.T) {
	var u DeomPrint

	if me, ok := u.(*Deom); ok {
		fmt.Println("success")
		me.Hello()
	}

}
