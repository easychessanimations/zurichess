// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build coach

package engine

import (
	. "bitbucket.org/zurichess/zurichess/board"
	"sync"
)

var (
	FeaturesMap     = make(map[string]*FeatureInfo)
	featuresMapLock sync.Mutex
)

type FeatureInfo struct {
	Name  string // Name of this feature.
	Start int    // Start position in the weights array.
	Num   int    // Number of weights used.
}

func getFeatureStart(feature string, num int) int {
	featuresMapLock.Lock()
	defer featuresMapLock.Unlock()

	if info, has := FeaturesMap[feature]; has {
		return info.Start
	}
	FeaturesMap[feature] = &FeatureInfo{
		Name:  feature,
		Start: len(Weights),
		Num:   num,
	}
	for i := 0; i < num; i++ {
		Weights = append(Weights, Score{M: 0, E: 0, I: len(Weights)})
	}
	return FeaturesMap[feature].Start
}

func groupByBoard(feature string, bb Bitboard, accum *Accum) {
	start := getFeatureStart(feature, 1)
	accum.addN(Weights[start], bb.Count())
}

func groupByFileSq(feature string, sq Square, accum *Accum) {
	start := getFeatureStart(feature, 8)
	accum.add(Weights[start+sq.File()])
}

func groupByRankSq(feature string, sq Square, accum *Accum) {
	start := getFeatureStart(feature, 8)
	accum.add(Weights[start+sq.Rank()])
}

func groupByRank(feature string, bb Bitboard, accum *Accum) {
	for bb != BbEmpty {
		sq := bb.Pop()
		groupByRankSq(feature, sq, accum)
	}
}
