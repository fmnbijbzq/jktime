package deom_slice_test

import (
	"deom_slice"
	"testing"
)

func BenchmarkDeleteS2(b *testing.B) {
	count := 100
	for i := 0; i < b.N; i++ {
		_, err := deom_slice.DeleteByInterception[int](deom_slice.GetSliceInt(count), 0)
		if err != nil {
		}

	}
}
