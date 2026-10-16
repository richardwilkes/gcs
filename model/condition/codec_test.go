// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"encoding/json/v2"
	"errors"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/check"
)

// nested returns levels groups, each holding the next, around leaf.
func nested(levels int, leaf string) string {
	return strings.Repeat(`{"type":"group","mode":"all","children":[`, levels) + leaf + strings.Repeat("]}", levels)
}

// TestDecodeRoundTrip checks that each input decodes to a tree that encodes in its canonical form, which decodes back
// to itself and hashes the same as the input.
func TestDecodeRoundTrip(t *testing.T) {
	deepest := nested(MaxDepth-1, `{"type":"const","value":true}`)
	for _, tc := range []struct {
		name, in, out string
	}{
		{"const", `{"type":"const","value":false}`, ""},
		{"whitespace", ` { "type" : "const" , "value" : true } `, `{"type":"const","value":true}`},
		{"check", `{"type":"check","field":"hidden"}`, ""},
		{"check text", `{"type":"check","field":"tags","not":true,"text":{"compare":"contains","qualifier":"@Weapon@"}}`, ""},
		{"check number", `{"type":"check","field":"level","number":{"compare":"at_least","qualifier":2.5}}`, ""},
		{
			"check weight", `{"type":"check","field":"weight","weight":{"compare":"at_most","qualifier":"10.79 lb"}}`,
			`{"type":"check","field":"weight","weight":{"compare":"at_most","qualifier":"10.79 lb"}}`,
		},
		{
			"keys ignore case", `{"type":"check","field":"name","text":{"compare":"Starts_With","qualifier":"F"}}`,
			`{"type":"check","field":"name","text":{"compare":"starts_with","qualifier":"F"}}`,
		},
		{
			"group",
			`{"type":"group","mode":"at_least","n":2,"not":true,"only_when":{"type":"const","value":true},"children":[{"type":"const","value":true},{"type":"script","script":"x"},{"type":"group","mode":"any"}]}`,
			"",
		},
		{
			"empty guard dropped", `{"type":"group","mode":"all","only_when":{"type":"group","mode":"any"},"children":[{"type":"const","value":true}]}`,
			`{"type":"group","mode":"all","children":[{"type":"const","value":true}]}`,
		},
		{"empty root", `{"type":"group","mode":"any","not":true}`, "null"},
		{"null", `null`, ""},
		{
			"count",
			`{"type":"count","of":"spells","include_self":true,"where":{"type":"check","field":"college","text":{"compare":"is","qualifier":"Fire"}},"quantity":{"compare":"at_least","qualifier":5},"group_by":"college","per_group":{"compare":"at_least","qualifier":2},"exclusive":true,"not":true}`,
			"",
		},
		{
			"count at least 1", `{"type":"count","of":"spells","quantity":{"compare":"at_least","qualifier":1},"per_group":{"compare":"at_least","qualifier":1},"group_by":"college"}`,
			`{"type":"count","of":"spells","group_by":"college"}`,
		},
		{"count any", `{"type":"count","of":"spells","quantity":{}}`, ""},
		{
			"compare",
			`{"type":"compare","left":{"aggregate":"sum","of":"traits","include_self":true,"where":{"type":"check","field":"name"},"field":"level","scale":0,"offset":-1},"op":"at_most","right":{"sum":[{"attribute":"iq"},{"entity":"tech_level"},{"value":0},{"self":"level"}]}}`,
			"",
		},
		{
			"compare weight", `{"type":"compare","left":{"weight":"2 lb","scale":2},"op":"less_than","right":{"aggregate":"max","of":"carried","field":"weight"}}`,
			`{"type":"compare","left":{"weight":"2 lb","scale":2},"op":"less_than","right":{"aggregate":"max","of":"carried","field":"weight"}}`,
		},
		{"absent operand", `{"type":"compare","op":"is","right":{"value":1}}`, `{"type":"compare","left":{},"op":"is","right":{"value":1}}`},
		{
			"script", `{"type":"script","name":"Has @Thing@","script":"entity.has(\"@Thing@\")","lang":"cel"}`,
			`{"type":"script","name":"Has @Thing@","script":"entity.has(\"@Thing@\")","lang":"cel"}`,
		},
		{"deepest", deepest, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			want := tc.out
			if want == "" {
				want = strings.Join(strings.Fields(tc.in), "")
			}
			n := decode(t, tc.in)
			got := encode(t, Ref{Node: n})
			c.Equal(want, got)
			again := decode(t, got)
			c.Equal(got, encode(t, Ref{Node: again}))
			c.Equal(nodeHash(n), nodeHash(again))
		})
	}
}

