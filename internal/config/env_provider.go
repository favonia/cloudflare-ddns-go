package config

import (
	"github.com/favonia/cloudflare-ddns/internal/ipnet"
	"github.com/favonia/cloudflare-ddns/internal/pp"
	"github.com/favonia/cloudflare-ddns/internal/provider"
)

// ReadProvider reads an environment variable and parses it as a provider.
//
// policyKey was the name of the deprecated parameters IP4/6_POLICY.
// use1001 indicates whether 1.0.0.1 should be used instead of 1.1.1.1.
//
//nolint:funlen
func ReadProvider(ppfmt pp.PP, use1001 bool, key, keyDeprecated string, field *provider.Provider) bool {
	if val := Getenv(key); val == "" {
		// parsing of the deprecated parameter
		switch valPolicy := Getenv(keyDeprecated); valPolicy {
		case "":
			ppfmt.Infof(pp.EmojiBullet, "Use default %s=%s", key, provider.Name(*field))
			return true
		case "cloudflare":
			ppfmt.Warningf(
				pp.EmojiUserWarning,
				`%s=cloudflare is deprecated; use %s=cloudflare.trace or %s=cloudflare.doh`,
				keyDeprecated, key, key,
			)
			*field = provider.NewCloudflareTrace(use1001)
			return true
		case "cloudflare.trace":
			ppfmt.Warningf(
				pp.EmojiUserWarning,
				`%s is deprecated; use %s=%s`,
				keyDeprecated, key, valPolicy,
			)
			*field = provider.NewCloudflareTrace(use1001)
			return true
		case "cloudflare.doh":
			ppfmt.Warningf(
				pp.EmojiUserWarning,
				`%s is deprecated; use %s=%s`,
				keyDeprecated, key, valPolicy,
			)
			*field = provider.NewCloudflareDOH(use1001)
			return true
		case "ipify":
			ppfmt.Warningf(
				pp.EmojiUserWarning,
				`%s=ipify is deprecated; use %s=cloudflare.trace or %s=cloudflare.doh`,
				keyDeprecated, key, key,
			)
			*field = provider.NewIpify()
			return true
		case "local":
			ppfmt.Warningf(
				pp.EmojiUserWarning,
				`%s is deprecated; use %s=%s`,
				keyDeprecated, key, valPolicy,
			)
			*field = provider.NewLocal(use1001)
			return true
		case "unmanaged":
			ppfmt.Warningf(
				pp.EmojiUserWarning,
				`%s is deprecated; use %s=none`,
				keyDeprecated, key,
			)
			*field = nil
			return true
		default:
			ppfmt.Errorf(pp.EmojiUserError, "%s (%q) is not a valid provider", keyDeprecated, valPolicy)
			return false
		}
	} else {
		if Getenv(keyDeprecated) != "" {
			ppfmt.Errorf(
				pp.EmojiUserError,
				`Cannot have both %s and %s set`,
				key, keyDeprecated,
			)
			return false
		}

		switch val {
		case "cloudflare":
			ppfmt.Errorf(
				pp.EmojiUserError,
				`%s=cloudflare is invalid; use %s=cloudflare.trace or %s=cloudflare.doh`,
				key, key, key,
			)
			return false
		case "cloudflare.trace":
			*field = provider.NewCloudflareTrace(use1001)
			return true
		case "cloudflare.doh":
			*field = provider.NewCloudflareDOH(use1001)
			return true
		case "ipify":
			ppfmt.Warningf(
				pp.EmojiUserWarning,
				`%s=ipify is deprecated; use %s=cloudflare.trace or %s=cloudflare.doh`,
				key, key, key,
			)
			*field = provider.NewIpify()
			return true
		case "local":
			*field = provider.NewLocal(use1001)
			return true
		case "none":
			*field = nil
			return true
		default:
			ppfmt.Errorf(pp.EmojiUserError, "%s (%q) is not a valid provider", key, val)
			return false
		}
	}
}

// ReadProviderMap reads the environment variables IP4_PROVIDER and IP6_PROVIDER,
// with support of deprecated environment variables IP4_POLICY and IP6_POLICY.
func ReadProviderMap(ppfmt pp.PP, use1001 bool, field *map[ipnet.Type]provider.Provider) bool {
	ip4Provider := (*field)[ipnet.IP4]
	ip6Provider := (*field)[ipnet.IP6]

	if !ReadProvider(ppfmt, use1001, "IP4_PROVIDER", "IP4_POLICY", &ip4Provider) ||
		!ReadProvider(ppfmt, use1001, "IP6_PROVIDER", "IP6_POLICY", &ip6Provider) {
		return false
	}

	*field = map[ipnet.Type]provider.Provider{
		ipnet.IP4: ip4Provider,
		ipnet.IP6: ip6Provider,
	}
	return true
}
