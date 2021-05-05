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

package com.trollworks.gcs.expression.operator;

import com.trollworks.gcs.expression.EvaluationException;

public abstract class Operator {
    private String  mSymbol;
    private int     mPrecedence;
    private boolean mUnary;

    protected Operator(String symbol, int precedence) {
        mSymbol = symbol;
        mPrecedence = precedence;
    }

    protected Operator(String symbol, int precedence, boolean unary) {
        mSymbol = symbol;
        mPrecedence = precedence;
        mUnary = unary;
    }

    public abstract Object evaluate(Object left, Object right) throws EvaluationException;

    public abstract Object evaluate(Object operand) throws EvaluationException;

    public final String getSymbol() {
        return mSymbol;
    }

    public final int getPrecedence() {
        return mPrecedence;
    }

    public final int getLength() {
        return mSymbol.length();
    }

    public final boolean isUnary() {
        return mUnary;
    }

    @Override
    public final boolean equals(Object object) {
        if (object == this) {
            return true;
        }
        if (object == null) {
            return false;
        }
        if (object instanceof Operator) {
            return mSymbol.equals(((Operator) object).getSymbol());
        }
        return false;
    }

    @Override
    public final int hashCode() {
        return mSymbol.hashCode();
    }

    @Override
    public final String toString() {
        return getSymbol();
    }
}
