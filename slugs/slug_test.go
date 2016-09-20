package slugs_test

import (
	"testing"
	"github.com/ssttevee/go-utils/slugs"
)

func TestMake(t *testing.T) {
	slug := slugs.Make("*Steve	+ __\n_--LAm swn lEfegs  TheEifle77Tower66-")

	if slug != "steve-l-am-swn-l-efegs-the-eifle-77-tower-66" {
		t.Fatalf("got unexpected result: %s", slug)
	}
}