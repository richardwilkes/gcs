/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package model

import (
	"context"
	"io/fs"
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xmath/rand"
)

type nameGenCharThreshold struct {
	ch        rune
	threshold int
}

// NameGenerator holds the data for name generation.
type NameGenerator struct {
	initialized  bool
	Type         NameGenerationType `json:"type"`
	TrainingData []string           `json:"training_data"`
	min          int
	max          int
	entries      map[string][]nameGenCharThreshold
}

// NewNameGeneratorFromFS creates a new NameGenerator from a file.
func NewNameGeneratorFromFS(fileSystem fs.FS, filePath string) (*NameGenerator, error) {
	var generator NameGenerator
	if err := jio.LoadFromFS(context.Background(), fileSystem, filePath, &generator); err != nil {
		return nil, err
	}
	return &generator, nil
}

func (n *NameGenerator) initializeIfNeeded() {
	if !n.initialized {
		list := make([]string, 0, len(n.TrainingData))
		for _, one := range n.TrainingData {
			one = strings.ToLower(strings.TrimSpace(one))
			if utf8.RuneCountInString(one) >= 2 {
				list = append(list, one)
			}
		}
		n.TrainingData = list
		if n.Type == MarkovChainNameGenerationType {
			n.min = 20
			n.max = 2
			builders := make(map[string]map[rune]int)
			for _, one := range n.TrainingData {
				runes := []rune(one)
				length := len(runes)
				if n.min > length {
					n.min = length
				}
				if n.max < length {
					n.max = length
				}
				for i := 2; i < length; i++ {
					charGroup := string(runes[i-2 : i])
					occurrences, exists := builders[charGroup]
					if !exists {
						occurrences = make(map[rune]int)
						builders[charGroup] = occurrences
					}
					occurrences[runes[i]]++
				}
			}
			n.entries = make(map[string][]nameGenCharThreshold)
			for k, v := range builders {
				n.entries[k] = n.makeCharThresholdEntry(v)
			}
		}
		n.initialized = true
	}
}

// Generate a name.
func (n *NameGenerator) Generate() string {
	n.initializeIfNeeded()
	rnd := rand.NewCryptoRand()
	switch n.Type {
	case SimpleNameGenerationType:
		if len(n.TrainingData) == 0 {
			return ""
		}
		return txt.FirstToUpper(n.TrainingData[rnd.Intn(len(n.TrainingData))])
	case MarkovChainNameGenerationType:
		var buffer strings.Builder
		var sub []rune
		for k := range n.entries {
			buffer.WriteString(txt.FirstToUpper(k))
			sub = []rune(k)
			break // Only want one, which is random
		}
		targetSize := n.min + rnd.Intn(n.max+1-n.min)
		for i := 2; i < targetSize; i++ {
			entry, exists := n.entries[string(sub)]
			if !exists {
				break
			}
			next := n.chooseCharacter(entry)
			if next == 0 {
				break
			}
			buffer.WriteRune(next)
			//goland:noinspection GoNilness
			sub[0] = sub[1]
			//goland:noinspection GoNilness
			sub[1] = next
		}
		return buffer.String()
	default:
		return ""
	}
}

func (n *NameGenerator) makeCharThresholdEntry(occurrences map[rune]int) []nameGenCharThreshold {
	ct := make([]nameGenCharThreshold, len(occurrences))
	i := 0
	for k, v := range occurrences {
		ct[i].ch = k
		if i > 0 {
			v += ct[i-1].threshold
		}
		ct[i].threshold = v
		i++
	}
	return ct
}

func (n *NameGenerator) chooseCharacter(ct []nameGenCharThreshold) rune {
	threshold := rand.NewCryptoRand().Intn(ct[len(ct)-1].threshold + 1)
	for i := range ct {
		if ct[i].threshold >= threshold {
			return ct[i].ch
		}
	}
	return 0
}
