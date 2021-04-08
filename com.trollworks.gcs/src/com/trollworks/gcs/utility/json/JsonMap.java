/*
 * Copyright ©1998-2021 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility.json;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

/** Represents a map in JSON. */
public class JsonMap extends JsonCollection {
    private Map<String, Object> mMap = new HashMap<>();

    /**
     * @param key The key to check for.
     * @return {@code true} if the key is present in the map.
     */
    public boolean has(String key) {
        return mMap.containsKey(key);
    }

    public boolean isEmpty() {
        return mMap.isEmpty();
    }

    /**
     * @return The set of keys in this map.
     */
    public Set<String> keySet() {
        return mMap.keySet();
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key or {@code null} if no key matches.
     */
    public Object get(String key) {
        return key == null ? null : mMap.get(key);
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key or {@code false} if no key matches or the value
     *         cannot be converted to a boolean.
     */
    public boolean getBoolean(String key) {
        return Json.asBoolean(get(key));
    }

    public boolean getBooleanWithDefault(String key, boolean def) {
        Object value = get(key);
        return value != null ? Json.asBoolean(value) : def;
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key or zero if no key matches or the value cannot be
     *         converted to an integer.
     */
    public int getInt(String key) {
        return Json.asInt(get(key));
    }

    public int getIntWithDefault(String key, int def) {
        Object value = get(key);
        return value != null ? Json.asInt(value) : def;
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key or zero if no key matches or the value cannot be
     *         converted to a long.
     */
    public long getLong(String key) {
        return Json.asLong(get(key));
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key or zero if no key matches or the value cannot be
     *         converted to a double.
     */
    public double getDouble(String key) {
        return Json.asDouble(get(key));
    }

    public double getDoubleWithDefault(String key, double def) {
        Object value = get(key);
        return value != null ? Json.asDouble(value) : def;
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key.
     */
    public String getString(String key) {
        return Json.asString(get(key));
    }

    public String getStringWithDefault(String key, String def) {
        Object value = get(key);
        return value != null ? Json.asString(value) : def;
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key.
     */
    public JsonArray getArray(String key) {
        return Json.asArray(get(key));
    }

    /**
     * @param key The key to retrieve.
     * @return The value associated with the key.
     */
    public JsonMap getMap(String key) {
        return Json.asMap(get(key));
    }

    /**
     * @param key   The key to store the value with.
     * @param value The value to store.
     */
    public void put(String key, Object value) {
        if (key != null) {
            mMap.put(key, Json.wrap(value));
        }
    }

    @Override
    public void appendTo(Appendable buffer, boolean compact, int depth) throws IOException {
        boolean needComma = false;
        buffer.append('{');
        List<String> keys = new ArrayList<>(mMap.keySet());
        Collections.sort(keys);
        depth++;
        for (String key : keys) {
            if (needComma) {
                buffer.append(',');
            } else {
                needComma = true;
            }
            if (!compact) {
                buffer.append('\n');
                indent(buffer, false, depth);
            }
            buffer.append(Json.quote(key));
            if (compact) {
                buffer.append(':');
            } else {
                buffer.append(" : ");
            }
            Object value = mMap.get(key);
            if (value instanceof JsonCollection) {
                ((JsonCollection) value).appendTo(buffer, compact, depth);
            } else {
                buffer.append(Json.toString(value));
            }
        }
        if (!compact && !keys.isEmpty()) {
            buffer.append('\n');
            indent(buffer, false, depth - 1);
        }
        buffer.append('}');
    }
}
