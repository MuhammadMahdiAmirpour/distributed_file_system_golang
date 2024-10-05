package p2p

import "io"

type Decoder interface {
	Decode(reader io.Reader, msg any) error
}
