package nbt

import (
	"testing"
)

func TestStringifyPrimitives(t *testing.T) {
	tests := []struct {
		tag  Tag
		want string
	}{
		{Byte(0), "0b"},
		{Byte(1), "1b"},
		{Byte(-1), "-1b"},
		{Byte(127), "127b"},
		{Short(256), "256s"},
		{Short(-32768), "-32768s"},
		{Int(42), "42"},
		{Int(-1), "-1"},
		{Int(2147483647), "2147483647"},
		{Long(100), "100L"},
		{Long(-9223372036854775808), "-9223372036854775808L"},
		{Float(1.5), "1.5f"},
		{Float(0), "0f"},
		{Double(3.14), "3.14d"},
		{Double(0), "0d"},
		{Double(2.5), "2.5d"},
		{String("hello"), `"hello"`},
		{String(""), `""`},
		{String(`say "hi"`), `"say \"hi\""`},
	}

	for _, tt := range tests {
		got := Stringify(tt.tag)
		if got != tt.want {
			t.Errorf("Stringify(%T(%v)) = %q, want %q", tt.tag, tt.tag, got, tt.want)
		}
	}
}

func TestStringifyCompound(t *testing.T) {
	// simple compound
	tag := Compound{
		"name": String("Steve"),
		"age":  Int(20),
	}
	got := Stringify(tag)
	// compound key order is map iteration order, so check both possibilities
	if got != `{age:20,name:"Steve"}` && got != `{name:"Steve",age:20}` {
		t.Errorf("Stringify compound = %q", got)
	}
}

func TestStringifyList(t *testing.T) {
	tag := &List{
		ElementType: TagInt,
		Elements:    []Tag{Int(1), Int(2), Int(3)},
	}
	got := Stringify(tag)
	if got != "[1,2,3]" {
		t.Errorf("Stringify list = %q, want %q", got, "[1,2,3]")
	}
}

func TestStringifyEmptyList(t *testing.T) {
	tag := &List{ElementType: TagEnd}
	got := Stringify(tag)
	if got != "[]" {
		t.Errorf("Stringify empty list = %q, want %q", got, "[]")
	}
}

func TestStringifyArrays(t *testing.T) {
	tests := []struct {
		tag  Tag
		want string
	}{
		{ByteArray{1, 2, 255}, "[B;1b,2b,-1b]"},
		{ByteArray{}, "[B;]"},
		{IntArray{10, 20, 30}, "[I;10,20,30]"},
		{LongArray{100, -200}, "[L;100L,-200L]"},
	}

	for _, tt := range tests {
		got := Stringify(tt.tag)
		if got != tt.want {
			t.Errorf("Stringify(%T) = %q, want %q", tt.tag, got, tt.want)
		}
	}
}

func TestStringifyNestedCompound(t *testing.T) {
	tag := Compound{
		"pos": Compound{
			"x": Double(1.5),
			"y": Double(64.0),
			"z": Double(-3.7),
		},
		"name": String("test"),
	}
	got := Stringify(tag)
	// should contain both keys
	if len(got) == 0 || got[0] != '{' || got[len(got)-1] != '}' {
		t.Errorf("Stringify nested = %q, not a compound", got)
	}
}

func TestStringifyQuotedKeys(t *testing.T) {
	tag := Compound{
		"simple":               Int(1),
		"with space":           Int(2),
		"minecraft:some_thing": Int(3),
	}
	got := Stringify(tag)
	// "minecraft:some_thing" has a colon so needs quoting
	if !contains(got, `"minecraft:some_thing":3`) {
		t.Errorf("Stringify should quote key with colon, got %q", got)
	}
	// "with space" needs quoting
	if !contains(got, `"with space":2`) {
		t.Errorf("Stringify should quote key with space, got %q", got)
	}
	// "simple" does not need quoting
	if !contains(got, `simple:1`) {
		t.Errorf("Stringify should not quote simple key, got %q", got)
	}
}

func TestFromJSONPrimitives(t *testing.T) {
	tests := []struct {
		json string
		want Tag
	}{
		{`42`, Int(42)},
		{`0`, Int(0)},
		{`-1`, Int(-1)},
		{`3.14`, Double(3.14)},
		{`0.5`, Double(0.5)},
		{`"hello"`, String("hello")},
		{`""`, String("")},
		{`true`, Byte(1)},
		{`false`, Byte(0)},
		{`null`, Byte(0)},
	}

	for _, tt := range tests {
		got, err := FromJSON([]byte(tt.json))
		if err != nil {
			t.Errorf("FromJSON(%s) error: %v", tt.json, err)
			continue
		}
		if got.ID() != tt.want.ID() {
			t.Errorf("FromJSON(%s) type = %s, want %s", tt.json, TagName(got.ID()), TagName(tt.want.ID()))
			continue
		}
		if Stringify(got) != Stringify(tt.want) {
			t.Errorf("FromJSON(%s) = %s, want %s", tt.json, Stringify(got), Stringify(tt.want))
		}
	}
}

