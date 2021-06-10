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

package com.trollworks.gcs.ui;

import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;
import com.trollworks.gcs.utility.text.Enums;

import java.awt.Font;
import java.io.IOException;

public class FontDesc {
    private static final String KEY_NAME  = "name";
    private static final String KEY_STYLE = "style";
    private static final String KEY_SIZE  = "size";

    public String    mName;
    public FontStyle mStyle;
    public int       mSize;

    public FontDesc(Font font) {
        mName = font.getName();
        mStyle = FontStyle.from(font);
        mSize = font.getSize();
    }

    public FontDesc(JsonMap m) {
        mName = m.getStringWithDefault(KEY_NAME, ThemeFont.ROBOTO);
        mStyle = Enums.extract(m.getString(KEY_STYLE), FontStyle.values(), FontStyle.PLAIN);
        mSize = m.getIntWithDefault(KEY_SIZE, 9);
        if (mSize < 1) {
            mSize = 1;
        } else if (mSize > 1024) {
            mSize = 1024;
        }
    }

    public Font create() {
        return new Font(mName, mStyle.ordinal(), mSize);
    }

    public void toJSON(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_NAME, mName);
        w.keyValue(KEY_STYLE, Enums.toId(mStyle));
        w.keyValue(KEY_SIZE, mSize);
        w.endMap();
    }
}
