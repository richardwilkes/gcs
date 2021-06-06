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

package com.trollworks.gcs.expression;

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Numbers;

import java.util.Enumeration;

public class ArgumentTokenizer implements Enumeration<String> {
    private String mArguments;

    public ArgumentTokenizer(String arguments) {
        mArguments = arguments;
    }

    @Override
    public final boolean hasMoreElements() {
        return hasMoreTokens();
    }

    public final boolean hasMoreTokens() {
        return !mArguments.isEmpty();
    }

    @Override
    public final String nextElement() {
        return nextToken();
    }

    public final String nextToken() {
        int length = mArguments.length();
        int parens = 0;
        for (int i = 0; i < length; i++) {
            char ch = mArguments.charAt(i);
            if (ch == '(') {
                parens++;
            } else if (ch == ')') {
                parens--;
            } else if (ch == ',' && parens == 0) {
                String token = mArguments.substring(0, i);
                mArguments = mArguments.substring(i + 1);
                return token;
            }
        }
        String token = mArguments;
        mArguments = "";
        return token;
    }

    public static final double getForcedDouble(Object arg) {
        if (arg instanceof Double) {
            return ((Double) arg).doubleValue();
        }
        return Numbers.extractDouble(arg.toString(), 0, false);
    }

    public static final double getDouble(Object arg) {
        if (arg instanceof Double) {
            return ((Double) arg).doubleValue();
        }
        double value = Numbers.extractDouble(arg.toString(), Double.MAX_VALUE, false);
        if (value == Double.MAX_VALUE) {
            throw new NumberFormatException(arg.toString());
        }
        return value;
    }

    public static final double getDoubleOperand(Object arg) throws EvaluationException {
        try {
            return getDouble(arg);
        } catch (Exception exception) {
            throw new EvaluationException(I18n.text("Invalid operand: ") + arg, exception);
        }
    }

    public static final double getDoubleArgument(Evaluator evaluator, String arguments) throws EvaluationException {
        try {
            return getDouble(new Evaluator(evaluator).evaluate(arguments));
        } catch (Exception exception) {
            throw new EvaluationException(I18n.text("Invalid argument: ") + arguments, exception);
        }
    }
}
