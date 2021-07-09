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
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Enums;
import com.trollworks.gcs.utility.units.LengthUnits;

public class SSRT implements ExpressionFunction {
    public static final Fixed6 ONE_AND_A_HALF = new Fixed6("1.5", false);
    public static final Fixed6 TWO            = new Fixed6(2);
    public static final Fixed6 THREE          = new Fixed6(3);
    public static final Fixed6 FIVE           = new Fixed6(5);
    public static final Fixed6 SEVEN          = new Fixed6(7);
    public static final Fixed6 EIGHT          = new Fixed6(8);
    public static final Fixed6 TEN            = new Fixed6(10);
    public static final Fixed6 ONE_FIFTH      = Fixed6.ONE.div(FIVE);
    public static final Fixed6 ONE_THIRD      = Fixed6.ONE.div(THREE);
    public static final Fixed6 HALF           = Fixed6.ONE.div(TWO);
    public static final Fixed6 TWO_THIRDS     = TWO.div(THREE);

    @Override
    public String getName() {
        return "ssrt";
    }

    @Override
    public Object execute(Evaluator evaluator, String arguments) throws EvaluationException {
        try {
            Evaluator         ev        = new Evaluator(evaluator);
            ArgumentTokenizer tokenizer = new ArgumentTokenizer(arguments);
            double            length    = ArgumentTokenizer.getDouble(ev.evaluate(tokenizer.nextToken()));
            LengthUnits       units     = Enums.extract(ev.evaluate(tokenizer.nextToken()).toString(), LengthUnits.values(), LengthUnits.YD);
            boolean           wantSize  = ArgumentTokenizer.getDouble(ev.evaluate(tokenizer.nextToken())) != 0;
            int               value     = yardsToValue(LengthUnits.YD.convert(units, new Fixed6(length)), wantSize);
            return Integer.valueOf(wantSize ? value : -value);
        } catch (Exception exception) {
            throw new EvaluationException(I18n.text("Three arguments are required: a length, the units of the length, and non-zero value to return the size or a zero value to return the speed/range"), exception);
        }
    }

    private static int yardsToValue(Fixed6 yards, boolean allowNegative) {
        if (allowNegative) {
            Fixed6 inches = LengthUnits.IN.convert(LengthUnits.YD, yards);
            if (inches.lessThanOrEqual(ONE_FIFTH)) { // 1/5 in
                return -15;
            }
            if (inches.lessThanOrEqual(ONE_THIRD)) { // 1/3 in
                return -14;
            }
            if (inches.lessThanOrEqual(HALF)) { // 1/2 in
                return -13;
            }
            if (inches.lessThanOrEqual(TWO_THIRDS)) { // 2/3 in
                return -12;
            }
            if (inches.lessThanOrEqual(Fixed6.ONE)) { // 1 in
                return -11;
            }
            if (inches.lessThanOrEqual(ONE_AND_A_HALF)) { // 1.5 in
                return -10;
            }
            if (inches.lessThanOrEqual(TWO)) { // 2 in
                return -9;
            }
            if (inches.lessThanOrEqual(THREE)) { // 3 in
                return -8;
            }
            if (inches.lessThanOrEqual(FIVE)) { // 5 in
                return -7;
            }
            if (inches.lessThanOrEqual(EIGHT)) { // 8 in
                return -6;
            }
            Fixed6 feet = LengthUnits.FT.convert(LengthUnits.YD, yards);
            if (feet.lessThanOrEqual(Fixed6.ONE)) { // 1 ft
                return -5;
            }
            if (feet.lessThanOrEqual(ONE_AND_A_HALF)) { // 1.5 ft
                return -4;
            }
            if (feet.lessThanOrEqual(TWO)) { // 2 ft
                return -3;
            }
            if (yards.lessThanOrEqual(Fixed6.ONE)) { // 1 yd
                return -2;
            }
            if (yards.lessThanOrEqual(ONE_AND_A_HALF)) { // 1.5 yd
                return -1;
            }
        }
        if (yards.lessThanOrEqual(TWO)) { // 2 yd
            return 0;
        }
        int amt = 0;
        while (yards.greaterThan(TEN)) {
            yards = yards.div(TEN);
            amt += 6;
        }
        if (yards.greaterThan(SEVEN)) {
            return amt + 4;
        }
        if (yards.greaterThan(FIVE)) {
            return amt + 3;
        }
        if (yards.greaterThan(THREE)) {
            return amt + 2;
        }
        if (yards.greaterThan(TWO)) {
            return amt + 1;
        }
        if (yards.greaterThan(ONE_AND_A_HALF)) {
            return amt;
        }
        return amt - 1;
    }
}
