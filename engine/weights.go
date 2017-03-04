// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was generated by bitbucket.org/zurichess/tuner.

// +build !coach

package engine

// Weights stores the network parameters.
// Network has train error 0.06266765 and validation error 0.06288316.
var Weights = [...]Score{
	{M: 15, E: 0}, {M: 9364, E: 17330}, {M: 60891, E: 46088}, {M: 69074, E: 50978}, {M: 79204, E: 93681}, {M: 194980, E: 157978}, {M: 43, E: -21}, {M: -4204, E: -2060},
	{M: -147, E: -85}, {M: 11, E: 1944}, {M: 1780, E: 3197}, {M: 3134, E: 2151}, {M: 2414, E: 684}, {M: 1753, E: -165}, {M: -150, E: -3398}, {M: -193, E: -3561},
	{M: -42, E: -66}, {M: 778, E: 1728}, {M: 3113, E: 3992}, {M: 3640, E: 3951}, {M: 922, E: 2256}, {M: -536, E: -39}, {M: -1, E: -3822}, {M: 1708, E: 327},
	{M: -3959, E: -663}, {M: 659, E: -37}, {M: 842, E: 617}, {M: -852, E: 1514}, {M: 52, E: 1579}, {M: -1, E: 729}, {M: 3015, E: -928}, {M: -2199, E: -603},
	{M: 16, E: -1966}, {M: 266, E: -367}, {M: 809, E: 922}, {M: -910, E: 1128}, {M: -886, E: 1324}, {M: 787, E: 1096}, {M: -5, E: -18}, {M: -426, E: -1180},
	{M: 1727, E: 300}, {M: -1041, E: 525}, {M: -248, E: 947}, {M: -48, E: 1146}, {M: 1815, E: 9}, {M: 2325, E: -926}, {M: 3724, E: -1229}, {M: -506, E: 25},
	{M: 1101, E: -1982}, {M: 2184, E: -1371}, {M: -33, E: 918}, {M: 16, E: 38}, {M: -8, E: -6}, {M: 15, E: -51}, {M: 102, E: 249}, {M: -300, E: 1494},
	{M: 2069, E: -729}, {M: 1587, E: 330}, {M: 8619, E: -1476}, {M: 2805, E: 11}, {M: -1937, E: -3407}, {M: -2624, E: -1333}, {M: -212, E: -1865}, {M: -2, E: 58},
	{M: 178, E: 1768}, {M: 266, E: 1773}, {M: 2343, E: 1064}, {M: 2336, E: 2077}, {M: 1565, E: -1843}, {M: 399, E: -338}, {M: 242, E: 1201}, {M: -2476, E: 4008},
	{M: -1965, E: 4476}, {M: -14, E: 1607}, {M: -19, E: -58}, {M: 2037, E: -1304}, {M: 652, E: 1315}, {M: 90, E: -2}, {M: -35, E: -36}, {M: 15, E: 16},
	{M: 212, E: -53}, {M: 111, E: -68}, {M: -63, E: 204}, {M: -33, E: 66}, {M: -35, E: -57}, {M: 1081, E: -286}, {M: 902, E: -284}, {M: 571, E: 172},
	{M: 662, E: 638}, {M: 747, E: 862}, {M: 1687, E: -50}, {M: 1797, E: -265}, {M: 1261, E: -820}, {M: 937, E: 68}, {M: 79, E: 434}, {M: 725, E: 5},
	{M: 547, E: 470}, {M: 1062, E: -72}, {M: 1237, E: -89}, {M: 1707, E: -462}, {M: 1306, E: -34}, {M: 889, E: 854}, {M: 799, E: 457}, {M: 1094, E: 113},
	{M: 1980, E: -559}, {M: 1863, E: -417}, {M: 1478, E: -328}, {M: 783, E: 325}, {M: 857, E: 809}, {M: 891, E: 242}, {M: 423, E: 115}, {M: 1146, E: -447},
	{M: 1757, E: -762}, {M: 2025, E: -769}, {M: 1506, E: -1102}, {M: 414, E: -143}, {M: 1025, E: 3}, {M: 845, E: -812}, {M: 216, E: -189}, {M: 886, E: -895},
	{M: 646, E: -156}, {M: 1112, E: 259}, {M: 558, E: -227}, {M: 1400, E: -370}, {M: 1357, E: -672}, {M: 693, E: -221}, {M: 336, E: -159}, {M: -10, E: 639},
	{M: 1241, E: -1356}, {M: 884, E: 20}, {M: 3205, E: -221}, {M: 2254, E: -3868}, {M: 36, E: 59}, {M: -115, E: -89}, {M: 173, E: -30}, {M: 31, E: 161},
	{M: -107, E: 22}, {M: -18, E: -111}, {M: 41, E: -31}, {M: -13, E: 113}, {M: 144, E: 116}, {M: -2957, E: -628}, {M: 1973, E: -945}, {M: -2660, E: -2616},
	{M: -2109, E: -1744}, {M: -5177, E: -415}, {M: 4418, E: -1107}, {M: 2959, E: 1001}, {M: -6735, E: 2922}, {M: 1773, E: 1478}, {M: -2497, E: 2772}, {M: 6806, E: -578},
	{M: -1830, E: -426}, {M: -656, E: -3024}, {M: 2870, E: -683}, {M: -37, E: 1903}, {M: -399, E: 2749}, {M: -4, E: 2594}, {M: -30, E: 2290}, {M: 2894, E: -66},
	{M: 1, E: -1886}, {M: -3546, E: 1781}, {M: 3578, E: -1441}, {M: -28, E: 23}, {M: -17, E: 12093}, {M: 901, E: 9561}, {M: 798, E: 6261}, {M: 1363, E: 6378},
	{M: 909, E: 9892}, {M: 57, E: 12588}, {M: 139, E: -71},
}
