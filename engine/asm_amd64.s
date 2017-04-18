// Copyright 2014-2017 The Zurichess Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "funcdata.h"

TEXT ·prefetch(SB), $0-8
       MOVQ       e+0(FP), AX
       PREFETCHNTA (AX)
       RET

