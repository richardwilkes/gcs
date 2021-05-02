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

package com.trollworks.gcs.expression.function;

import com.trollworks.gcs.expression.ArgumentTokenizer;
import com.trollworks.gcs.expression.EvaluationException;
import com.trollworks.gcs.expression.Evaluator;

public class Ceil implements ExpressionFunction {
    @Override
    public final String getName() {
        return "ceil";
    }

    @Override
    public final Object execute(Evaluator evaluator, String arguments) throws EvaluationException {
        return Double.valueOf(Math.ceil(ArgumentTokenizer.getDoubleArgument(evaluator, arguments)));
    }
}
