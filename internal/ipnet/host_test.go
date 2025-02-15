// vim: nowrap
package ipnet_test

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/favonia/cloudflare-ddns/internal/ipnet"
)

func TestHostIDDescribe(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		input    ipnet.HostID
		expected string
	}{
		"ip6suffix": {
			ipnet.IP6Suffix{0x00, 0x00, 0x00, 0x00, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			"::4455:6677:8899:aabb:ccdd:eeff",
		},
		"mac": {
			ipnet.EUI48{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			"aa:bb:cc:dd:ee:ff",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.expected, tc.input.Describe())
		})
	}
}

func TestHostIDWithPrefix(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		input  ipnet.HostID
		prefix netip.Prefix
		ok     bool
		addr   netip.Addr
	}{
		"ip6suffix": {
			ipnet.IP6Suffix{0x00, 0x00, 0x00, 0x00, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			netip.MustParsePrefix("1122::/40"),
			true,
			netip.MustParseAddr("1122::55:6677:8899:aabb:ccdd:eeff"),
		},
		"mac": {
			ipnet.EUI48{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			netip.MustParsePrefix("1122::/24"),
			true,
			netip.MustParseAddr("1122::a8bb:ccff:fedd:eeff"),
		},
		"mac/96": {
			ipnet.EUI48{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
			netip.MustParsePrefix("1122::/96"),
			false,
			netip.Addr{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			addr, ok := tc.input.WithPrefix(tc.prefix)
			require.Equal(t, tc.ok, ok)
			require.Equal(t, tc.addr, addr)
		})
	}
}

func TestParseHostID(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		input     string
		prefixLen int
		err       error
		hostID    ipnet.HostID
	}{
		"empty": {"", 40, nil, nil},
		"ip6": {
			"11:2233:4455:6677:8899:aabb:ccdd:eeff",
			40,
			nil,
			ipnet.IP6Suffix{0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		},
		"ip6/zone": {
			"11:2233:4455:6677:8899:aabb:ccdd:eeff%eth0",
			40,
			ipnet.ErrHostIDHasIP6Zone,
			nil,
		},
		"mac": {
			"aa:bb:cc:dd:ee:ff",
			40,
			nil,
			ipnet.EUI48{0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff},
		},
		"mac/96": {
			"aa:bb:cc:dd:ee:ff",
			96,
			ipnet.ErrIP6SubnetTooSmall,
			nil,
		},
		"-1":         {"", -1, ipnet.ErrInvalidPrefixLength, nil},
		"ip4":        {"1.1.1.1", 40, ipnet.ErrNotHostID, nil},
		"ill-formed": {"1:1:1:1", 40, ipnet.ErrNotHostID, nil},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			hostID, err := ipnet.ParseHost(tc.input, tc.prefixLen)
			require.Equal(t, tc.hostID, hostID)
			require.ErrorIs(t, err, tc.err)
		})
	}
}
