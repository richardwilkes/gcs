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

package com.trollworks.gcs.ui.print;

class ObjectWrapper<T> {
    private String mTitle;
    private T      mObject;

    ObjectWrapper(String title, T object) {
        mTitle = title;
        mObject = object;
    }

    /** @return The object. */
    public T getObject() {
        return mObject;
    }

    @Override
    public String toString() {
        return mTitle;
    }
}
