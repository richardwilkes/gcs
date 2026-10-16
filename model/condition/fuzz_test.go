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
	"math/rand/v2"
	"testing"
)

// FuzzDecode checks, for any input, that decoding never panics and either fails with a Problem or gives a tree whose
// encoding decodes to a tree with the same encoding and hash, as does a clone. A decoded tree is then evaluated on a
// random host, seeded from the input, where every mode must agree on the state.
func FuzzDecode(f *testing.F) {
	for _, seed := range []string{
		`{"type":"group","mode":"at_least","n":1,"only_when":{"type":"const","value":true},"children":[{"type":"check","field":"name","text":{"compare":"is","qualifier":"{self.name}"}}]}`,
		`{"type":"count","of":"kids","where":{"type":"check","field":"level","number":{"compare":"at_least","qualifier":2}},"quantity":{"compare":"is","qualifier":2},"group_by":"tags","per_group":{"compare":"at_least","qualifier":1},"exclusive":true}`,
		`{"type":"group","mode":"all","exclusive":true,"children":[{"type":"count","of":"kids","quantity":{"compare":"at_least","qualifier":2}},{"type":"count","of":"kids"}]}`,
		`{"type":"compare","left":{"aggregate":"sum","of":"kids","field":"level","scale":2},"op":"at_most","right":{"sum":[{"attribute":"st"},{"value":1}]}}`,
		`{"type":"script","script":"met"}`,
		`{"type":"check","field":"tags","not":true,"text":{"compare":"is_not","qualifier":"a, b"}}`,
	} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		n, err := Decode(data)
		if err != nil {
			if _, ok := err.(Problem); !ok { //nolint:errorlint // Decode returns a Problem itself.
				t.Fatalf("not a Problem: %v", err)
			}
			return
		}
		first := encode(t, Ref{Node: n})
		again, err := Decode([]byte(first))
		if err != nil {
			t.Fatalf("decode of encoded data: %v\n%s", err, first)
		}
		if second := encode(t, Ref{Node: again}); first != second {
			t.Fatalf("not idempotent:\n%s\n%s", first, second)
		}
		if nodeHash(n) != nodeHash(again) || nodeHash(n) != nodeHash(CloneNode(n)) {
			t.Fatalf("hash changed:\n%s", first)
		}
		var r Ref
		if err = json.Unmarshal(data, &r); err != nil || r.Err() != nil {
			t.Fatalf("a Ref fails where Decode doesn't: %v %v", err, r.Err())
		}
		if n != nil {
			Validate(n)
			n.FillWithNameableKeys(make(map[string]string), nil)
			h, self := randomHost(rand.New(rand.NewPCG(uint64(len(data)), nodeHash(n)))) //nolint:gosec // Not for security.
			evalAll(t, n, h, self)
		}
	})
}
