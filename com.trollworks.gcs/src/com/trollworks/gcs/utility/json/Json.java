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

import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.UrlUtils;

import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.io.Reader;
import java.io.StringReader;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;

/** Json utilities. */
public class Json {
    private Reader  mReader;
    private int     mIndex;
    private int     mCharacter = 1;
    private int     mLine      = 1;
    private char    mPrevious;
    private boolean mEOF;
    private boolean mUsePrevious;

    /**
     * @param reader A {@link Reader} to load JSON data from.
     * @return The result of loading the data.
     */
    public static final Object parse(Reader reader) throws IOException {
        return new Json(reader).nextValue();
    }

    /**
     * @param url A {@link URL} to load JSON data from.
     * @return The result of loading the data.
     */
    public static final Object parse(URL url) throws IOException {
        try (InputStream in = UrlUtils.setupConnection(url).getInputStream()) {
            return parse(in);
        }
    }

    /**
     * @param stream An {@link InputStream} to load JSON data from. {@link StandardCharsets#UTF_8}
     *               will be used as the encoding when reading from the stream.
     * @return The result of loading the data.
     */
    public static final Object parse(InputStream stream) throws IOException {
        return parse(new InputStreamReader(stream, StandardCharsets.UTF_8));
    }

    /**
     * @param string A {@link String} to load JSON data from.
     * @return The result of loading the data.
     */
    public static final Object parse(String string) throws IOException {
        return parse(new StringReader(string));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code false} if the object is {@code null}
     *         or the value cannot be converted to a boolean.
     */
    public static final boolean asBoolean(Object obj) {
        return Boolean.TRUE.equals(obj) || obj instanceof String && Boolean.TRUE.toString().equalsIgnoreCase((String) obj);
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code false} if the object is {@code null}
     *         or the value cannot be converted to a {@link Boolean}.
     */
    public static final Boolean asBooleanObject(Object obj) {
        return asBoolean(obj) ? Boolean.TRUE : Boolean.FALSE;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a byte.
     */
    public static final byte asByte(Object obj) {
        if (obj instanceof Number) {
            return ((Number) obj).byteValue();
        }
        if (obj instanceof String) {
            try {
                return Byte.parseByte((String) obj);
            } catch (Exception exception) {
                return 0;
            }
        }
        return 0;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a {@link Byte}.
     */
    public static final Byte asByteObject(Object obj) {
        return Byte.valueOf(asByte(obj));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a char.
     */
    public static final char asChar(Object obj) {
        if (obj instanceof Number) {
            return (char) ((Number) obj).intValue();
        }
        if (obj instanceof String) {
            String str = (String) obj;
            if (!str.isEmpty()) {
                return str.charAt(0);
            }
        }
        return 0;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a {@link Character}.
     */
    public static final Character asCharObject(Object obj) {
        return Character.valueOf(asChar(obj));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a short.
     */
    public static final short asShort(Object obj) {
        if (obj instanceof Number) {
            return ((Number) obj).shortValue();
        }
        if (obj instanceof String) {
            try {
                return Short.parseShort((String) obj);
            } catch (Exception exception) {
                return 0;
            }
        }
        return 0;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a {@link Short}.
     */
    public static final Short asShortObject(Object obj) {
        return Short.valueOf(asShort(obj));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to an int.
     */
    public static final int asInt(Object obj) {
        if (obj instanceof Number) {
            return ((Number) obj).intValue();
        }
        if (obj instanceof String) {
            try {
                return Integer.parseInt((String) obj);
            } catch (Exception exception) {
                return 0;
            }
        }
        return 0;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to an {@link Integer}.
     */
    public static final Integer asIntObject(Object obj) {
        return Integer.valueOf(asInt(obj));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a long.
     */
    public static final long asLong(Object obj) {
        if (obj instanceof Number) {
            return ((Number) obj).longValue();
        }
        if (obj instanceof String) {
            try {
                return Long.parseLong((String) obj);
            } catch (Exception exception) {
                return 0;
            }
        }
        return 0;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a {@link Long}.
     */
    public static final Long asLongObject(Object obj) {
        return Long.valueOf(asLong(obj));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a float.
     */
    public static final float asFloat(Object obj) {
        if (obj instanceof Number) {
            return ((Number) obj).floatValue();
        }
        if (obj instanceof String) {
            try {
                return Float.parseFloat((String) obj);
            } catch (Exception exception) {
                return 0;
            }
        }
        return 0;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a {@link Float}.
     */
    public static final Float asFloatObject(Object obj) {
        return Float.valueOf(asFloat(obj));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a double.
     */
    public static final double asDouble(Object obj) {
        if (obj instanceof Number) {
            return ((Number) obj).doubleValue();
        }
        if (obj instanceof String) {
            try {
                return Double.parseDouble((String) obj);
            } catch (Exception exception) {
                return 0;
            }
        }
        return 0;
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object or {@code 0} if the object is {@code null} or
     *         the value cannot be converted to a {@link Double}.
     */
    public static final Double asDoubleObject(Object obj) {
        return Double.valueOf(asDouble(obj));
    }

    /**
     * @param obj An object to process.
     * @return The value associated with the object.
     */
    public static final String asString(Object obj) {
        return JsonNull.INSTANCE.equals(obj) ? "" : obj.toString();
    }

    /**
     * @param obj An object to process.
     * @return The {@link JsonArray}.
     */
    public static final JsonArray asArray(Object obj) {
        return (obj instanceof JsonArray) ? (JsonArray) obj : new JsonArray();
    }

    /**
     * @param obj An object to process.
     * @return The {@link JsonMap}.
     */
    public static final JsonMap asMap(Object obj) {
        return (obj instanceof JsonMap) ? (JsonMap) obj : new JsonMap();
    }

    /**
     * @param value The value to encode as a JSON string.
     * @return The encoded {@link String}.
     */
    public static final String toString(Object value) {
        if (JsonNull.INSTANCE.equals(value)) {
            return JsonNull.INSTANCE.toString();
        }
        if (value instanceof Number) {
            String str = value.toString();
            if (str.indexOf('.') > 0 && str.indexOf('e') < 0 && str.indexOf('E') < 0) {
                while (str.endsWith("0")) {
                    str = str.substring(0, str.length() - 1);
                }
                if (str.endsWith(".")) {
                    str = str.substring(0, str.length() - 1);
                }
            }
            return str;
        }
        if (value instanceof Boolean || value instanceof JsonCollection) {
            return value.toString();
        }
        if (value instanceof Map || value instanceof List || value.getClass().isArray()) {
            return wrap(value).toString();
        }
        return quote(value.toString());
    }

    /**
     * @param object The object to wrap for storage inside a {@link JsonCollection}.
     * @return The wrapped version of the object, which may be the original object passed in.
     */
    public static final Object wrap(Object object) {
        if (JsonNull.INSTANCE.equals(object)) {
            return JsonNull.INSTANCE;
        }
        if (object instanceof JsonCollection || object instanceof Boolean || object instanceof Byte || object instanceof Character || object instanceof Short || object instanceof Integer || object instanceof Long || object instanceof Float || object instanceof Double || object instanceof String) {
            return object;
        }
        if (object instanceof List) {
            JsonArray array = new JsonArray();
            for (Object one : (List<?>) object) {
                array.put(wrap(one));
            }
            return array;
        }
        Class<?> type = object.getClass();
        if (type.isArray()) {
            JsonArray array = new JsonArray();
            if (object instanceof boolean[]) {
                for (boolean value : (boolean[]) object) {
                    array.put(value);
                }
            } else if (object instanceof byte[]) {
                for (byte value : (byte[]) object) {
                    array.put(value);
                }
            } else if (object instanceof char[]) {
                for (char value : (char[]) object) {
                    array.put(value);
                }
            } else if (object instanceof short[]) {
                for (short value : (short[]) object) {
                    array.put(value);
                }
            } else if (object instanceof int[]) {
                for (int value : (int[]) object) {
                    array.put(value);
                }
            } else if (object instanceof long[]) {
                for (long value : (long[]) object) {
                    array.put(value);
                }
            } else if (object instanceof float[]) {
                for (float value : (float[]) object) {
                    array.put(value);
                }
            } else if (object instanceof double[]) {
                for (double value : (double[]) object) {
                    array.put(value);
                }
            } else {
                for (Object obj : (Object[]) object) {
                    array.put(wrap(obj));
                }
            }
            return array;
        }
        if (object instanceof Map) {
            JsonMap map = new JsonMap();
            for (Map.Entry<?, ?> entry : ((Map<?, ?>) object).entrySet()) {
                map.put(entry.getKey().toString(), entry.getValue());
            }
            return map;
        }
        return object.toString();
    }

    /**
     * @param string The string to quote.
     * @return The quoted {@link String}, suitable for storage inside a JSON object.
     */
    public static final String quote(String string) {
        int len;
        if (string == null || (len = string.length()) == 0) {
            return "\"\"";
        }
        StringBuilder buffer = new StringBuilder(len + 4);
        buffer.append('"');
        char ch = 0;
        for (int i = 0; i < len; i++) {
            char last = ch;
            ch = string.charAt(i);
            switch (ch) {
            case '\\':
            case '"':
                buffer.append('\\');
                buffer.append(ch);
                break;
            case '/':
                if (last == '<') {
                    buffer.append('\\');
                }
                buffer.append(ch);
                break;
            case '\b':
                buffer.append("\\b");
                break;
            case '\t':
                buffer.append("\\t");
                break;
            case '\n':
                buffer.append("\\n");
                break;
            case '\f':
                buffer.append("\\f");
                break;
            case '\r':
                buffer.append("\\r");
                break;
            default:
                if (ch < 0x20) {
                    String hex = "000" + Integer.toHexString(ch);
                    buffer.append("\\u").append(hex.substring(hex.length() - 4));
                } else {
                    buffer.append(ch);
                }
                break;
            }
        }
        buffer.append('"');
        return buffer.toString();
    }

    private Json(Reader reader) {
        mReader = reader;
    }

    private char next() throws IOException {
        int c;
        if (mUsePrevious) {
            mUsePrevious = false;
            c = mPrevious;
        } else {
            c = mReader.read();
            if (c <= 0) { // End of stream
                mEOF = true;
                c = 0;
            }
        }
        mIndex++;
        if (mPrevious == '\r') {
            mLine++;
            mCharacter = c == '\n' ? 0 : 1;
        } else if (c == '\n') {
            mLine++;
            mCharacter = 0;
        } else {
            mCharacter++;
        }
        mPrevious = (char) c;
        return mPrevious;
    }

    private char nextSkippingWhitespace() throws IOException {
        for (; ; ) {
            char c = next();
            if (c == 0 || c > ' ') {
                return c;
            }
        }
    }

    private Object nextValue() throws IOException {
        char   c = nextSkippingWhitespace();
        String s;

        switch (c) {
        case '"':
        case '\'':
            return nextString(c);
        case '{':
            back();
            return nextMap();
        case '[':
        case '(':
            back();
            return nextArray();
        default:
            break;
        }

        StringBuilder sb = new StringBuilder();
        while (c >= ' ' && ",:]}/\\\"[{;=#".indexOf(c) < 0) {
            sb.append(c);
            c = next();
        }
        back();

        s = sb.toString().trim();
        if (s.isEmpty()) {
            throw syntaxError("missing value");
        }
        if ("true".equalsIgnoreCase(s)) {
            return Boolean.TRUE;
        }
        if ("false".equalsIgnoreCase(s)) {
            return Boolean.FALSE;
        }
        if ("null".equalsIgnoreCase(s)) {
            return JsonNull.INSTANCE;
        }

        char b = s.charAt(0);
        if (b >= '0' && b <= '9' || b == '.' || b == '-' || b == '+') {
            if (b == '0' && s.length() > 2 && (s.charAt(1) == 'x' || s.charAt(1) == 'X')) {
                try {
                    return Integer.valueOf(Integer.parseInt(s.substring(2), 16));
                } catch (Exception ignore) {
                    Log.error(ignore);
                }
            }
            try {
                if (s.indexOf('.') > -1 || s.indexOf('e') > -1 || s.indexOf('E') > -1) {
                    return Double.valueOf(s);
                }
                Long myLong = Long.valueOf(s);
                if (myLong.longValue() == myLong.intValue()) {
                    return Integer.valueOf(myLong.intValue());
                }
                return myLong;
            } catch (Exception ignore) {
                Log.error(ignore);
            }
        }
        return s;
    }

    private JsonArray nextArray() throws IOException {
        char c = nextSkippingWhitespace();
        char q;
        if (c == '[') {
            q = ']';
        } else if (c == '(') {
            q = ')';
        } else {
            throw syntaxError("a JSONArray text must start with '['");
        }
        JsonArray array = new JsonArray();
        if (nextSkippingWhitespace() == ']') {
            return array;
        }
        back();
        for (; ; ) {
            if (nextSkippingWhitespace() == ',') {
                back();
                array.put(null);
            } else {
                back();
                array.put(nextValue());
            }
            c = nextSkippingWhitespace();
            switch (c) {
            case ';':
            case ',':
                if (nextSkippingWhitespace() == ']') {
                    return array;
                }
                back();
                break;
            case ']':
            case ')':
                if (q != c) {
                    throw syntaxError("expected a '" + Character.toString(q) + "'");
                }
                return array;
            default:
                throw syntaxError("expected a ',' or ']'");
            }
        }
    }

    private JsonMap nextMap() throws IOException {
        char   c;
        String key;

        if (nextSkippingWhitespace() != '{') {
            throw syntaxError("JSON object text must begin with '{'");
        }
        JsonMap map = new JsonMap();
        while (true) {
            c = nextSkippingWhitespace();
            switch (c) {
            case 0:
                throw syntaxError("JSON object text must end with '}'");
            case '}':
                return map;
            default:
                back();
                key = nextValue().toString();
            }

            c = nextSkippingWhitespace();
            if (c == '=') {
                if (next() != '>') {
                    back();
                }
            } else if (c != ':') {
                throw syntaxError("expected a ':' after a key");
            }
            if (map.has(key)) {
                throw new IOException("duplicate key \"" + key + "\"");
            }
            map.put(key, nextValue());

            switch (nextSkippingWhitespace()) {
            case ';':
            case ',':
                if (nextSkippingWhitespace() == '}') {
                    return map;
                }
                back();
                break;
            case '}':
                return map;
            default:
                throw syntaxError("expected a ',' or '}'");
            }
        }
    }

    private String nextString(char quote) throws IOException {
        char          c;
        StringBuilder buffer = new StringBuilder();
        while (true) {
            c = next();
            switch (c) {
            case 0:
            case '\n':
            case '\r':
                throw syntaxError("unterminated string");
            case '\\':
                c = next();
                switch (c) {
                case 'b':
                    buffer.append('\b');
                    break;
                case 't':
                    buffer.append('\t');
                    break;
                case 'n':
                    buffer.append('\n');
                    break;
                case 'f':
                    buffer.append('\f');
                    break;
                case 'r':
                    buffer.append('\r');
                    break;
                case 'u':
                    buffer.append((char) Integer.parseInt(next(4), 16));
                    break;
                case '"':
                case '\'':
                case '\\':
                case '/':
                    buffer.append(c);
                    break;
                default:
                    throw syntaxError("illegal escape");
                }
                break;
            default:
                if (c == quote) {
                    return buffer.toString();
                }
                buffer.append(c);
            }
        }
    }

    private void back() {
        if (mUsePrevious || mIndex <= 0) {
            throw new IllegalStateException("stepping back two steps is not supported");
        }
        mIndex--;
        mCharacter--;
        mUsePrevious = true;
        mEOF = false;
    }

    private String next(int n) throws IOException {
        if (n == 0) {
            return "";
        }
        char[] buffer = new char[n];
        int    pos    = 0;
        while (pos < n) {
            buffer[pos] = next();
            if (mEOF && !mUsePrevious) {
                throw syntaxError("substring bounds error");
            }
            pos++;
        }
        return new String(buffer);
    }

    private IOException syntaxError(String message) {
        return new IOException(message + toString());
    }

    @Override
    public String toString() {
        return " at " + mIndex + " [character " + mCharacter + " line " + mLine + "]";
    }
}
