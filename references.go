// Package govar provides a powerful and highly configurable pretty-printer for Go
// data structures. This file contains all the data structures and methods related
// to the sophisticated ID/back-reference system, which detects and visualizes
// pointer relationships and cycles within the dumped data.
package govar

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

// canonicalKey is a unique identifier for a value in memory, based on its address and type.
// It serves as the primary key for tracking references.
type canonicalKey struct {
	addr uintptr
	typ  reflect.Type
}

// definitionPoint stores information about which variable instance is chosen
// to be the "definition" that gets an ID printed next to it.
type definitionPoint struct {
	instanceKey      canonicalKey
	isPointerRef     bool
	indirectionLevel int
	level            int
	valueType        reflect.Type
}

// RefStats collects statistics about references to a value during the analysis pass.
type RefStats struct {
	pointerReferencesCount, definitionLevel, minPointerRefLevel, totalReferencesCount int
	valueKind                                                                         reflect.Kind
	isPrimitive                                                                       bool
	value                                                                             interface{}
}

// queueItem is used for the Breadth-First Search (BFS) traversal of the value graph.
type queueItem struct {
	v     reflect.Value
	level int
}

// addChildrenToQueue adds all child elements of a composite type (struct, slice, map) to the BFS queue.
func (d *Dumper) addChildrenToQueue(queue []queueItem, v reflect.Value, level int) []queueItem {
	v = deref(v)
	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			queue = append(queue, queueItem{v.Field(i), level + 1})
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			queue = append(queue, queueItem{v.Index(i), level + 1})
		}
	case reflect.Map:
		keys := sortMapKeys(v) // Use govar's stable map key sorting
		for _, key := range keys {
			queue = append(queue, queueItem{key, level + 1})
			queue = append(queue, queueItem{v.MapIndex(key), level + 1})
		}
	}
	return queue
}

// assignReferenceIDs is the third analysis pass. It iterates through the merged stats
// and assigns an ID (e.g., "&1") to any value that requires one.
func (d *Dumper) assignReferenceIDs() {
	mergedStats := d.getMergedStats()
	idCounter := 1

	// Sort roots to ensure deterministic ID assignment across runs.
	sortedRoots := make([]canonicalKey, 0, len(mergedStats))
	for key := range mergedStats {
		sortedRoots = append(sortedRoots, key)
	}
	sort.Slice(sortedRoots, func(i, j int) bool {
		if sortedRoots[i].addr != sortedRoots[j].addr {
			return sortedRoots[i].addr < sortedRoots[j].addr
		}
		return sortedRoots[i].typ.String() < sortedRoots[j].typ.String()
	})

	for _, rootKey := range sortedRoots {
		stats := mergedStats[rootKey]
		// An ID is needed if a value is pointed to by more than one pointer...
		isReferencedByMultiplePointers := stats.pointerReferencesCount > 1
		// ...or if it's referenced by both a pointer and a value copy.
		isMixedValueAndPointerReference := stats.pointerReferencesCount > 0 && stats.totalReferencesCount > stats.pointerReferencesCount

		if isReferencedByMultiplePointers || isMixedValueAndPointerReference {
			if _, exists := d.referenceIDs[rootKey]; !exists {
				d.referenceIDs[rootKey] = "&" + strconv.Itoa(idCounter)
				idCounter++
			}
		}
	}
}

