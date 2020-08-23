package gobayeux

import "strings"

// Channel represents a Bayeux Channel which is defined as "a string that
// looks like a URL path such as `/foo/bar`, `/meta/connect`, or
// `/service/chat`."
//
// See also: https://docs.cometd.org/current/reference/#_concepts_channels
type Channel string

const (
	// MetaHandshake is the Channel for the first message a new client sends.
	MetaHandshake Channel = "/meta/handshake"
	// MetaConnect is the Channel used for connect messages after a successful
	// handshake.
	MetaConnect Channel = "/meta/connect"
	// MetaDisconnect is the Channel used for disconnect messages.
	MetaDisconnect Channel = "/meta/disconnect"
	// MetaSubscribe is the Channel used by a client to subscribe to channels.
	MetaSubscribe Channel = "/meta/subscribe"
	// MetaUnsubscribe is the Channel used by a client to unsubscribe to
	// channels.
	MetaUnsubscribe Channel = "/meta/unsubscribe"
	emptyChannel    Channel = ""
)

// TODO: Determine if supporting parameters in channels would be useful:
// https://docs.cometd.org/current/reference/#_concepts_channels_parameters

// ChannelType is used to define the three types of channels:
// - meta channels, channels starting with `/meta/`
// - service channels, channels starting with `/service/`
// - broadcast channels, all other channels
type ChannelType string

const (
	// MetaChannel represents the `/meta/` channel type
	MetaChannel ChannelType = "meta"
	// ServiceChannel represents the `/service/` channel type
	ServiceChannel ChannelType = "service"
	// BroadcastChannel represents all other channels
	BroadcastChannel ChannelType = "broadcast"
)

const (
	metaPrefix    string = "/meta/"
	servicePrefix string = "/service/"
)

// Type provides the type of Channel this struct represents
func (c Channel) Type() ChannelType {
	s := string(c)
	switch {
	case strings.HasPrefix(s, metaPrefix):
		return MetaChannel
	case strings.HasPrefix(s, servicePrefix):
		return ServiceChannel
	default:
		return BroadcastChannel
	}
}

// HasWildcard indicates whether the Channel ends with * or **
//
// See also: https://docs.cometd.org/current/reference/#_concepts_channels_wild
func (c Channel) HasWildcard() bool {
	s := string(c)
	return strings.HasSuffix(s, "*")
}

// IsValid does its best to check the validity of a Channel
func (c Channel) IsValid() bool {
	s := string(c)
	if strings.Contains(s, "*") && !c.HasWildcard() {
		return false
	}

	if !strings.HasPrefix(s, "/") {
		return false
	}

	return true
}

// Match checks if a given Channel matches this Channel.
// Note wildcards are only valid after the last /.
//
// See also: https://docs.cometd.org/current/reference/#_concepts_channels_wild
func (c Channel) Match(other Channel) bool {
	return c.MatchString(string(other))
}

// MatchString checks if a given string matches this Channel.
// Note wildcards are only valid after the last /.
//
// See also: https://docs.cometd.org/current/reference/#_concepts_channels_wild
func (c Channel) MatchString(other string) bool {
	if c.HasWildcard() {
		return c.matchAgainstWildcards(other)
	}
	return string(c) == other
}

func (c Channel) matchAgainstWildcards(other string) bool {
	// Preconditions: If we're here, it's because we've detected a valid
	// wildcard at the end of this string. Let's do two things, determine how
	// many wildcards are in this Channel and then do our best to match it
	// against other.
	self := string(c)
	index := strings.LastIndexByte(self, '/')
	if index == -1 {
		return false
	}
	prefix := self[:index]
	if !strings.HasPrefix(other, prefix) {
		// If other doesn't start with our prefix then let's just bail now
		return false
	}

	// At this point, our Channel and our other channel have the same prefix.
	// It's also important to note that index above is thus also the length of
	// the prefix
	wildcards := self[index+1:]
	startMatchingWildcards := other[index+1:]

	switch wildcards {
	case "*":
		return strings.Count(startMatchingWildcards, "/") == 0
	case "**":
		return true
	default:
		// Assuming the end of our current channel has something other than *
		// or ** then we can't match it. I don't think we can ever hit this
		// case, but it may as well be here to be safe
		return false
	}
}
