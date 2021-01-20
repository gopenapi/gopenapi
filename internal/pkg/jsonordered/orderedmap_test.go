package jsonordered

import (
	"encoding/json"
	"testing"
)

// TestJson test for MapSlice
func TestJson(t *testing.T) {
	// 不支持根不是对象的yaml
	cases := []string{
		`{"a":1,"b":"abc2","d":[{"x":2,"a":1}]}`,
		`{"z":{"b":1,"d":[{"x":1,"b":2,"f":1.244555},{"d":"fsdaf","c":"123"}]},"o":{"x":{"d":1,"c":2,"x":3,"a":4}}}`,
	}

	for _, c := range cases {
		i := MapSlice{}
		err := json.Unmarshal([]byte(c), &i)
		if err != nil {
			t.Fatalf("Unmarshal '%s' err: %s", c, err)
		}

		bs, err := json.Marshal(i)
		if err != nil {
			t.Fatal()
		}
		if string(bs) != c {
			t.Errorf("Unexpected result on JSON: '%s'", cases)
			t.Log("Got:     ", string(bs))
			t.Log("Expected:", c)
		}
	}

	t.Log("ok")
}
