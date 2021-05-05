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

import com.trollworks.gcs.expression.function.Abs;
import com.trollworks.gcs.expression.function.Cbrt;
import com.trollworks.gcs.expression.function.Ceil;
import com.trollworks.gcs.expression.function.DisplayDice;
import com.trollworks.gcs.expression.function.Exp;
import com.trollworks.gcs.expression.function.ExpressionFunction;
import com.trollworks.gcs.expression.function.Floor;
import com.trollworks.gcs.expression.function.If;
import com.trollworks.gcs.expression.function.Log;
import com.trollworks.gcs.expression.function.Log10;
import com.trollworks.gcs.expression.function.Log1p;
import com.trollworks.gcs.expression.function.Max;
import com.trollworks.gcs.expression.function.Min;
import com.trollworks.gcs.expression.function.Pow;
import com.trollworks.gcs.expression.function.Roll;
import com.trollworks.gcs.expression.function.Round;
import com.trollworks.gcs.expression.function.Signed;
import com.trollworks.gcs.expression.function.Sqrt;
import com.trollworks.gcs.expression.operator.Add;
import com.trollworks.gcs.expression.operator.And;
import com.trollworks.gcs.expression.operator.CloseParen;
import com.trollworks.gcs.expression.operator.Divide;
import com.trollworks.gcs.expression.operator.Equal;
import com.trollworks.gcs.expression.operator.GreaterThan;
import com.trollworks.gcs.expression.operator.GreaterThanOrEqual;
import com.trollworks.gcs.expression.operator.LessThan;
import com.trollworks.gcs.expression.operator.LessThanOrEqual;
import com.trollworks.gcs.expression.operator.Mod;
import com.trollworks.gcs.expression.operator.Multiply;
import com.trollworks.gcs.expression.operator.Not;
import com.trollworks.gcs.expression.operator.NotEqual;
import com.trollworks.gcs.expression.operator.OpenParen;
import com.trollworks.gcs.expression.operator.Operator;
import com.trollworks.gcs.expression.operator.Or;
import com.trollworks.gcs.expression.operator.Power;
import com.trollworks.gcs.expression.operator.Subtract;
import com.trollworks.gcs.utility.I18n;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/** A simple expression evaluator. */
public class Evaluator {
    private static List<Operator>                  DEFAULT_OPERATORS = new ArrayList<>();
    private static Map<String, ExpressionFunction> DEFAULT_FUNCTIONS = new HashMap<>();
    private        VariableResolver                mVariableResolver;
    private        List<Operator>                  mOperators        = new ArrayList<>();
    private        Map<String, ExpressionFunction> mFunctions        = new HashMap<>();
    private        Stack<ExpressionOperator>       mOperatorStack;
    private        Stack<Object>                   mOperandStack;

    static {
        addDefaultOperator(new OpenParen());
        addDefaultOperator(new CloseParen());
        addDefaultOperator(new Add());
        addDefaultOperator(new Subtract());
        addDefaultOperator(new Power());
        addDefaultOperator(new Multiply());
        addDefaultOperator(new Divide());
        addDefaultOperator(new Equal());
        addDefaultOperator(new NotEqual());
        addDefaultOperator(new LessThanOrEqual());
        addDefaultOperator(new LessThan());
        addDefaultOperator(new GreaterThanOrEqual());
        addDefaultOperator(new GreaterThan());
        addDefaultOperator(new And());
        addDefaultOperator(new Or());
        addDefaultOperator(new Not());
        addDefaultOperator(new Mod());

        addDefaultFunction(new Abs());
        addDefaultFunction(new Cbrt());
        addDefaultFunction(new Ceil());
        addDefaultFunction(new DisplayDice());
        addDefaultFunction(new Exp());
        addDefaultFunction(new Floor());
        addDefaultFunction(new If());
        addDefaultFunction(new Log());
        addDefaultFunction(new Log1p());
        addDefaultFunction(new Log10());
        addDefaultFunction(new Max());
        addDefaultFunction(new Min());
        addDefaultFunction(new Pow());
        addDefaultFunction(new Roll());
        addDefaultFunction(new Round());
        addDefaultFunction(new Signed());
        addDefaultFunction(new Sqrt());
    }

