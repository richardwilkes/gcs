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
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;
import java.util.List;

public class DisplayDice implements ExpressionFunction {
    @Override
    public String getName() {
        return "dice";
    }

    @Override
    public Object execute(Evaluator evaluator, String arguments) throws EvaluationException {
        try {
            Evaluator         ev        = new Evaluator(evaluator);
            ArgumentTokenizer tokenizer = new ArgumentTokenizer(arguments);
            List<Integer>     args      = new ArrayList<>();
            while (tokenizer.hasMoreTokens()) {
                args.add(Integer.valueOf((int) ArgumentTokenizer.getDouble(ev.evaluate(tokenizer.nextToken()))));
            }
            Dice dice = switch (args.size()) {
                case 1 -> new Dice(1, args.get(0).intValue(), 0, 1); // sides
                case 2 -> new Dice(args.get(0).intValue(), args.get(1).intValue(), 0, 1); // count, sides
                case 3 -> new Dice(args.get(0).intValue(), args.get(1).intValue(), args.get(2).intValue(), 1); // count, sides, modifier
                case 4 -> new Dice(args.get(0).intValue(), args.get(1).intValue(), args.get(2).intValue(), args.get(3).intValue()); // count, sides, modifier, multiplier
                default -> throw new Exception();
            };
            return dice.toString();
        } catch (EvaluationException exception) {
            throw exception;
        } catch (Exception exception) {
            throw new EvaluationException(String.format(I18n.text("Invalid dice specification: %s"), arguments));
        }
    }
}
