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

package com.trollworks.gcs.calendar;

import com.trollworks.gcs.utility.I18n;

/** LeapYear holds parameters for determining leap years. */
public class LeapYear {
    public int mMonth;
    public int mEvery;
    public int mExcept;
    public int mUnless;

    public LeapYear(int month, int every, int except, int unless) {
        mMonth = month;
        mEvery = every;
        mExcept = except;
        mUnless = unless;
    }

    /** @return null if the leap year data is usable for the given calendar. */
    public String checkValidity(Calendar cal) {
        if (mMonth < 1 || mMonth > cal.mMonths.size()) {
            return I18n.text("Leap Year Month must specify a valid month");
        }
        if (mEvery < 2) {
            return I18n.text("Leap Year Every may not be less than 2");
        }
        if (mExcept != 0) {
            if (mExcept <= mEvery) {
                return I18n.text("Leap Year Except must be greater than Leap Year Every if not 0");
            }
            if ((mExcept / mEvery) * mEvery != mExcept) {
                return I18n.text("Leap Year Except must be a multiple of Leap Year Every");
            }
        }
        if (mUnless != 0) {
            if (mExcept == 0) {
                return I18n.text("Leap Year Unless may not be set if Leap Year Except is 0");
            }
            if (mUnless <= mExcept) {
                return I18n.text("Leap Year Unless must be greater than Leap Year Except if not 0");
            }
            if ((mUnless / mExcept) * mExcept != mUnless) {
                return I18n.text("Leap Year Unless must be a multiple of Leap Year Except");
            }
        }
        return null;
    }

    /** @return true if the year is a leap year. */
    public boolean is(int year) {
        if (year < 1) {
            year++; // account for gap, since there is no year 0
        }
        if (year % mEvery == 0) {
            if (mExcept != 0) {
                if (year % mExcept == 0) {
                    if (mUnless != 0) {
                        return year % mUnless == 0;
                    }
                    return false;
                }
            }
            return true;
        }
        return false;
    }

    /**
     * @return the number of leap years that have occurred between year 1 and the specified year,
     *         exclusive.
     */
    public int since(int year) {
        if (year == -1) {
            return 0;
        }
        int delta = year;
        if (delta < 1) {
            delta = -(delta + 1); // make it positive and account for gap, since there is no year 0
        }
        int count = delta / mEvery;
        if (mExcept != 0) {
            count -= delta / mExcept;
            if (mUnless != 0) {
                count += delta / mUnless;
            }
        }
        if (is(year)) {
            count--;
        }
        if (year < -1) {
            count++;
        }
        return count;
    }
}
