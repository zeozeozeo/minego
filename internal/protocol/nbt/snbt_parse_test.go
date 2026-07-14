package nbt

import "testing"

func TestParseEntityNBT(t *testing.T) {
	tag, err := Parse(`{CustomName:'"Bob"',NoGravity:1b,Health:20.0f,Count:5,Pos:[0.0d,1.5d,2.0d]}`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	c, ok := tag.(Compound)
	if !ok {
		t.Fatalf("root is %T, want Compound", tag)
	}
	if got := c.GetString("CustomName"); got != `"Bob"` {
		t.Errorf("CustomName = %q, want %q", got, `"Bob"`)
	}
	if got := c.GetByte("NoGravity"); got != 1 {
		t.Errorf("NoGravity = %d, want 1", got)
	}
	if got := c.GetFloat("Health"); got != 20.0 {
		t.Errorf("Health = %v, want 20", got)
	}
	if got := c.GetInt("Count"); got != 5 {
		t.Errorf("Count = %d, want 5", got)
	}
	pos := c.GetList("Pos")
	if pos.Len() != 3 || pos.ElementType != TagDouble {
		t.Fatalf("Pos list = %+v, want 3 doubles", pos)
	}
	if d, _ := pos.Get(1).(Double); d != 1.5 {
		t.Errorf("Pos[1] = %v, want 1.5", d)
	}
}

func TestParseArraysAndNesting(t *testing.T) {
	tag, err := Parse(`{ints:[I;1,2,3],bytes:[B;1b,2b],nested:{a:{b:{c:7L}}},list:[{k:1},{k:2}]}`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	c := tag.(Compound)
	if ia := c.GetIntArray("ints"); len(ia) != 3 || ia[2] != 3 {
		t.Errorf("ints = %v", ia)
	}
	if ba := c.GetByteArray("bytes"); len(ba) != 2 || ba[0] != 1 {
		t.Errorf("bytes = %v", ba)
	}
	if got := c.GetCompound("nested").GetCompound("a").GetCompound("b").GetLong("c"); got != 7 {
		t.Errorf("nested.a.b.c = %d, want 7", got)
	}
	if l := c.GetList("list"); l.Len() != 2 || l.ElementType != TagCompound {
		t.Errorf("list = %+v", l)
	}
}

func TestParseRoundTripBinary(t *testing.T) {
	// parse -> binary encode -> decode -> compare a value
	tag, err := Parse(`{x:42,name:"hi",f:3.5f}`)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	data, err := Encode(tag, "", false)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, _, err := Decode(data, false)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	c := got.(Compound)
	if c.GetInt("x") != 42 || c.GetString("name") != "hi" || c.GetFloat("f") != 3.5 {
		t.Errorf("round-trip mismatch: %+v", c)
	}
}

func TestParseErrors(t *testing.T) {
	for _, bad := range []string{`{`, `{a:}`, `[1,2`, `{a:1 b:2}`, ``, `{a:1},`} {
		if _, err := Parse(bad); err == nil {
			t.Errorf("Parse(%q) = nil error, want error", bad)
		}
	}
}
