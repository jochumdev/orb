// Package codecs is provides an interface to encode and decode content types
// to and from byte sequences.
package codecs

import (
	"errors"
	"io"
)

var (
	// ErrNoFileMarshaler happens when we haven't found a marshaler for the given file.
	ErrNoFileMarshaler = errors.New("no marshaler for the given file found")
)

// Map is an alias for an codec map.
// Common keys to use here are either the plugin name or mime types.
type Map map[string]Marshaler

// Marshaler is able to encode/decode a content type to/from a byte sequence.
type Marshaler interface {
	// Marshal encodes "v" into byte sequence.
	// "v" must be a pointer value.
	Marshal(v any) ([]byte, error)

	// Unmarshal decodes "data" into "v".
	// "v" must be a pointer value.
	Unmarshal(data []byte, v any) error

	// Marshals returns if this codec is able to encode the given type.
	Marshals(v any) bool

	// Unmarshals returns if this codec is able to decode the given type.
	Unmarshals(v any) bool

	// NewDecoder returns a Decoder which reads byte sequence from "r".
	NewDecoder(r io.Reader) Decoder

	// NewEncoder returns an Encoder which writes bytes sequence into "w".
	NewEncoder(w io.Writer) Encoder

	// ContentTypes returns the list of content types this codec is able to
	// output.
	ContentTypes() []string

	// Name returns the codec name.
	Name() string

	// Exts returns the common file extensions for this encoder.
	Exts() []string
}

// Decoder decodes a byte sequence.
type Decoder interface {
	Decode(v any) error
}

// Encoder encodes payloads / fields into byte sequence.
type Encoder interface {
	Encode(v any) error
}

// DecoderFunc adapts an decoder function into Decoder.
type DecoderFunc func(v any) error

// Decode delegates invocations to the underlying function itself.
func (f DecoderFunc) Decode(v any) error { return f(v) }

// EncoderFunc adapts an encoder function into Encoder.
type EncoderFunc func(v any) error

// Encode delegates invocations to the underlying function itself.
func (f EncoderFunc) Encode(v any) error { return f(v) }
