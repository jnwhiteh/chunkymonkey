package main

import (
	"encoding/hex"
	"io"
	"log"
	"os"

	"chunkymonkey/proto"
	. "chunkymonkey/types"
)

// Hex dumps the input to the log
func (p *MessageParser) dumpInput(logPrefix string, reader io.Reader) {
	buf := make([]byte, 16, 16)
	for {
		_, err := io.ReadAtLeast(reader, buf, 1)
		if err != nil {
			return
		}

		hexData := hex.EncodeToString(buf)
		p.printf("Unparsed data: %s", hexData)
	}
}

// Consumes data from reader until an error occurs
func (p *MessageParser) consumeUnrecognizedInput(reader io.Reader) {
	p.printf("Lost packet sync. Ignoring further data.")
	buf := make([]byte, 4096)
	for {
		_, err := io.ReadAtLeast(reader, buf, 1)
		if err != nil {
			return
		}
	}
}

type MessageParser struct {
	logger *log.Logger
}

func (p *MessageParser) printf(format string, v ...interface{}) {
	p.logger.Printf(format, v...)
}

func (p *MessageParser) PacketKeepAlive() {
	// Not logging this packet as it's a bit spammy
}

func (p *MessageParser) PacketChatMessage(message string) {
	p.printf("PacketChatMessage(%q)", message)
}

func (p *MessageParser) PacketRespawn(dimension DimensionId) {
	p.printf("PacketRespawn(dimension=%d)")
}

func (p *MessageParser) PacketPlayer(onGround bool) {
	// Not logging this packet as it's a bit spammy
}

func (p *MessageParser) PacketPlayerPosition(position *AbsXyz, stance AbsCoord, onGround bool) {
	p.printf("PacketPlayerPosition(position=%v, stance=%v, onGround=%t)", position, stance, onGround)
}

func (p *MessageParser) PacketPlayerLook(look *LookDegrees, onGround bool) {
	p.printf("PacketPlayerLook(look=%v, onGround=%t)", look, onGround)
}

func (p *MessageParser) PacketPlayerBlockHit(status DigStatus, blockLoc *BlockXyz, face Face) {
	p.printf("PacketPlayerBlockHit(status=%v, blockLoc=%v, face=%v)", status, blockLoc, face)
}

func (p *MessageParser) PacketPlayerBlockInteract(itemId ItemTypeId, blockLoc *BlockXyz, face Face, amount ItemCount, data ItemData) {
	p.printf("PacketPlayerBlockInteract(itemId=%d, blockLoc=%v, face=%d, amount=%d, data=%d)",
		itemId, blockLoc, face, amount, data)
}

func (p *MessageParser) PacketHoldingChange(slotId SlotId) {
	p.printf("PacketHoldingChange(slotId=%d)", slotId)
}

func (p *MessageParser) PacketBedUse(flag bool, bedLoc *BlockXyz) {
	p.printf("PacketBedUse(flag=%v, bedLoc=%v)", flag, bedLoc)
}

func (p *MessageParser) PacketEntityAnimation(entityId EntityId, animation EntityAnimation) {
	p.printf("PacketEntityAnimation(entityId=%d, animation=%v)", entityId, animation)
}

func (p *MessageParser) PacketEntityAction(entityId EntityId, action EntityAction) {
	p.printf("PacketEntityAction(entityId=%d, action=%d)",
		entityId, action)
}

func (p *MessageParser) PacketSignUpdate(position *BlockXyz, lines [4]string) {
	p.printf("PacketSignUpdate(position=%v, lines=[%q, %q, %q, %q])",
		position,
		lines[0], lines[1], lines[2], lines[3])
}

func (p *MessageParser) ClientPacketLogin(entityId EntityId, mapSeed RandomSeed, dimension DimensionId) {
	p.printf("PacketLogin(entityId=%d, mapSeed=%d, dimension=%d)",
		entityId, mapSeed, dimension)
}

