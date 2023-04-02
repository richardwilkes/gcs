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
	"bufio"
	"bytes"
	"io/fs"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/richardwilkes/json"
	"github.com/richardwilkes/rpgtools/names"
	"github.com/richardwilkes/rpgtools/names/namesets"
	"github.com/richardwilkes/rpgtools/names/namesets/american"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/txt"
	"golang.org/x/exp/maps"
)

const unweighted = "unweighted"

// NewNameGeneratorFromFS creates a new NameGenerator from a file.
func NewNameGeneratorFromFS(fileSystem fs.FS, filePath string) (names.Namer, error) {
	data, err := fs.ReadFile(fileSystem, filePath)
	if err != nil {
		return nil, errs.Wrap(err)
	}
	if len(data) == 0 {
		return nil, errs.New("file is empty")
	}

	if data[0] == '{' {
		var oldNameGenData struct {
			Type         string   `json:"type"`
			TrainingData []string `json:"training_data"`
		}
		if err = json.Unmarshal(data, &oldNameGenData); err != nil {
			return nil, errs.Wrap(err)
		}
		if oldNameGenData.Type == "markov_chain" {
			return names.NewMarkovLetterUnweightedNamer(3, oldNameGenData.TrainingData), nil
		}
		return names.NewSimpleUnweightedNamer(oldNameGenData.TrainingData), nil
	}

	var line string
	line, data = nextNonEmptyLine(data)
	typeParts := strings.Split(line, ":")
	if len(typeParts) < 2 {
		return nil, errs.New("missing generator type")
	}
	line, data = nextNonEmptyLine(data)
	dataParts := strings.Split(line, ":")
	if len(dataParts) < 2 {
		return nil, errs.New("missing generator data")
	}
	var trainingData map[string]int
	switch typeParts[1] {
	case "simple":
		if trainingData, err = loadTrainingData(dataParts[1], data); err != nil {
			return nil, err
		}
		if len(typeParts) > 2 && typeParts[2] == unweighted {
			return names.NewSimpleUnweightedNamer(maps.Keys(trainingData)), nil
		}
		return names.NewSimpleNamer(trainingData), nil
	case "compound":
		if dataParts[1] != "inline" {
			return nil, errs.New("compound data kind must be 'inline'")
		}
		sep := " "
		lowered := false
		if len(typeParts) > 2 {
			if sep, err = url.PathUnescape(typeParts[2]); err != nil {
				return nil, errs.NewWithCause("invalid separator spec", err)
			}
			if len(typeParts) > 3 {
				lowered = txt.IsTruthy(typeParts[3])
			}
		}
		var namers []names.Namer
		scanner := bufio.NewScanner(bytes.NewReader(data))
		for scanner.Scan() {
			if line = strings.TrimSpace(scanner.Text()); line != "" {
				var n names.Namer
				if n, err = NewNameGeneratorFromFS(fileSystem, line); err != nil {
					return nil, errs.NewWithCause("unable to load name generator file: "+line, err)
				}
				namers = append(namers, n)
			}
		}
		if len(namers) == 0 {
			return nil, errs.New("must specify at least one name generation file for compound name generator")
		}
		return names.NewCompoundNamer(sep, lowered, namers...), nil
	case "markov_letter":
		depth := 3
		if len(typeParts) > 2 {
			if depth, err = strconv.Atoi(typeParts[2]); err != nil {
				return nil, errs.New("invalid depth option: " + typeParts[2])
			}
		}
		if trainingData, err = loadTrainingData(dataParts[1], data); err != nil {
			return nil, err
		}
		if len(typeParts) > 3 && typeParts[3] == unweighted {
			return names.NewMarkovLetterUnweightedNamer(depth, maps.Keys(trainingData)), nil
		}
		return names.NewMarkovLetterNamer(depth, trainingData), nil
	case "markov_run":
		if trainingData, err = loadTrainingData(dataParts[1], data); err != nil {
			return nil, err
		}
		if len(typeParts) > 2 && typeParts[2] == unweighted {
			return names.NewMarkovRunUnweightedNamer(maps.Keys(trainingData)), nil
		}
		return names.NewMarkovRunNamer(trainingData), nil
	default:
		return nil, errs.New("invalid name generator type: " + typeParts[1])
	}
}

func nextNonEmptyLine(in []byte) (line string, data []byte) {
	var buffer []byte
	data = in
	for {
		if i := bytes.IndexByte(data, '\n'); i == -1 {
			buffer = data
			data = nil
		} else {
			buffer = data[:i]
			data = data[i+1:]
		}
		buffer = bytes.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, buffer)
		if len(buffer) != 0 || data == nil {
			return string(buffer), data
		}
	}
}

func loadTrainingData(kind string, data []byte) (map[string]int, error) {
	switch kind {
	case "inline":
		trainingData, err := namesets.LoadFromReader(bytes.NewReader(data))
		if err != nil {
			return nil, errs.NewWithCause("invalid data", err)
		}
		return trainingData, nil
	case "american_male":
		return american.Male(), nil
	case "american_female":
		return american.Female(), nil
	case "american_last":
		return american.Last(), nil
	default:
		return nil, errs.New("invalid name generator data kind: " + kind)
	}
}
