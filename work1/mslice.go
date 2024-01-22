package deom_slice

import (
	"errors"
	"fmt"
	"math/rand"
)

var OutRange error = errors.New("数组下坐标越界")

func DeleteByInterception[T any](slices []T, idx int) ([]T, error) {
	if len(slices)+1 <= idx || idx < 0 {
		return nil, fmt.Errorf("error:%w, 下标超出范围, 长度%d, 下标%d", OutRange, len(slices), idx)
	}
	slices = append(slices[:idx], slices[idx+1:]...)
	return slices, nil
}

// 缩容
func Shink[T any](slices []T) []T {
	cur_cap, flag := SliceCapAnalysis(slices)
	if !flag {
		return slices
	}
	cur_slices := make([]T, 0, cur_cap)
	cur_slices = append(cur_slices, slices...)
	return cur_slices
}

func SliceCapAnalysis[T any](slices []T) (int, bool) {
	caps := cap(slices)
	length := len(slices)

	// 参考答案设定阈值进行缩容
	// 使用的是1.18以后的go版本
	ratios := float64(caps) / float64(length)
	if caps <= 512 && ratios >= 2 {
		return int(float64(caps) * 0.5), true
	} else if caps > 512 && float64(ratios) > 1.25 {
		return int(float64(caps) * 0.8), true
	}
	return caps, false
}

func GetSliceInt(n int) []int {
	a := make([]int, 0, n)
	for i := 0; i < n; i++ {
		a = append(a, rand.Int())
	}
	return a
}

func GetSliceFloat(n int) []float64 {
	a := make([]float64, 0, n)
	for i := 0; i < n; i++ {
		a = append(a, rand.Float64())
	}
	return a
}
