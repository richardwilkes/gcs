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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.expression.EvaluationException;
import com.trollworks.gcs.expression.Evaluator;
import com.trollworks.gcs.expression.VariableResolver;

public class LookupAdvantageLevel implements ExpressionFunction {
    @Override
    public String getName() {
        return "advantage_level";
    }

    @Override
    public Object execute(Evaluator evaluator, String arguments) {
        VariableResolver resolver = evaluator.getVariableResolver();
        if (resolver instanceof GURPSCharacter) {
            if (arguments.startsWith("\"") && arguments.endsWith("\"")) {
                arguments = arguments.substring(1, arguments.length() - 1);
            }
            GURPSCharacter character = (GURPSCharacter) resolver;
            for (Advantage advantage : character.getAdvantagesIterator(false)) {
                if (advantage.getName().equalsIgnoreCase(arguments)) {
                    if (advantage.isLeveled()) {
                        double levels = advantage.getLevels();
                        if (advantage.allowHalfLevels() && advantage.hasHalfLevel()) {
                            levels += 0.5;
                        }
                        return Double.valueOf(levels);
                    }
                    break;
                }
            }
        }
        return Double.valueOf(-1);
    }
}
