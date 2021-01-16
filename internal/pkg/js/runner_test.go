package js

import (
	"fmt"
	"testing"
)

func TestRunJs(t *testing.T) {
	getter := func(name string) (interface{}, error) {
		switch name {
		case "a":
			return 2, nil
		case "obj":
			return map[string]interface{}{
				"a": 3,
			}, nil
		case "getter":
			return func(k string) (v interface{}) {
				switch k {
				case "get1000":
					return 1000
				}
				return 0
			}, nil
		}
		return nil, nil
	}

	cases := []struct {
		js   string
		want string
	}{
		{
			js:   "obj.a+1",
			want: "4 float64",
		},
		{
			js:   "getter.get1000 + 1",
			want: "1001 float64",
		},
	}

	for _, c := range cases {
		v, err := RunJs(c.js, getter)
		if err != nil {
			t.Fatal(err)
		}

		got := fmt.Sprintf("%+v %T", v, v)
		if got != c.want {
			t.Fatalf("want: %s, got: %s", c.want, got)
		}
	}

	t.Log("ok")
}
