package jsonordered

import (
	"encoding/json"
	"github.com/buger/jsonparser"
)

// MapSlice encodes and decodes as a JSON map.
// The order of keys is preserved when encoding and decoding.
// Just like yaml.MapSlice
type MapSlice []MapItem

// MapItem is an item in a MapSlice.
type MapItem struct {
	Key string
	Val interface{}
}

func (j *MapSlice) UnmarshalJSON(bytes []byte) error {
	err := jsonparser.ObjectEach(bytes, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		item, err := unmarshal(value, dataType)
		if err != nil {
			return err
		}

		*j = append(*j, MapItem{
			Key: string(key),
			Val: item,
		})

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func unmarshal(bs []byte, dataType jsonparser.ValueType) (interface{}, error) {
	switch dataType {
	case jsonparser.Object:
		i := MapSlice{}
		err := json.Unmarshal(bs, &i)
		if err != nil {
			return nil, err
		}

		return i, nil

	case jsonparser.Array:
		kvs := make([]interface{}, 0)
		_, err := jsonparser.ArrayEach(bs, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			item, err := unmarshal(value, dataType)
			if err != nil {
				return
			}
			kvs = append(kvs, item)
			return
		})
		if err != nil {
			return nil, err
		}

		return kvs, nil
	case jsonparser.String:
		// append "
		bsc := make([]byte, len(bs)+2)
		copy(bsc[1:], bs)
		bsc[0] = '"'
		bsc[len(bsc)-1] = '"'

		bs = bsc
		fallthrough
	default:
		var i interface{}
		err := json.Unmarshal(bs, &i)
		if err != nil {
			return nil, err
		}
		return i, nil
	}
}

func (j MapSlice) MarshalJSON() ([]byte, error) {
	var bs = []byte(`{}`)
	var err error
	for _, item := range j {
		itembs, err := json.Marshal(item.Val)
		if err != nil {
			return nil, err
		}
		bs, err = jsonparser.Set(bs, itembs, item.Key)
		if err != nil {
			return nil, err
		}
	}

	return bs, err
}

func (j MapSlice) Get(key string) (v interface{}, exist bool) {
	for _, item := range j {
		if item.Key == key {
			return item.Val, true
		}
	}

	return nil, false
}
