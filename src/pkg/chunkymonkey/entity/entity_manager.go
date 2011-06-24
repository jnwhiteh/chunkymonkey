package entity

import (
	"sync"

	. "chunkymonkey/types"
)

// TODO EntityManager should be a service in its own right, able to hand out
// blocks of IDs and running its own goroutine (potentially shardable by
// entityId if necessary). Right now taking the easy option of using a simple
// lock.
type EntityManager struct {
	nextEntityId EntityId
	entities     map[EntityId]bool
	lock         sync.Mutex
}

func (mgr *EntityManager) Init() {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	mgr.nextEntityId = 0
	mgr.entities = make(map[EntityId]bool)
}

func (mgr *EntityManager) createEntityId() EntityId {
	// Search for next free ID
	entityId := mgr.nextEntityId
	_, exists := mgr.entities[entityId]
	for exists {
		entityId++
		if entityId == mgr.nextEntityId {
			// TODO Better handling of this? It shouldn't happen, realistically - but
			// neither should it explode.
			panic("EntityId space exhausted")
		}
		_, exists = mgr.entities[entityId]
	}
	mgr.nextEntityId = entityId + 1

	return entityId
}

// AddEntity adds an entity to the manager, and assigns it a world-unique
// EntityId.
// NewEntity creates a world-unique entityId in the manager and returns it.
func (mgr *EntityManager) NewEntity() EntityId {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	entityId := mgr.createEntityId()
	mgr.entities[entityId] = true
	return entityId
}

// RemoveEntity removes an entity from the manager.
func (mgr *EntityManager) RemoveEntityById(entityId EntityId) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	mgr.entities[entityId] = false, false
}
