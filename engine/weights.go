// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file was generated by bitbucket.org/zurichess/tuner.

// +build !coach

package engine

// Weights stores the network parameters.
// Network has train error 0.05706097 and validation error 0.05722470.
var Weights = [...]Score{
	{M: -129, E: 103}, {M: 13805, E: 13672}, {M: 64241, E: 48840}, {M: 66765, E: 52132}, {M: 85974, E: 96899}, {M: 206476, E: 163627}, {M: 31, E: -12}, {M: 1181, E: 2161},
	{M: -8071, E: -103}, {M: -9867, E: 1723}, {M: -2792, E: 220}, {M: -1717, E: 114}, {M: -4436, E: -1914}, {M: -511, E: -75}, {M: -13, E: 1855}, {M: 1347, E: 3271},
	{M: 1956, E: 2726}, {M: 2171, E: 1267}, {M: 964, E: 4}, {M: -1285, E: -3014}, {M: -288, E: -2540}, {M: -11, E: -81}, {M: 2575, E: 1585}, {M: 4671, E: 3014},
	{M: 5911, E: 3464}, {M: 5016, E: 846}, {M: -1341, E: -77}, {M: -14735, E: -1843}, {M: 1550, E: 436}, {M: -2932, E: -15}, {M: 1394, E: -272}, {M: 1212, E: 126},
	{M: -659, E: 1171}, {M: 21, E: 1291}, {M: 458, E: 453}, {M: 3811, E: -893}, {M: -1359, E: -98}, {M: -726, E: -384}, {M: 1005, E: -407}, {M: 2598, E: 521},
	{M: 275, E: 765}, {M: 326, E: 860}, {M: -82, E: 21}, {M: -9218, E: 1334}, {M: -8637, E: -130}, {M: 1563, E: 617}, {M: -956, E: 156}, {M: -44, E: 387},
	{M: 568, E: 609}, {M: 1920, E: -1}, {M: 2452, E: -834}, {M: 3488, E: -1294}, {M: -1142, E: 271}, {M: -1, E: -1387}, {M: 2273, E: -2144}, {M: -626, E: -816},
	{M: -15, E: -1100}, {M: -218, E: -13}, {M: 113, E: 825}, {M: 407, E: 594}, {M: 1765, E: 1453}, {M: -84, E: 1968}, {M: 1288, E: 512}, {M: 7643, E: -1626},
	{M: 1747, E: 1894}, {M: -959, E: -4319}, {M: -318, E: -3412}, {M: -293, E: -1254}, {M: -23, E: 11}, {M: 456, E: 2003}, {M: 1644, E: 791}, {M: 3229, E: 500},
	{M: 3779, E: 873}, {M: 4035, E: -11019}, {M: 3159, E: -8099}, {M: 1445, E: -2857}, {M: -544, E: 118}, {M: -2529, E: 2530}, {M: 150, E: 67}, {M: -6046, E: 3251},
	{M: 51, E: 217}, {M: 523, E: 1478}, {M: 1229, E: 3459}, {M: 6082, E: 8417}, {M: -2983, E: 876}, {M: -331, E: 12}, {M: 18081, E: -4282}, {M: 31654, E: -168},
	{M: 44, E: 56}, {M: 13, E: -59}, {M: 23, E: -47}, {M: 4, E: 76}, {M: -44, E: 67}, {M: 39, E: -27}, {M: -50, E: 51}, {M: -152, E: -151},
	{M: -1764, E: 23}, {M: -1354, E: -717}, {M: -1210, E: 623}, {M: -588, E: -98}, {M: -1367, E: 1156}, {M: 4551, E: -413}, {M: 4009, E: -2002}, {M: -977, E: -2358},
	{M: -1477, E: -756}, {M: -3067, E: -351}, {M: -183, E: -1398}, {M: -1467, E: 85}, {M: 17, E: 24}, {M: 199, E: -371}, {M: 2412, E: -2404}, {M: -847, E: -1640},
	{M: -1257, E: 1575}, {M: -2939, E: 818}, {M: 702, E: -973}, {M: 1322, E: -1731}, {M: 1268, E: -698}, {M: 823, E: -1270}, {M: -1915, E: -195}, {M: -2307, E: -107},
	{M: -436, E: 2423}, {M: 137, E: 666}, {M: -29, E: 24}, {M: 956, E: -1042}, {M: 1, E: 98}, {M: 956, E: 16}, {M: -155, E: 1410}, {M: -2385, E: 1961},
	{M: 2109, E: 5081}, {M: 147, E: 4286}, {M: 2576, E: 2198}, {M: 847, E: 47}, {M: 4662, E: 17}, {M: 12133, E: -20}, {M: 2298, E: 4449}, {M: 386, E: 4848},
	{M: 202, E: 665}, {M: 18, E: -5}, {M: 33, E: 2}, {M: -14, E: -1520}, {M: 88, E: 34}, {M: -102, E: -45}, {M: -27, E: -20}, {M: 55, E: 1458},
	{M: 93, E: -66}, {M: -44, E: 0}, {M: 14, E: 37}, {M: 88, E: 158}, {M: -143, E: 101}, {M: -102, E: 63}, {M: -6, E: 62}, {M: -69, E: -181},
	{M: -2591, E: -2063}, {M: 2076, E: 968}, {M: -1698, E: 0}, {M: -1533, E: -1324}, {M: -1317, E: -3100}, {M: 4388, E: -2025}, {M: 703, E: 243}, {M: -8381, E: 2168},
	{M: -526, E: 1101}, {M: -4636, E: 2639}, {M: 4081, E: -24}, {M: 1352, E: -2206}, {M: -2952, E: -1990}, {M: 94, E: -784}, {M: -763, E: 340}, {M: -3547, E: 39},
	{M: 197, E: 511}, {M: 6544, E: -25}, {M: 25, E: 598}, {M: -58, E: -1246}, {M: -1394, E: 995}, {M: 412, E: -472}, {M: 3507, E: -2087}, {M: 2346, E: -53},
	{M: -78, E: 28}, {M: 1942, E: 12742}, {M: -88, E: 6836}, {M: -13, E: 827}, {M: -2591, E: -5}, {M: -2289, E: -571}, {M: 113, E: -660}, {M: -18, E: -957},
	{M: -16, E: -60}, {M: -544, E: 22761}, {M: -5815, E: 17255}, {M: -663, E: 7142}, {M: 1118, E: -586}, {M: 942, E: -1843}, {M: -2, E: -2647}, {M: 2370, E: -4},
	{M: 234, E: 94}, {M: 62, E: 2445}, {M: -55, E: 2787}, {M: -1308, E: 6939}, {M: 715, E: 15309}, {M: 2305, E: 28586}, {M: 10950, E: 44822}, {M: 87, E: -85},
}