// TestDecodeErrors checks that a tree with anything decoding doesn't understand is rejected, with a Problem naming
// where and what.
func TestDecodeErrors(t *testing.T) {
	tooDeep := strings.TrimSuffix(strings.Repeat("children[0].", MaxDepth), ".")
	for _, tc := range []struct {
		name, in, want string
	}{
		{"not json", `{"type"`, "not valid JSON"},
		{"not an object", `[1,2,3]`, "a node is not an object"},
		{"no type", `{"value":true}`, `missing "type"`},
		{"type not a string", `{"type":5}`, `invalid value for "type"`},
		{"unknown type", `{"type":"future","x":1}`, `unknown node type "future"`},
		{"unknown member", `{"type":"check","field":"name","future":1}`, `unknown member "future"`},
		{"wrong kind", `{"type":"check","field":5}`, `invalid value for "field"`},
		{
			"criterion member", `{"type":"check","field":"x","text":{"compare":"is","qualifier":"a","case":true}}`,
			`unknown member "text.case"`,
		},
		{
			"unknown comparison", `{"type":"check","field":"x","text":{"compare":"matches","qualifier":"F.*"}}`,
			`unknown "text.compare" value "matches"`,
		},
		{"no comparison", `{"type":"check","field":"x","number":{"qualifier":3}}`, `missing "number.compare"`},
		{"compares nothing", `{"type":"check","field":"x","weight":{"compare":""}}`, `"weight" compares nothing`},
		{
			"bad qualifier", `{"type":"check","field":"x","number":{"compare":"is","qualifier":"many"}}`,
			`invalid value for "number.qualifier"`,
		},
		{"no mode", `{"type":"group","children":[]}`, `missing "mode"`},
		{"group member", `{"type":"group","mode":"all","children":[],"kids":[]}`, `unknown member "kids"`},
		{"compare member", `{"type":"compare","op":"is","not":true}`, `unknown member "not"`},
		{"script member", `{"type":"script","script":"x","language":"js"}`, `unknown member "language"`},
		{"const member", `{"type":"const","value":true,"legacy":{}}`, `unknown member "legacy"`},
		{"unknown mode", `{"type":"group","mode":"most"}`, `unknown "mode" value "most"`},
		{
			"bad child", `{"type":"group","mode":"all","children":[{"type":"const","value":true},{"type":"nope"}]}`,
			`children[1]: unknown node type "nope"`,
		},
		{"null child", `{"type":"group","mode":"any","children":[null]}`, "children[0]: missing node"},
		{
			"bad guard", `{"type":"group","mode":"all","only_when":{"type":"const","value":"yes"}}`,
			`only_when: invalid value for "value"`,
		},
		{"count member", `{"type":"count","of":"x","where":{"type":"check"},"exclusiv":true}`, `unknown member "exclusiv"`},
		{
			"nested where", `{"type":"count","of":"x","where":{"type":"count","of":"y","where":[]}}`,
			"where.where: a node is not an object",
		},
		{
			"unknown quantity comparison", `{"type":"count","of":"x","quantity":{"compare":"around","qualifier":3}}`,
			`unknown "quantity.compare" value "around"`,
		},
		{
			"quantity qualifier only", `{"type":"count","of":"x","per_group":{"qualifier":3}}`,
			`"per_group" has a qualifier but compares nothing`,
		},
		{"no op", `{"type":"compare","left":{"value":1},"right":{"value":2}}`, `missing "op"`},
		{
			"unknown op", `{"type":"compare","left":{"value":1},"op":"approximately","right":{"value":2}}`,
			`unknown "op" value "approximately"`,
		},
		{
			"operand not an object", `{"type":"compare","left":7,"op":"is","right":{"value":2}}`,
			"left: an operand is not an object",
		},
		{
			"operand member", `{"type":"compare","left":{"aggr":"count","of":"x"},"op":"is","right":{"value":2}}`,
			`left: unknown member "aggr"`,
		},
		{
			"unknown aggregate", `{"type":"compare","left":{"value":1},"op":"is","right":{"aggregate":"min","of":"x","field":"y"}}`,
			`right: unknown "aggregate" value "min"`,
		},
		{
			"operand where", `{"type":"compare","left":{"aggregate":"count","of":"x","where":{"type":"x"}},"op":"is","right":{"value":2}}`,
			`left.where: unknown node type "x"`,
		},
		{
			"sum", `{"type":"compare","left":{"value":1},"op":"is","right":{"sum":[{"value":1},{"value":true}]}}`,
			`right.sum[1]: invalid value for "value"`,
		},
		{"too deep", nested(MaxDepth, `{"type":"const","value":true}`), tooDeep + ": nested more than 32 deep"},
		{
			"operand too deep", nested(MaxDepth-2, `{"type":"compare","left":{"sum":[{"value":1}]},"op":"is","right":{"value":1}}`),
			strings.Repeat("children[0].", MaxDepth-2) + "left.sum[0]: nested more than 32 deep",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			n, err := Decode([]byte(tc.in))
			c.Nil(n)
			var p Problem
			c.True(errors.As(err, &p))
			c.Equal(tc.want, err.Error())
		})
	}
}

