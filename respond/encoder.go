package respond

import "encoding/json"

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
	ContentType() string
}

type jsonEncoder struct {}

var JSON Encoder = (*jsonEncoder)(nil)

func (*jsonEncoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (*jsonEncoder) ContentType() string {
	return "application/json; charset=utf-8"
}