func (p *MessageParser) PacketTimeUpdate(time Ticks) {
	p.printf("PacketTime(time=%d)", time)
}

func (p *MessageParser) PacketEntityEquipment(entityId EntityId, slot SlotId, itemId ItemTypeId, data ItemData) {
	p.printf("PacketEntityEquipment(entityId=%d, slot=%d, itemId=%d, data=%d)",
		entityId, slot, itemId, data)
}

func (p *MessageParser) PacketSpawnPosition(position *BlockXyz) {
	p.printf("PacketSpawnPosition(position=%v)", position)
}

func (p *MessageParser) PacketUseEntity(user EntityId, target EntityId, leftClick bool) {
	p.printf("PacketUseEntity(user=%d, target=%d, leftClick=%t)", user, target, leftClick)
}

func (p *MessageParser) PacketUpdateHealth(health Health) {
	p.printf("PacketUpdateHealth(health=%d)", health)
}

func (p *MessageParser) PacketNamedEntitySpawn(entityId EntityId, name string, position *AbsIntXyz, look *LookBytes, currentItem ItemTypeId) {
	p.printf("PacketNamedEntitySpawn(entityId=%d, name=%q, position=%v, look=%v, currentItem=%d)",
		entityId, name, position, look, currentItem)
}

func (p *MessageParser) PacketItemSpawn(entityId EntityId, itemId ItemTypeId, count ItemCount, data ItemData, location *AbsIntXyz, orientation *OrientationBytes) {
	p.printf("PacketItemSpawn(entityId=%d, itemId=%d, count=%d, data=%d, location=%v, orientation=%v)",
		entityId, itemId, count, data, location, orientation)
}

func (p *MessageParser) PacketItemCollect(collectedItem EntityId, collector EntityId) {
	p.printf("PacketItemCollect(collectedItem=%d, collector=%d)",
		collectedItem, collector)
}

func (p *MessageParser) PacketObjectSpawn(entityId EntityId, objType ObjTypeId, position *AbsIntXyz, objectData *proto.ObjectData) {
	p.printf("PacketObjectSpawn(entityId=%d, objType=%d, position=%v, objectData=%#v)",
		entityId, objType, position, objectData)
}

func (p *MessageParser) PacketEntitySpawn(entityId EntityId, mobType EntityMobType, position *AbsIntXyz, look *LookBytes, metadata []proto.EntityMetadata) {
	p.printf("PacketEntitySpawn(entityId=%d, mobType=%d, position=%v, look=%v, metadata=%v)",
		entityId, mobType, position, look, metadata)
}

func (p *MessageParser) PacketPaintingSpawn(entityId EntityId, title string, position *BlockXyz, paintingType PaintingTypeId) {
	p.printf("PacketPaintingSpawn(entityId=%d, title=%q, position=%v, paintingType=%d)",
		entityId, title, position, position, paintingType)
}

func (p *MessageParser) PacketUnknown0x1b(field1, field2 float32, field3, field4 bool, field5, field6 float32) {
	p.printf(
		"PacketUnknown0x1b(field1=%v, field2=%v, field3=%t, field4=%t, field5=%v, field6=%v)",
		field1, field2, field3, field4, field5, field6)
}

func (p *MessageParser) PacketEntityVelocity(entityId EntityId, velocity *Velocity) {
	p.printf("PacketEntityVelocity(entityId=%d, velocity=%v)",
		entityId, velocity)
}

func (p *MessageParser) PacketEntityDestroy(entityId EntityId) {
	p.printf("PacketEntityDestroy(entityId=%d)", entityId)
}

func (p *MessageParser) PacketEntity(entityId EntityId) {
	p.printf("PacketEntity(entityId=%d)", entityId)
}

func (p *MessageParser) PacketEntityRelMove(entityId EntityId, movement *RelMove) {
	p.printf("PacketEntityRelMove(entityId=%d, movement=%v)",
		entityId, movement)
}

