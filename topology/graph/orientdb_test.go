/*
 * Copyright (C) 2017 Red Hat, Inc.
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
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/skydive-project/skydive/common"
	"github.com/skydive-project/skydive/filters"
	"github.com/skydive-project/skydive/storage/orientdb"
)

type op struct {
	name string
	data interface{}
}

type fakeOrientDBClient struct {
	ops    []op
	result orientdb.Result
}

func (f *fakeOrientDBClient) getOps() []op {
	return f.ops
}

func (f *fakeOrientDBClient) Request(method string, url string, body io.Reader) (*http.Response, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) DeleteDocument(id string) error {
	return nil
}
func (f *fakeOrientDBClient) GetDocument(id string) (*orientdb.Result, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) CreateDocument(doc interface{}) (*orientdb.Result, error) {
	// reconvert data into map as it is in json.RawMessage
	rd := make(map[string]interface{})
	if err := json.Unmarshal([]byte(doc.(json.RawMessage)), &rd); err != nil {
		return nil, err
	}

	f.ops = append(f.ops, op{name: "CreateDocument", data: rd})
	return nil, nil
}
func (f *fakeOrientDBClient) Upsert(class string, doc interface{}, idkey string, idval string) (*orientdb.Result, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) GetDocumentClass(name string) (*orientdb.DocumentClass, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) AlterProperty(className string, prop orientdb.Property) error {
	return nil
}
func (f *fakeOrientDBClient) CreateProperty(className string, prop orientdb.Property) error {
	return nil
}
func (f *fakeOrientDBClient) CreateClass(class orientdb.ClassDefinition) error {
	return nil
}
func (f *fakeOrientDBClient) CreateIndex(className string, index orientdb.Index) error {
	return nil
}
func (f *fakeOrientDBClient) CreateDocumentClass(class orientdb.ClassDefinition) error {
	return nil
}
func (f *fakeOrientDBClient) DeleteDocumentClass(name string) error {
	return nil
}
func (f *fakeOrientDBClient) GetDatabase() (*orientdb.Result, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) CreateDatabase() (*orientdb.Result, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) SQL(query string) (*orientdb.Result, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) Search(query string) (*orientdb.Result, error) {
	f.ops = append(f.ops, op{name: "Search", data: query})
	return &f.result, nil
}
func (f *fakeOrientDBClient) Query(obj string, query *filters.SearchQuery) (*orientdb.Result, error) {
	return nil, nil
}
func (f *fakeOrientDBClient) Connect() error {
	return nil
}

func newOrientDBGraph(t *testing.T) (*Graph, *fakeOrientDBClient) {
	client := &fakeOrientDBClient{}
	b, err := newOrientDBBackend(client)

	if err != nil {
		t.Error(err)
	}

	return NewGraphFromConfig(b, common.UnknownService), client
}

// test history when doing local modification
func TestLocalHistory(t *testing.T) {
	g, client := newOrientDBGraph(t)

	client.result.Body = []byte(`{"result": [{"value": 1}]}`)

	node := g.CreateNode("aaa", Metadata{"MTU": 1500}, Unix(1, 0), "host1")
	g.AddNode(node)
	g.addMetadata(node, "MTU", 1510, Unix(2, 0))

	origin := common.UnknownService.String() + ".host1"

	expected := []op{
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(1000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(1),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1500),
				},
			},
		},
		{
			name: "Search",
			data: "UPDATE Node SET ArchivedAt = 2000 WHERE DeletedAt IS NULL AND ArchivedAt IS NULL AND ID = 'aaa'",
		},
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(2000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(2),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1510),
				},
			},
		},
	}

	if !reflect.DeepEqual(client.getOps()[0].data, expected[0].data) {
		t.Fatalf("Expected orientdb records not found: \nexpected: %s\ngot: %s", spew.Sdump(expected), spew.Sdump(client.getOps()))
	}

	g.addMetadata(node, "MTU", 1520, Unix(3, 0))

	expected = []op{
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(1000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(1),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1500),
				},
			},
		},
		{
			name: "Search",
			data: "UPDATE Node SET ArchivedAt = 2000 WHERE DeletedAt IS NULL AND ArchivedAt IS NULL AND ID = 'aaa'",
		},
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(2000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(2),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1510),
				},
			},
		},
		{
			name: "Search",
			data: "UPDATE Node SET ArchivedAt = 3000 WHERE DeletedAt IS NULL AND ArchivedAt IS NULL AND ID = 'aaa'",
		},
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(3000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(3),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1520),
				},
			},
		},
	}

	if !reflect.DeepEqual(client.getOps(), expected) {
		t.Fatalf("Expected orientdb records not found: \nexpected: %s\ngot: %s", spew.Sdump(expected), spew.Sdump(client.getOps()))
	}

	client.result.Body = []byte(`{"result": [{"ID": "bbb", "Parent": "123", "Child": "456"}]}`)

	g.delNode(node, Unix(4, 0))

	expected = []op{
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(1000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(1),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1500),
				},
			},
		},
		{
			name: "Search",
			data: "UPDATE Node SET ArchivedAt = 2000 WHERE DeletedAt IS NULL AND ArchivedAt IS NULL AND ID = 'aaa'",
		},
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(2000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(2),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1510),
				},
			},
		},
		{
			name: "Search",
			data: "UPDATE Node SET ArchivedAt = 3000 WHERE DeletedAt IS NULL AND ArchivedAt IS NULL AND ID = 'aaa'",
		},
		{
			name: "CreateDocument",
			data: map[string]interface{}{
				"UpdatedAt": float64(3000),
				"CreatedAt": float64(1000),
				"DeletedAt": nil,
				"Revision":  float64(3),
				"@class":    "Node",
				"ID":        "aaa",
				"Host":      "host1",
				"Origin":    origin,
				"Metadata": map[string]interface{}{
					"MTU": float64(1520),
				},
			},
		},
		{
			name: "Search",
			data: "SELECT FROM Link WHERE ArchivedAt is NULL AND (Parent = 'aaa' OR Child = 'aaa')",
		},
		{
			name: "Search",
			data: "UPDATE Link SET DeletedAt = 4000, ArchivedAt = 4000 WHERE DeletedAt IS NULL AND ArchivedAt IS NULL AND ID = 'bbb'",
		},
		{
			name: "Search",
			data: "UPDATE Node SET DeletedAt = 4000, ArchivedAt = 4000 WHERE DeletedAt IS NULL AND ArchivedAt IS NULL AND ID = 'aaa'",
		},
	}

	if !reflect.DeepEqual(client.getOps(), expected) {
		t.Fatalf("Expected orientdb records not found: \nexpected: %s\ngot: %s", spew.Sdump(expected), spew.Sdump(client.getOps()))
	}
}