// TestBrokenRef checks that a Ref whose tree can't be decoded doesn't fail what holds it, and that it keeps what it
// read, writes it back as it was, reports the error, and evaluates as Failed with the error as the reason.
func TestBrokenRef(t *testing.T) {
	c := check.New(t)
	var holder struct {
		Before string `json:"before"`
		Tree   Ref    `json:"tree"`
		After  string `json:"after"`
	}
	in := `{"before":"a","tree":{"type":"group","mode":"all","children":[{"type":"nope","x":[1,{}]}]},"after":"b"}`
	c.NoError(json.Unmarshal([]byte(in), &holder))
	c.Equal("a", holder.Before)
	c.Equal("b", holder.After)
	c.Equal(`children[0]: unknown node type "nope"`, holder.Tree.Err().Error())
	c.Equal(in, encode(t, holder))

	r := eval(t, holder.Tree.Node, newHost(), nil, collect)
	c.Equal(Failed, r.State)
	c.Equal(holder.Tree.Err().Error(), r.Reason)
	c.Equal(r.Reason, holder.Tree.Node.eval(nil, frame{}).Reason)
	c.Equal([]Problem{{Path: Path{{Kind: SlotChild}}, Msg: `unknown node type "nope"`}}, Validate(holder.Tree.Node))

	clone := CloneNode(holder.Tree.Node)
	c.Equal(nodeHash(holder.Tree.Node), nodeHash(clone))
	c.Equal(encode(t, holder.Tree), encode(t, Ref{Node: clone}))
	var other Ref
	c.NoError(json.Unmarshal([]byte(`{"type":"other"}`), &other))
	c.NotEqual(nodeHash(holder.Tree.Node), nodeHash(other.Node))
	c.Equal(`unknown node type "other"`, other.Err().Error())
	c.Equal(1, len(Validate(other.Node)))
	other.Node.FillWithNameableKeys(nil, nil)
	Walk(other.Node, func(Path, Node) bool { return true })

	var good Ref
	c.NoError(json.Unmarshal([]byte(`{"type":"const","value":true}`), &good))
	c.NoError(good.Err())
	c.NoError(json.Unmarshal([]byte(`null`), &good))
	c.Nil(good.Node)
	c.NotNil(json.Unmarshal([]byte(`{"tree":`), &holder))
}

// TestEncodeErrors checks that a mode or operator outside its enum, which only code can build, fails to encode
// rather than being written as something else.
func TestEncodeErrors(t *testing.T) {
	c := check.New(t)
	for _, n := range []Node{
		&Group{Mode: LastGroupMode + 1},
		&Compare{Left: value(1), Op: LastCompareOp + 1, Right: value(1)},
		&Group{Children: Nodes{&Group{Mode: LastGroupMode + 1}}},
	} {
		_, err := json.Marshal(n)
		c.NotNil(err)
	}
	c.Equal(`{"type":"group","mode":"all","children":[null]}`, encode(t, &Group{Children: Nodes{nil}}))
}

// TestCanonicalEquivalents checks that forms meaning the same thing encode and hash the same.
func TestCanonicalEquivalents(t *testing.T) {
	c := check.New(t)
	stale := criteria.Text{Qualifier: "stale"}
	for _, pair := range [][2]Node{
		{&Count{Of: "x"}, &Count{Of: "x", Quantity: atLeast(1)}},
		{&Count{Of: "x", GroupBy: "y"}, &Count{Of: "x", GroupBy: "y", PerGroup: atLeast(1)}},
		{&Check{Field: "x"}, &Check{Field: "x", Text: stale}},
		{&Count{Of: "x"}, &Count{Of: "x", Where: Ref{Node: &Group{Mode: GroupAny}}}},
	} {
		c.Equal(encode(t, pair[0]), encode(t, pair[1]))
		c.Equal(nodeHash(pair[0]), nodeHash(pair[1]))
	}
	c.Equal(nodeHash(nil), hashWith(Ref{Node: &Group{}}.hash))
	c.NotEqual(nodeHash(&Count{Of: "x"}), nodeHash(&Count{Of: "x", Quantity: &criteria.Number{}}))
}
