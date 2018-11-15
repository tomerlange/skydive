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

package peering

import (
	"github.com/skydive-project/skydive/topology"
	"github.com/skydive-project/skydive/topology/graph"
)

// Probe describes graph peering based on MAC address and graph events
type Probe struct {
	graph              *graph.Graph
	peerIntfMACIndexer *graph.MetadataIndexer
	macIndexer         *graph.MetadataIndexer
	linker             *graph.MetadataIndexerLinker
}

// Start the MAC peering resolver probe
func (p *Probe) Start() {
	p.peerIntfMACIndexer.Start()
	p.macIndexer.Start()
	p.linker.Start()
}

// Stop the probe
func (p *Probe) Stop() {
	p.peerIntfMACIndexer.Stop()
	p.macIndexer.Stop()
	p.linker.Stop()
}

// NewProbe creates a new graph node peering probe
func NewProbe(g *graph.Graph) *Probe {
	peerIntfMACIndexer := graph.NewMetadataIndexer(g, g, nil, "PeerIntfMAC")
	macIndexer := graph.NewMetadataIndexer(g, g, nil, "MAC")
	probe := &Probe{
		graph:              g,
		peerIntfMACIndexer: peerIntfMACIndexer,
		macIndexer:         macIndexer,
		linker:             graph.NewMetadataIndexerLinker(g, peerIntfMACIndexer, macIndexer, graph.Metadata{"RelationType": topology.Layer2Link}),
	}

	return probe
}
