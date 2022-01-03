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
import com.trollworks.gcs.expression.EvaluationException;

public class Not extends Operator {
    public Not() {
        super("!", 0, true);
    }

    @Override
    public final Object evaluate(Object left, Object right) {
        return null;
    }

    @Override
    public final Object evaluate(Object operand) throws EvaluationException {
        return Double.valueOf(ArgumentTokenizer.getDoubleOperand(operand) != 0 ? 0 : 1);
    }
}
