// Package ordereduuid is a SQL type converter from String to Binary
// by way of reordering the the most significant digits first as inspired by
// https://www.percona.com/blog/2014/12/19/store-uuid-optimized-way/
package ordereduuid

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"errors"

	_uuid "github.com/pborman/uuid"
)

const dash = rune('-')

var (
	errUUIDFormat  = errors.New("ordereduuid: cannot convert type")
	errUUIDInvalid = errors.New("ordereduuid: invalid uuid value")
)

// OrderedUUID is the SQL type that is convertable from binary/bytes
// to string
type OrderedUUID struct {
	UUID  string
	Valid bool
}

// UUID exports underlying pborman/uuid package
func UUID() string {
	return _uuid.New()
}

// New returns a valid OrderedUUID
func New() *OrderedUUID {
	return &OrderedUUID{
		UUID:  UUID(),
		Valid: true,
	}
}

// Parse returns a validated OrderedUUID
func Parse(u string) *OrderedUUID {
	o := &OrderedUUID{UUID: u}
	o.orderedUUID()
	return o
}

// Scan implements the Scanner interface.
func (o *OrderedUUID) Scan(value interface{}) error {
	h, ok := value.([]byte)
	if !ok || len(h) != 16 {
		return errUUIDFormat
	}
	dst := make([]byte, hex.EncodedLen(len(h)))
	hex.Encode(dst, h)

	return o.formatUUID(dst)
}

// String returns the UUID value
func (o *OrderedUUID) String() string {
	return o.UUID
}

// Value implements the driver Valuer interface.
func (o *OrderedUUID) Value() (driver.Value, error) {
	h := o.orderedUUID()
	if len(h) != 32 {
		return nil, errUUIDInvalid
	}
	dst := make([]byte, hex.DecodedLen(len(h)))
	if _, err := hex.Decode(dst, h); err != nil {
		return nil, err
	}

	return dst, nil
}

// RETURN CONCAT(
//        SUBSTR(LCASE(HEX(uuid)), 9, 8),  "-",
//        SUBSTR(LCASE(HEX(uuid)), 5, 4),  "-",
//        SUBSTR(LCASE(HEX(uuid)), 1, 4),  "-",
//        SUBSTR(LCASE(HEX(uuid)), 17, 4), "-",
//        SUBSTR(LCASE(HEX(uuid)) , 21)
// )
func (o *OrderedUUID) formatUUID(h []byte) error {
	h = bytes.ToLower(h)

	if len(h) != 32 {
		return errUUIDFormat
	}

	buf := new(bytes.Buffer)
	buf.Write(h[8:16])
	buf.WriteRune(dash)
	buf.Write(h[4:8])
	buf.WriteRune(dash)
	buf.Write(h[0:4])
	buf.WriteRune(dash)
	buf.Write(h[16:20])
	buf.WriteRune(dash)
	buf.Write(h[20:])

	o.Valid = true
	o.UUID = buf.String()
	return nil
}

// RETURN UNHEX(CONCAT(
//     SUBSTR(uuid, 15, 4), SUBSTR(uuid, 10,4),
//     SUBSTR(uuid, 1, 8),  SUBSTR(uuid, 20, 4),
//     SUBSTR(uuid, 25)
// ))
func (o *OrderedUUID) orderedUUID() []byte {
	if o == nil {
		return []byte(nil)
	}

	h := bytes.ToUpper(o.stripDash())

	if len(h) != 32 {
		return h
	}

	buf := new(bytes.Buffer)
	buf.Write(h[12:16])
	buf.Write(h[8:12])
	buf.Write(h[0:8])
	buf.Write(h[16:20])
	buf.Write(h[20:])

	o.Valid = true
	return buf.Bytes()
}

func (o *OrderedUUID) stripDash() []byte {
	return bytes.Replace([]byte(o.UUID), []byte("-"), []byte{}, -1)
}
