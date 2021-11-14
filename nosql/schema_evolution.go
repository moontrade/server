package nosql

import (
	"context"
	"fmt"
	"github.com/moontrade/mdbx-go"
	"sync"
	"time"
)

// evolution is a set of ChangeActions required to perform in order to get the
// schema consistent.
type evolution struct {
	store         *schemasStore
	started       time.Time
	completed     time.Time
	from          *SchemaMeta
	to            *SchemaMeta
	schema        *Schema
	creates       []*CollectionCreate
	drops         []*CollectionDrop
	indexCreates  []*IndexCreate
	indexRebuilds []*IndexRebuild
	indexDrops    []*IndexDrop
	progress      EvolutionProgress
	chProgress    chan EvolutionProgress
	err           error
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
}

type EvolutionState int

const (
	EvolutionStatePreparing          EvolutionState = 0
	EvolutionStatePrepared           EvolutionState = 1
	EvolutionStateDroppingIndex      EvolutionState = 2
	EvolutionStateCreatingIndex      EvolutionState = 3
	EvolutionStateDroppingCollection EvolutionState = 4
	EvolutionStateCompleted          EvolutionState = 10
	EvolutionStateError              EvolutionState = 100
)

type EvolutionProgress struct {
	ProgressStat
	Time                 time.Time
	Started              time.Time
	Prepared             time.Duration
	IndexesDroppedIn     time.Duration
	IndexesCreatedIn     time.Duration
	CollectionsDroppedIn time.Duration
	State                EvolutionState
	BatchSize            int64
	IndexDrops           ProgressStat
	IndexCreates         ProgressStat
	CollectionDrops      ProgressStat
	Err                  error
}

type ProgressStat struct {
	Total     int64
	Completed int64
}

func (p ProgressStat) Pct() float64 {
	if p.Total == 0 {
		return 1
	}
	if p.Completed == 0 {
		return 0
	}
	return float64(p.Completed) / float64(p.Total)
}

func (evol *evolution) NeedsApply() bool {
	evol.mu.Lock()
	defer evol.mu.Unlock()
	if !evol.completed.IsZero() {
		return false
	}
	return evol.needsApply()
}

func (evol *evolution) needsApply() bool {
	return len(evol.creates) == 0 ||
		len(evol.drops) > 0 ||
		len(evol.indexCreates) > 0 ||
		len(evol.indexRebuilds) > 0 ||
		len(evol.indexDrops) > 0
}

type ChangeKind int

const (
	ChangeCreate       ChangeKind = 1
	ChangeDrop         ChangeKind = 2
	ChangeCreateIndex  ChangeKind = 3
	ChangeRebuildIndex ChangeKind = 4
	ChangeDropIndex    ChangeKind = 5
)

type CollectionCreate struct {
	meta  CollectionMeta
	store *collectionStore
}

// CollectionDrop action to delete all of a collection's documents and remove
// the metadata from the schema.
type CollectionDrop struct {
	meta  CollectionMeta
	store *collectionStore
}

type IndexCreate struct {
	meta  IndexMeta
	store *indexStore
}

type IndexRebuild struct {
	from  IndexMeta
	to    IndexMeta
	store *indexStore
}

// IndexDrop represents an action to drop an index from a schema, delete
// all data in the index database and remove the metadata from the saved schema.
type IndexDrop struct {
	meta  IndexMeta
	store *indexStore
}

