// Defines interfaces for "objects" in the world (including pick up items,
// mobs, players).
package object

import (
	"io"
	"os"

	"chunkymonkey/physics"
	. "chunkymonkey/types"
)

// ISpawn represents common elements to all types of entities that can be
// present in a chunk.
type ISpawn interface {
	GetEntityId() EntityId
	SendSpawn(io.Writer) os.Error
	SendUpdate(io.Writer) os.Error
	Position() *AbsXyz
}

type INonPlayerEntity interface {
	ISpawn
	SetEntityId(EntityId)
	Tick(physics.BlockQueryFn) (leftBlock bool)
}