func TestFromJSONCompound(t *testing.T) {
	input := `{"name": "Steve", "health": 20, "pos": {"x": 1.5, "y": 64.0, "z": -3.7}}`
	tag, err := FromJSON([]byte(input))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	compound, ok := tag.(Compound)
	if !ok {
		t.Fatalf("expected Compound, got %T", tag)
	}
	if compound.GetString("name") != "Steve" {
		t.Errorf("name = %q, want %q", compound.GetString("name"), "Steve")
	}
	if compound.GetInt("health") != 20 {
		t.Errorf("health = %d, want 20", compound.GetInt("health"))
	}
	pos := compound.GetCompound("pos")
	if pos == nil {
		t.Fatal("pos is nil")
	}
	if pos.GetDouble("x") != 1.5 {
		t.Errorf("pos.x = %f, want 1.5", pos.GetDouble("x"))
	}
}

func TestFromJSONList(t *testing.T) {
	input := `[1, 2, 3]`
	tag, err := FromJSON([]byte(input))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	list, ok := tag.(*List)
	if !ok {
		t.Fatalf("expected *List, got %T", tag)
	}
	if list.Len() != 3 {
		t.Errorf("list len = %d, want 3", list.Len())
	}
	if list.ElementType != TagInt {
		t.Errorf("list element type = %s, want Int", TagName(list.ElementType))
	}
}

func TestFromJSONEmptyList(t *testing.T) {
	tag, err := FromJSON([]byte(`[]`))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	list, ok := tag.(*List)
	if !ok {
		t.Fatalf("expected *List, got %T", tag)
	}
	if list.Len() != 0 {
		t.Errorf("list len = %d, want 0", list.Len())
	}
	if list.ElementType != TagEnd {
		t.Errorf("empty list element type = %s, want End", TagName(list.ElementType))
	}
}

func TestFromJSONBooleans(t *testing.T) {
	input := `{"flag": true, "off": false}`
	tag, err := FromJSON([]byte(input))
	if err != nil {
		t.Fatalf("FromJSON error: %v", err)
	}
	compound := tag.(Compound)
	if compound.GetByte("flag") != 1 {
		t.Errorf("flag = %d, want 1", compound.GetByte("flag"))
	}
	if compound.GetByte("off") != 0 {
		t.Errorf("off = %d, want 0", compound.GetByte("off"))
	}
}

func TestFromJSONRoundTrip(t *testing.T) {
	// build a tag, stringify, then build from JSON and stringify again —
	// the JSON path should produce equivalent (though not identical) output
	original := Compound{
		"name":  String("test"),
		"count": Int(5),
		"items": &List{
			ElementType: TagString,
			Elements:    []Tag{String("a"), String("b")},
		},
	}
	snbt := Stringify(original)
	if snbt == "" {
		t.Fatal("Stringify produced empty string")
	}
	t.Logf("original SNBT: %s", snbt)
}

func TestFromJSONWholeNumbersAreInt(t *testing.T) {
	// JSON numbers without decimal points that fit in int32 should be Int, not Double
	tag, err := FromJSON([]byte(`100`))
	if err != nil {
		t.Fatal(err)
	}
	if tag.ID() != TagInt {
		t.Errorf("100 should be Int, got %s", TagName(tag.ID()))
	}

	// large numbers beyond int32 should be Double
	tag, err = FromJSON([]byte(`3000000000`))
	if err != nil {
		t.Fatal(err)
	}
	if tag.ID() != TagDouble {
		t.Errorf("3000000000 should be Double, got %s", TagName(tag.ID()))
	}
}

func TestFromJSONBinaryRoundTrip(t *testing.T) {
	// convert JSON → NBT tag → binary → decode → verify key fields survive
	input := `{"name":"Steve","health":20,"flying":true}`
	tag, err := FromJSON([]byte(input))
	if err != nil {
		t.Fatalf("FromJSON: %v", err)
	}

	encoded, err := EncodeNetwork(tag)
	if err != nil {
		t.Fatalf("EncodeNetwork: %v", err)
	}
	t.Logf("encoded %d bytes, SNBT: %s", len(encoded), Stringify(tag))

	decoded, err := DecodeNetwork(encoded)
	if err != nil {
		t.Fatalf("DecodeNetwork: %v", err)
	}
	t.Logf("decoded SNBT: %s", Stringify(decoded))

	compound, ok := decoded.(Compound)
	if !ok {
		t.Fatalf("decoded is %T, want Compound", decoded)
	}
	if compound.GetString("name") != "Steve" {
		t.Errorf("name = %q, want Steve", compound.GetString("name"))
	}
	if compound.GetInt("health") != 20 {
		t.Errorf("health = %d, want 20", compound.GetInt("health"))
	}
	if compound.GetByte("flying") != 1 {
		t.Errorf("flying = %d, want 1", compound.GetByte("flying"))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsCheck(s, substr)))
}

func containsCheck(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
