package stats

import (
	"testing"
)

func TestCompare_basic(t *testing.T) {
	a := Totals{Prompts: 100}
	b := Totals{Prompts: 150}
	cmp := Compare(a, b)
	d := cmp.Prompts
	if d.A != 100 {
		t.Errorf("A: got %v, want 100", d.A)
	}
	if d.B != 150 {
		t.Errorf("B: got %v, want 150", d.B)
	}
	if d.Delta != 50 {
		t.Errorf("Delta: got %v, want 50", d.Delta)
	}
	if d.Pct != 50.0 {
		t.Errorf("Pct: got %v, want 50.0", d.Pct)
	}
	if !d.Gt {
		t.Errorf("Gt: got false, want true")
	}
}

func TestCompare_zeroA(t *testing.T) {
	a := Totals{Prompts: 0}
	b := Totals{Prompts: 10}
	cmp := Compare(a, b)
	d := cmp.Prompts
	if d.Delta != 10 {
		t.Errorf("Delta: got %v, want 10", d.Delta)
	}
	if d.Pct != 0.0 {
		t.Errorf("Pct: got %v, want 0.0 (zero-A guard)", d.Pct)
	}
	if !d.Gt {
		t.Errorf("Gt: got false, want true")
	}
}

func TestCompare_equal(t *testing.T) {
	a := Totals{Prompts: 42}
	b := Totals{Prompts: 42}
	cmp := Compare(a, b)
	d := cmp.Prompts
	if d.Delta != 0 {
		t.Errorf("Delta: got %v, want 0", d.Delta)
	}
	if d.Pct != 0 {
		t.Errorf("Pct: got %v, want 0", d.Pct)
	}
	if d.Gt {
		t.Errorf("Gt: got true, want false")
	}
}

func TestCompare_bLessThanA(t *testing.T) {
	a := Totals{Prompts: 200}
	b := Totals{Prompts: 100}
	cmp := Compare(a, b)
	d := cmp.Prompts
	if d.Delta != -100 {
		t.Errorf("Delta: got %v, want -100", d.Delta)
	}
	if d.Pct != -50.0 {
		t.Errorf("Pct: got %v, want -50.0", d.Pct)
	}
	if d.Gt {
		t.Errorf("Gt: got true, want false")
	}
}
