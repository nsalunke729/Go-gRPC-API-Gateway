// Package codec registers a JSON codec that replaces the default protobuf
// codec for all gRPC calls within this process. This lets us define plain
// Go structs as message types without running protoc. Import this package
// with a blank identifier in every binary's main.go.
package codec

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(JSON{})
}

// JSON is a gRPC codec that serialises messages as JSON.
type JSON struct{}

func (JSON) Marshal(v interface{}) ([]byte, error)        { return json.Marshal(v) }
func (JSON) Unmarshal(data []byte, v interface{}) error   { return json.Unmarshal(data, v) }
func (JSON) Name() string                                 { return "proto" } // override default codec
