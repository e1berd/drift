// Package typeid encodes UUIDs as TypeIDs: a lowercase type prefix, an underscore and the
// 128-bit UUID in 26 Crockford base32 characters (https://github.com/jetify-com/typeid/tree/main/spec).
package typeid

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	alphabet      = "0123456789abcdefghjkmnpqrstvwxyz"
	suffixLength  = 26
	MessagePrefix = "msg"
)

var ErrInvalid = errors.New("invalid typeid")

var decodeTable = func() [256]byte {
	var table [256]byte
	for i := range table {
		table[i] = 0xff
	}
	for i := range len(alphabet) {
		table[alphabet[i]] = byte(i)
	}
	return table
}()

// Encode returns "<prefix>_<suffix>" for the UUID.
func Encode(prefix string, id [16]byte) string {
	var suffix [suffixLength]byte
	carry := uint(0)
	bits := uint(2)
	position := 0
	for _, value := range id {
		carry = carry<<8 | uint(value)
		bits += 8
		for bits >= 5 {
			bits -= 5
			suffix[position] = alphabet[carry>>bits&31]
			position++
		}
	}
	return prefix + "_" + string(suffix[:])
}

// Decode parses "<prefix>_<suffix>" and rejects any other prefix.
func Decode(prefix, value string) ([16]byte, error) {
	var id [16]byte
	suffix, found := strings.CutPrefix(value, prefix+"_")
	if !found || len(suffix) != suffixLength || suffix[0] > '7' {
		return id, ErrInvalid
	}
	carry := uint(0)
	bits := uint(0)
	position := 0
	for index := range suffixLength {
		digit := decodeTable[suffix[index]]
		if digit == 0xff {
			return id, ErrInvalid
		}
		carry = carry<<5 | uint(digit)
		bits += 5
		if index == 0 {
			bits -= 2
		}
		if bits >= 8 {
			bits -= 8
			id[position] = byte(carry >> bits)
			position++
		}
	}
	return id, nil
}

// MessageID renders a stored message UUID as its public "msg_…" id.
func MessageID(id pgtype.UUID) string {
	return Encode(MessagePrefix, id.Bytes)
}

// ParseMessageID accepts a "msg_…" id or, for ids written before TypeIDs, a plain UUID.
func ParseMessageID(value string) (pgtype.UUID, error) {
	if bytes, err := Decode(MessagePrefix, value); err == nil {
		return pgtype.UUID{Bytes: bytes, Valid: true}, nil
	}
	var id pgtype.UUID
	if err := id.Scan(value); err != nil {
		return id, ErrInvalid
	}
	return id, nil
}

// NormalizeMessageID turns a stored reference, old plain UUID or new TypeID, into the public id.
func NormalizeMessageID(value string) string {
	id, err := ParseMessageID(value)
	if err != nil {
		return value
	}
	return MessageID(id)
}
