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

package com.trollworks.gcs.ancestry;

import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.lang.reflect.Constructor;
import java.util.ArrayList;
import java.util.List;

public abstract class WeightedOption<T> {
    protected static final String KEY_WEIGHT = "weight";
    protected static final String KEY_VALUE  = "value";
    public                 int    mWeight;
    public                 T      mValue;

    protected WeightedOption(int weight, T value) {
        mWeight = weight;
        mValue = value;
    }

    protected WeightedOption(JsonMap m) {
        mWeight = m.getInt(KEY_WEIGHT);
        loadValue(m);
    }

    protected abstract void loadValue(JsonMap m);

    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_WEIGHT, mWeight);
        saveValue(w);
        w.endMap();
    }

    protected abstract void saveValue(JsonWriter w) throws IOException;

    public boolean isValid() {
        return mWeight > 0;
    }

    public static final <T extends WeightedOption<E>, E> List<T> loadList(JsonMap m, String key, Class<T> cls) {
        List<T> list = new ArrayList<>();
        if (m.has(key)) {
            try {
                Constructor<T> constructor = cls.getConstructor(JsonMap.class);
                JsonArray      a           = m.getArray(key);
                int            count       = a.size();
                for (int i = 0; i < count; i++) {
                    T option = constructor.newInstance(a.getMap(i));
                    if (option.isValid()) {
                        list.add(option);
                    }
                }
            } catch (Exception e) {
                Log.error(e);
            }
        }
        return list;
    }

    public static final <T extends WeightedOption<E>, E> void saveList(JsonWriter w, String key, List<T> options) throws IOException {
        if (options != null && !options.isEmpty()) {
            w.key(key);
            w.startArray();
            for (T option : options) {
                option.save(w);
            }
            w.endArray();
        }
    }

    public static final <T extends WeightedOption<E>, E> T choose(List<T> options) {
        int total = 0;
        for (T option : options) {
            total += option.mWeight;
        }
        if (total > 0) {
            int choice = 1 + Dice.RANDOM.nextInt(total);
            for (T option : options) {
                choice -= option.mWeight;
                if (choice < 1) {
                    return option;
                }
            }
        }
        return null;
    }
}