func (p *MessageParser) PacketEntityLook(entityId EntityId, look *LookBytes) {
	p.printf("PacketEntityLook(entityId=%d, look=%v)",
		entityId, look)
}

func (p *MessageParser) PacketEntityTeleport(entityId EntityId, position *AbsIntXyz, look *LookBytes) {
	p.printf("PacketEntityTeleport(entityId=%d, position=%v, look=%v)",
		entityId, position, look)
}

func (p *MessageParser) PacketEntityStatus(entityId EntityId, status EntityStatus) {
	p.printf("PacketEntityStatus(entityId=%d, status=%d)",
		entityId, status)
}

func (p *MessageParser) PacketEntityMetadata(entityId EntityId, metadata []proto.EntityMetadata) {
	p.printf("PacketEntityMetadata(entityId=%d, metadata=%v)", entityId, metadata)
}

func (p *MessageParser) PacketPreChunk(position *ChunkXz, mode ChunkLoadMode) {
	p.printf("PacketPreChunk(position=%v, mode=%d)", position, mode)
}

func (p *MessageParser) PacketMapChunk(position *BlockXyz, size *SubChunkSize, data []byte) {
	p.printf("PacketMapChunk(position=%v, size=%v, len(data)=%d)",
		position, size, len(data))
}

func (p *MessageParser) PacketBlockChangeMulti(chunkLoc *ChunkXz, blockCoords []SubChunkXyz, blockTypes []BlockId, blockMetaData []byte) {
	p.printf("PacketBlockChangeMulti(chunkLoc=%v, blockCoords=(%d) %v, blockTypes=%v, blockMetaData=%v)",
		chunkLoc, len(blockCoords), blockCoords, blockTypes, blockMetaData)
}

func (p *MessageParser) PacketBlockChange(blockLoc *BlockXyz, blockType BlockId, blockMetaData byte) {
	p.printf("PacketBlockChange(blockLoc=%v, blockType=%d, blockMetaData=%d)",
		blockLoc, blockType, blockMetaData)
}

func (p *MessageParser) PacketNoteBlockPlay(position *BlockXyz, instrument InstrumentId, pitch NotePitch) {
	p.printf("PacketNoteBlockPlay(position=%v, instrument=%d, pitch=%d)",
		position, instrument, pitch)
}

func (p *MessageParser) PacketExplosion(position *AbsXyz, power float32, blockOffsets []proto.ExplosionOffsetXyz) {
	p.printf("PacketExplosion(position=%v, power=%f, blockOffsets=(%d) %v)",
		position, power, len(blockOffsets), blockOffsets)
}

func (p *MessageParser) PacketUnknown0x3d(field1, field2 int32, field3 int8, field4, field5 int32) {
	p.printf(
		"PacketUnknown0x3d(field1=%d, field2=%d, field3=%d, field4=%d, field5=%d)",
		field1, field2, field3, field4, field5)
}

func (p *MessageParser) PacketBedInvalid(field1 byte) {
	p.printf("PacketBedInvalid(field1=%t)", field1)
}

func (p *MessageParser) PacketWeather(entityId EntityId, raining bool, position *AbsIntXyz) {
	p.printf("PacketWeather(entityId=%d, raining=%t, position=%#v)",
		entityId, raining, position)
}

func (p *MessageParser) PacketWindowOpen(windowId WindowId, invTypeId InvTypeId, windowTitle string, numSlots byte) {
	p.printf("PacketWindowOpen(windowId=%d, invTypeId=%d, windowTitle=%q, numSlots=%d)",
		windowId, invTypeId, windowTitle, numSlots)
}

func (p *MessageParser) PacketWindowClose(windowId WindowId) {
	p.printf("PacketWindowClose(windowId=%d)", windowId)
}

