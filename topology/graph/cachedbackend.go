/*
 * Copyright (C) 2016 Red Hat, Inc.
 *
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 *
 */

package graph

import (
	"sync/atomic"

	"github.com/skydive-project/skydive/common"
)

const (
	CACHE_ONLY_MODE int = iota
	PERSISTENT_ONLY_MODE
	DEFAULT_MODE
)

type CachedBackend struct {
	memory     *MemoryBackend
	persistent GraphBackend
	cacheMode  atomic.Value
}

func (c *CachedBackend) SetMode(mode int) {
	c.cacheMode.Store(mode)
}

func (c *CachedBackend) NodeAdded(n *Node) bool {
	mode := c.cacheMode.Load()

	r := false
	if mode != PERSISTENT_ONLY_MODE {
		r = c.memory.NodeAdded(n)
	}

	if mode != CACHE_ONLY_MODE {
		r = c.persistent.NodeAdded(n)
	}

	return r
}

func (c *CachedBackend) NodeDeleted(n *Node) bool {
	mode := c.cacheMode.Load()

	r := false
	if mode != PERSISTENT_ONLY_MODE {
		r = c.memory.NodeDeleted(n)
	}

	if mode != CACHE_ONLY_MODE {
		r = c.persistent.NodeDeleted(n)
	}

	return r
}

func (c *CachedBackend) GetNode(i Identifier, t *common.TimeSlice) []*Node {
	mode := c.cacheMode.Load()

	if t == nil && mode != PERSISTENT_ONLY_MODE {
		return c.memory.GetNode(i, t)
	}

	if mode != CACHE_ONLY_MODE {
		return c.persistent.GetNode(i, t)
	}

	return nil
}

func (c *CachedBackend) GetNodeEdges(n *Node, t *common.TimeSlice, m Metadata) (edges []*Edge) {
	mode := c.cacheMode.Load()

	if t == nil && mode != PERSISTENT_ONLY_MODE {
		return c.memory.GetNodeEdges(n, t, m)
	}

	if mode != CACHE_ONLY_MODE {
		return c.persistent.GetNodeEdges(n, t, m)
	}

	return edges
}

func (c *CachedBackend) EdgeAdded(e *Edge) bool {
	mode := c.cacheMode.Load()

	r := false
	if mode != PERSISTENT_ONLY_MODE {
		r = c.memory.EdgeAdded(e)
	}

	if mode != CACHE_ONLY_MODE {
		r = c.persistent.EdgeAdded(e)
	}

	return r
}

func (c *CachedBackend) EdgeDeleted(e *Edge) bool {
	mode := c.cacheMode.Load()

	r := false
	if mode != PERSISTENT_ONLY_MODE {
		r = c.memory.EdgeDeleted(e)
	}

	if mode != CACHE_ONLY_MODE {
		r = c.persistent.EdgeDeleted(e)
	}

	return r
}

func (c *CachedBackend) GetEdge(i Identifier, t *common.TimeSlice) []*Edge {
	mode := c.cacheMode.Load()

	if t == nil && mode != PERSISTENT_ONLY_MODE {
		return c.memory.GetEdge(i, t)
	}

	if mode != CACHE_ONLY_MODE {
		return c.persistent.GetEdge(i, t)
	}

	return nil
}

func (c *CachedBackend) GetEdgeNodes(e *Edge, t *common.TimeSlice, parentMetadata, childMetadata Metadata) ([]*Node, []*Node) {
	mode := c.cacheMode.Load()

	if t == nil && mode != PERSISTENT_ONLY_MODE {
		return c.memory.GetEdgeNodes(e, t, parentMetadata, childMetadata)
	}

	if mode != CACHE_ONLY_MODE {
		return c.persistent.GetEdgeNodes(e, t, parentMetadata, childMetadata)
	}

	return nil, nil
}

func (c *CachedBackend) MetadataUpdated(i interface{}) bool {
	mode := c.cacheMode.Load()

	r := false
	if mode != CACHE_ONLY_MODE {
		r = c.persistent.MetadataUpdated(i)
	}

	if mode != PERSISTENT_ONLY_MODE {
		r = c.memory.MetadataUpdated(i)
	}

	return r
}

func (c *CachedBackend) GetNodes(t *common.TimeSlice, m Metadata) []*Node {
	mode := c.cacheMode.Load()

	if t == nil && mode != PERSISTENT_ONLY_MODE {
		return c.memory.GetNodes(t, m)
	}

	if mode != CACHE_ONLY_MODE {
		return c.persistent.GetNodes(t, m)
	}

	return []*Node{}
}

func (c *CachedBackend) GetEdges(t *common.TimeSlice, m Metadata) []*Edge {
	mode := c.cacheMode.Load()

	if t == nil && mode != PERSISTENT_ONLY_MODE {
		return c.memory.GetEdges(t, m)
	}

	if mode != CACHE_ONLY_MODE {
		return c.persistent.GetEdges(t, m)
	}

	return []*Edge{}
}

func (c *CachedBackend) WithContext(graph *Graph, context GraphContext) (*Graph, error) {
	return c.persistent.WithContext(graph, context)
}

func NewCachedBackend(persistent GraphBackend) (*CachedBackend, error) {
	memory, err := NewMemoryBackend()
	if err != nil {
		return nil, err
	}

	sb := &CachedBackend{
		persistent: persistent,
		memory:     memory,
	}

	return sb, nil
}
