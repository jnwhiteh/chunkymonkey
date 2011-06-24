package player

import (
	"bytes"
	"expvar"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"sync"

	"chunkymonkey/gamerules"
	"chunkymonkey/proto"
	"chunkymonkey/slot"
	"chunkymonkey/stub"
	. "chunkymonkey/types"
	"chunkymonkey/window"
)

var (
	expVarPlayerConnectionCount    *expvar.Int
	expVarPlayerDisconnectionCount *expvar.Int
)

const StanceNormal = 1.62

func init() {
	expVarPlayerConnectionCount = expvar.NewInt("player-connection-count")
	expVarPlayerDisconnectionCount = expvar.NewInt("player-disconnection-count")
}

type Player struct {
	EntityId
	shardReceiver  shardPlayerClient
	shardConnecter stub.IShardConnecter
	conn           net.Conn
	name           string
	position       AbsXyz
	look           LookDegrees
	chunkSubs      chunkSubscriptions

	cursor       slot.Slot // Item being moved by mouse cursor.
	inventory    window.PlayerInventory
	curWindow    window.IWindow
	nextWindowId WindowId
	remoteInv    *RemoteInventory

	gameRules *gamerules.GameRules

	mainQueue chan func(*Player)
	txQueue   chan []byte

	// TODO remove this lock, packet handling shouldn't use a lock, it should use
	// a channel instead (ideally).
	lock sync.Mutex

	onDisconnect chan<- EntityId
}

func NewPlayer(entityId EntityId, shardConnecter stub.IShardConnecter, gameRules *gamerules.GameRules, conn net.Conn, name string, position AbsXyz, onDisconnect chan<- EntityId) *Player {
	player := &Player{
		EntityId:       entityId,
		shardConnecter: shardConnecter,
		conn:           conn,
		name:           name,
		position:       position,
		look:           LookDegrees{0, 0},

		curWindow:    nil,
		nextWindowId: WindowIdFreeMin,

		gameRules: gameRules,

		mainQueue: make(chan func(*Player), 128),
		txQueue:   make(chan []byte, 128),

		onDisconnect: onDisconnect,
	}

	player.shardReceiver.Init(player)
	player.cursor.Init()
	player.inventory.Init(player.EntityId, player, gameRules.Recipes)

	return player
}

func (player *Player) getHeldItemTypeId() ItemTypeId {
	heldSlot, _ := player.inventory.HeldItem()
	heldItemId := heldSlot.GetItemTypeId()
	if heldItemId < 0 {
		return 0
	}
	return heldItemId
}

func (player *Player) Start() {
	go player.receiveLoop()
	go player.transmitLoop()
	go player.mainLoop()
}

// Start of packet handling code
// Note: any packet handlers that could change the player state or read a
// changeable state must use player.lock

func (player *Player) PacketKeepAlive() {
}

func (player *Player) PacketChatMessage(message string) {
	if message[0] == '/' {
		runCommand(player, message[1:])
	} else {
		player.sendChatMessage(message)
	}
}

func (player *Player) PacketEntityAction(entityId EntityId, action EntityAction) {
}

func (player *Player) PacketUseEntity(user EntityId, target EntityId, leftClick bool) {
}

func (player *Player) PacketRespawn(dimension DimensionId) {
}

func (player *Player) PacketPlayer(onGround bool) {
}

func (player *Player) PacketPlayerPosition(position *AbsXyz, stance AbsCoord, onGround bool) {
	player.lock.Lock()
	defer player.lock.Unlock()

	var delta = AbsXyz{position.X - player.position.X,
		position.Y - player.position.Y,
		position.Z - player.position.Z}
	distance := math.Sqrt(float64(delta.X*delta.X + delta.Y*delta.Y + delta.Z*delta.Z))
	if distance > 10 {
		log.Printf("Discarding player position that is too far removed (%.2f, %.2f, %.2f)",
			position.X, position.Y, position.Z)
		return
	}
	player.position = *position
	player.chunkSubs.Move(position)

	// TODO: Should keep track of when players enter/leave their mutual radius
	// of "awareness". I.e a client should receive a RemoveEntity packet when
	// the player walks out of range, and no longer receive WriteEntityTeleport
	// packets for them. The converse should happen when players come in range
	// of each other.
}

