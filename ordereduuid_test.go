package ordereduuid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testUUIDs = []struct {
	u        *OrderedUUID
	valid    bool
	original string
	stripped []byte
	ordered  []byte
}{
	{&OrderedUUID{UUID: "c5cd73e6-b074-48ac-85f3-e55fbd199a31"}, true,
		"c5cd73e6-b074-48ac-85f3-e55fbd199a31", []byte("c5cd73e6b07448ac85f3e55fbd199a31"),
		[]byte("48ACB074C5CD73E685F3E55FBD199A31")},
	{&OrderedUUID{UUID: "e70d8211-3b40-4883-9ee9-a11f97d984c4"}, true,
		"e70d8211-3b40-4883-9ee9-a11f97d984c4", []byte("e70d82113b4048839ee9a11f97d984c4"),
		[]byte("48833B40E70D82119EE9A11F97D984C4")},
	{&OrderedUUID{UUID: "8494415c-6724-ca42-bfba-9ab275b3112b"}, true,
		"8494415c-6724-ca42-bfba-9ab275b3112b", []byte("8494415c6724ca42bfba9ab275b3112b"),
		[]byte("CA4267248494415CBFBA9AB275B3112B")},
	{&OrderedUUID{UUID: "0\x01-CA42672-48494415CBFBA-9AB275B310\x01"}, false,
		"0\x01-CA42672-48494415CBFBA-9AB275B310\x01", []byte("0\x01CA4267248494415CBFBA9AB275B310\x01"),
		[]byte("0\x01CA4267248494415CBFBA9AB275B310\x01")},
	{&OrderedUUID{UUID: "abcd"}, false, "abcd", []byte("abcd"), []byte("ABCD")},
	{&OrderedUUID{UUID: "0"}, false, "0", []byte("0"), []byte("0")},
	{&OrderedUUID{}, false, "", []byte(nil), []byte("")},
}

func TestOrderedUUID_New(t *testing.T) {
	u := New()
	assert.NotNil(t, u)
	assert.True(t, u.Valid)
}

func TestOrderedUUID_Parse(t *testing.T) {
	for _, uuid := range testUUIDs {
		u := Parse(uuid.original)
		assert.NotNil(t, u)
		if assert.Equal(t, u.Valid, uuid.valid, uuid.original) {
			assert.Equal(t, u.UUID, uuid.u.UUID)
		}
	}
}

func TestOrderedUUID_Convert(t *testing.T) {
	for _, uuid := range testUUIDs {
		v, err := uuid.u.Value()
		assert.Equal(t, uuid.valid, uuid.u.Valid)
		if err != nil {
			assert.Equal(t, err, errUUIDInvalid)
			assert.Nil(t, v)
			continue
		}

		assert.NoError(t, err)
		assert.NotNil(t, v)
		assert.Equal(t, uuid.u.Scan(v) == nil, uuid.valid)
	}
}

func TestOrderedUUID_BadScan(t *testing.T) {
	for _, uuid := range testUUIDs {
		if !uuid.valid {
			if err := uuid.u.Scan(interface{}(uuid.ordered)); assert.Error(t, err) {
				assert.Equal(t, err, errUUIDFormat)
			}
		}
	}

	u := &OrderedUUID{}
	if err := u.Scan(interface{}('r')); assert.Error(t, err) {
		assert.Equal(t, err, errUUIDFormat)
	}
}

// Test the case where a nested struct does not initialize OrderedUUID
func TestOrderedUUID_BadValue(t *testing.T) {
	u := func() *OrderedUUID { return nil }()
	v, err := u.Value()
	assert.Empty(t, v)
	assert.Equal(t, err, errUUIDInvalid)
}

func TestOrderedUUID_formats(t *testing.T) {
	for i, uuid := range testUUIDs {
		assert.Equal(t, uuid.stripped, uuid.u.stripDash(), "unexpected error on %v", i+1)
		assert.Equal(t, uuid.ordered, uuid.u.orderedUUID())

		err := uuid.u.formatUUID(uuid.ordered)
		if !uuid.valid {
			if assert.Error(t, err) {
				assert.Equal(t, err, errUUIDFormat)
			}
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, uuid.original, uuid.u.String())
	}
}
