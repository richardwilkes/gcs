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

package com.trollworks.gcs.ui.widget;

import com.trollworks.gcs.ui.Colors;

import java.awt.Color;

public enum MessageType {
    ERROR {
        @Override
        String getText() {
            return "\uf06a";
        }

        @Override
        Color getColor() {
            return Colors.ERROR;
        }
    },
    WARNING {
        @Override
        String getText() {
            return "\uf071";
        }

        @Override
        Color getColor() {
            return Colors.WARNING;
        }
    },
    QUESTION {
        @Override
        String getText() {
            return "\uf059";
        }
    },
    NONE;

    String getText() {
        return "";
    }

    Color getColor() {
        return Colors.ON_BACKGROUND;
    }
}
