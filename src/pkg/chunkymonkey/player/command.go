package player

import (
	"bytes"
	"regexp"
	"strconv"

	"chunkymonkey/proto"
	"chunkymonkey/slot"
	. "chunkymonkey/types"
)

var cmdPartRegexp = regexp.MustCompile("[^ ]+")

func runCommand(player *Player, command string) {
	cmdParts := cmdPartRegexp.FindAllString(command, -1)
	// TODO Permissions.
	var msg string
	switch cmdParts[0] {
	case "give":
		msg = cmdGive(player, cmdParts[1:])
	}

	if msg != "" {
		buf := new(bytes.Buffer)
		proto.WriteChatMessage(buf, msg)
		player.TransmitPacket(buf.Bytes())
	}
}

const giveUsage = "/give <item ID> [<quantity> [<data>]]"

func cmdGive(player *Player, cmdParts []string) string {
	if len(cmdParts) < 1 || len(cmdParts) > 3 {
		return giveUsage
	}

	// TODO First argument should be player to receive item. Right now it just
	// gives it to the current player.
	itemId, err := strconv.Atoi(cmdParts[0])
	if err != nil {
		return giveUsage
	}

	quantity := 1
	if len(cmdParts) >= 2 {
		quantity, err = strconv.Atoi(cmdParts[1])
		if err != nil {
			return giveUsage
		}
	}

	data := 0
	if len(cmdParts) >= 3 {
		data, err = strconv.Atoi(cmdParts[2])
		if err != nil {
			return giveUsage
		}
	}

	itemType, ok := player.gameRules.ItemTypes[ItemTypeId(itemId)]
	if !ok {
		return "Unknown item ID"
	}

	item := slot.Slot{
		ItemType: itemType,
		Count:    ItemCount(quantity),
		Data:     ItemData(data),
	}

	player.reqGiveItem(&player.position, &item)

	return ""
}
