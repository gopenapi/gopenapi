package openapi

import (
	"gopkg.in/yaml.v2"
	"testing"
)

type TestMergeJsonMapCase struct {
	Name string
	A    []yaml.MapItem
	R    string
}

func TestMergeJsonMap(t *testing.T) {
	cases := []TestMergeJsonMapCase{
		{
			Name: "base",
			A: []yaml.MapItem{
				yaml.MapItem{
					Key:   "a",
					Value: 1,
				},
				yaml.MapItem{
					Key:   "b",
					Value: 2,
				},
				yaml.MapItem{
					Key:   "c",
					Value: 1,
				},
				yaml.MapItem{
					Key:   "b",
					Value: 3,
				},
			},
			R: `a: 1
b: 3
c: 1
`,
		},
		{
			Name: "nested",
			A: []yaml.MapItem{
				yaml.MapItem{
					Key:   "a",
					Value: 1,
				},
				yaml.MapItem{
					Key: "b",
					Value: []yaml.MapItem{
						yaml.MapItem{
							Key:   "b-1",
							Value: []interface{}{"a"},
						},
						yaml.MapItem{
							Key:   "b-2",
							Value: 2,
						},
					},
				},
				yaml.MapItem{
					Key:   "c",
					Value: 1,
				},
				yaml.MapItem{
					Key: "b",
					Value: []yaml.MapItem{
						yaml.MapItem{
							Key:   "b-1",
							Value: []interface{}{"b"},
						},
						yaml.MapItem{
							Key:   "b-3",
							Value: 2,
						},
					},
				},
			},
			R: `a: 1
b:
  b-1:
  - a
  - b
  b-2: 2
  b-3: 2
c: 1
`,
		},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			bs, _ := yaml.Marshal(mergeYamlMapKey(c.A))
			if string(bs) != c.R {
				t.Fatalf("Unexpected result on test '%s', expected: %s, got: %s", c.Name, c.R, bs)
			}
		})
	}

	t.Logf("OK")
}
