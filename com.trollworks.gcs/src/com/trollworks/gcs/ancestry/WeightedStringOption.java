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

package com.trollworks.gcs.ancestry;

import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;

public class WeightedStringOption extends WeightedOption<String> {
    public WeightedStringOption(int weight, String value) {
        super(weight, value);
    }

    public WeightedStringOption(JsonMap m) {
        super(m);
    }

    @Override
    protected void loadValue(JsonMap m) {
        mValue = m.getString(KEY_VALUE);
    }

    @Override
    protected void saveValue(JsonWriter w) throws IOException {
        w.keyValue(KEY_VALUE, mValue);
    }
}