    public static final void addDefaultOperator(Operator operator) {
        DEFAULT_OPERATORS.add(operator);
    }

    public static final void addDefaultFunction(ExpressionFunction function) {
        DEFAULT_FUNCTIONS.put(function.getName(), function);
    }

    /** Creates a new {@link Evaluator} that does not do variable resolution. */
    public Evaluator() {
        this((VariableResolver) null);
    }

    /**
     * Creates a new {@link Evaluator}.
     *
     * @param variableResolver The {@link VariableResolver} to use.
     */
    public Evaluator(VariableResolver variableResolver) {
        mVariableResolver = variableResolver;
        mOperators.addAll(DEFAULT_OPERATORS);
        mFunctions.putAll(DEFAULT_FUNCTIONS);
    }

    /**
     * Creates a new {@link Evaluator}.
     *
     * @param other An {@link Evaluator} to copy the {@link VariableResolver} from.
     */
    public Evaluator(Evaluator other) {
        mVariableResolver = other.mVariableResolver;
        mOperators.addAll(other.mOperators);
        mFunctions.putAll(other.mFunctions);
    }

    /** @return The current variable resolver. */
    public VariableResolver getVariableResolver() {
        return mVariableResolver;
    }

    /** @param variableResolver The {@link VariableResolver} to use. */
    public void setVariableResolver(VariableResolver variableResolver) {
        mVariableResolver = variableResolver;
    }

    /**
     * Add a new {@link Operator}.
     *
     * @param operator The {@link Operator}.
     */
    public void addOperator(Operator operator) {
        mOperators.add(operator);
    }

    /**
     * Add a new {@link ExpressionFunction}.
     *
     * @param function The {@link ExpressionFunction}.
     */
    public void addFunction(ExpressionFunction function) {
        mFunctions.put(function.getName(), function);
    }

    /**
     * Evaluate an expression and return a result.
     *
     * @param expression The expression to evaluate.
     * @return The result. May be a {@link String} or a {@link Double}.
     */
    public final Object evaluate(String expression) throws EvaluationException {
        parse(expression);
        while (!mOperatorStack.isEmpty()) {
            processTree(mOperandStack, mOperatorStack);
        }
        if (mOperandStack.isEmpty()) {
            return "";
        }
        return evaluateOperand(mOperandStack.pop());
    }

    /**
     * Evaluate an expression and return a number result.
     *
     * @param expression The expression to evaluate.
     * @return The result. If the result can't be translated to a number, an EvaluationException
     *         will be thrown.
     */
    public final double evaluateToNumber(String expression) throws EvaluationException {
        try {
            return ArgumentTokenizer.getForcedDouble(evaluate(expression));
        } catch (Exception exception) {
            throw new EvaluationException(exception);
        }
    }

    /**
     * Evaluate an expression and return a integer result, truncating any fractions.
     *
     * @param expression The expression to evaluate.
     * @return The result. If the result can't be translated to an integer, an EvaluationException
     *         will be thrown.
     */
    public final int evaluateToInteger(String expression) throws EvaluationException {
        try {
            return (int) Math.floor(ArgumentTokenizer.getForcedDouble(evaluate(expression)));
        } catch (Exception exception) {
            throw new EvaluationException(exception);
        }
    }

    private void processTree(Stack<Object> operandStack, Stack<ExpressionOperator> operatorStack) {
        Object rightOperand = operandStack.isEmpty() ? null : operandStack.pop();
        Object leftOperand  = operandStack.isEmpty() ? null : operandStack.pop();
        operandStack.push(new ExpressionTree(this, leftOperand, rightOperand, operatorStack.pop().mOperator, null));
    }

