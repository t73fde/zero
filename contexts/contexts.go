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

// Package contexts provides some elements to work easier with storing things in a context.
package contexts

import "context"

// WithValueFunc is a function that returns a new context with a given value.
type WithValueFunc[T any] func(parent context.Context, val T) context.Context

// ValueFunc is a function that returns a value from a context, and a boolean
// to singal that the value was stored in the context, e.g. with [WithValueFunc],
// and that the value if of the correct type.
type ValueFunc[T any] func(context.Context) (T, bool)

// WithValue returns a [WithValueFunc] for the given key.
func WithValue[T any](key any) WithValueFunc[T] {
	return func(ctx context.Context, val T) context.Context {
		return context.WithValue(ctx, key, val)
	}
}

// Value returns a [ValueFunc] for the given key.
func Value[T any](key any) ValueFunc[T] {
	return func(ctx context.Context) (T, bool) {
		val, ok := ctx.Value(key).(T)
		return val, ok
	}
}

// WithAndValue returns a [WithValueFunc] and a [ValueFunc] for the given key.
func WithAndValue[T any](key any) (WithValueFunc[T], ValueFunc[T]) {
	return WithValue[T](key), Value[T](key)
}
