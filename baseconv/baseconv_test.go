package baseconv_test

import (
	"golang.ssttevee.com/utils/baseconv"
	"testing"
)

const chars36 = "0123456789abcdefghijklmnopqrstuvwxyz"

func TestNewBaseMap(t *testing.T) {
	base, err := baseconv.NewBaseMap("wagdd")
	if err != baseconv.ErrDuplicateCharacter {
		t.Fail()
	}

	if base != nil {
		t.Fail()
	}
}

func TestBaseMap_Parse(t *testing.T) {
	num, err := baseconv.Base16.Parse("36c9")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(num)

	str := num.Format(baseconv.Base10)
	t.Log(str)
}

func TestConvert(t *testing.T) {
	out, err := baseconv.Convert("6353835517599558185862", chars36, chars36[:10])
	if err != nil {
		t.Fatal(err)
	}

	t.Log(out)
}

func BenchmarkFormat(b *testing.B) {
	num, _ := baseconv.Base36.Parse("6353835517599558185862")
    for i := 0; i < b.N; i++ {
		num.Format(baseconv.Base10)
    }
}