// determineDefinitionPoints is the final analysis pass. It traverses the graph again
// to find the "best" place to print the ID for each value that has one.
// The "best" place is determined by a set of priority rules.
func (d *Dumper) determineDefinitionPoints(val reflect.Value) {
	queue := []queueItem{{val, 0}}
	traversed := make(map[canonicalKey]bool)
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		v := item.v
		if !v.IsValid() {
			continue
		}

		rawKey, keyOK := d.getRawKey(v)
		if !keyOK {
			continue
		}
		rootKey := d.findRoot(rawKey)
		// Only consider values that were assigned an ID.
		if _, hasID := d.referenceIDs[rootKey]; hasID {
			incumbent, exists := d.definitionPoints[rootKey]
			isBetter := false

			currentIsPointerRef := isPointerRef(v)

			if !exists {
				isBetter = true
			} else {
				// Priority Rules for choosing the definition point:
				// 1. A non-pointer value beats a pointer value.
				if incumbent.isPointerRef && !currentIsPointerRef {
					isBetter = true
				} else if !incumbent.isPointerRef && currentIsPointerRef {
					isBetter = false
				} else {
					// Kinds are the same (both pointers or both non-pointers).
					if currentIsPointerRef { // Both are pointers
						// 2. Lower pointer indirection level wins (e.g., *T beats **T).
						currentIndirection := getIndirectionLevel(v)
						if currentIndirection < incumbent.indirectionLevel {
							isBetter = true
						} else if currentIndirection == incumbent.indirectionLevel {
							// 3. Lower nesting depth wins.
							if item.level < incumbent.level {
								isBetter = true
							}
						}
					} else { // Both are non-pointers
						// 3. Lower nesting depth wins.
						if item.level < incumbent.level {
							isBetter = true
						}
					}
				}
			}

			// If the current location is better, update the definition point.
			if isBetter {
				if instKey, instKeyOK := d.getInstanceKey(v); instKeyOK {
					d.definitionPoints[rootKey] = definitionPoint{
						instanceKey:      instKey,
						isPointerRef:     currentIsPointerRef,
						indirectionLevel: getIndirectionLevel(v),
						level:            item.level,
						valueType:        deref(v).Type(),
					}
				}
			}
		}

		// Continue traversal.
		traversalTarget := deref(v)
		if !traversalTarget.IsValid() || !isCompositeOrInterface(traversalTarget.Kind()) {
			continue
		}
		traversalKey, ok := d.getRawKey(traversalTarget)
		if !ok || traversed[traversalKey] {
			continue
		}
		traversed[traversalKey] = true
		queue = d.addChildrenToQueue(queue, v, item.level)
	}
}

// findRoot is part of the union-find algorithm. It finds the root representative
// for a given key, applying path compression for efficiency.
func (d *Dumper) findRoot(k canonicalKey) canonicalKey {
	if parent, ok := d.canonicalRoots[k]; !ok || parent == k {
		d.canonicalRoots[k] = k
		return k
	}
	root := d.findRoot(d.canonicalRoots[k])
	d.canonicalRoots[k] = root
	return root
}

// getInstanceKey generates a canonicalKey for the value v itself, without dereferencing.
// This is used to uniquely identify a specific variable instance (e.g., a pointer variable)
// to mark it as the definition point.
func (d *Dumper) getInstanceKey(v reflect.Value) (canonicalKey, bool) {
	if !v.IsValid() {
		return canonicalKey{}, false
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return canonicalKey{}, false
		}
		return d.getInstanceKey(v.Elem())
	}
	if v.CanAddr() {
		return canonicalKey{addr: v.UnsafeAddr(), typ: v.Type()}, true
	}
	switch v.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		if v.IsNil() {
			return canonicalKey{}, false
		}
		return canonicalKey{addr: v.Pointer(), typ: v.Type()}, true
	}
	if isPrimitiveKind(v.Kind()) {
		exportedV := tryExport(v)
		if !exportedV.CanInterface() {
			return canonicalKey{}, false
		}
		val := exportedV.Interface()
		if addr, ok := d.fakeAddrs[val]; ok {
			return canonicalKey{addr: addr, typ: v.Type()}, true
		}
		addr := uintptr(0xffffff00 + len(d.fakeAddrs))
		d.fakeAddrs[val] = addr
		return canonicalKey{addr: addr, typ: v.Type()}, true
	}
	return canonicalKey{}, false
}

