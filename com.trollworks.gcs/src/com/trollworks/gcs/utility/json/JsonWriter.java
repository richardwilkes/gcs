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

import java.io.FilterWriter;
import java.io.IOException;
import java.io.Writer;
import java.util.Objects;

public class JsonWriter extends FilterWriter {
    private String  mIndent;
    private int     mDepth;
    private boolean mCompact;
    private boolean mNeedComma;
    private boolean mNeedIndent;

    public JsonWriter(Writer writer, String indent) {
        super(writer);
        mIndent = indent;
        mCompact = indent.isEmpty();
    }

    private void indent() throws IOException {
        for (int i = 0; i < mDepth; i++) {
            write(mIndent);
        }
        mNeedIndent = false;
    }

    public void startMap() throws IOException {
        if (mNeedComma) {
            mNeedComma = false;
            write(',');
            if (!mCompact) {
                write('\n');
                indent();
            }
        } else if (mNeedIndent) {
            indent();
        }
        write('{');
        if (!mCompact) {
            write('\n');
            mDepth++;
            mNeedIndent = true;
        }
    }

    public void endMap() throws IOException {
        if (!mCompact) {
            write('\n');
            mDepth--;
            indent();
        }
        write('}');
        mNeedComma = true;
    }

    public void startArray() throws IOException {
        if (mNeedComma) {
            mNeedComma = false;
            write(',');
            if (!mCompact) {
                write('\n');
                indent();
            }
        } else if (mNeedIndent) {
            indent();
        }
        write('[');
        if (!mCompact) {
            write('\n');
            mDepth++;
            mNeedIndent = true;
        }
    }

    public void endArray() throws IOException {
        if (!mCompact) {
            write('\n');
            mDepth--;
            indent();
        }
        write(']');
        mNeedComma = true;
    }

    public void key(String key) throws IOException {
        if (mNeedComma) {
            mNeedComma = false;
            write(',');
            if (!mCompact) {
                write('\n');
                indent();
            }
        } else if (!mCompact) {
            indent();
        }
        write(Json.quote(key));
        write(':');
        if (!mCompact) {
            write(' ');
        }
    }

    public void value(String value) throws IOException {
        commaIfNeeded();
        write(Json.quote(value));
    }

    public void value(Number value) throws IOException {
        commaIfNeeded();
        write(Json.toString(value));
    }

    public void value(boolean value) throws IOException {
        commaIfNeeded();
        write(value ? "true" : "false");
    }

    public void value(short value) throws IOException {
        commaIfNeeded();
        write(Short.toString(value));
    }

    public void value(int value) throws IOException {
        commaIfNeeded();
        write(Integer.toString(value));
    }

    public void value(long value) throws IOException {
        commaIfNeeded();
        write(Long.toString(value));
    }

    public void value(float value) throws IOException {
        value(Float.valueOf(value));
    }

    public void value(double value) throws IOException {
        value(Double.valueOf(value));
    }

    public void keyValue(String key, String value) throws IOException {
        key(key);
        write(Json.quote(value));
        mNeedComma = true;
    }

    public void keyValueNot(String key, String value, String not) throws IOException {
        if (!Objects.equals(value, not)) {
            key(key);
            write(Json.quote(value));
            mNeedComma = true;
        }
    }

    public void keyValue(String key, Number value) throws IOException {
        key(key);
        write(Json.toString(value));
        mNeedComma = true;
    }

    public void keyValueNot(String key, Number value, Number not) throws IOException {
        if (!Objects.equals(value, not)) {
            key(key);
            write(Json.toString(value));
            mNeedComma = true;
        }
    }

    public void keyValue(String key, boolean value) throws IOException {
        key(key);
        write(value ? "true" : "false");
        mNeedComma = true;
    }

    public void keyValueNot(String key, boolean value, boolean not) throws IOException {
        if (value != not) {
            key(key);
            write(value ? "true" : "false");
            mNeedComma = true;
        }
    }

    public void keyValue(String key, short value) throws IOException {
        key(key);
        write(Short.toString(value));
        mNeedComma = true;
    }

    public void keyValueNot(String key, short value, short not) throws IOException {
        if (value != not) {
            key(key);
            write(Short.toString(value));
            mNeedComma = true;
        }
    }

    public void keyValue(String key, int value) throws IOException {
        key(key);
        write(Integer.toString(value));
        mNeedComma = true;
    }

    public void keyValueNot(String key, int value, int not) throws IOException {
        if (value != not) {
            key(key);
            write(Integer.toString(value));
            mNeedComma = true;
        }
    }

    public void keyValue(String key, long value) throws IOException {
        key(key);
        write(Long.toString(value));
        mNeedComma = true;
    }

    public void keyValueNot(String key, long value, long not) throws IOException {
        if (value != not) {
            key(key);
            write(Long.toString(value));
            mNeedComma = true;
        }
    }

    public void keyValue(String key, float value) throws IOException {
        keyValue(key, Float.valueOf(value));
    }

    public void keyValueNot(String key, float value, float not) throws IOException {
        if (value != not) {
            keyValue(key, Float.valueOf(value));
        }
    }

    public void keyValue(String key, double value) throws IOException {
        keyValue(key, Double.valueOf(value));
    }

    public void keyValueNot(String key, double value, double not) throws IOException {
        if (value != not) {
            keyValue(key, Double.valueOf(value));
        }
    }

    private void commaIfNeeded() throws IOException {
        if (mNeedComma) {
            write(',');
            if (!mCompact) {
                write('\n');
                indent();
            }
        } else {
            if (!mCompact) {
                indent();
            }
            mNeedComma = true;
        }
    }

    @Override
    public void close() throws IOException {
        if (!mCompact) {
            write('\n');
        }
        super.close();
    }
}
