package irc

import (
	"strconv"
	"strings"
	"time"
)

// parseMessage parses a raw IRC message into a Message struct.
// Format: [@tags] [:prefix] COMMAND [params...] [:trailing]
func parseMessage(raw string) *Message {
	msg := &Message{
		Raw:  raw,
		Tags: make(map[string]string),
	}

	raw = strings.TrimSpace(raw)
	if raw == "" {
		return msg
	}

	pos := 0

	// Parse tags (starts with @)
	if raw[pos] == '@' {
		end := strings.Index(raw, " ")
		if end == -1 {
			return msg
		}
		msg.Tags = parseTags(raw[1:end])
		pos = end + 1
	}

	// Skip whitespace
	for pos < len(raw) && raw[pos] == ' ' {
		pos++
	}

	// Parse prefix (starts with :)
	if pos < len(raw) && raw[pos] == ':' {
		end := strings.Index(raw[pos:], " ")
		if end == -1 {
			msg.Prefix = raw[pos+1:]
			return msg
		}
		msg.Prefix = raw[pos+1 : pos+end]
		pos = pos + end + 1
	}

	// Skip whitespace
	for pos < len(raw) && raw[pos] == ' ' {
		pos++
	}

	// Parse command
	end := strings.Index(raw[pos:], " ")
	if end == -1 {
		msg.Command = raw[pos:]
		return msg
	}
	msg.Command = raw[pos : pos+end]
	pos = pos + end + 1

	// Skip whitespace
	for pos < len(raw) && raw[pos] == ' ' {
		pos++
	}

	// Parse params
	for pos < len(raw) {
		if raw[pos] == ':' {
			// Trailing parameter (rest of message)
			msg.Trailing = raw[pos+1:]
			break
		}

		end := strings.Index(raw[pos:], " ")
		if end == -1 {
			msg.Params = append(msg.Params, raw[pos:])
			break
		}
		msg.Params = append(msg.Params, raw[pos:pos+end])
		pos = pos + end + 1

		// Skip whitespace
		for pos < len(raw) && raw[pos] == ' ' {
			pos++
		}
	}

	return msg
}

// parseTags parses IRCv3 tags from a tag string.
func parseTags(tagStr string) map[string]string {
	tags := make(map[string]string)
	if tagStr == "" {
		return tags
	}

	pairs := strings.Split(tagStr, ";")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		eqIdx := strings.Index(pair, "=")
		if eqIdx == -1 {
			tags[pair] = ""
			continue
		}
		key := pair[:eqIdx]
		value := unescapeTagValue(pair[eqIdx+1:])
		tags[key] = value
	}

	return tags
}

// unescapeTagValue unescapes IRC tag values.
func unescapeTagValue(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '\\' {
			switch s[i+1] {
			case ':':
				result.WriteByte(';')
			case 's':
				result.WriteByte(' ')
			case '\\':
				result.WriteByte('\\')
			case 'r':
				result.WriteByte('\r')
			case 'n':
				result.WriteByte('\n')
			default:
				result.WriteByte(s[i+1])
			}
			i += 2
		} else {
			result.WriteByte(s[i])
			i++
		}
	}

	return result.String()
}

// parseEmotes parses the emotes tag into a slice of Emote.
// Format: emote_id:start-end,start-end/emote_id:start-end
func parseEmotes(emoteStr string) []Emote {
	if emoteStr == "" {
		return nil
	}

	var emotes []Emote
	emoteParts := strings.Split(emoteStr, "/")

	for _, part := range emoteParts {
		if part == "" {
			continue
		}

		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			continue
		}

		emoteID := part[:colonIdx]
		positionsStr := part[colonIdx+1:]
		positions := strings.Split(positionsStr, ",")

		for _, posStr := range positions {
			dashIdx := strings.Index(posStr, "-")
			if dashIdx == -1 {
				continue
			}

			start, err1 := strconv.Atoi(posStr[:dashIdx])
			end, err2 := strconv.Atoi(posStr[dashIdx+1:])
			if err1 != nil || err2 != nil {
				continue
			}

			emotes = append(emotes, Emote{
				ID:    emoteID,
				Start: start,
				End:   end,
				Count: 1,
			})
		}
	}

	return emotes
}