func (player *Player) PacketPlayerLook(look *LookDegrees, onGround bool) {
	player.lock.Lock()
	defer player.lock.Unlock()

	// TODO input validation
	player.look = *look

	buf := new(bytes.Buffer)
	proto.WriteEntityLook(buf, player.EntityId, look.ToLookBytes())

	// TODO update playerData on current chunk

	player.chunkSubs.curShard.ReqMulticastPlayers(
		player.chunkSubs.curChunkLoc,
		player.EntityId,
		buf.Bytes(),
	)
}

func (player *Player) PacketPlayerBlockHit(status DigStatus, target *BlockXyz, face Face) {
	player.lock.Lock()
	defer player.lock.Unlock()

	// TODO validate that the player is actually somewhere near the block

	// TODO measure the dig time on the target block and relay to the shard to
	// stop speed hacking (based on block type and tool used - non-trivial).

	shardClient, _, ok := player.chunkSubs.ShardClientForBlockXyz(target)
	if ok {
		held, _ := player.inventory.HeldItem()
		shardClient.ReqHitBlock(held, *target, status, face)
	}
}

func (player *Player) PacketPlayerBlockInteract(itemId ItemTypeId, target *BlockXyz, face Face, amount ItemCount, uses ItemData) {
	if face < FaceMinValid || face > FaceMaxValid {
		// TODO sometimes FaceNull means something. This case should be covered.
		log.Printf("Player/PacketPlayerBlockInteract: invalid face %d", face)
		return
	}

	player.lock.Lock()
	defer player.lock.Unlock()

	shardClient, _, ok := player.chunkSubs.ShardClientForBlockXyz(target)
	if ok {
		held, _ := player.inventory.HeldItem()
		shardClient.ReqInteractBlock(held, *target, face)
	}
}

func (player *Player) PacketHoldingChange(slotId SlotId) {
	player.lock.Lock()
	defer player.lock.Unlock()
	player.inventory.SetHolding(slotId)
}

func (player *Player) PacketEntityAnimation(entityId EntityId, animation EntityAnimation) {
}

func (player *Player) PacketUnknown0x1b(field1, field2 float32, field3, field4 bool, field5, field6 float32) {
	log.Printf(
		"PacketUnknown0x1b(field1=%v, field2=%v, field3=%t, field4=%t, field5=%v, field6=%v)",
		field1, field2, field3, field4, field5, field6)
}

func (player *Player) PacketUnknown0x3d(field1, field2 int32, field3 int8, field4, field5 int32) {
	// TODO Remove this method if it's S->C only.
	log.Printf(
		"PacketUnknown0x3d(field1=%d, field2=%d, field3=%d, field4=%d, field5=%d)",
		field1, field2, field3, field4, field5)
}

func (player *Player) PacketWindowClose(windowId WindowId) {
	player.lock.Lock()
	defer player.lock.Unlock()

	player.closeCurrentWindow(false)
}

func (player *Player) PacketWindowClick(windowId WindowId, slotId SlotId, rightClick bool, txId TxId, shiftClick bool, expectedSlot *proto.WindowSlot) {
	player.lock.Lock()
	defer player.lock.Unlock()

	// Note that the expectedSlot parameter is currently ignored. The item(s)
	// involved are worked out from the server-side data.
	// TODO use the expectedSlot as a conditions for the click, and base the
	// transaction result on that.

	// Determine which inventory window is involved.
	// TODO support for more windows

	var clickedWindow window.IWindow
	if windowId == WindowIdInventory {
		clickedWindow = &player.inventory
	} else if player.curWindow != nil && player.curWindow.GetWindowId() == windowId {
		clickedWindow = player.curWindow
	} else {
		log.Printf(
			"Warning: ignored window click on unknown window ID %d",
			windowId)
	}

	expectedItemType, ok := player.gameRules.ItemTypes[expectedSlot.ItemTypeId]
	if !ok {
		return
	}

	expectedSlotContent := &slot.Slot{
		ItemType: expectedItemType,
		Count:    expectedSlot.Count,
		Data:     expectedSlot.Data,
	}
	// The client tends to send item IDs even when the count is zero.
	expectedSlotContent.Normalize()

	txState := TxStateRejected

	if clickedWindow != nil {
		txState = clickedWindow.Click(slotId, &player.cursor, rightClick, shiftClick, txId, expectedSlotContent)
	}

	switch txState {
	case TxStateAccepted, TxStateRejected:
		// Inform client of operation status.
		buf := new(bytes.Buffer)
		proto.WriteWindowTransaction(buf, windowId, txId, txState == TxStateAccepted)
		player.cursor.SendUpdate(buf, WindowIdCursor, SlotIdCursor)
		player.TransmitPacket(buf.Bytes())
	case TxStateDeferred:
		// The remote inventory should send the transaction outcome.
	}
}

