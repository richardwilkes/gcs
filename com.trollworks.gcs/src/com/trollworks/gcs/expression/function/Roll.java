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

package com.trollworks.gcs.expression.function;

import com.trollworks.gcs.expression.EvaluationException;
import com.trollworks.gcs.expression.Evaluator;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;

public class Roll implements ExpressionFunction {
    @Override
    public final String getName() {
        return "roll";
    }

    @Override
    public final Object execute(Evaluator evaluator, String arguments) throws EvaluationException {
        try {
            Dice dice = new Dice(arguments.contains("(") ? new Evaluator(evaluator).evaluate(arguments).toString() : arguments);
            return Double.valueOf(dice.roll(false));
        } catch (Exception exception) {
            throw new EvaluationException(String.format(I18n.text("Invalid dice specification: %s"), arguments));
        }
    }
}
