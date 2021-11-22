package nosql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/moontrade/mdbx-go"
	"sort"
	"sync"
	"time"
)

var (
	ErrEvolutionInProcess = errors.New("evolution in process")
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
	Context              context.Context
	Cancel               context.CancelFunc
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

func (ev *evolution) NeedsApply() bool {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	if !ev.completed.IsZero() {
		return false
	}
	return ev.needsApply()
}

func (ev *evolution) needsApply() bool {
	return len(ev.creates) == 0 ||
		len(ev.drops) > 0 ||
		len(ev.indexCreates) > 0 ||
		len(ev.indexRebuilds) > 0 ||
		len(ev.indexDrops) > 0
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

func (ss *schemasStore) hydrate(
	ctx context.Context,
	schema *Schema,
) (<-chan EvolutionProgress, error) {
	ctx, cancel := context.WithCancel(ctx)
	ev := &evolution{
		store:      ss,
		schema:     schema,
		ctx:        ctx,
		cancel:     cancel,
		chProgress: make(chan EvolutionProgress, 1),
	}
	schema.store = ss.store
	ss.mu.Lock()
	ev.from = ss.schemasByUID[schema.Meta.UID]
	existing := ss.evolutions[schema.Meta.UID]
	if existing != nil {
		ss.mu.Unlock()
		return nil, ErrEvolutionInProcess
	}
	ss.evolutions[schema.Meta.UID] = ev
	ss.mu.Unlock()

	createCollection := func(col Collection) (Collection, error) {
		if col.collectionStore == nil {
			col.collectionStore = &collectionStore{
				store: ss.store,
			}
		}
		col.collectionStore.store = ss.store

		collectionMeta := col.CollectionMeta
		collectionMeta.Id = ss.nextCollectionID()
		collectionMeta.Owner = schema.Meta.Id
		collectionMeta.Name = col.Name
		if collectionMeta.Id == 0 {
			return col, ErrCollectionIDExhaustion
		}
		collectionMeta.Indexes = make([]IndexMeta, len(col.indexes))
		for i := 0; i < len(col.indexes); i++ {
			index := col.indexes[i]
			idxStore := index.getStore()
			if idxStore == nil {
				idxStore = &indexStore{}
			}
			idxStore.store = ss.store
			idxStore.collection = col.collectionStore
			idxStore.index = index
			index.setStore(idxStore)

			indexMeta := index.Meta()
			indexMeta.ID = ss.nextIndexID()
			if indexMeta.ID == 0 {
				return col, ErrIndexIDExhaustion
			}
			indexMeta.Owner = collectionMeta.Id
			collectionMeta.Indexes[i] = indexMeta
			index.setMeta(indexMeta)
		}

		col.CollectionMeta = collectionMeta
		ev.creates = append(ev.creates, &CollectionCreate{
			meta:  collectionMeta,
			store: col.collectionStore,
		})
		return col, nil
	}

	if ev.from != nil {
		schema.Meta.Id = ev.from.Id
		existingCollections := make(map[string]CollectionMeta)
		for _, col := range ev.from.Collections {
			existingCollections[col.Name] = col
		}

		nextCollections := make(map[string]*collectionStore)
		for i := 0; i < len(schema.Collections); i++ {
			collection := schema.Collections[i]
			if nextCollections[collection.Name] != nil {
				return nil, fmt.Errorf("duplicate collection name used: %s", collection.Name)
			}
			collection.collectionStore.store = ss.store
			nextCollections[collection.Name] = collection.collectionStore
		}

		for i := 0; i < len(schema.Collections); i++ {
			col := schema.Collections[i]
			existingCollection, ok := existingCollections[col.Name]
			if !ok {
				var err error
				if col, err = createCollection(col); err != nil {
					return nil, err
				}
				schema.Collections[i] = col
			} else {
				col.CollectionMeta.Id = existingCollection.Id
				col.CollectionMeta.Owner = existingCollection.Owner
				col.CollectionMeta.Name = col.Name

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
				for i := 0; i < len(col.indexes); i++ {
					index := col.indexes[i]
					idxStore := index.getStore()
					if idxStore == nil {
						idxStore = &indexStore{}
						index.setStore(idxStore)
					}
					idxStore.store = ss.store
					idxStore.collection = col.collectionStore
					idxStore.index = index
					index.setStore(idxStore)

					to := index.Meta()
					to.Owner = col.Id
					from, ok := existingIndexes[index.Name()]
					if !ok {
						to.ID = ss.nextIndexID()
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
						to.Owner = from.Owner
						index.setMeta(to)
						if !to.equals(from) {
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
					ev.indexCreates = append(ev.indexCreates, indexCreates...)
				}
				if len(indexRebuilds) > 0 {
					ev.indexRebuilds = append(ev.indexRebuilds, indexRebuilds...)
				}
				if len(indexDrops) > 0 {
					ev.indexDrops = append(ev.indexDrops, indexDrops...)
				}

				// Remove from existingCollections map.
				delete(existingCollections, col.Name)

				schema.Collections[i] = col
			}
		}

		// Anything remaining in existingCollections map needs to be dropped.
		if len(existingCollections) > 0 {
			for _, collection := range existingCollections {
				ev.drops = append(ev.drops, &CollectionDrop{
					meta:  collection,
					store: nil,
				})

				// Drop all indexes
				if len(collection.Indexes) > 0 {
					for _, index := range collection.Indexes {
						ev.indexDrops = append(ev.indexDrops, &IndexDrop{
							meta:  index,
							store: nil,
						})
					}
				}
			}
		}

		//if !ev.needsApply() {
		//	ss.mu.Lock()
		//	delete(ss.evolutions, schema.Meta.UID)
		//	ss.mu.Unlock()
		//	return nil, nil
		//}
	} else {
		schema.Meta.Id = ss.nextIndexID()
		ev.creates = make([]*CollectionCreate, 0, len(schema.Collections))
		for i := 0; i < len(schema.Collections); i++ {
			col := schema.Collections[i]
			var err error
			if col, err = createCollection(col); err != nil {
				return nil, err
			}
			schema.Collections[i] = col
		}
	}

	ev.to = schema.buildMeta()

	return ev.apply()
}

func (ev *evolution) Close() error {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	if ev.cancel != nil {
		ev.cancel()
		ev.cancel = nil
	}
	return nil
}

func (ev *evolution) apply() (<-chan EvolutionProgress, error) {
	ev.mu.Lock()
	defer ev.mu.Unlock()
	chProgress := ev.chProgress
	go func() {
		var (
			err       error
			uid       = ev.to.UID
			batchSize = int64(100000)
			nstore    = ev.store.store
			store     = ev.store.store.store
			docsDBI   = ev.store.store.documentsDBI
			indexDBI  = ev.store.store.indexDBI
			progress  = EvolutionProgress{
				Context: ev.ctx,
				Cancel:  ev.cancel,
				State:   EvolutionStatePreparing,
				Started: time.Now(),
			}
			now = time.Now()
		)

		defer func() {
			recover()
			ev.mu.Lock()
			defer ev.mu.Unlock()

			delete(ev.store.evolutions, uid)
			if ev.err == nil {
				ev.store.schemasByUID[ev.to.UID] = ev.to
			}
			ev.completed = time.Now()
			if ev.chProgress != nil {
				close(ev.chProgress)
			}
			if ev.ctx != nil {
				ev.ctx.Value(progress)
			}
			if ev.cancel != nil {
				ev.cancel()
				ev.cancel = nil
			}
		}()

		ev.ctx.Value(progress)
		chProgress <- progress

		var (
			collectionsToLoad = make(map[CollectionID]*collectionStore)
			collectionDrops   = make(map[CollectionID]*CollectionDrop)
			indexDrops        = make(map[uint32]*IndexDrop)
			indexCreates      = make(map[uint32]*IndexCreate)
		)

		if len(ev.drops) > 0 {
			for _, drop := range ev.drops {
				collectionsToLoad[drop.meta.Id] = drop.store
				collectionDrops[drop.meta.Id] = drop
			}
		}

		if len(ev.indexCreates) > 0 {
			for _, create := range ev.indexCreates {
				if _, ok := collectionDrops[create.meta.Owner]; ok {
					continue
				}
				collectionsToLoad[create.meta.Owner] = create.store.collection
				indexCreates[create.meta.ID] = create
			}
		}

		if len(ev.indexDrops) > 0 {
			for _, drop := range ev.indexDrops {
				collectionsToLoad[drop.meta.Owner] = drop.store.collection
				indexDrops[drop.meta.ID] = drop
			}
		}

		if len(ev.indexRebuilds) > 0 {
			for _, rebuild := range ev.indexRebuilds {
				if _, ok := collectionDrops[rebuild.to.Owner]; ok {
					continue
				}
				collectionsToLoad[rebuild.to.Owner] = rebuild.store.collection
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
				collectionsToLoad[drop.meta.Id] = drop.store
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
							ev.indexDrops = append(ev.indexDrops, indexDrop)
							indexDrops[meta.ID] = indexDrop
						}
					}
				}
			}
		}

		if len(collectionsToLoad) > 0 {
			// Sort by ID
			list := make([]*collectionStore, 0, len(collectionsToLoad))
			for _, cs := range collectionsToLoad {
				list = append(list, cs)
			}
			sort.Sort(collectionStoresSlice(list))

			if err = store.View(func(tx *mdbx.Tx) error {
				var (
					begin, end *mdbx.Cursor
					err        error
				)
				begin, err = tx.OpenCursor(docsDBI)
				if err != mdbx.ErrSuccess {
					return err
				}
				err = nil
				defer begin.Close()

				end, err = tx.OpenCursor(docsDBI)
				if err != mdbx.ErrSuccess {
					return err
				}
				err = nil
				defer end.Close()

				// collection counts
				for _, cs := range list {
					_, _, _, err := cs.load(begin, end)
					if err != nil {
						continue
					}
				}

				return nil
			}); err != nil && err != mdbx.ErrSuccess {
				goto DONE
			}
			err = nil
		}

		progress.IndexDrops.Total = int64(len(indexDrops))
		for _, drop := range indexDrops {
			progress.IndexDrops.Total += collectionsToLoad[drop.meta.Owner].estimated
			progress.Total += collectionsToLoad[drop.meta.Owner].estimated
		}
		progress.IndexCreates.Total = int64(len(indexCreates))
		for _, create := range indexCreates {
			progress.IndexCreates.Total += collectionsToLoad[create.meta.Owner].estimated
			progress.Total += collectionsToLoad[create.meta.Owner].estimated
		}
		progress.CollectionDrops.Total = int64(len(collectionDrops))
		for _, drop := range collectionDrops {
			progress.CollectionDrops.Total += collectionsToLoad[drop.meta.Id].estimated
			progress.Total += collectionsToLoad[drop.meta.Id].estimated
		}
		now = time.Now()
		progress.Prepared = now.Sub(progress.Started)
		progress.State = EvolutionStatePrepared
		chProgress <- progress

		// Cancelled?
		select {
		case <-ev.ctx.Done():
			goto DONE
		default:
		}

		// Drop indexes?
		if len(indexDrops) > 0 {
			// Cancelled?
			select {
			case <-ev.ctx.Done():
				goto DONE
			default:
			}

			// drop indexes
			progress.State = EvolutionStateDroppingIndex
			chProgress <- progress

			for _, drop := range indexDrops {
				// Cancelled?
				select {
				case <-ev.ctx.Done():
					goto DONE
				default:
				}

				var (
					collectionID = drop.meta.Owner
					prefix       = drop.meta.ID
					key          = mdbx.U32(&prefix)
					data         = mdbx.Val{}
					count        = int64(0)
					total        = int64(0)
					estimated    = collectionsToLoad[collectionID].estimated
				)
				for err == nil {
					// Cancelled?
					select {
					case <-ev.ctx.Done():
						goto DONE
					default:
					}

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
					if err == mdbx.ErrSuccess {
						err = nil
					}

					progress.IndexDrops.Completed += count
					total += count
					count = 0
					chProgress <- progress
				}
				if err == mdbx.ErrSuccess || err == mdbx.ErrNotFound {
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
			// Cancelled?
			select {
			case <-ev.ctx.Done():
				goto DONE
			default:
			}

			progress.State = EvolutionStateCreatingIndex
			chProgress <- progress

			for _, create := range indexCreates {
				// Cancelled?
				select {
				case <-ev.ctx.Done():
					goto DONE
				default:
				}

				var (
					index        = create.store.index
					collectionID = create.meta.Owner
					docID        = NewDocID(create.meta.Owner, 0)
					docKey       = docID.Key()
					docData      = mdbx.Val{}
					count        = int64(0)
					total        = int64(0)
					estimated    = collectionsToLoad[collectionID].estimated
				)
				for err == nil {
					// Cancelled?
					select {
					case <-ev.ctx.Done():
						goto DONE
					default:
					}

					err = store.Update(func(tx *mdbx.Tx) error {
						var (
							docsCursor *mdbx.Cursor
							e          error
						)

						nstore.tx.Tx = tx
						nstore.tx.index, e = tx.OpenCursor(indexDBI)
						if e != mdbx.ErrSuccess {
							return e
						}
						defer func() {
							nstore.tx.Tx = nil
						}()
						defer nstore.tx.index.Close()

						docsCursor, e = tx.OpenCursor(docsDBI)
						if e != mdbx.ErrSuccess {
							return e
						}
						defer docsCursor.Close()

						for {
							if e = docsCursor.Get(&docKey, &docData, mdbx.CursorNextNoDup); e != mdbx.ErrSuccess {
								return e
							}

							if docKey.Len < 8 {
								return mdbx.ErrNotFound
							}
							docID = DocID(docKey.U64())
							if docID.CollectionID() != collectionID {
								return mdbx.ErrNotFound
							}

							// Insert index
							if e = index.doInsert(nstore.tx); e != nil {
								return e
							}

							count++
							if count >= batchSize {
								return nil
							}
						}
					})
					if err == mdbx.ErrSuccess {
						err = nil
					}

					progress.IndexCreates.Completed += count
					total += count
					count = 0
					chProgress <- progress
				}
				if err == mdbx.ErrSuccess || err == mdbx.ErrNotFound {
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
			// Cancelled?
			select {
			case <-ev.ctx.Done():
				goto DONE
			default:
			}

			progress.State = EvolutionStateDroppingCollection
			chProgress <- progress

			for _, drop := range collectionDrops {
				// Cancelled?
				select {
				case <-ev.ctx.Done():
					goto DONE
				default:
				}

				var (
					collectionID = drop.meta.Id
					k            = NewDocID(collectionID, 0)
					key          = k.Key()
					data         = mdbx.Val{}
					id           DocID
					count        = int64(0)
					total        = int64(0)
					estimated    = collectionsToLoad[collectionID].estimated
				)
				for err == nil {
					// Cancelled?
					select {
					case <-ev.ctx.Done():
						goto DONE
					default:
					}

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
					if err == mdbx.ErrSuccess {
						err = nil
					}

					progress.CollectionDrops.Completed += count
					progress.Completed += count
					total += count
					count = 0
					chProgress <- progress
				}
				if err == mdbx.ErrSuccess || err == mdbx.ErrNotFound {
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

		// Cancelled?
		select {
		case <-ev.ctx.Done():
			goto DONE
		default:
		}

		// Save schema.
		if err = store.Update(func(tx *mdbx.Tx) error {
			var (
				bytes, err = json.Marshal(ev.to)
				k          = NewDocID(schemaCollectionID, uint64(ev.to.Id))
				key        = k.Key()
				data       = mdbx.Bytes(&bytes)
			)

			if err != nil {
				return err
			}

			if e := tx.Put(docsDBI, &key, &data, 0); e != mdbx.ErrSuccess {
				return e
			}
			return nil
		}); err != nil {
			goto DONE
		}

		// force sync to disk
		if err = store.Sync(); err != mdbx.ErrSuccess && err != mdbx.ErrResultTrue && err != nil {
			goto DONE
		}
		if err = store.Sync(); err != mdbx.ErrSuccess && err != mdbx.ErrResultTrue && err != nil {
			goto DONE
		}

		for _, drop := range indexDrops {
			_ = drop
		}
		for _, drop := range collectionDrops {
			drop.store.CollectionMeta = drop.meta
		}

	DONE:
		if err == nil {
			err = ev.ctx.Err()
		}
		if err != nil {
			progress.Err = err
			progress.State = EvolutionStateError
			ev.err = err
			ev.chProgress <- progress
		} else {
			progress.State = EvolutionStateCompleted
			ev.chProgress <- progress
		}
	}()

	return chProgress, nil
}