    private void parse(String expression) throws EvaluationException {
        try {
            mOperandStack = new Stack<>();
            mOperatorStack = new Stack<>();
            boolean  haveOperand   = false;
            boolean  haveOperator  = false;
            Operator unaryOperator = null;
            int      max           = expression.length();
            int      i             = 0;
            while (i < max) {
                if (Character.isWhitespace(expression.charAt(i))) {
                    i++;
                } else {
                    Operator     operator     = null;
                    int          opIndex      = -1;
                    NextOperator nextOperator = nextOperator(expression, i, null);
                    if (nextOperator != null) {
                        operator = nextOperator.mOperator;
                        opIndex = nextOperator.mIndex;
                    }
                    if (opIndex > i || opIndex == -1) {
                        i = processOperand(expression, i, opIndex, mOperandStack, unaryOperator);
                        haveOperand = true;
                        haveOperator = false;
                        unaryOperator = null;
                    }
                    if (opIndex == i) {
                        if (nextOperator != null && nextOperator.mOperator.isUnary() && (haveOperator || i == 0)) {
                            i = opIndex + nextOperator.mOperator.getSymbol().length();
                            if (unaryOperator == null) {
                                unaryOperator = nextOperator.mOperator;
                            } else {
                                throw new EvaluationException(String.format(I18n.Text("Consecutive unary operators are not allowed (index=%d)"), Integer.valueOf(i)));
                            }
                        } else {
                            i = processOperator(expression, opIndex, operator, mOperatorStack, mOperandStack, haveOperand, unaryOperator);
                            unaryOperator = null;
                        }
                        if (!(nextOperator != null && nextOperator.mOperator instanceof CloseParen)) {
                            haveOperand = false;
                            haveOperator = true;
                        }
                    }
                }
            }
        } catch (EvaluationException evalEx) {
            throw evalEx;
        } catch (Exception exception) {
            throw new EvaluationException(exception.getMessage(), exception);
        }
    }

    private static String invalidExpressionText() {
        return I18n.Text("Expression is invalid");
    }

    private static int processOperand(String expression, int start, int operatorIndex, Stack<Object> operandStack, Operator unaryOperator) throws EvaluationException {
        String text;
        int    result;
        if (operatorIndex == -1) {
            text = expression.substring(start).trim();
            result = expression.length();
        } else {
            text = expression.substring(start, operatorIndex).trim();
            result = operatorIndex;
        }
        if (text.isEmpty()) {
            throw new EvaluationException(invalidExpressionText());
        }
        operandStack.push(new ExpressionOperand(text, unaryOperator));
        return result;
    }

    private int processOperator(String expression, int index, Operator operator, Stack<ExpressionOperator> operatorStack, Stack<Object> operandStack, boolean haveOperand, Operator unaryOperator) throws EvaluationException {
        if (haveOperand && operator instanceof OpenParen) {
            NextOperator nextOperator = processFunction(expression, index, operandStack);
            operator = nextOperator.mOperator;
            index = nextOperator.mIndex + operator.getLength();
            nextOperator = nextOperator(expression, index, null);
            if (nextOperator == null) {
                return index;
            }
            operator = nextOperator.mOperator;
            index = nextOperator.mIndex;
        }
        if (operator instanceof OpenParen) {
            operatorStack.push(new ExpressionOperator(operator, unaryOperator));
        } else if (operator instanceof CloseParen) {
            ExpressionOperator stackOp = operatorStack.isEmpty() ? null : operatorStack.peek();
            while (stackOp != null && !(stackOp.mOperator instanceof OpenParen)) {
                processTree(operandStack, operatorStack);
                stackOp = operatorStack.isEmpty() ? null : operatorStack.peek();
            }
            if (operatorStack.isEmpty()) {
                throw new EvaluationException(invalidExpressionText());
            }
            ExpressionOperator exop = operatorStack.pop();
            if (!(exop.mOperator instanceof OpenParen)) {
                throw new EvaluationException(invalidExpressionText());
            }
            if (exop.mUnaryOperator != null) {
                operandStack.push(new ExpressionTree(this, operandStack.pop(), null, null, exop.mUnaryOperator));
            }
        } else {
            if (!operatorStack.isEmpty()) {
                ExpressionOperator stackOp = operatorStack.peek();
                while (stackOp != null && stackOp.mOperator.getPrecedence() >= operator.getPrecedence()) {
                    processTree(operandStack, operatorStack);
                    stackOp = operatorStack.isEmpty() ? null : operatorStack.peek();
                }
            }
            operatorStack.push(new ExpressionOperator(operator, unaryOperator));
        }
        return index + operator.getLength();
    }

