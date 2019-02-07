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

package k8s

import (
	"fmt"

	"github.com/skydive-project/skydive/graffiti/graph"
	"github.com/skydive-project/skydive/probe"

	"k8s.io/api/apps/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type statefulSetHandler struct {
}

func (h *statefulSetHandler) Dump(obj interface{}) string {
	ss := obj.(*v1beta1.StatefulSet)
	return fmt.Sprintf("statefulset{Namespace: %s, Name: %s}", ss.Namespace, ss.Name)
}

func (h *statefulSetHandler) Map(obj interface{}) (graph.Identifier, graph.Metadata) {
	ss := obj.(*v1beta1.StatefulSet)

	m := NewMetadataFields(&ss.ObjectMeta)
	m.SetField("DesiredReplicas", int32ValueOrDefault(ss.Spec.Replicas, 1))
	m.SetField("ServiceName", ss.Spec.ServiceName) // FIXME: replace by link to Service
	m.SetField("Replicas", ss.Status.Replicas)
	m.SetField("ReadyReplicas", ss.Status.ReadyReplicas)
	m.SetField("CurrentReplicas", ss.Status.CurrentReplicas)
	m.SetField("UpdatedReplicas", ss.Status.UpdatedReplicas)
	m.SetField("CurrentRevision", ss.Status.CurrentRevision)
	m.SetField("UpdateRevision", ss.Status.UpdateRevision)

	return graph.Identifier(ss.GetUID()), NewMetadata(Manager, "statefulset", m, ss, ss.Name)
}

func newStatefulSetProbe(client interface{}, g *graph.Graph) Subprobe {
	return NewResourceCache(client.(*kubernetes.Clientset).AppsV1beta1().RESTClient(), &v1beta1.StatefulSet{}, "statefulsets", g, &statefulSetHandler{})
}

func statefulSetPodAreLinked(a, b interface{}) bool {
	return matchLabelSelector(b.(*v1.Pod), a.(*v1beta1.StatefulSet).Spec.Selector)
}

func statefuleSetPodMetadata(a, b interface{}, typeA, typeB, manager string) graph.Metadata {
        return NewEdgeMetadata(manager, typeA)
}

func newStatefulSetPodLinker(g *graph.Graph) probe.Probe {
	return NewABLinker(g, Manager, "statefulset", Manager, "pod", statefuleSetPodMetadata, statefulSetPodAreLinked)
}
