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

package crc

import (
	"hash/crc64"
)

var table = crc64.MakeTable(crc64.ECMA)

// Bytes returns the CRC-64 value for the given data starting with the given crc value.
func Bytes(crc uint64, data []byte) uint64 {
	return crc64.Update(crc, table, data)
}

// String returns the CRC-64 value for the given data starting with the given crc value.
func String(crc uint64, data string) uint64 {
	return crc64.Update(crc, table, []byte(data))
}

// Byte returns the CRC-64 value for the given data starting with the given crc value.
func Byte(crc uint64, data byte) uint64 {
	var buffer [1]byte
	buffer[0] = data
	return crc64.Update(crc, table, buffer[:])
}

// Number returns the CRC-64 value for the given data starting with the given crc value.
func Number[T ~int64 | ~uint64 | ~int | ~uint](crc uint64, data T) uint64 {
	var buffer [8]byte
	d := uint64(data)
	buffer[0] = byte(d)
	buffer[1] = byte(d >> 8)
	buffer[2] = byte(d >> 16)
	buffer[3] = byte(d >> 24)
	buffer[4] = byte(d >> 32)
	buffer[5] = byte(d >> 40)
	buffer[6] = byte(d >> 48)
	buffer[7] = byte(d >> 56)
	return crc64.Update(crc, table, buffer[:])
}
