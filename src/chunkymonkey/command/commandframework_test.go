package command

import (
	"testing"

	"gomock.googlecode.com/hg/gomock"

	"chunkymonkey/gamerules"
	"testmatcher"
)

func TestCommandFramework(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGame := gamerules.NewMockIGame(mockCtrl)

	cf := NewCommandFramework("/")

	mockGame.EXPECT().BroadcastMessage("§dthis is a broadcast")
	cf.Process("thePlayer", "/say this is a broadcast", mockGame)

	mockGame.EXPECT().IsValidPlayerName("thePlayer").Return(true)
	mockGame.EXPECT().IsValidItemId(1).Return(true)
	mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Giving 64 of 1 to thePlayer")
	mockGame.EXPECT().GiveItem("thePlayer", 1, 64, 0)
	cf.Process("thePlayer", "/give thePlayer 1 64", mockGame)

	mockGame.EXPECT().IsValidPlayerName("otherPlayer").Return(false)
	mockGame.EXPECT().SendMessageToPlayer("thePlayer", "'otherPlayer' is not logged in")
	cf.Process("thePlayer", "/give otherPlayer 1 64", mockGame)

	mockGame.EXPECT().IsValidPlayerName("otherPlayer").Return(true)
	mockGame.EXPECT().IsValidItemId(1).Return(false)
	mockGame.EXPECT().SendMessageToPlayer("thePlayer", "'1' is not a valid item id")
	cf.Process("thePlayer", "/give otherPlayer 1 64", mockGame)

	mockGame.EXPECT().IsValidPlayerName("otherPlayer").Return(true)
	mockGame.EXPECT().IsValidItemId(1).Return(true)
	mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Cannot give more than 512 items at once")
	cf.Process("thePlayer", "/give otherPlayer 1 513", mockGame)

	mockGame.EXPECT().SendMessageToPlayer("thePlayer", &testmatcher.StringPrefix{"Commands:"})
	cf.Process("thePlayer", "/help", mockGame)

	gomock.InOrder(
		mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Command: /help"),
		mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Usage: help|?"),
		mockGame.EXPECT().SendMessageToPlayer("thePlayer", "Description: Shows a list of all commands."),
	)
	cf.Process("thePlayer", "/help help", mockGame)
}
