package excel

import (
	pmcpb "pulse/helper/utils/pmcpb/protobuf"
	"pulse/helper/utils/server/encoding"
	"fmt"
)

func init() {
	encoding.RegisterCodec(codec{})
}

type codec struct{}

func (codec) Marshal(v any) ([]byte, error) {
	f, ok := v.(*pmcpb.FileInfo)
	if !ok {
		return nil, fmt.Errorf("excel codec: expected []byte input")
	}
	return f.FileData, nil
}

func (codec) Unmarshal(data []byte, v any) error {
	if b, ok := v.(*[]byte); ok {
		*b = data
		return nil
	}
	return nil
}

func (codec) Name() string {
	return "vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}