func (p *MessageParser) PacketWindowClick(windowId WindowId, slot SlotId, rightClick bool, txId TxId, shiftClick bool, expectedSlot *proto.WindowSlot) {
	p.printf("PacketWindowClick(windowId=%d, slot=%d, rightClick=%t, txId=%d, shiftClick=%t, expectedSlot=%#v)",
		windowId, slot, rightClick, txId, shiftClick, expectedSlot)
}

func (p *MessageParser) PacketWindowSetSlot(windowId WindowId, slot SlotId, itemId ItemTypeId, amount ItemCount, data ItemData) {
	p.printf("PacketWindowSetSlot(windowId=%d, slot=%d, itemId=%d, amount=%d, data=%d)",
		windowId, slot, itemId, amount, data)
}

func (p *MessageParser) PacketWindowItems(windowId WindowId, items []proto.WindowSlot) {
	p.printf("PacketWindowItems(windowId=%d, items=(%d) %v)",
		windowId, len(items), items)
}

func (p *MessageParser) PacketWindowProgressBar(windowId WindowId, prgBarId PrgBarId, value PrgBarValue) {
	p.printf("PacketWindowProgressBar(windowId=%d, prgBarId=%d, value=%d)",
		windowId, prgBarId, value)
}

func (p *MessageParser) PacketWindowTransaction(windowId WindowId, txId TxId, accepted bool) {
	p.printf("PacketWindowTransaction(windowId=%d, txId=%d, accepted=%t)",
		windowId, txId, accepted)
}

func (p *MessageParser) PacketIncrementStatistic(statisticId StatisticId, delta int8) {
	p.printf("PacketIncrementStatistic(statisticId=%d, delta=%d)",
		statisticId, delta)
}

func (p *MessageParser) PacketUnknown0x83(field1, field2 int16, field3 string) {
	p.printf("PacketUnknown0x83(field1=%d, field2=%d, field3=%q)",
		field1, field2, field3)
}

func (p *MessageParser) PacketDisconnect(reason string) {
	p.printf("PacketDisconnect(%q)", reason)
}

// Parses messages from the client
func (p *MessageParser) CsParse(reader io.Reader, logger *log.Logger) {
	p.logger = logger

	// If we return, we should consume all input to avoid blocking the pipe
	// we're listening on. TODO Maybe we could just close it?
	defer p.consumeUnrecognizedInput(reader)

	defer func() {
		if err := recover(); err != nil {
			p.printf("Parsing failed: %v", err)
		}
	}()

	username, err := proto.ServerReadHandshake(reader)
	if err != nil {
		p.printf("ServerReadHandshake error: %v", err)
		return
	}
	p.printf("ServerReadHandshake(username=%v)", username)

	loginUsername, err := proto.ServerReadLogin(reader)
	if err != nil {
		p.printf("ServerReadLogin error: %v", err)
		return
	}
	p.printf("ServerReadLogin(username=%v)", loginUsername)

	for {
		err := proto.ServerReadPacket(reader, p)
		if err != nil {
			if err != os.EOF {
				p.printf("ReceiveLoop failed: %v", err)
			} else {
				p.printf("ReceiveLoop hit EOF")
			}
			return
		}
	}
}

// Parses messages from the server
func (p *MessageParser) ScParse(reader io.Reader, logger *log.Logger) {
	p.logger = logger

	// If we return, we should consume all input to avoid blocking the pipe
	// we're listening on. TODO Maybe we could just close it?
	defer p.consumeUnrecognizedInput(reader)

	defer func() {
		if err := recover(); err != nil {
			p.printf("Parsing failed: %v", err)
		}
	}()

	serverId, err := proto.ClientReadHandshake(reader)
	if err != nil {
		p.printf("ClientReadHandshake error: %v", err)
		return
	}
	p.printf("ClientReadHandshake(serverId=%v)", serverId)

	for {
		err := proto.ClientReadPacket(reader, p)
		if err != nil {
			if err != os.EOF {
				p.printf("ReceiveLoop failed: %v", err)
			} else {
				p.printf("ReceiveLoop hit EOF")
			}
			return
		}
	}
}