func (player *Player) PacketWindowTransaction(windowId WindowId, txId TxId, accepted bool) {
	// TODO investigate when this packet is sent from the client and what it
	// means when it does get sent.
	log.Printf(
		"Got PacketWindowTransaction from player %q: windowId=%d txId=%d accepted=%t",
		player.name, windowId, txId, accepted)
}

func (player *Player) PacketSignUpdate(position *BlockXyz, lines [4]string) {
}

func (player *Player) PacketDisconnect(reason string) {
	log.Printf("Player %s disconnected reason=%s", player.name, reason)

	player.sendChatMessage(fmt.Sprintf("%s has left", player.name))

	player.onDisconnect <- player.EntityId
	player.txQueue <- nil
	player.mainQueue <- nil
	player.conn.Close()
}

func (player *Player) receiveLoop() {
	for {
		err := proto.ServerReadPacket(player.conn, player)
		if err != nil {
			if err != os.EOF {
				log.Print("ReceiveLoop failed: ", err.String())
			}
			return
		}
	}
}

// End of packet handling code

func (player *Player) transmitLoop() {
	for {
		bs, ok := <-player.txQueue

		if !ok || bs == nil {
			return // txQueue closed
		}
		_, err := player.conn.Write(bs)
		if err != nil {
			if err != os.EOF {
				log.Print("TransmitLoop failed: ", err.String())
			}
			return
		}
	}
}

func (player *Player) TransmitPacket(packet []byte) {
	if packet == nil {
		return // skip empty packets
	}
	player.txQueue <- packet
}

func (player *Player) runQueuedCall(f func(*Player)) {
	player.lock.Lock()
	defer player.lock.Unlock()
	f(player)
}

func (player *Player) mainLoop() {
	expVarPlayerConnectionCount.Add(1)
	defer expVarPlayerDisconnectionCount.Add(1)

	player.chunkSubs.Init(player)
	defer player.chunkSubs.Close()

	player.postLogin()

	for {
		f, ok := <-player.mainQueue
		if !ok || f == nil {
			return
		}
		player.runQueuedCall(f)
	}
}

func (player *Player) postLogin() {
	// TODO Old version of chunkSubscriptions.Move() that was called here had
	// stuff for a callback when the nearest chunks had been sent so that player
	// position would only be sent when nearby chunks were out. Some replacement
	// for this will be needed. Possibly a message could be queued to the current
	// shard following on from chunkSubscriptions's initialization that would ask
	// the shard to send out the following packets - this would result in them
	// being sent at least after the chunks that are in the current shard have
	// been sent.

	player.sendChatMessage(fmt.Sprintf("%s has joined", player.name))

	// Send player start position etc.
	buf := new(bytes.Buffer)
	proto.ServerWritePlayerPositionLook(
		buf,
		&player.position, player.position.Y+StanceNormal,
		&player.look, false)
	player.inventory.WriteWindowItems(buf)
	packet := buf.Bytes()

	// Enqueue on the shard as a hacky way to defer the packet send until after
	// the initial chunk data has been sent.
	player.chunkSubs.curShard.Enqueue(func() {
		player.TransmitPacket(packet)
	})
}

func (player *Player) reqInventorySubscribed(block *BlockXyz, invTypeId InvTypeId, slots []proto.WindowSlot) {
	if player.remoteInv != nil {
		player.closeCurrentWindow(true)
	}

	remoteInv := NewRemoteInventory(block, &player.chunkSubs, slots)

	window := player.inventory.NewWindow(invTypeId, player.nextWindowId, remoteInv)
	if window == nil {
		return
	}

	player.remoteInv = remoteInv
	player.curWindow = window

	if player.nextWindowId >= WindowIdFreeMax {
		player.nextWindowId = WindowIdFreeMin
	} else {
		player.nextWindowId++
	}

	buf := new(bytes.Buffer)
	window.WriteWindowOpen(buf)
	window.WriteWindowItems(buf)
	player.TransmitPacket(buf.Bytes())
}

