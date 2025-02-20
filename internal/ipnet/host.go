package ipnet

import (
	"errors"
	"fmt"
	"net"
	"net/netip"
)

// HostID is the host part of an IPv6 address.
type HostID interface {
	// Describe prints the HostID.
	Describe() string

	// WithPrefix calculates the new address with a prefix.
	WithPrefix(prefix netip.Prefix) (netip.Addr, bool)
}

// mask gives a bitwise mask:
// - mask(0): 11111111.
// - mask(1): 01111111.
// - mask(2): 00111111.
// - mask(3): 00011111.
// - mask(4): 00001111.
// - mask(5): 00000111.
// - mask(6): 00000011.
// - mask(7): 00000001.
func mask(s int) byte {
	return ^byte(0) >> s
}

// IP6Suffix represents a suffix of an IPv6 address.
type IP6Suffix [16]byte

// Describe prints the suffix as an IPv6 address.
func (r IP6Suffix) Describe() string { return netip.AddrFrom16(r).String() }

func (r IP6Suffix) mask(prefixLen int) IP6Suffix {
	for i := range prefixLen / 8 {
		r[i] = 0
	}
	r[prefixLen/8] &= mask(prefixLen % 8)
	return r
}

// WithPrefix combines a prefix and a host ID to construct an IPv6 address.
func (r IP6Suffix) WithPrefix(prefix netip.Prefix) (netip.Addr, bool) {
	ip := r.mask(prefix.Bits())
	prefixAsBytes := prefix.Masked().Addr().As16()
	for i := range 128 / 8 {
		ip[i] |= prefixAsBytes[i]
	}
	return netip.AddrFrom16(ip), true
}

// EUI48 represents a MAC (EUI-48) address.
type EUI48 [6]byte

// Describe prints the suffix as a MAC address.
func (e EUI48) Describe() string { return net.HardwareAddr(e[:]).String() }

// WithPrefix combines a prefix and a host ID to construct an IPv6 address.
func (e EUI48) WithPrefix(prefix netip.Prefix) (netip.Addr, bool) {
	if prefix.Bits() > 64 {
		return netip.Addr{}, false
	}
	prefixAsBytes := prefix.Masked().Addr().As16()

	bytes := [16]byte{
		prefixAsBytes[0],
		prefixAsBytes[1],
		prefixAsBytes[2],
		prefixAsBytes[3],
		prefixAsBytes[4],
		prefixAsBytes[5],
		prefixAsBytes[6],
		prefixAsBytes[7],
		e[0] ^ 0x02, // flip the global-local bit
		e[1],
		e[2],
		0xff,
		0xfe,
		e[3],
		e[4],
		e[5],
	}
	return netip.AddrFrom16(bytes), true
}

// Errors from ParseHost.
var (
	ErrNotHostID        = errors.New("not an IPv6 or MAC (EUI-48) address")
	ErrHostIDHasIP6Zone = errors.New("IPv6 address as a host ID should not have IPv6 zone")
)

// ParseHost parses a host ID for an IPv6 address.
func ParseHost(s string) (HostID, error) {
	if s == "" {
		return nil, nil //nolint:nilnil
	}

	ip, errIP := netip.ParseAddr(s)
	if errIP == nil {
		if !ip.Is6() {
			return nil, ErrNotHostID
		}
		if ip.Zone() != "" {
			return nil, ErrHostIDHasIP6Zone
		}

		return IP6Suffix(ip.As16()), nil
	}

	// Possible formats for MAC (EUI-48)
	// 00:00:5e:00:53:01
	// 00-00-5e-00-53-01
	// 0000.5e00.5301
	mac, errMAC := net.ParseMAC(s)
	if errMAC != nil || len(mac) != 6 {
		return nil, fmt.Errorf("%w: as IPv6 address, %w; as EUI48 address, %w", ErrNotHostID, errIP, errMAC)
	}
	return EUI48(mac), nil
}
