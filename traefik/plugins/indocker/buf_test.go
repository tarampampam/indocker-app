package indocker_test

import (
	"reflect"
	"testing"

	"indocker"
)

func TestNewSizeLimitedBuf_Add(t *testing.T) {
	t.Parallel()

	var b = indocker.NewSizeLimitedBuf(3)

	if d := b.Get(); len(d) != 0 {
		t.Errorf("expected empty data, got: %v", len(d))
	}

	b.Add(1)
	b.Add(2)

	if d := b.Get(); len(d) != 2 {
		t.Errorf("expected 2 items, got: %v", len(d))
	}

	if v := b.Get()[0]; v != 1 {
		t.Error("expected 1")
	}

	if v := b.Get()[1]; v != 2 {
		t.Error("expected 2")
	}

	b.Add(3)
	b.Add(4)

	if d := b.Get(); len(d) != 3 {
		t.Errorf("expected 3 items, got: %v", len(d))
	}

	if v := b.Get()[0]; v != 2 {
		t.Errorf("want: 2, got: %v", v)
	}

	if v := b.Get()[1]; v != 3 {
		t.Errorf("want: 3, got: %v", v)
	}

	if v := b.Get()[2]; v != 4 {
		t.Errorf("want: 4, got: %v", v)
	}

	for i := 1; i < 100; i++ {
		b.Add(i + 4)
	}

	if d := b.Get(); len(d) != 3 {
		t.Errorf("expected 3 items, got: %v", len(d))
	}

	if v := b.Get()[0]; v != 101 {
		t.Errorf("want: 101, got: %v", v)
	}

	if v := b.Get()[2]; v != 103 {
		t.Errorf("want: 103, got: %v", v)
	}

	var d1, d2 = b.Get(), b.Get()

	if !reflect.DeepEqual(d1, d2) {
		t.Error("expected equal")
	}
}
