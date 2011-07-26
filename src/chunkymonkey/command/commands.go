package command

import (
	"fmt"
	"strconv"
	"strings"

	"chunkymonkey/gamerules"
)

func getCommands() map[string]*Command {
	cmds := map[string]*Command{}
	cmds[sayCmd] = NewCommand(sayCmd, sayDesc, sayUsage, cmdSay)
	cmds[tpCmd] = NewCommand(tpCmd, tpDesc, tpUsage, cmdTp)
	cmds[killCmd] = NewCommand(killCmd, killDesc, killUsage, cmdKill)
	cmds[tellCmd] = NewCommand(tellCmd, tellDesc, tellUsage, cmdTell)
	cmds[giveCmd] = NewCommand(giveCmd, giveDesc, giveUsage, cmdGive)
	return cmds
}

const msgNotImplemented = "We are sorry. This command is not yet implemented."
const msgUnknownItem = "Unknown item ID"
// say message
const sayCmd = "say"
const sayUsage = "say <message>"
const sayDesc = "Broadcasts a message to all players without showing a player name. The message is colored pink."

func cmdSay(player, message string, cmdHandler gamerules.IGame) {
	args := strings.Split(message, " ", -1)
	if len(args) < 2 {
		cmdHandler.SendMessageToPlayer(player, sayUsage)
		return
	}
	msg := strings.Join(args[1:], " ")
	cmdHandler.BroadcastMessage("§d" + msg)
}

// tp player1 player2

const tpCmd = "tp"
const tpUsage = "tp <player1> <player2>"
const tpDesc = "Teleports player1 to player2."

func cmdTp(player, message string, cmdHandler gamerules.IGame) {
	args := strings.Split(message, " ", -1)
	if len(args) < 3 {
		cmdHandler.SendMessageToPlayer(player, tpUsage)
		return
	}

	teleportee := args[1]
	destination := args[2]
	if !cmdHandler.IsValidPlayerName(teleportee) {
		msg := fmt.Sprintf("'%s' is not logged in", teleportee)
		cmdHandler.SendMessageToPlayer(player, msg)
		return
	}
	if !cmdHandler.IsValidPlayerName(destination) {
		msg := fmt.Sprintf("'%s' is not logged in", destination)
		cmdHandler.SendMessageToPlayer(player, msg)
		return
	}

	cmdHandler.TeleportToPlayer(teleportee, destination)
}

// /kill
const killCmd = "kill"
const killUsage = "kill"
const killDesc = "Inflicts damage to self. Useful when lost or stuck."

func cmdKill(player, message string, cmdHandler gamerules.IGame) {
	// TODO inflict damage to player
	cmdHandler.SendMessageToPlayer(player, msgNotImplemented)
}

// /tell player message
const tellCmd = "tell"
const tellUsage = "tell <player> <message>"
const tellDesc = "Tells a player a message."

func cmdTell(player, message string, cmdHandler gamerules.IGame) {
	args := strings.Split(message, " ", -1)
	if len(args) < 3 {
		cmdHandler.SendMessageToPlayer(player, tellUsage)
		return
	}
	/* TODO Get player to send message, too
	player := args[1]
	message := strings.Join(args[2:], " ")
	*/
	cmdHandler.SendMessageToPlayer(player, msgNotImplemented)
}

const helpShortCmd = "?"
const helpCmd = "help"
const helpUsage = "help|?"
const helpDesc = "Shows a list of all commands."
const msgUnknownCommand = "Command not available."

func cmdHelp(player, message string, cmdFramework *CommandFramework, cmdHandler gamerules.IGame) {
	args := strings.Split(message, " ", -1)
	if len(args) > 2 {
		cmdHandler.SendMessageToPlayer(player, helpUsage)
		return
	}
	cmds := cmdFramework.Commands()
	if len(args) == 2 {
		cmd := args[1]
		if command, ok := cmds[cmd]; ok {
			cmdHandler.SendMessageToPlayer(player, "Command: "+cmdFramework.Prefix()+command.Trigger)
			cmdHandler.SendMessageToPlayer(player, "Usage: "+command.Usage)
			cmdHandler.SendMessageToPlayer(player, "Description: "+command.Description)
			return
		}
		cmdHandler.SendMessageToPlayer(player, msgUnknownCommand)
		return
	}
	var resp string
	if len(cmds) == 0 {
		resp = "No commands available."
	} else {
		resp = "Commands:"
		for trigger, _ := range cmds {
			resp += " " + trigger + ","
		}
		resp = resp[:len(resp)-1]
	}
	cmdHandler.SendMessageToPlayer(player, resp)
}

const giveCmd = "give"
const giveUsage = "give <player> <item ID> [<quantity> [<data>]]"
const giveDesc = "Gives x amount of y items to player."

func cmdGive(player, message string, cmdHandler gamerules.IGame) {
	args := strings.Split(message, " ", -1)
	if len(args) < 3 || len(args) > 5 {
		cmdHandler.SendMessageToPlayer(player, giveUsage)
		return
	}
	args = args[1:]

	// Check to make sure this is a valid player name (at the moment command
	// is being run, the player may log off before command completes).
	target := args[0]
	if !cmdHandler.IsValidPlayerName(target) {
		msg := fmt.Sprintf("'%s' is not logged in", target)
		cmdHandler.SendMessageToPlayer(player, msg)
		return
	}

	itemId, err := strconv.Atoi(args[1])
	if err != nil || !cmdHandler.IsValidItemId(itemId) {
		msg := fmt.Sprintf("'%s' is not a valid item id", args[1])
		cmdHandler.SendMessageToPlayer(player, msg)
		return
	}

	quantity := 1
	if len(args) >= 3 {
		quantity, err = strconv.Atoi(args[2])
		if err != nil {
			cmdHandler.SendMessageToPlayer(player, giveUsage)
			return
		}

		if quantity > 512 {
			msg := "Cannot give more than 512 items at once"
			cmdHandler.SendMessageToPlayer(player, msg)
			return
		}
	}

	data := 0
	if len(args) >= 4 {
		data, err = strconv.Atoi(args[2])
		if err != nil {
			cmdHandler.SendMessageToPlayer(player, giveUsage)
			return
		}
	}

	// TODO: How can we get the name of this item without a dependency loop?
	msg := fmt.Sprintf("Giving %d of %d to %s", quantity, itemId, target)
	cmdHandler.SendMessageToPlayer(player, msg)
	cmdHandler.GiveItem(target, itemId, quantity, data)
	if player != target {
		msg = fmt.Sprintf("%s gave you %d of %d", player, quantity, itemId)
		cmdHandler.SendMessageToPlayer(target, msg)
	}
}
