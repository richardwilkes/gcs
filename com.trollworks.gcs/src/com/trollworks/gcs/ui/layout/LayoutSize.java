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

package com.trollworks.gcs.ui.layout;

import java.awt.Component;
import java.awt.Dimension;

/** A convenience for retrieving minimum/maximum/preferred component sizes. */
public enum LayoutSize {
    /** The preferred size. */
    PREFERRED {
        @Override
        public Dimension get(Component component) {
            return sanitizeSize(component.getPreferredSize());
        }
    },
    /** The minimum size. */
    MINIMUM {
        @Override
        public Dimension get(Component component) {
            return sanitizeSize(component.getMinimumSize());
        }
    },
    /** The maximum size. */
    MAXIMUM {
        @Override
        public Dimension get(Component component) {
            return sanitizeSize(component.getMaximumSize());
        }
    };

    /** The maximum size to allow. */
    public static final int MAXIMUM_SIZE = Integer.MAX_VALUE / 512;

    /**
     * @param component The {@link Component} to return the size for.
     * @return The size desired by the {@link Component}.
     */
    public abstract Dimension get(Component component);

    /**
     * Ensures the size is within reasonable parameters.
     *
     * @param size The size to check.
     * @return The passed-in {@link Dimension} object, for convenience.
     */
    public static Dimension sanitizeSize(Dimension size) {
        if (size.width < 0) {
            size.width = 0;
        } else if (size.width > MAXIMUM_SIZE) {
            size.width = MAXIMUM_SIZE;
        }
        if (size.height < 0) {
            size.height = 0;
        } else if (size.height > MAXIMUM_SIZE) {
            size.height = MAXIMUM_SIZE;
        }
        return size;
    }
}