// parseBadges parses the badges tag into a map.
// Format: badge/version,badge/version
func parseBadges(badgeStr string) map[string]string {
	badges := make(map[string]string)
	if badgeStr == "" {
		return badges
	}

	parts := strings.Split(badgeStr, ",")
	for _, part := range parts {
		if part == "" {
			continue
		}
		slashIdx := strings.Index(part, "/")
		if slashIdx == -1 {
			badges[part] = ""
			continue
		}
		badges[part[:slashIdx]] = part[slashIdx+1:]
	}

	return badges
}

// parseTimestamp parses the tmi-sent-ts tag into a time.Time.
func parseTimestamp(ts string) time.Time {
	if ts == "" {
		return time.Now()
	}
	ms, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.UnixMilli(ms)
}

// parseBool parses a "0" or "1" string into a bool.
func parseBool(s string) bool {
	return s == "1"
}

// parseInt parses a string into an int, returning 0 on error.
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// parseChannel removes the # prefix from a channel name.
func parseChannel(s string) string {
	return strings.TrimPrefix(s, "#")
}

// parseUserFromPrefix extracts the username from a prefix.
// Format: nick!user@host
func parseUserFromPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}
	bangIdx := strings.Index(prefix, "!")
	if bangIdx == -1 {
		return prefix
	}
	return prefix[:bangIdx]
}

// parseChatMessage converts a Message into a ChatMessage.
func parseChatMessage(msg *Message) *ChatMessage {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	badges := parseBadges(msg.Tags["badges"])

	return &ChatMessage{
		ID:            msg.Tags["id"],
		Channel:       channel,
		User:          msg.Tags["login"],
		UserID:        msg.Tags["user-id"],
		Message:       msg.Trailing,
		Emotes:        parseEmotes(msg.Tags["emotes"]),
		Badges:        badges,
		BadgeInfo:     parseBadges(msg.Tags["badge-info"]),
		Color:         msg.Tags["color"],
		DisplayName:   msg.Tags["display-name"],
		IsMod:         parseBool(msg.Tags["mod"]),
		IsVIP:         badges["vip"] != "",
		IsSubscriber:  parseBool(msg.Tags["subscriber"]),
		IsBroadcaster: badges["broadcaster"] != "",
		Bits:          parseInt(msg.Tags["bits"]),
		FirstMessage:  parseBool(msg.Tags["first-msg"]),
		ReturningChatter: parseBool(msg.Tags["returning-chatter"]),
		ReplyParentMsgID:       msg.Tags["reply-parent-msg-id"],
		ReplyParentUserID:      msg.Tags["reply-parent-user-id"],
		ReplyParentUserLogin:   msg.Tags["reply-parent-user-login"],
		ReplyParentDisplayName: msg.Tags["reply-parent-display-name"],
		ReplyParentMsgBody:     msg.Tags["reply-parent-msg-body"],
		Timestamp:     parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:           msg.Raw,
	}
}

// parseUserNotice converts a Message into a UserNotice.
func parseUserNotice(msg *Message) *UserNotice {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	// Extract msg-param-* tags
	msgParams := make(map[string]string)
	for key, value := range msg.Tags {
		if strings.HasPrefix(key, "msg-param-") {
			paramName := strings.TrimPrefix(key, "msg-param-")
			msgParams[paramName] = value
		}
	}

	return &UserNotice{
		Type:          msg.Tags["msg-id"],
		Channel:       channel,
		User:          msg.Tags["login"],
		UserID:        msg.Tags["user-id"],
		DisplayName:   msg.Tags["display-name"],
		Message:       msg.Trailing,
		SystemMessage: msg.Tags["system-msg"],
		MsgParams:     msgParams,
		Badges:        parseBadges(msg.Tags["badges"]),
		BadgeInfo:     parseBadges(msg.Tags["badge-info"]),
		Color:         msg.Tags["color"],
		Emotes:        parseEmotes(msg.Tags["emotes"]),
		Timestamp:     parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:           msg.Raw,
	}
}

