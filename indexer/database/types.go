package database

import (
	"database/sql/driver"
	"errors"
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/jackc/pgtype"
)

var u256BigIntOverflow = new(big.Int).Exp(big.NewInt(2), big.NewInt(256), nil)
var big10 = big.NewInt(10)

var ErrU256Overflow = errors.New("number exceeds u256")
var ErrU256ContainsDecimal = errors.New("number contains fractional digits")
var ErrU256Null = errors.New("number cannot be null")

// U256 is a wrapper over big.Int that conforms to the database U256 numeric domain type
type U256 struct {
	Int *big.Int
}

// Scan implements the database/sql Scanner interface.
func (u256 *U256) Scan(src interface{}) error {
	// deserialize as a numeric
	var numeric pgtype.Numeric
	err := numeric.Scan(src)
	if err != nil {
		return err
	} else if numeric.Exp < 0 {
		return ErrU256ContainsDecimal
	} else if numeric.Status == pgtype.Null {
		return ErrU256Null
	}

	// factor in the powers of 10
	num := numeric.Int
	if numeric.Exp > 0 {
		factor := new(big.Int).Exp(big10, big.NewInt(int64(numeric.Exp)), nil)
		num.Mul(num, factor)
	}

	// check bounds before setting the u256
	if num.Cmp(u256BigIntOverflow) >= 0 {
		return ErrU256Overflow
	} else {
		u256.Int = num
	}

	return nil
}

// Value implements the database/sql/driver Valuer interface.
func (u256 U256) Value() (driver.Value, error) {
	// check bounds
	if u256.Int == nil {
		return nil, ErrU256Null
	} else if u256.Int.Cmp(u256BigIntOverflow) >= 0 {
		return nil, ErrU256Overflow
	}

	// simply encode as a numeric with no Exp set (non-decimal)
	numeric := pgtype.Numeric{Int: u256.Int, Status: pgtype.Present}
	return numeric.Value()
}

type RLPHeader types.Header

func (h *RLPHeader) EncodeRLP(w io.Writer) error {
	return types.NewBlockWithHeader((*types.Header)(h)).EncodeRLP(w)
}

func (h *RLPHeader) DecodeRLP(s *rlp.Stream) error {
	block := new(types.Block)
	err := block.DecodeRLP(s)
	if err != nil {
		return err
	}

	header := block.Header()
	*h = (RLPHeader)(*header)
	return nil
}

func (h *RLPHeader) Header() *types.Header {
	return (*types.Header)(h)
}

func (h *RLPHeader) Hash() common.Hash {
	return h.Header().Hash()
}

type Bytes []byte

func (b Bytes) Bytes() []byte {
	return b[:]
}
func (b *Bytes) SetBytes(bytes []byte) {
	*b = bytes
}
