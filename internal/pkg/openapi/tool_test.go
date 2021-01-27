package openapi

import (
	"encoding/json"
	"github.com/zbysir/gopenapi/internal/pkg/jsonordered"
	"testing"
)

type TestMergeJsonMapCase struct {
	Name string
	A    jsonordered.MapSlice
	B    jsonordered.MapSlice
	R    string
}

func TestMergeJsonMap(t *testing.T) {
	cases := []TestMergeJsonMapCase{
		{
			Name: "base",
			A: jsonordered.MapSlice{
				jsonordered.MapItem{
					Key: "a",
					Val: 1,
				},
				jsonordered.MapItem{
					Key: "b",
					Val: 2,
				},
			},
			B: jsonordered.MapSlice{
				jsonordered.MapItem{
					Key: "c",
					Val: 1,
				},
				jsonordered.MapItem{
					Key: "b",
					Val: 3,
				},
			},
			R: `{"a":1,"b":3,"c":1}`,
		},
		{
			Name: "nested",
			A: jsonordered.MapSlice{
				jsonordered.MapItem{
					Key: "a",
					Val: 1,
				},
				jsonordered.MapItem{
					Key: "b",
					Val: jsonordered.MapSlice{
						jsonordered.MapItem{
							Key: "b-1",
							Val: "a",
						},
						jsonordered.MapItem{
							Key: "b-2",
							Val: 2,
						},
					},
				},
			},
			B: jsonordered.MapSlice{
				jsonordered.MapItem{
					Key: "c",
					Val: 1,
				},
				jsonordered.MapItem{
					Key: "b",
					Val: jsonordered.MapSlice{
						jsonordered.MapItem{
							Key: "b-1",
							Val: "b",
						},
						jsonordered.MapItem{
							Key: "b-3",
							Val: 2,
						},
					},
				},
			},
			R: `{"a":1,"b":{"b-1":"b","b-2":2,"b-3":2},"c":1}`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			bs, _ := json.Marshal(mergeJsonMap(c.A, c.B))
			if string(bs) != c.R {
				t.Fatalf("Unexpected result on test '%s', expected: %s, got: %s", c.Name, c.R, bs)
			}
		})
	}

	t.Logf("OK")
}