// parseRoomState converts a Message into a RoomState.
func parseRoomState(msg *Message) *RoomState {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	followersOnly := -1
	if fo, ok := msg.Tags["followers-only"]; ok {
		followersOnly = parseInt(fo)
	}

	return &RoomState{
		Channel:       channel,
		EmoteOnly:     parseBool(msg.Tags["emote-only"]),
		FollowersOnly: followersOnly,
		R9K:           parseBool(msg.Tags["r9k"]),
		Slow:          parseInt(msg.Tags["slow"]),
		SubsOnly:      parseBool(msg.Tags["subs-only"]),
		RoomID:        msg.Tags["room-id"],
		Raw:           msg.Raw,
	}
}

// parseNotice converts a Message into a Notice.
func parseNotice(msg *Message) *Notice {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	return &Notice{
		Channel: channel,
		Message: msg.Trailing,
		MsgID:   msg.Tags["msg-id"],
		Raw:     msg.Raw,
	}
}

// parseClearChat converts a Message into a ClearChat.
func parseClearChat(msg *Message) *ClearChat {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	return &ClearChat{
		Channel:      channel,
		User:         msg.Trailing,
		BanDuration:  parseInt(msg.Tags["ban-duration"]),
		RoomID:       msg.Tags["room-id"],
		TargetUserID: msg.Tags["target-user-id"],
		Timestamp:    parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:          msg.Raw,
	}
}

// parseClearMessage converts a Message into a ClearMessage.
func parseClearMessage(msg *Message) *ClearMessage {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	return &ClearMessage{
		Channel:     channel,
		User:        msg.Tags["login"],
		Message:     msg.Trailing,
		TargetMsgID: msg.Tags["target-msg-id"],
		Timestamp:   parseTimestamp(msg.Tags["tmi-sent-ts"]),
		Raw:         msg.Raw,
	}
}

// parseWhisper converts a Message into a Whisper.
func parseWhisper(msg *Message) *Whisper {
	to := ""
	if len(msg.Params) > 0 {
		to = msg.Params[0]
	}

	return &Whisper{
		From:        parseUserFromPrefix(msg.Prefix),
		FromID:      msg.Tags["user-id"],
		To:          to,
		Message:     msg.Trailing,
		DisplayName: msg.Tags["display-name"],
		Color:       msg.Tags["color"],
		Badges:      parseBadges(msg.Tags["badges"]),
		Emotes:      parseEmotes(msg.Tags["emotes"]),
		MessageID:   msg.Tags["message-id"],
		ThreadID:    msg.Tags["thread-id"],
		Raw:         msg.Raw,
	}
}

// parseGlobalUserState converts a Message into a GlobalUserState.
func parseGlobalUserState(msg *Message) *GlobalUserState {
	emoteSets := []string{}
	if es := msg.Tags["emote-sets"]; es != "" {
		emoteSets = strings.Split(es, ",")
	}

	return &GlobalUserState{
		UserID:      msg.Tags["user-id"],
		DisplayName: msg.Tags["display-name"],
		Color:       msg.Tags["color"],
		Badges:      parseBadges(msg.Tags["badges"]),
		BadgeInfo:   parseBadges(msg.Tags["badge-info"]),
		EmoteSets:   emoteSets,
		Raw:         msg.Raw,
	}
}

// parseUserState converts a Message into a UserState.
func parseUserState(msg *Message) *UserState {
	channel := ""
	if len(msg.Params) > 0 {
		channel = parseChannel(msg.Params[0])
	}

	emoteSets := []string{}
	if es := msg.Tags["emote-sets"]; es != "" {
		emoteSets = strings.Split(es, ",")
	}

	badges := parseBadges(msg.Tags["badges"])

	return &UserState{
		Channel:      channel,
		DisplayName:  msg.Tags["display-name"],
		Color:        msg.Tags["color"],
		Badges:       badges,
		BadgeInfo:    parseBadges(msg.Tags["badge-info"]),
		EmoteSets:    emoteSets,
		IsMod:        parseBool(msg.Tags["mod"]),
		IsSubscriber: parseBool(msg.Tags["subscriber"]),
		Raw:          msg.Raw,
	}
}
