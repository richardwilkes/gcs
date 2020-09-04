/*
 * Copyright ©1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.utility.text;

import java.awt.Toolkit;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.text.DecimalFormat;
import java.text.DecimalFormatSymbols;
import java.text.NumberFormat;
import javax.swing.JTextField;

/** A standard numeric key entry filter. */
public class NumberFilter implements KeyListener {
    private static final char       GROUP_CHAR;
    private static final char       DECIMAL_CHAR;
    private              JTextField mField;
    private              boolean    mAllowDecimal;
    private              boolean    mAllowSign;
    private              boolean    mAllowGroup;
    private              int        mMaxDigits;

    static {
        DecimalFormatSymbols symbols = ((DecimalFormat) NumberFormat.getNumberInstance()).getDecimalFormatSymbols();
        GROUP_CHAR = symbols.getGroupingSeparator();
        DECIMAL_CHAR = symbols.getDecimalSeparator();
    }

    /**
     * Applies a new numeric key entry filter to a field.
     *
     * @param field        The {@link JTextField} to filter.
     * @param allowDecimal Pass in {@code true} to allow floating point.
     * @param allowSign    Pass in {@code true} to allow sign characters.
     * @param allowGroup   Pass in {@code true} to allow group characters.
     * @param maxDigits    The maximum number of digits (not necessarily characters) the field can
     *                     have.
     */
    public static void apply(JTextField field, boolean allowDecimal, boolean allowSign, boolean allowGroup, int maxDigits) {
        field.addKeyListener(new NumberFilter(field, allowDecimal, allowSign, allowGroup, maxDigits));
    }

    private NumberFilter(JTextField field, boolean allowDecimal, boolean allowSign, boolean allowGroup, int maxDigits) {
        mField = field;
        mAllowDecimal = allowDecimal;
        mAllowSign = allowSign;
        mAllowGroup = allowGroup;
        mMaxDigits = maxDigits;
        for (KeyListener listener : mField.getKeyListeners()) {
            if (listener instanceof NumberFilter) {
                mField.removeKeyListener(listener);
            }
        }
        mField.addKeyListener(this);
    }

    @Override
    public void keyPressed(KeyEvent event) {
        // Not used.
    }

    @Override
    public void keyReleased(KeyEvent event) {
        // Not used.
    }

    @Override
    public void keyTyped(KeyEvent event) {
        if (!event.isConsumed() && (event.getModifiersEx() & mField.getToolkit().getMenuShortcutKeyMaskEx()) == 0) {
            char ch = event.getKeyChar();
            if (ch != '\n' && ch != '\r' && ch != '\t' && ch != '\b' && ch != KeyEvent.VK_DELETE) {
                if (mAllowGroup && ch == GROUP_CHAR || ch >= '0' && ch <= '9' || mAllowSign && (ch == '-' || ch == '+') || mAllowDecimal && ch == DECIMAL_CHAR) {
                    StringBuilder buffer = new StringBuilder(mField.getText());
                    int           start  = mField.getSelectionStart();
                    int           end    = mField.getSelectionEnd();
                    if (start != end) {
                        buffer.delete(start, end);
                    }
                    if (ch >= '0' && ch <= '9') {
                        int length = buffer.length();
                        int count  = 0;
                        for (int i = 0; i < length; i++) {
                            char one = buffer.charAt(i);
                            if (one >= '0' && one <= '9') {
                                count++;
                            }
                        }
                        if (count >= mMaxDigits) {
                            filter(event);
                            return;
                        }
                    }
                    if (ch == GROUP_CHAR || ch >= '0' && ch <= '9') {
                        if (mAllowSign && start == 0 && !buffer.isEmpty() && (buffer.charAt(0) == '-' || buffer.charAt(0) == '+')) {
                            filter(event);
                        }
                    } else if (ch == '-' || ch == '+') {
                        if (start != 0) {
                            filter(event);
                        }
                    } else if (buffer.indexOf("" + DECIMAL_CHAR) != -1 || mAllowSign && start == 0 && !buffer.isEmpty() && (buffer.charAt(0) == '-' || buffer.charAt(0) == '+')) {
                        filter(event);
                    }
                } else {
                    filter(event);
                }
            }
        }
    }

    private static void filter(KeyEvent event) {
        Toolkit.getDefaultToolkit().beep();
        event.consume();
    }
}
