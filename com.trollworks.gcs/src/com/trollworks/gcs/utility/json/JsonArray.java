/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility.json;

import java.util.ArrayList;
import java.util.List;

/** Represents an array in JSON. */
public class JsonArray extends JsonCollection {
    private List<Object> mList = new ArrayList<>();

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or {@code null} if no such index exists.
     */
    public Object get(int index) {
        return index < 0 || index >= size() ? null : mList.get(index);
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or {@code false} if no such index exists or the
     *         value cannot be converted to a boolean.
     */
    public boolean getBoolean(int index) {
        return Json.asBoolean(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or zero if no such index exists or the value
     *         cannot be converted to a byte.
     */
    public byte getByte(int index) {
        return Json.asByte(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or zero if no such index exists or the value
     *         cannot be converted to a char.
     */
    public char getChar(int index) {
        return Json.asChar(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or zero if no such index exists or the value
     *         cannot be converted to a short.
     */
    public short getShort(int index) {
        return Json.asShort(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or zero if no such index exists or the value
     *         cannot be converted to an integer.
     */
    public int getInt(int index) {
        return Json.asInt(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or zero if no such index exists or the value
     *         cannot be converted to a long.
     */
    public long getLong(int index) {
        return Json.asLong(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or zero if no such index exists or the value
     *         cannot be converted to a float.
     */
    public float getFloat(int index) {
        return Json.asFloat(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index or zero if no such index exists or the value
     *         cannot be converted to a double.
     */
    public double getDouble(int index) {
        return Json.asDouble(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index.
     */
    public String getString(int index) {
        return Json.asString(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index.
     */
    public JsonArray getArray(int index) {
        return Json.asArray(get(index));
    }

    /**
     * @param index The index to retrieve.
     * @return The value associated with the index.
     */
    public JsonMap getMap(int index) {
        return Json.asMap(get(index));
    }

    /** @return The number of elements in the array. */
    public int size() {
        return mList.size();
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(Object value) {
        mList.add(Json.wrap(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(boolean value) {
        put(Boolean.valueOf(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(byte value) {
        put(Byte.valueOf(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(char value) {
        put(Character.valueOf(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(short value) {
        put(Short.valueOf(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(int value) {
        put(Integer.valueOf(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(long value) {
        put(Long.valueOf(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(float value) {
        put(Float.valueOf(value));
    }

    /**
     * Adds a value to the end of the array.
     *
     * @param value The value to store.
     */
    public void put(double value) {
        put(Double.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, Object value) {
        if (index >= 0) {
            value = Json.wrap(value);
            if (index < size()) {
                mList.set(index, value);
            } else {
                while (index != size()) {
                    put(JsonNull.INSTANCE);
                }
                put(value);
            }
        }
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, boolean value) {
        put(index, Boolean.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, byte value) {
        put(index, Byte.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, char value) {
        put(index, Character.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, short value) {
        put(index, Short.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, int value) {
        put(index, Integer.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, long value) {
        put(index, Long.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, float value) {
        put(index, Float.valueOf(value));
    }

    /**
     * Adds a value to the array.
     *
     * @param index The index to insert the value at. Must be greater than or equal to zero. If the
     *              index is past the end of the current set of values, {@code null}'s will be
     *              inserted as padding.
     * @param value The value to store.
     */
    public void put(int index, double value) {
        put(index, Double.valueOf(value));
    }

    /**
     * Removes a value from the array.
     *
     * @param index The index of the value to remove.
     */
    public void remove(int index) {
        if (index >= 0 && index < mList.size()) {
            mList.remove(index);
        }
    }

    @Override
    public StringBuilder appendTo(StringBuilder buffer, boolean compact, int depth) {
        int len = size();
        buffer.append('[');
        depth++;
        for (int i = 0; i < len; i++) {
            if (i > 0) {
                buffer.append(',');
            }
            if (!compact) {
                buffer.append('\n');
                indent(buffer, false, depth);
            }
            Object value = mList.get(i);
            if (value instanceof JsonCollection) {
                ((JsonCollection) value).appendTo(buffer, compact, depth);
            } else {
                buffer.append(Json.toString(value));
            }
        }
        if (!compact && len > 0) {
            buffer.append('\n');
            indent(buffer, false, depth - 1);
        }
        buffer.append(']');
        return buffer;
    }
}
