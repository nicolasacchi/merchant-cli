package output

import (
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("json encode: %w", err)
	}
	return nil
}

// ProtoToRaw renders a protobuf message as canonical protobuf-JSON.
func ProtoToRaw(m proto.Message) (json.RawMessage, error) {
	b, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("protojson: %w", err)
	}
	return b, nil
}

// PrintProto prints one protobuf message as JSON.
func PrintProto(m proto.Message) error {
	raw, err := ProtoToRaw(m)
	if err != nil {
		return err
	}
	return PrintJSON(raw)
}

// PrintProtos prints {key: [...], count, nextPageToken}.
func PrintProtos[T proto.Message](key string, items []T, next string) error {
	raws := make([]json.RawMessage, 0, len(items))
	for _, it := range items {
		raw, err := ProtoToRaw(it)
		if err != nil {
			return err
		}
		raws = append(raws, raw)
	}
	return PrintJSON(map[string]any{key: raws, "totalResults": len(raws), "nextPageToken": next})
}