    private NextOperator processFunction(String expression, int operatorIndex, Stack<Object> operandStack) throws EvaluationException {
        int          parens       = 1;
        NextOperator nextOperator = null;
        int          next         = operatorIndex;
        while (parens > 0) {
            nextOperator = nextOperator(expression, next + 1, null);
            if (nextOperator == null) {
                throw new EvaluationException(I18n.Text("Function not closed"));
            } else if (nextOperator.mOperator instanceof OpenParen) {
                parens++;
            } else if (nextOperator.mOperator instanceof CloseParen) {
                parens--;
            }
            next = nextOperator.mIndex;
        }
        ExpressionOperand  operand  = (ExpressionOperand) operandStack.pop();
        ExpressionFunction function = mFunctions.get(operand.mValue);
        if (function == null) {
            throw new EvaluationException(String.format(I18n.Text("Function not defined: %s"), operand.mValue));
        }
        operandStack.push(new ParsedFunction(function, expression.substring(operatorIndex + 1, next), operand.mUnaryOperator));
        return nextOperator;
    }

    private NextOperator nextOperator(String expression, int start, Operator match) {
        int length = expression.length();
        for (int i = start; i < length; i++) {
            if (match != null) {
                NextOperator next = nextOperator(expression, i, length, match);
                if (next != null) {
                    return next;
                }
            } else {
                for (Operator operator : mOperators) {
                    NextOperator next = nextOperator(expression, i, length, operator);
                    if (next != null) {
                        return next;
                    }
                }
            }
        }
        return null;
    }

    private static NextOperator nextOperator(String expression, int start, int max, Operator operator) {
        int    length = operator.getLength();
        String symbol = operator.getSymbol();
        if (length == 1) {
            char symbolChar = symbol.charAt(0);
            if (expression.charAt(start) == symbolChar) {
                // Hack to allow negative exponents on floating point numbers (i.e. 1.2e-2)
                if (symbolChar == '-' && start > 1) {
                    if (expression.charAt(start - 1) == 'e' && Character.isDigit(expression.charAt(start - 2))) {
                        return null;
                    }
                }
                return new NextOperator(operator, start);
            }
        } else {
            if (expression.regionMatches(start, symbol, 0, start + length <= max ? length : max - start)) {
                return new NextOperator(operator, start);
            }
        }
        return null;
    }

    final Object evaluateOperand(Object operand) throws EvaluationException {
        if (operand instanceof ExpressionTree) {
            return ((ExpressionTree) operand).evaluate();
        } else if (operand instanceof ExpressionOperand) {
            ExpressionOperand exop  = (ExpressionOperand) operand;
            Object            value = replaceVariables(exop.mValue);
            Operator          unary = exop.mUnaryOperator;
            return unary != null ? unary.evaluate(value) : value;
        } else if (operand instanceof ParsedFunction) {
            ParsedFunction function = (ParsedFunction) operand;
            Object         value    = function.mFunction.execute(this, replaceVariables(function.mArguments));
            if (function.mUnaryOperator != null) {
                value = function.mUnaryOperator.evaluate(value);
            }
            return value;
        } else if (operand != null) {
            throw new EvaluationException(invalidExpressionText());
        }
        return null;
    }

    private String replaceVariables(String expression) throws EvaluationException {
        int dollar = expression.indexOf('$');
        while (dollar >= 0) {
            int last = dollar;
            int max  = expression.length();
            for (int i = dollar + 1; i < max; i++) {
                char ch = expression.charAt(i);
                if (ch == '_' || ch == '.' || ch == '#' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' || i != dollar + 1 && ch >= '0' && ch <= '9') {
                    last = i;
                } else {
                    break;
                }
            }
            if (dollar == last) {
                throw new EvaluationException(String.format(I18n.Text("Invalid variable at index %d"), Integer.valueOf(dollar)));
            } else {
                String name  = expression.substring(dollar + 1, last + 1);
                String value = null;
                if (mVariableResolver != null) {
                    value = mVariableResolver.resolveVariable(name);
                }
                if (value == null || value.trim().isEmpty()) {
                    throw new EvaluationException(String.format(I18n.Text("Unable to resolve variable $%s"), name));
                }
                StringBuilder buffer = new StringBuilder();
                if (dollar > 0) {
                    buffer.append(expression, 0, dollar);
                }
                buffer.append(value);
                if (last + 1 < max) {
                    buffer.append(expression.substring(last + 1));
                }
                expression = buffer.toString();
            }
            dollar = expression.indexOf('$');
        }
        return expression;
    }
}
