package openapi

import "github.com/zbysir/gopenapi/internal/pkg/jsonordered"

// 合并两个mapSlice
func mergeJsonMap(a, b jsonordered.MapSlice) jsonordered.MapSlice {
	keyIndex := map[string]int{}
	for i, item := range a {
		keyIndex[item.Key] = i
	}
	var r = make(jsonordered.MapSlice, len(a))
	copy(r, a)

	for _, item := range b {
		if index, ok := keyIndex[item.Key]; ok {
			av, avIsMap := a[index].Val.(jsonordered.MapSlice)
			if !avIsMap {
				r[index] = item
				continue
			}

			bv, bvIsMap := item.Val.(jsonordered.MapSlice)
			if !bvIsMap {
				r[index] = item
				continue
			}

			r[index] = jsonordered.MapItem{
				Key: item.Key,
				Val: mergeJsonMap(av, bv),
			}
			continue
		}

		r = append(r, item)
	}

	return r
}
