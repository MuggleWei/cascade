package common

import (
	"encoding/json"
	"errors"
)

// common stream data
type StreamData struct {
	Op   string      `json:"op"`   // operation command
	Data interface{} `json:"data"` // data
}

// parse stream data
func ParseStreamData(bytes []byte) (string, []byte, error) {
	var objmap map[string]*json.RawMessage
	err := json.Unmarshal(bytes, &objmap)
	if err != nil {
		return "", nil, err
	}

	var op_bytes []byte
	if op, ok := objmap["op"]; !ok {
		return "", nil, errors.New("without op field")
	} else {
		if op == nil {
			return "", nil, errors.New("nil op field")
		}

		op_bytes, err = op.MarshalJSON()
		if err != nil {
			return "", nil, err
		}
		if len(op_bytes) < 3 {
			return "", nil, errors.New("empty op field")
		}
	}

	var data_bytes []byte
	if data, ok := objmap["data"]; !ok {
		return "", nil, errors.New("without data field")
	} else {
		if data != nil {
			data_bytes, err = data.MarshalJSON()
			if err != nil {
				return "", nil, err
			}
		}
	}

	op_name := string(op_bytes[1 : len(op_bytes)-1])
	return op_name, data_bytes, nil
}
