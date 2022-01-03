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

import com.trollworks.gcs.expression.ArgumentTokenizer;
import com.trollworks.gcs.expression.EvaluationException;
import com.trollworks.gcs.expression.Evaluator;
import com.trollworks.gcs.utility.I18n;

public class Pow implements ExpressionFunction {
    @Override
    public final String getName() {
        return "pow";
    }

    @Override
    public final Object execute(Evaluator evaluator, String arguments) throws EvaluationException {
        try {
            Evaluator         ev        = new Evaluator(evaluator);
            ArgumentTokenizer tokenizer = new ArgumentTokenizer(arguments);
            double            arg1      = ArgumentTokenizer.getDouble(ev.evaluate(tokenizer.nextToken()));
            double            arg2      = ArgumentTokenizer.getDouble(ev.evaluate(tokenizer.nextToken()));
            return Double.valueOf(Math.pow(arg1, arg2));
        } catch (Exception exception) {
            throw new EvaluationException(I18n.text("Two numeric arguments are required"), exception);
        }
    }
}