// getMergedStats aggregates the stats from all members of a unified set into a single
// RefStats struct for the root of that set.
func (d *Dumper) getMergedStats() map[canonicalKey]*RefStats {
	rootToMembers := make(map[canonicalKey][]canonicalKey)
	for key := range d.referenceStats {
		root := d.findRoot(key)
		rootToMembers[root] = append(rootToMembers[root], key)
	}

	mergedStats := make(map[canonicalKey]*RefStats)
	for root, members := range rootToMembers {
		// Start with the stats of the root itself.
		newStats := &RefStats{}
		if rootStat, ok := d.referenceStats[root]; ok {
			newStats.value = rootStat.value
			newStats.valueKind = rootStat.valueKind
			newStats.isPrimitive = rootStat.isPrimitive
		}

		// Add stats from all other members of the set.
		for _, memberKey := range members {
			if memberStats, ok := d.referenceStats[memberKey]; ok {
				newStats.totalReferencesCount += memberStats.totalReferencesCount
				newStats.pointerReferencesCount += memberStats.pointerReferencesCount
			}
		}
		mergedStats[root] = newStats
	}
	return mergedStats
}

// getOrCreateStats retrieves or initializes a RefStats struct for a given key.
func (d *Dumper) getOrCreateStats(key canonicalKey, v reflect.Value, level int) *RefStats {
	if stats, ok := d.referenceStats[key]; ok {
		return stats
	}
	stats := &RefStats{definitionLevel: level, minPointerRefLevel: -1, valueKind: v.Kind(), isPrimitive: isPrimitiveKind(v.Kind())}
	if v.IsValid() {
		exportedV := tryExport(v)
		if exportedV.CanInterface() {
			stats.value = exportedV.Interface()
		} else {
			stats.value = "<unexported>"
		}
	}
	d.referenceStats[key] = stats
	return stats
}

// getRawKey generates a canonicalKey for the underlying value that v points or refers to.
// It dereferences pointers and interfaces to find the "raw" value.
// For primitives, it uses a map of fake addresses to ensure distinct values get distinct keys.
func (d *Dumper) getRawKey(v reflect.Value) (canonicalKey, bool) {
	targetVal := deref(v)
	if !targetVal.IsValid() {
		return canonicalKey{}, false
	}

	if targetVal.CanAddr() {
		return canonicalKey{addr: targetVal.UnsafeAddr(), typ: targetVal.Type()}, true
	}

	switch targetVal.Kind() {
	case reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		if targetVal.IsNil() {
			return canonicalKey{}, false
		}
		return canonicalKey{addr: targetVal.Pointer(), typ: targetVal.Type()}, true
	}

	if isPrimitiveKind(targetVal.Kind()) {
		exportedV := tryExport(targetVal)
		if !exportedV.CanInterface() {
			return canonicalKey{}, false
		}
		val := exportedV.Interface()

		if addr, ok := d.fakeAddrs[val]; ok {
			return canonicalKey{addr: addr, typ: targetVal.Type()}, true
		}

		addr := uintptr(0xffffff00 + len(d.fakeAddrs))
		d.fakeAddrs[val] = addr
		return canonicalKey{addr: addr, typ: targetVal.Type()}, true
	}

	return canonicalKey{}, false
}

// processValue updates reference counts and statistics for a single value encountered during the scan.
func (d *Dumper) processValue(v reflect.Value, level int) {
	isPtr := isPointerRef(v)
	// Treat non-nil slices and maps as if they were pointers for reference counting purposes.
	if !isPtr {
		k := v.Kind()
		if (k == reflect.Slice || k == reflect.Map) && !v.IsNil() {
			isPtr = true
		}
	}

	targetVal := deref(v)

	if targetVal.IsValid() {
		key, ok := d.getRawKey(targetVal)
		if !ok {
			return
		}
		stats := d.getOrCreateStats(key, targetVal, level)
		stats.totalReferencesCount++

		if isPtr {
			stats.pointerReferencesCount++
			if stats.minPointerRefLevel == -1 || level < stats.minPointerRefLevel {
				stats.minPointerRefLevel = level
			}
		}

		if stats.isPrimitive {
			if _, exists := d.primitiveInstances[key]; !exists {
				d.primitiveInstances[key] = stats.value
			}
		}
	}
}

