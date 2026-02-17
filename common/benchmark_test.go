package common

import (
	"testing"
)

func BenchmarkBytesToUnit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesToUnit(1073741824) // 1 GB
	}
}

func BenchmarkConvertBytesToUnit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConvertBytesToUnit(1073741824, "MB")
	}
}

func BenchmarkConvertToReadableUnit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ConvertToReadableUnit(uint64(1073741824))
	}
}

func BenchmarkRoundFloat64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RoundFloat64(3.14159265, 2)
	}
}
