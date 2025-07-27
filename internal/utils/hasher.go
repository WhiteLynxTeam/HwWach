package utils

type Hasher interface {
	Hasher(data []byte) []byte
}