func (player *Player) reqInventorySlotUpdate(block *BlockXyz, slot *slot.Slot, slotId SlotId) {
	if player.remoteInv == nil || !player.remoteInv.IsForBlock(block) {
		return
	}

	player.remoteInv.slotUpdate(slot, slotId)
}

func (player *Player) reqInventoryProgressUpdate(block *BlockXyz, prgBarId PrgBarId, value PrgBarValue) {
	if player.remoteInv == nil || !player.remoteInv.IsForBlock(block) {
		return
	}

	player.remoteInv.progressUpdate(prgBarId, value)
}

func (player *Player) reqInventoryCursorUpdate(block *BlockXyz, cursor *slot.Slot) {
	if player.remoteInv == nil || !player.remoteInv.IsForBlock(block) {
		return
	}

	player.cursor = *cursor
	buf := new(bytes.Buffer)
	player.cursor.SendUpdate(buf, WindowIdCursor, SlotIdCursor)
	player.TransmitPacket(buf.Bytes())
}

func (player *Player) reqInventoryTxState(block *BlockXyz, txId TxId, accepted bool) {
	if player.remoteInv == nil || !player.remoteInv.IsForBlock(block) || player.curWindow == nil {
		return
	}

	buf := new(bytes.Buffer)
	proto.WriteWindowTransaction(buf, player.curWindow.GetWindowId(), txId, accepted)
	player.TransmitPacket(buf.Bytes())
}

func (player *Player) reqInventoryUnsubscribed(block *BlockXyz) {
	if player.remoteInv == nil || !player.remoteInv.IsForBlock(block) {
		return
	}

	player.closeCurrentWindow(true)
}

func (player *Player) reqPlaceHeldItem(target *BlockXyz, wasHeld *slot.Slot) {
	curHeld, _ := player.inventory.HeldItem()

	// Currently held item has changed since chunk saw it.
	// TODO think about having the slot index passed as well so if that changes,
	// we can still track the original item and improve placement success rate.
	if curHeld.ItemType != wasHeld.ItemType || curHeld.Data != wasHeld.Data {
		return
	}

	shardClient, _, ok := player.chunkSubs.ShardClientForBlockXyz(target)
	if ok {
		var into slot.Slot
		into.Init()

		player.inventory.TakeOneHeldItem(&into)

		shardClient.ReqPlaceItem(*target, into)
	}
}

// Used to receive items picked up from chunks. It is synchronous so that the
// passed item can be looked at by the caller afterwards to see if it has been
// consumed.
func (player *Player) reqOfferItem(fromChunk *ChunkXz, entityId EntityId, item *slot.Slot) {
	if player.inventory.CanTakeItem(item) {
		shardClient, ok := player.chunkSubs.ShardClientForChunkXz(fromChunk)
		if ok {
			shardClient.ReqTakeItem(*fromChunk, entityId)
		}
	}

	return
}

func (player *Player) reqGiveItem(atPosition *AbsXyz, item *slot.Slot) {
	defer func() {
		// Check if item not fully consumed. If it is not, then throw the remains
		// back to the chunk.
		if item.Count > 0 {
			chunkLoc := atPosition.ToChunkXz()
			shardClient, ok := player.chunkSubs.ShardClientForChunkXz(&chunkLoc)
			if ok {
				shardClient.ReqDropItem(*item, *atPosition, AbsVelocity{})
			}
		}
	}()

	player.inventory.PutItem(item)
}

// Enqueue queues a function to run with the player lock within the player's
// mainloop.
func (player *Player) Enqueue(f func(*Player)) {
	if f == nil {
		return
	}
	player.mainQueue <- f
}

func (player *Player) sendChatMessage(message string) {
	buf := new(bytes.Buffer)
	proto.WriteChatMessage(buf, message)

	player.chunkSubs.curShard.ReqMulticastPlayers(
		player.chunkSubs.curChunkLoc,
		player.EntityId,
		buf.Bytes(),
	)
}

// closeCurrentWindow closes any open window. It must be called with
// player.lock held.
func (player *Player) closeCurrentWindow(sendClosePacket bool) {
	if player.curWindow != nil {
		player.curWindow.Finalize(sendClosePacket)
		player.curWindow = nil
	}

	if player.remoteInv != nil {
		player.remoteInv.Close()
		player.remoteInv = nil
	}

	player.inventory.Resubscribe()
}
