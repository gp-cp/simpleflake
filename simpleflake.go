package simpleflake

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"strconv"
	"time"
)

type SimpleflakeId uint64

const (
	nano = 1000 * 1000
)

var (
	// default epoch: 2000-01-01 00:00:00 +0000 UTC = 946684800000
	epoch         int64  = int64(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()) / nano
	timestampBits uint32 = 41
	randomBits    uint32 = 64 - timestampBits
)

// Generate a new 64-bit, roughly-ordered, unique ID
func NewSimpleflake() (id SimpleflakeId, err error) {
	seq, err := randomSequence()
	if err != nil {
		return
	}
	id = buildId(customTimestamp(time.Now()), seq)
	return
}

// Parse a previously generated ID
func Parse(id SimpleflakeId) [2]SimpleflakeId {
	return [2]SimpleflakeId{
		extractBits(id, randomBits, timestampBits) + SimpleflakeId(epoch), // timestamp
		extractBits(id, 0, randomBits),                                    // sequence
	}
}

// Set the epoch to a custom time
func SetCustomEpoch(t time.Time) {
	epoch = t.UTC().UnixNano() / nano
}

// Set the precision level of the timestamp
func SetCustomPrecision(bits uint32) {
	timestampBits = bits
	// reset random bit length
	randomBits = 64 - timestampBits
}

// Build a new 64-bit ID from the timestamp and random sequence
func buildId(ts int64, seq SimpleflakeId) SimpleflakeId {
	return (SimpleflakeId(ts) << randomBits) | seq
}

// Get a custom timestamp to be used to generate a new ID
func customTimestamp(t time.Time) int64 {
	return t.UnixNano()/nano - epoch
}

// Extract bits from a simpleflakeId
func extractBits(data SimpleflakeId, shift uint32, length uint32) SimpleflakeId {
	bitmask := SimpleflakeId(((1 << length) - 1) << shift)
	return ((data & bitmask) >> shift)
}

// Get a random sequence to be used to generate a new ID
func randomSequence() (seq SimpleflakeId, err error) {
	// the maximum random sequence we can generate is 2^randomBits-1
	max := big.NewInt(int64((math.Pow(2, float64(randomBits))) - 1))
	random, err := rand.Int(rand.Reader, max)
	if err == nil {
		seq = SimpleflakeId(random.Uint64())
	}
	return
}

func (u SimpleflakeId) MarshalJSON() ([]byte, error) {
	n := uint64(u)
	s := strconv.FormatUint(n, 10)

	j, e := json.Marshal(s)

	if e != nil {
		return nil, e
	}

	return j, nil

}

func (u *SimpleflakeId) UnmarshalJSON(bs []byte) error {
	var i uint64
	if err := json.Unmarshal(bs, &i); err == nil {
		*u = SimpleflakeId(i)
		return nil
	}
	var s string
	if err := json.Unmarshal(bs, &s); err != nil {
		return errors.New("expected a string or an integer")
	}
	if err := json.Unmarshal([]byte(s), &i); err != nil {
		return err
	}
	*u = SimpleflakeId(i)
	return nil
}

func SimpleflakeIdToString(id SimpleflakeId) string {
	return strconv.FormatUint(uint64(id), 10)
}

func SimpleflakeIdFromString(s string) (SimpleflakeId, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return SimpleflakeId(i), nil
}
