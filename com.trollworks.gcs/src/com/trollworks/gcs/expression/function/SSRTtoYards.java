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
import com.trollworks.gcs.utility.Fixed6;
import com.trollworks.gcs.utility.units.LengthUnits;

public class SSRTtoYards implements ExpressionFunction {
    @Override
    public String getName() {
        return "ssrt_to_yards";
    }

    @Override
    public Object execute(Evaluator evaluator, String arguments) throws EvaluationException {
        int v = (int) ArgumentTokenizer.getDoubleArgument(evaluator, arguments);
        if (v < -15) {
            v = -15;
        }
        switch (v) {
            case -15:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.ONE_FIFTH).asDouble());
            case -14:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.ONE_THIRD).asDouble());
            case -13:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.HALF).asDouble());
            case -12:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.TWO_THIRDS).asDouble());
            case -11:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, Fixed6.ONE).asDouble());
            case -10:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.ONE_AND_A_HALF).asDouble());
            case -9:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.TWO).asDouble());
            case -8:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.THREE).asDouble());
            case -7:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.FIVE).asDouble());
            case -6:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.IN, SSRT.EIGHT).asDouble());
            case -5:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.FT, Fixed6.ONE).asDouble());
            case -4:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.FT, SSRT.ONE_AND_A_HALF).asDouble());
            case -3:
                return Double.valueOf(LengthUnits.YD.convert(LengthUnits.FT, SSRT.TWO).asDouble());
            case -2:
                return Double.valueOf(1);
            case -1:
                return Double.valueOf(1.5);
            case 0:
                return Double.valueOf(2);
            case 1:
                return Double.valueOf(3);
            case 2:
                return Double.valueOf(5);
            case 3:
                return Double.valueOf(7);
            default:
                v -= 4;
                long multiplier = 1;
                for (int i = 0; i < v / 6; i++) {
                    multiplier *= 10;
                }
                v = switch (v % 6) {
                    case 0 -> 10;
                    case 1 -> 15;
                    case 2 -> 20;
                    case 3 -> 30;
                    case 4 -> 50;
                    case 5 -> 70;
                    default -> 0; // can't actually happen
                };
                return Double.valueOf(v * multiplier);
        }
    }
}
