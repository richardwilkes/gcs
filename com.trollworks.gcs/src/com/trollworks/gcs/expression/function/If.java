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

public class If implements ExpressionFunction {
    @Override
    public final String getName() {
        return "if";
    }

    @Override
    public final Object execute(Evaluator evaluator, String arguments) throws EvaluationException {
        ArgumentTokenizer tokenizer = new ArgumentTokenizer(arguments);
        Evaluator         ev        = new Evaluator(evaluator);
        Object            result    = ev.evaluate(tokenizer.nextToken());
        if (result instanceof Double) {
            if (((Double) result).doubleValue() == 0) {
                tokenizer.nextToken();
            }
        } else {
            String str = result.toString();
            if (str.isEmpty()) {
                tokenizer.nextToken();
            } else {
                try {
                    if (Double.parseDouble(str) == 0) {
                        tokenizer.nextToken();
                    }
                } catch (NumberFormatException nfe) {
                    // Treat as true
                }
            }
        }
        return ev.evaluate(tokenizer.nextToken());
    }
}
