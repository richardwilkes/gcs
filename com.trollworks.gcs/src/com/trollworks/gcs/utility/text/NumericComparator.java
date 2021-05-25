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
        if (mCaseless ? left.equalsIgnoreCase(right) : left.equals(right)) {
            return 0;
        }
        if (naturalLess(left, right, mCaseless)) {
            return -1;
        }
        return 1;
    }

    // NaturalLess compares two strings using natural ordering. This means that
    // "a2" < "a12".
    //
    // Non-digit sequences and numbers are compared separately. The former are
    // compared byte-wise, while the latter are compared numerically (except that
    // the number of leading zeros is used as a tie-breaker, so "2" < "02").
    //
    // Limitations:
    //  - only ASCII digits (0-9) are considered.
    //
    // Original algorithm:
    //    https://github.com/fvbommel/util/blob/master/sortorder/natsort.go
    private static boolean naturalLess(String s1, String s2, boolean caseInsensitive) {
        int l1 = s1.length();
        int l2 = s2.length();
        int i1 = 0;
        int i2 = 0;
        while (i1 < l1 && i2 < l2) {
            char    c1 = s1.charAt(i1);
            char    c2 = s2.charAt(i2);
            boolean d1 = isDigit(c1);
            boolean d2 = isDigit(c2);
            if (d1 != d2) { // Digits before other characters
                return d1; // True if LHS is a digit, false if the RHS is one.
            }
            if (!d1) {
                if (caseInsensitive) {
                    if (c1 >= 'a' && c1 <= 'z') {
                        c1 -= 'a' - 'A';
                    }
                    if (c2 >= 'a' && c2 <= 'z') {
                        c2 -= 'a' - 'A';
                    }
                }
                if (c1 != c2) {
                    return c1 < c2;
                }
                i1++;
                i2++;
            } else { // Digits
                // Eat zeros.
                while (i1 < l1 && s1.charAt(i1) == '0') {
                    i1++;
                }
                while (i2 < l2 && s2.charAt(i2) == '0') {
                    i2++;
                }
                // Eat all digits.
                int nz1 = i1;
                int nz2 = i2;
                while (i1 < l1 && isDigit(s1.charAt(i1))) {
                    i1++;
                }
                while (i2 < l2 && isDigit(s2.charAt(i2))) {
                    i2++;
                }
                // If lengths of numbers with non-zero prefix differ, the shorter
                // one is less.
                int len1 = i1 - nz1;
                int len2 = i2 - nz2;
                if (len1 != len2) {
                    return len1 < len2;
                }
                // If they're not equal, string comparison is correct.
                String nr1 = s1.substring(nz1, i1);
                String nr2 = s2.substring(nz2, i2);
                if (!nr1.equals(nr2)) {
                    return nr1.compareTo(nr2) < 0;
                }
                // Otherwise, the one with less zeros is less.
                // Because everything up to the number is equal, comparing the index
                // after the zeros is sufficient.
                if (nz1 != nz2) {
                    return nz1 < nz2;
                }
            }
            // They're identical so far, so continue comparing.
        }
        // So far they are identical. At least one is ended. If the other continues,
        // it sorts last. If the are the same length and the caseInsensitive flag
        // was set, compare again, but without the flag.
        if (caseInsensitive && l1 == l2) {
            return naturalLess(s1, s2, false);
        }
        return l1 < l2;
    }

    private static boolean isDigit(char ch) {
        return ch >= '0' && ch <= '9';
    }
}
