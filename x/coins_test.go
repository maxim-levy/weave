package x

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mustCombineCoins has one return value for tests...
func mustCombineCoins(cs ...Coin) Coins {
	s, err := CombineCoins(cs...)
	if err != nil {
		panic(err)
	}
	return s
}

func TestMakeCoins(t *testing.T) {
	// TODO: verify constructor checks well for errors

	cases := []struct {
		inputs   []Coin
		isEmpty  bool
		isNonNeg bool
		has      []Coin // <= the wallet
		dontHave []Coin // > or outside the wallet
		isErr    bool
	}{
		// empty
		{
			nil,
			true,
			true,
			nil,
			[]Coin{NewCoin(0, 0, "")},
			false,
		},
		// ignore 0
		{
			[]Coin{NewCoin(0, 0, "FOO")},
			true,
			true,
			nil,
			[]Coin{NewCoin(0, 0, "FOO")},
			false,
		},
		// simple
		{
			[]Coin{NewCoin(40, 0, "FUD")},
			false,
			true,
			[]Coin{NewCoin(10, 0, "FUD"), NewCoin(40, 0, "FUD")},
			[]Coin{NewCoin(40, 1, "FUD"), NewCoin(40, 0, "FUN")},
			false,
		},
		// simple with issuer
		{
			[]Coin{NewCoin(40, 0, "FUD").WithIssuer("johnny")},
			false,
			true,
			[]Coin{NewCoin(37, 0, "FUD").WithIssuer("johnny")},
			[]Coin{NewCoin(10, 0, "FUD")},
			false,
		},
		// out of order, with negative
		{
			[]Coin{NewCoin(-20, -3, "FIN"), NewCoin(40, 5, "BON")},
			false,
			false,
			[]Coin{NewCoin(40, 4, "BON"), NewCoin(-30, 0, "FIN")},
			[]Coin{NewCoin(40, 6, "BON"), NewCoin(-20, 0, "FIN")},
			false,
		},
		// out of order, with different issuers
		{
			[]Coin{
				NewCoin(200, 0, "FIN").WithIssuer("chain-2"),
				NewCoin(100, 0, "FIN").WithIssuer("chain-1"),
			},
			false,
			true,
			// make sure both match
			[]Coin{
				NewCoin(100, 0, "FIN").WithIssuer("chain-1"),
				NewCoin(200, 0, "FIN").WithIssuer("chain-2"),
			},
			// don't combine the two issuers
			[]Coin{NewCoin(200, 0, "FIN").WithIssuer("chain-1")},
			false,
		},
		// combine and remove
		{
			[]Coin{NewCoin(-123, -456, "BOO"), NewCoin(123, 456, "BOO")},
			true,
			true,
			nil,
			[]Coin{NewCoin(0, 0, "BOO")},
			false,
		},
		// safely combine
		{
			[]Coin{NewCoin(12, 0, "ADA"), NewCoin(-123, -456, "BOO"), NewCoin(124, 756, "BOO")},
			false,
			true,
			[]Coin{NewCoin(12, 0, "ADA"), NewCoin(1, 300, "BOO")},
			[]Coin{NewCoin(13, 0, "ADA"), NewCoin(1, 400, "BOO")},
			false,
		},
		// verify invalid input cur -> error
		{
			[]Coin{NewCoin(1, 2, "AL2")},
			false, false, nil, nil,
			true,
		},
		// verify invalid input values -> error
		{
			[]Coin{NewCoin(MaxInt+3, 2, "AND")},
			false, false, nil, nil,
			true,
		},
		// if we can combine invalid inputs, then acceptable?
		{
			[]Coin{NewCoin(MaxInt+3, 2, "AND"), NewCoin(-10, 0, "AND")},
			false,
			true,
			[]Coin{NewCoin(MaxInt-8, 0, "AND")},
			nil,
			false,
		},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("case-%d", idx), func(t *testing.T) {
			s, err := CombineCoins(tc.inputs...)
			if tc.isErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NoError(t, s.Validate())
			assert.Equal(t, tc.isEmpty, s.IsEmpty())
			assert.Equal(t, tc.isNonNeg, s.IsNonNegative())

			for _, h := range tc.has {
				assert.True(t, s.Contains(h))
			}
			for _, d := range tc.dontHave {
				assert.False(t, s.Contains(d))
			}
		})
	}
}

// TestCombine checks combine and equals
// and thereby checks add
func TestCombine(t *testing.T) {
	cases := []struct {
		a, b  Coins
		comb  Coins
		isErr bool
	}{
		// empty
		{
			mustCombineCoins(), mustCombineCoins(), mustCombineCoins(), false,
		},
		// one plus one
		{
			mustCombineCoins(NewCoin(MaxInt, 5, "ABC")),
			mustCombineCoins(NewCoin(-MaxInt, -4, "ABC")),
			mustCombineCoins(NewCoin(0, 1, "ABC")),
			false,
		},
		// multiple
		{
			mustCombineCoins(NewCoin(7, 8, "FOO"), NewCoin(8, 9, "BAR")),
			mustCombineCoins(NewCoin(5, 4, "APE"), NewCoin(2, 1, "FOO")),
			mustCombineCoins(NewCoin(5, 4, "APE"), NewCoin(8, 9, "BAR"), NewCoin(9, 9, "FOO")),
			false,
		},
		// overflows
		{
			mustCombineCoins(NewCoin(MaxInt, 0, "ADA")),
			mustCombineCoins(NewCoin(2, 0, "ADA")),
			Coins{},
			true,
		},
	}

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("case-%d", idx), func(t *testing.T) {

			ac := tc.a.Count()
			bc := tc.b.Count()

			res, err := tc.a.Combine(tc.b)
			// don't modify original Coinss
			assert.Equal(t, ac, tc.a.Count())
			assert.Equal(t, bc, tc.b.Count())
			if tc.isErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NoError(t, res.Validate())
			assert.True(t, tc.comb.Equals(res))
			// result should only be the same as an input
			// if the other input was empty
			assert.Equal(t, tc.a.IsEmpty(),
				tc.b.Equals(res))
			assert.Equal(t, tc.b.IsEmpty(),
				tc.a.Equals(res))
		})
	}
}