func (m *schemasStore) hydrate(ctx context.Context, nextSchema *Schema) (chan EvolutionProgress, error) {
	ctx, cancel := context.WithCancel(ctx)
	cs := &evolution{
		store:  m,
		to:     nextSchema.buildMeta(),
		ctx:    ctx,
		cancel: cancel,
	}
	m.mu.Lock()
	cs.from = m.schemasByUID[nextSchema.Meta.UID]
	m.mu.Unlock()

	if cs.from != nil {
		existingCollections := make(map[string]CollectionMeta)
		for _, col := range cs.from.Collections {
			existingCollections[col.Name] = col
		}

		nextCollections := make(map[string]*collectionStore)
		collections := make([]*collectionStore, len(nextSchema.Collections))
		for i, col := range nextSchema.Collections {
			collections[i] = col.collectionStore
			if collections[i] == nil {
				return nil, ErrCollectionStore
			}
			if nextCollections[col.Name] != nil {
				return nil, fmt.Errorf("duplicate collection name used: %s", col.Name)
			}
			nextCollections[col.Name] = col.collectionStore
		}

		for _, col := range collections {
			existingCollection, ok := existingCollections[col.Name]
			if !ok {
				col.CollectionMeta.Id = m.nextCollectionID()
				if col.CollectionMeta.Id == 0 {
					return nil, ErrCollectionIDExhaustion
				}
				cs.creates = append(cs.creates, &CollectionCreate{
					meta:  col.CollectionMeta,
					store: col,
				})
			} else {
				// Was anything changed on collection?
				if !col.CollectionMeta.Equals(&existingCollection.collectionDescriptor) {
					// NOOP
				}

				existingIndexes := make(map[string]IndexMeta)
				for _, index := range existingCollection.Indexes {
					existingIndexes[index.Name] = index
				}
				var (
					indexCreates  []*IndexCreate
					indexRebuilds []*IndexRebuild
					indexDrops    []*IndexDrop
				)
				for _, index := range col.indexes {
					to := index.Meta()
					from, ok := existingIndexes[index.Name()]
					if !ok {
						to.ID = m.nextIndexID()
						if to.ID == 0 {
							return nil, ErrIndexIDExhaustion
						}
						index.setMeta(to)
						indexCreates = append(indexCreates, &IndexCreate{
							meta:  to,
							store: index.getStore(),
						})
					} else {
						delete(existingIndexes, index.Name())
						to.ID = from.ID
						if to.didChange(from) {
							indexRebuilds = append(indexRebuilds, &IndexRebuild{
								from:  from,
								to:    to,
								store: index.getStore(),
							})
						}
					}
				}

				for _, index := range existingIndexes {
					indexDrops = append(indexDrops, &IndexDrop{
						meta:  index,
						store: nil,
					})
				}

				if len(indexCreates) > 0 {
					cs.indexCreates = append(cs.indexCreates, indexCreates...)
				}
				if len(indexRebuilds) > 0 {
					cs.indexRebuilds = append(cs.indexRebuilds, indexRebuilds...)
				}
				if len(indexDrops) > 0 {
					cs.indexDrops = append(cs.indexDrops, indexDrops...)
				}

				// Remove from existingCollections map.
				delete(existingCollections, col.Name)
			}
		}

		// Anything remaining in existingCollections map needs to be dropped.
		if len(existingCollections) > 0 {
			for _, collection := range existingCollections {
				cs.drops = append(cs.drops, &CollectionDrop{
					meta:  collection,
					store: nil,
				})

				// Drop all indexes
				if len(collection.Indexes) > 0 {
					for _, index := range collection.Indexes {
						cs.indexDrops = append(cs.indexDrops, &IndexDrop{
							meta:  index,
							store: nil,
						})
					}
				}
			}
		}
	} else {
		cs.to.Id = m.nextSchemaID()
		cs.creates = make([]*CollectionCreate, len(nextSchema.Collections))

		for i, col := range nextSchema.Collections {
			if col.collectionStore == nil {
				col.collectionStore = &collectionStore{
					store: m.store,
				}
			}
			col.collectionStore.store = m.store
			col.CollectionMeta.Id = m.nextCollectionID()

			cs.creates[i] = &CollectionCreate{
				meta:  col.CollectionMeta,
				store: col.collectionStore,
			}
		}
	}

	return cs.Apply()
}

