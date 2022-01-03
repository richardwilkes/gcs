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

package com.trollworks.gcs.ui;

import java.awt.Color;

public class DynamicColor extends Color {
    public interface Resolver {
        int getRGB();
    }

    private Resolver mResolver;

    public DynamicColor(Resolver resolver) {
        super(0, true);
        mResolver = resolver;
    }

    @Override
    public int getRGB() {
        return mResolver.getRGB();
    }

    @Override
    public boolean equals(Object obj) {
        return obj instanceof DynamicColor && ((DynamicColor) obj).mResolver == mResolver;
    }

    @Override
    public int hashCode() {
        return mResolver.hashCode();
    }
}
