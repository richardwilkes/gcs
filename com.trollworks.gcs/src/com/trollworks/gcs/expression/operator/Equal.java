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

package com.trollworks.gcs.expression.operator;

import com.trollworks.gcs.expression.ArgumentTokenizer;

public class Equal extends Operator {
    public Equal() {
        super("==", 3);
    }

    @Override
    public final Object evaluate(Object left, Object right) {
        try {
            return Double.valueOf(ArgumentTokenizer.getDouble(left) == ArgumentTokenizer.getDouble(right) ? 1 : 0);
        } catch (Exception exception) {
            return Double.valueOf(left.toString().equals(right.toString()) ? 1 : 0);
        }
    }

    @Override
    public final Object evaluate(Object operand) {
        return null;
    }
}