func (evol *evolution) Apply() (chan EvolutionProgress, error) {
	evol.mu.Lock()
	defer evol.mu.Unlock()

	if !evol.completed.IsZero() {
		return nil, nil
	}
	if evol.chProgress != nil {
		return evol.chProgress, nil
	}
	evol.chProgress = make(chan EvolutionProgress, 1)
	chProgress := evol.chProgress

	go func() {
		defer func() {
			recover()
			evol.mu.Lock()
			if evol.chProgress != nil {
				close(evol.chProgress)
			}
			if evol.cancel != nil {
				evol.cancel()
				evol.cancel = nil
			}
		}()

		var (
			err              error
			batchSize        = int64(100000)
			store            = evol.store.store.store
			docsDBI          = evol.store.store.documentsDBI
			indexDBI         = evol.store.store.indexDBI
			progress         = EvolutionProgress{}
			collectionCounts = make(map[CollectionID]int64)
			collectionDrops  = make(map[CollectionID]*CollectionDrop)
			indexDrops       = make(map[uint32]*IndexDrop)
			indexCreates     = make(map[uint32]*IndexCreate)
		)
		evol.ctx.Value(progress)
		progress.State = EvolutionStatePreparing
		progress.Started = time.Now()

		if len(evol.drops) > 0 {
			for _, drop := range evol.drops {
				collectionCounts[drop.meta.Id] = -1
				collectionDrops[drop.meta.Id] = drop
			}
		}

		if len(evol.indexCreates) > 0 {
			for _, create := range evol.indexCreates {
				if _, ok := collectionDrops[create.meta.Owner]; ok {
					continue
				}
				collectionCounts[create.meta.Owner] = -1
				indexCreates[create.meta.ID] = create
			}
		}

		if len(evol.indexDrops) > 0 {
			for _, drop := range evol.indexDrops {
				collectionCounts[drop.meta.Owner] = -1
				indexDrops[drop.meta.ID] = drop
			}
		}

		if len(evol.indexRebuilds) > 0 {
			for _, rebuild := range evol.indexRebuilds {
				if _, ok := collectionDrops[rebuild.to.Owner]; ok {
					continue
				}
				collectionCounts[rebuild.to.Owner] = -1
				if _, ok := indexDrops[rebuild.to.ID]; !ok {
					indexDrops[rebuild.to.ID] = &IndexDrop{
						meta:  rebuild.to,
						store: rebuild.store,
					}
				}
				if _, ok := indexCreates[rebuild.to.ID]; !ok {
					indexCreates[rebuild.to.ID] = &IndexCreate{
						meta:  rebuild.to,
						store: rebuild.store,
					}
				}
			}
		}

		if len(collectionDrops) > 0 {
			for _, drop := range collectionDrops {
				collectionCounts[drop.meta.Id] = -1
				if len(drop.meta.Indexes) > 0 {
					for _, index := range drop.store.indexes {
						meta := index.Meta()

						// Fix any index create actions
						delete(indexCreates, meta.ID)

						if _, ok := indexDrops[meta.ID]; !ok {
							indexDrop := &IndexDrop{
								meta:  meta,
								store: index.getStore(),
							}
							evol.indexDrops = append(evol.indexDrops, indexDrop)
							indexDrops[meta.ID] = indexDrop
						}
					}
				}
			}
		}

		// collection counts
		for collectionID := range collectionCounts {
			count, err := evol.store.store.EstimateCollectionCount(collectionID)
			if err != nil {
				continue
			}
			collectionCounts[collectionID] = count
		}
		progress.IndexDrops.Total = int64(len(indexDrops))
		for _, drop := range indexDrops {
			progress.IndexDrops.Total += collectionCounts[drop.meta.Owner]
			progress.Total += collectionCounts[drop.meta.Owner]
		}
		progress.IndexCreates.Total = int64(len(indexCreates))
		for _, create := range indexCreates {
			progress.IndexCreates.Total += collectionCounts[create.meta.Owner]
			progress.Total += collectionCounts[create.meta.Owner]
		}
		progress.CollectionDrops.Total = int64(len(collectionDrops))
		for _, drop := range collectionDrops {
			progress.CollectionDrops.Total += collectionCounts[drop.meta.Id]
			progress.Total += collectionCounts[drop.meta.Id]
		}
		now := time.Now()
		progress.Prepared = now.Sub(progress.Started)
		progress.State = EvolutionStatePrepared
		chProgress <- progress

		// Drop indexes?
		if len(indexDrops) > 0 {
			// drop indexes
			progress.State = EvolutionStateDroppingIndex
			chProgress <- progress

			for _, drop := range indexDrops {
				var (
					collectionID = drop.meta.Owner
					prefix       = drop.meta.ID
					key          = mdbx.U32(&prefix)
					data         = mdbx.Val{}
					count        = int64(0)
					total        = int64(0)
					estimated    = collectionCounts[collectionID]
				)
				for err == nil {
					err = store.Update(func(tx *mdbx.Tx) error {
						var (
							cursor *mdbx.Cursor
							e      error
						)

						cursor, e = tx.OpenCursor(indexDBI)
						if e != mdbx.ErrSuccess {
							return e
						}
						defer cursor.Close()

						for {
							if e = cursor.Get(&key, &data, mdbx.CursorNextNoDup); e != mdbx.ErrSuccess {
								return e
							}

							if key.Len < 4 {
								return mdbx.ErrNotFound
							}
							if key.U32() != prefix {
								return mdbx.ErrNotFound
							}
							if e = cursor.Delete(0); e != mdbx.ErrSuccess {
								return e
							}
							count++
							if count >= batchSize {
								return nil
							}
						}
					})

					progress.IndexDrops.Completed += count
					total += count
					count = 0
					chProgress <- progress
				}
				if err == mdbx.ErrNotFound {
					err = nil
				}
				if err != nil {
					goto DONE
				}
				// Readjust to actual count
				progress.IndexDrops.Total = progress.IndexDrops.Total - estimated + total
				progress.Total = progress.Total - estimated + total
			}

			progress.Time = time.Now()
			progress.IndexesDroppedIn = progress.Time.Sub(now)
			now = progress.Time
		}

		// Create indexes?
		if len(indexCreates) > 0 {
			progress.State = EvolutionStateCreatingIndex
			chProgress <- progress

			for _, create := range indexCreates {
				var (
					collectionID = create.meta.Owner
					prefix       = create.meta.ID
					key          = mdbx.U32(&prefix)
					data         = mdbx.Val{}
					count        = int64(0)
					total        = int64(0)
					estimated    = collectionCounts[collectionID]
				)
				for err == nil {
					err = store.Update(func(tx *mdbx.Tx) error {
						var (
							cursor *mdbx.Cursor
							e      error
						)

						cursor, e = tx.OpenCursor(indexDBI)
						if e != mdbx.ErrSuccess {
							return e
						}
						defer cursor.Close()

						for {
							if e = cursor.Get(&key, &data, mdbx.CursorNextNoDup); e != mdbx.ErrSuccess {
								return e
							}

							if key.Len < 4 {
								return mdbx.ErrNotFound
							}
							if key.U32() != prefix {
								return mdbx.ErrNotFound
							}
							count++
							if count >= batchSize {
								return nil
							}
						}
					})

					progress.IndexCreates.Completed += count
					total += count
					count = 0
					chProgress <- progress
				}
				if err == mdbx.ErrNotFound {
					err = nil
				}
				if err != nil {
					goto DONE
				}
				// Readjust to actual count
				progress.IndexCreates.Total = progress.IndexCreates.Total - estimated + total
				progress.Total = progress.Total - estimated + total
			}

			progress.Time = time.Now()
			progress.IndexesCreatedIn = progress.Time.Sub(now)
			now = progress.Time
		}

		// Drop collections?
		if len(collectionDrops) > 0 {
			progress.State = EvolutionStateDroppingCollection
			chProgress <- progress

			for _, drop := range collectionDrops {
				var (
					collectionID = drop.meta.Id
					k            = NewDocID(collectionID, 0)
					key          = k.Key()
					data         = mdbx.Val{}
					id           DocID
					count        = int64(0)
					total        = int64(0)
					estimated    = collectionCounts[collectionID]
				)
				for err == nil {
					err = store.Update(func(tx *mdbx.Tx) error {
						var (
							cursor *mdbx.Cursor
							e      error
						)

						cursor, e = tx.OpenCursor(docsDBI)
						if e != mdbx.ErrSuccess {
							return e
						}
						defer cursor.Close()

						for {
							if e = cursor.Get(&key, &data, mdbx.CursorNextNoDup); e != mdbx.ErrSuccess {
								return e
							}
							id = DocID(key.U64())
							if id.CollectionID() != collectionID {
								return mdbx.ErrNotFound
							}
							if e = cursor.Delete(0); e != mdbx.ErrSuccess {
								return e
							}
							count++
							if count >= batchSize {
								return nil
							}
						}
					})
					if err != nil {
						break
					}

					progress.CollectionDrops.Completed += count
					progress.Completed += count
					total += count
					count = 0
					chProgress <- progress
				}
				if err == mdbx.ErrNotFound {
					err = nil
				}
				if err != nil {
					goto DONE
				}
				// Readjust to actual count
				progress.CollectionDrops.Total = progress.CollectionDrops.Total - estimated + total
				progress.Total = progress.Total - estimated + total
			}

			progress.Time = time.Now()
			progress.CollectionsDroppedIn = progress.Time.Sub(now)
			now = progress.Time
		}

		// Save schema.
		err = store.Update(func(tx *mdbx.Tx) error {
			var (
				k    = NewDocID(schemaCollectionID, uint64(evol.to.Id))
				key  = k.Key()
				data = mdbx.Val{}
			)

			if e := tx.Put(docsDBI, &key, &data, 0); e != mdbx.ErrSuccess {
				return e
			}
			return nil
		})

		// force sync to disk
		if err = store.Sync(); err != nil {
			goto DONE
		}
		if err = store.Sync(); err != nil {
			goto DONE
		}

		for _, drop := range indexDrops {
			_ = drop
		}
		for _, drop := range collectionDrops {
			drop.store.CollectionMeta = drop.meta
		}

	DONE:
		if err != nil {
			progress.Err = err
			progress.State = EvolutionStateError
			evol.err = err
			evol.chProgress <- progress
		} else {
			progress.State = EvolutionStateCompleted
			evol.chProgress <- progress
		}
	}()

	return chProgress, nil
}
