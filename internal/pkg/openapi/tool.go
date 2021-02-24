package openapi

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
)

// 合并两个yamlItem
func mergeYamlItem(a, b yaml.MapItem) yaml.MapItem {
	if a.Key != b.Key {
		return a
	}
	switch av := a.Value.(type) {
	case []yaml.MapItem:
		// 如果b也是map, 则递归
		switch bv := b.Value.(type) {
		case []yaml.MapItem:
			rv := append(av, bv...)
			r := yaml.MapItem{
				Key:   a.Key,
				Value: nil,
			}
			r.Value = mergeYamlMap(rv)

			return r
		default:
			return a
		}
	case []interface{}:
		switch bv := b.Value.(type) {
		case []interface{}:
			r := yaml.MapItem{
				Key:   a.Key,
				Value: append(av, bv...),
			}
			return r
		default:
			return a
		}
	default:
		return b
	}
}

// 去掉重复key
func mergeYamlMap(a []yaml.MapItem) []yaml.MapItem {
	keyIndex := map[interface{}]int{}
	for i, item := range a {
		keyIndex[item.Key] = i
	}
	var r []yaml.MapItem

	has := map[interface{}]int{}
	for _, item := range a {
		if index, ok := has[item.Key]; ok {
			if ok {
				// key 重复了, 需要合并
				r[index] = mergeYamlItem(r[index], item)
				continue
			}
		}
		r = append(r, item)
		has[item.Key] = len(r) - 1
	}

	return r
}

// copyToBaseType copy any type data to base type
func copyToBaseType(a interface{}) (i interface{}) {
	bs, _ := json.Marshal(a)
	json.Unmarshal(bs, &i)
	return
}
