//-----------------------------------------------------------------------------
// Copyright (c) 2025-present Detlef Stern
//
// This file is part of Zero.
//
// Zero is licensed under the latest version of the EUPL (European Union Public
// License). Please see file LICENSE.txt for your rights and obligations under
// this license.
//
// SPDX-License-Identifier: EUPL-1.2
// SPDX-FileCopyrightText: 2025-present Detlef Stern
//-----------------------------------------------------------------------------

package context_test

import (
	"context"
	"testing"

	zerocontext "t73f.de/r/zero/context"
)

func TestContextValues(t *testing.T) {
	type keyType1 struct{}
	type keyType2 string

	with1, val1 := zerocontext.WithAndValue[string](keyType1{})
	with2, val2 := zerocontext.WithAndValue[int](keyType2("just-a-key"))

	ctx1 := with1(context.Background(), "abc")
	if v1, ok := val1(ctx1); ok {
		if v1 != "abc" {
			t.Errorf("abc expected but got %v", v1)
		}
	} else {
		t.Error("string value expected")
	}

	ctx2 := with2(context.Background(), 17)
	if v2, ok := val2(ctx2); ok {
		if v2 != 17 {
			t.Errorf("17 expected but got %v", v2)
		}
	} else {
		t.Error("int value expected")
	}

	if val, ok := val1(ctx2); ok {
		t.Errorf("must not have string value, but got: %v", val)
	}
	if val, ok := val2(ctx1); ok {
		t.Errorf("must not have int value, but got: %v", val)
	}
}
