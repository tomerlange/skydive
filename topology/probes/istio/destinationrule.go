/*
 * Copyright (C) 2018 IBM, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy ofthe License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specificlanguage governing permissions and
 * limitations under the License.
 *
 */

package istio

import (
	"fmt"

	kiali "github.com/kiali/kiali/kubernetes"
	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/probe"
	"github.com/skydive-project/skydive/topology/probes/k8s"
	"k8s.io/api/core/v1"
)

type destinationRuleHandler struct {
}

// Map graph node to k8s resource
func (h *destinationRuleHandler) Map(obj interface{}) (graph.Identifier, graph.Metadata) {
	dr := obj.(*kiali.DestinationRule)
	m := k8s.NewMetadataFields(&dr.ObjectMeta)
	return graph.Identifier(dr.GetUID()), k8s.NewMetadata(Manager, "destinationrule", m, dr, dr.Name)
}

// Dump k8s resource
func (h *destinationRuleHandler) Dump(obj interface{}) string {
	dr := obj.(*kiali.DestinationRule)
	return fmt.Sprintf("destinationrule{Namespace: %s, Name: %s}", dr.Namespace, dr.Name)
}

func newDestinationRuleProbe(client interface{}, g *graph.Graph) k8s.Subprobe {
	return k8s.NewResourceCache(client.(*kiali.IstioClient).GetIstioNetworkingApi(), &kiali.DestinationRule{}, "destinationrules", g, &destinationRuleHandler{})
}

type destinationRuleSpec struct {
	App string `mapstructure:"host"`
}

func destinationRuleServiceAreLinked(a, b interface{}) bool {
	dr := a.(*kiali.DestinationRule)
	service := b.(*v1.Service)
	return dr.Spec["host"] == service.Labels["app"]
}

func newDestinationRuleServiceLinker(g *graph.Graph) probe.Probe {
	return k8s.NewABLinker(g, Manager, "destinationrule", k8s.Manager, "service", destinationRuleServiceAreLinked)
}