// preScanBFS is the first analysis pass. It traverses the entire object graph using BFS
// to collect statistics about every value, such as reference counts and depth.
func (d *Dumper) preScanBFS(val reflect.Value) {
	queue := []queueItem{{val, 0}}
	traversed := make(map[canonicalKey]bool)
	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]
		if !item.v.IsValid() {
			continue
		}
		// Collect stats for the current value.
		d.processValue(item.v, item.level)

		// Continue traversal into composite types.
		targetVal := deref(item.v)
		if !targetVal.IsValid() || !isCompositeOrInterface(targetVal.Kind()) {
			continue
		}

		// Avoid re-traversing the same composite value.
		key, ok := d.getRawKey(targetVal)
		if !ok || traversed[key] {
			continue
		}
		traversed[key] = true
		queue = d.addChildrenToQueue(queue, item.v, item.level)
	}
}

// resetState clears all maps and slices used for reference tracking.
// It is called at the beginning of each top-level dump operation to ensure a clean slate.
func (d *Dumper) resetState() {
	d.referenceStats = make(map[canonicalKey]*RefStats)
	d.referenceIDs = make(map[canonicalKey]string)
	d.canonicalRoots = make(map[canonicalKey]canonicalKey)
	d.primitiveInstances = make(map[canonicalKey]any)
	d.definitionPoints = make(map[canonicalKey]definitionPoint)
	d.renderedIDs = make(map[canonicalKey]bool)
	d.fakeAddrs = make(map[any]uintptr)
}

// unifyAllCopies is the second analysis pass. It identifies values that are identical
// (e.g., a struct and a pointer to a copy of that struct) and merges them into a
// single logical group using the union-find structure.
func (d *Dumper) unifyAllCopies() {
	// Group values by their string representation. This is a heuristic to find potential copies.
	valueToKeys := make(map[string][]canonicalKey)
	for key, stats := range d.referenceStats {
		// Ignore zero-sized structs as they are always identical.
		if stats.valueKind == reflect.Struct && key.typ.Size() == 0 {
			continue
		}
		// NOTE: Using Sprintf is a heuristic. It's not foolproof but works well for many cases.
		valueStr := fmt.Sprintf("%#v", stats.value)
		valueToKeys[valueStr] = append(valueToKeys[valueStr], key)
	}

	for _, keys := range valueToKeys {
		if len(keys) < 2 {
			continue
		}

		// Separate keys into "sources" (referenced by a pointer) and "copies" (value-only).
		var sources, copies []canonicalKey
		for _, key := range keys {
			if stats, ok := d.referenceStats[key]; ok && stats.pointerReferencesCount > 0 {
				sources = append(sources, key)
			} else {
				copies = append(copies, key)
			}
		}

		isPrimitiveGroup := false
		if stats, ok := d.referenceStats[keys[0]]; ok {
			isPrimitiveGroup = stats.isPrimitive
		}

		if isPrimitiveGroup {
			if len(sources) > 1 && len(sources) == len(copies) {
				// Heuristic for ambiguous cases like TC#3.
				sort.Slice(sources, func(i, j int) bool { return sources[i].addr < sources[j].addr })
				sort.Slice(copies, func(i, j int) bool { return copies[i].addr < copies[j].addr })
				for i := 0; i < len(sources); i++ {
					d.union(sources[i], copies[i])
				}
			}
		} else { // Is a Composite Group
			if len(sources) > 1 && len(copies) == 0 {
				// Unify all sources together if they are identical reference types (fixes TC#19).
				rootSource := sources[0]
				for i := 1; i < len(sources); i++ {
					d.union(sources[i], rootSource)
				}
			}
		}

		// This general rule applies to both primitives and composites.
		if len(sources) == 1 && len(copies) > 0 {
			// Standard case: one source, multiple copies.
			rootSource := sources[0]
			for _, copyKey := range copies {
				d.union(copyKey, rootSource)
			}
		}
	}
}

// union merges the sets containing k1 and k2 in the union-find structure.
func (d *Dumper) union(k1, k2 canonicalKey) {
	root1, root2 := d.findRoot(k1), d.findRoot(k2)
	if root1 != root2 {
		// A simple heuristic to keep the tree balanced: merge smaller addr into larger.
		if root1.addr < root2.addr {
			d.canonicalRoots[root2] = root1
		} else {
			d.canonicalRoots[root1] = root2
		}
	}
}
