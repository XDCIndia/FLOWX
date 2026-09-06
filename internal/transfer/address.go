package transfer

import "regexp"

// destinationAddressRe matches EVM-style addresses used by the XDC network:
// 0x or xdc prefix followed by 40 hex characters.
var destinationAddressRe = regexp.MustCompile(`^(0x|xdc)[0-9a-fA-F]{40}$`)

// IsValidDestinationAddress reports whether addr is a raw on-chain address
// acceptable as an external payout destination.
func IsValidDestinationAddress(addr string) bool {
	return destinationAddressRe.MatchString(addr)
}
