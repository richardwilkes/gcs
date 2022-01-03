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

public class Divide extends Operator {
    public Divide() {
        super("/", 6);
    }

    @Override
    public final Object evaluate(Object left, Object right) throws EvaluationException {
        double a = ArgumentTokenizer.getDoubleOperand(left);
        double b = ArgumentTokenizer.getDoubleOperand(right);
        double r = 0;
        if (b != 0) {
            r = a / b;
        }
        return Double.valueOf(r);
    }

    @Override
    public final Object evaluate(Object operand) {
        return null;
    }
}
