package sample

import (
	"fmt"
	"testing"
)

var documents = map[int]string{}

// Bench runs bench on documents of 10 and 100 MB. -short skips the large
// one.
func Bench(b *testing.B, bench func(b *testing.B, doc string)) {
	for _, mb := range []int{10, 100} {
		b.Run(fmt.Sprintf("%dMB", mb), func(b *testing.B) {
			if mb > 10 && testing.Short() {
				b.Skip("large document")
			}
			if _, ok := documents[mb]; !ok {
				documents[mb] = XML(mb << 20)
			}
			bench(b, documents[mb])
		})
	}
}
