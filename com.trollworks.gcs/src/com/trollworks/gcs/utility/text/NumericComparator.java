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

import java.util.Comparator;

/**
 * A string comparator that will honor numeric values embedded in strings and treat them as numbers
 * for comparison purposes.
 */
public class NumericComparator implements Comparator<String> {
    /** The standard caseless, numeric-aware string comparator. */
    public static final NumericComparator CASELESS_COMPARATOR = new NumericComparator(true);
    /** The standard case-sensitive, numeric-aware string comparator. */
    public static final NumericComparator COMPARATOR          = new NumericComparator(false);
    private             boolean           mCaseless;

    private NumericComparator(boolean caseless) {
        mCaseless = caseless;
    }

    /**
     * Convenience method for caseless comparisons.
     *
     * @param s0 The first string.
     * @param s1 The second string.
     * @return A negative integer, zero, or a positive integer if the first argument is less than,
     *         equal to, or greater than the second.
     */
    public static final int caselessCompareStrings(String s0, String s1) {
        return CASELESS_COMPARATOR.compare(s0, s1);
    }

    /**
     * Convenience method for case-sensitive comparisons.
     *
     * @param s0 The first string.
     * @param s1 The second string.
     * @return A negative integer, zero, or a positive integer if the first argument is less than,
     *         equal to, or greater than the second.
     */
    public static final int compareStrings(String s0, String s1) {
        return COMPARATOR.compare(s0, s1);
    }

    @Override
    public int compare(String left, String right) {
        if (left == null) {
            left = "";
        }
        if (right == null) {
            right = "";
        }
        char[] chars0          = left.toCharArray();
        char[] chars1          = right.toCharArray();
        int    pos0            = 0;
        int    pos1            = 0;
        int    len0            = chars0.length;
        int    len1            = chars1.length;
        int    result          = 0;
        int    secondaryResult = 0;
        char   c0;
        char   c1;

        while (result == 0 && pos0 < len0 && pos1 < len1) {
            boolean normalCompare = true;

            c0 = chars0[pos0++];
            c1 = chars1[pos1++];

            if (isDigit(c0) && isDigit(c1)) {
                int count0 = 1;
                int count1 = 1;

                while (pos0 < len0) {
                    c0 = chars0[pos0];
                    if (isDigit(c0)) {
                        count0++;
                        pos0++;
                    } else {
                        break;
                    }
                }

                while (pos1 < len1) {
                    c1 = chars1[pos1];
                    if (isDigit(c1)) {
                        count1++;
                        pos1++;
                    } else {
                        break;
                    }
                }

                try {
                    normalCompare = false;
                    result = Long.compare(Long.parseLong(left.substring(pos0 - count0, pos0)), Long.parseLong(right.substring(pos1 - count1, pos1)));
                    if (result == 0 && secondaryResult == 0) {
                        secondaryResult = count0 - count1;
                    }
                } catch (NumberFormatException nfex) {
                    pos0 -= count0;
                    pos1 -= count1;
                    c0 = chars0[pos0++];
                    c1 = chars1[pos1++];
                }
            }

            if (normalCompare) {
                if (mCaseless) {
                    int c0Val = Character.isLowerCase(c0) ? Character.toUpperCase(c0) : c0;
                    int c1Val = Character.isLowerCase(c1) ? Character.toUpperCase(c1) : c1;
                    result = c0Val - c1Val;
                } else {
                    result = c0 - c1;
                }
            }
        }

        // string without suffix comes first
        if (result == 0) {
            result = len0 - pos0 - (len1 - pos1);
        }

        // tie breaker
        if (result == 0) {
            result = secondaryResult;
            if (result == 0 && mCaseless) {
                return compareStrings(left, right);
            }
        }

        if (result < 0) {
            result = -1;
        } else if (result > 0) {
            result = 1;
        }

        return result;
    }

    private static boolean isDigit(char ch) {
        return ch >= '0' && ch <= '9';
    }
}
