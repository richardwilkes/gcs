/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.io;

/** Utility method for reading and writing bytes as either little endian or big endian values. */
public class EndianUtils {
    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 2 bytes from the buffer, interpreted as a big-endian {@code short}.
     */
    public static final short readBEShort(byte[] buffer, int offset) {
        return (short) ((buffer[offset] & 0xFF) << 8 | (buffer[offset + 1] & 0xFF) << 0);
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 2 bytes from the buffer, interpreted as a big-endian unsigned {@code
     *         short}.
     */
    public static final int readBEUnsignedShort(byte[] buffer, int offset) {
        return (buffer[offset] & 0xFF) << 8 | (buffer[offset + 1] & 0xFF) << 0;
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 4 bytes from the buffer, interpreted as a big-endian {@code int}.
     */
    public static final int readBEInt(byte[] buffer, int offset) {
        return (buffer[offset] & 0xFF) << 24 | (buffer[offset + 1] & 0xFF) << 16 | (buffer[offset + 2] & 0xFF) << 8 | (buffer[offset + 3] & 0xFF) << 0;
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 8 bytes from the buffer, interpreted as a big-endian {@code long}.
     */
    public static final long readBELong(byte[] buffer, int offset) {
        return ((long) buffer[offset] & 0xFF) << 56 | ((long) buffer[offset + 1] & 0xFF) << 48 | ((long) buffer[offset + 2] & 0xFF) << 40 | ((long) buffer[offset + 3] & 0xFF) << 32 | ((long) buffer[offset + 4] & 0xFF) << 24 | ((long) buffer[offset + 5] & 0xFF) << 16 | ((long) buffer[offset + 6] & 0xFF) << 8 | ((long) buffer[offset + 7] & 0xFF) << 0;
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 4 bytes from the buffer, interpreted as a big-endian {@code float}.
     */
    public static final float readBEFloat(byte[] buffer, int offset) {
        return Float.intBitsToFloat(readBEInt(buffer, offset));
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 8 bytes from the buffer, interpreted as a big-endian {@code double}.
     */
    public static final double readBEDouble(byte[] buffer, int offset) {
        return Double.longBitsToDouble(readBELong(buffer, offset));
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeBEShort(int value, byte[] buffer, int offset) {
        buffer[offset] = (byte) (value >>> 8 & 0xFF);
        buffer[offset + 1] = (byte) (value >>> 0 & 0xFF);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeBEInt(int value, byte[] buffer, int offset) {
        buffer[offset] = (byte) (value >>> 24 & 0xFF);
        buffer[offset + 1] = (byte) (value >>> 16 & 0xFF);
        buffer[offset + 2] = (byte) (value >>> 8 & 0xFF);
        buffer[offset + 3] = (byte) (value >>> 0 & 0xFF);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeBELong(long value, byte[] buffer, int offset) {
        buffer[offset] = (byte) (value >>> 56 & 0xFF);
        buffer[offset + 1] = (byte) (value >>> 48 & 0xFF);
        buffer[offset + 2] = (byte) (value >>> 40 & 0xFF);
        buffer[offset + 3] = (byte) (value >>> 32 & 0xFF);
        buffer[offset + 4] = (byte) (value >>> 24 & 0xFF);
        buffer[offset + 5] = (byte) (value >>> 16 & 0xFF);
        buffer[offset + 6] = (byte) (value >>> 8 & 0xFF);
        buffer[offset + 7] = (byte) (value >>> 0 & 0xFF);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeBEFloat(float value, byte[] buffer, int offset) {
        writeBEInt(Float.floatToIntBits(value), buffer, offset);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeBEDouble(double value, byte[] buffer, int offset) {
        writeBELong(Double.doubleToLongBits(value), buffer, offset);
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 2 bytes from the buffer, interpreted as a little-endian {@code short}.
     */
    public static final short readLEShort(byte[] buffer, int offset) {
        return (short) ((buffer[offset] & 0xFF) << 0 | (buffer[offset + 1] & 0xFF) << 8);
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 2 bytes from the buffer, interpreted as a little-endian unsigned {@code
     *         short}.
     */
    public static final int readLEUnsignedShort(byte[] buffer, int offset) {
        return (buffer[offset] & 0xFF) << 0 | (buffer[offset + 1] & 0xFF) << 8;
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 4 bytes from the buffer, interpreted as a little-endian {@code int}.
     */
    public static final int readLEInt(byte[] buffer, int offset) {
        return (buffer[offset] & 0xFF) << 0 | (buffer[offset + 1] & 0xFF) << 8 | (buffer[offset + 2] & 0xFF) << 16 | (buffer[offset + 3] & 0xFF) << 24;
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 8 bytes from the buffer, interpreted as a little-endian {@code long}.
     */
    public static final long readLELong(byte[] buffer, int offset) {
        return ((long) buffer[offset] & 0xFF) << 0 | ((long) buffer[offset + 1] & 0xFF) << 8 | ((long) buffer[offset + 2] & 0xFF) << 16 | ((long) buffer[offset + 3] & 0xFF) << 24 | ((long) buffer[offset + 4] & 0xFF) << 32 | ((long) buffer[offset + 5] & 0xFF) << 40 | ((long) buffer[offset + 6] & 0xFF) << 48 | ((long) buffer[offset + 7] & 0xFF) << 56;
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 4 bytes from the buffer, interpreted as a little-endian {@code float}.
     */
    public static final float readLEFloat(byte[] buffer, int offset) {
        return Float.intBitsToFloat(readLEInt(buffer, offset));
    }

    /**
     * @param buffer The buffer to read from.
     * @param offset The offset within the buffer to start reading from.
     * @return The next 8 bytes from the buffer, interpreted as a little-endian {@code double}.
     */
    public static final double readLEDouble(byte[] buffer, int offset) {
        return Double.longBitsToDouble(readLELong(buffer, offset));
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeLEShort(int value, byte[] buffer, int offset) {
        buffer[offset] = (byte) (value >>> 0 & 0xFF);
        buffer[offset + 1] = (byte) (value >>> 8 & 0xFF);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeLEInt(int value, byte[] buffer, int offset) {
        buffer[offset] = (byte) (value >>> 0 & 0xFF);
        buffer[offset + 1] = (byte) (value >>> 8 & 0xFF);
        buffer[offset + 2] = (byte) (value >>> 16 & 0xFF);
        buffer[offset + 3] = (byte) (value >>> 24 & 0xFF);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeLELong(long value, byte[] buffer, int offset) {
        buffer[offset] = (byte) (value >>> 0 & 0xFF);
        buffer[offset + 1] = (byte) (value >>> 8 & 0xFF);
        buffer[offset + 2] = (byte) (value >>> 16 & 0xFF);
        buffer[offset + 3] = (byte) (value >>> 24 & 0xFF);
        buffer[offset + 4] = (byte) (value >>> 32 & 0xFF);
        buffer[offset + 5] = (byte) (value >>> 40 & 0xFF);
        buffer[offset + 6] = (byte) (value >>> 48 & 0xFF);
        buffer[offset + 7] = (byte) (value >>> 56 & 0xFF);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeLEFloat(float value, byte[] buffer, int offset) {
        writeLEInt(Float.floatToIntBits(value), buffer, offset);
    }

    /**
     * @param value  The value to write.
     * @param buffer The buffer to write to.
     * @param offset The offset within the buffer to start writing to.
     */
    public static final void writeLEDouble(double value, byte[] buffer, int offset) {
        writeLELong(Double.doubleToLongBits(value), buffer, offset);
    }
}
