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

package com.trollworks.gcs.calendar;

import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;

/** Season defines a seasonal period in the calendar. */
public class Season {
    private static final String KEY_NAME        = "name";
    private static final String KEY_START_MONTH = "start_month";
    private static final String KEY_START_DAY   = "start_day";
    private static final String KEY_END_MONTH   = "end_month";
    private static final String KEY_END_DAY     = "end_day";
    public               String mName;
    public               int    mStartMonth;
    public               int    mStartDay;
    public               int    mEndMonth;
    public               int    mEndDay;

    /** Create a new season. */
    public Season(String name, int startMonth, int startDay, int endMonth, int endDay) {
        mName = name;
        mStartMonth = startMonth;
        mStartDay = startDay;
        mEndMonth = endMonth;
        mEndDay = endDay;
    }

    /** Create a new season from json. */
    public Season(JsonMap m) {
        mName = m.getString(KEY_NAME);
        mStartMonth = m.getInt(KEY_START_MONTH);
        mStartDay = m.getInt(KEY_START_DAY);
        mEndMonth = m.getInt(KEY_END_MONTH);
        mEndDay = m.getInt(KEY_END_DAY);
    }

    /** Save the data as json. */
    public void save(JsonWriter w) throws IOException {
        w.startMap();
        w.keyValue(KEY_NAME, mName);
        w.keyValue(KEY_START_MONTH, mStartMonth);
        w.keyValue(KEY_START_DAY, mStartDay);
        w.keyValue(KEY_END_MONTH, mEndMonth);
        w.keyValue(KEY_END_DAY, mEndDay);
        w.endMap();
    }

    @SuppressWarnings("AutoBoxing")
    @Override
    public String toString() {
        if (mStartMonth == mEndMonth && mStartDay == mEndDay) {
            return String.format("%s (%d/%d)", mName, mStartMonth, mStartDay);
        }
        return String.format("%s (%d/%d-%d/%d)", mName, mStartMonth, mStartDay, mEndMonth, mEndDay);
    }

    /** @return null if the season data is usable for the given calendar. */
    public String checkValidity(Calendar cal) {
        if (mName == null || mName.isBlank()) {
            return I18n.text("Calendar season names must not be empty");
        }
        if (mStartMonth < 1 || mStartMonth > cal.mMonths.size()) {
            return I18n.text("Calendar seasons must start in a valid month");
        }
        if (mStartDay < 1 || mStartDay > cal.mMonths.get(mStartMonth - 1).mDays) {
            return I18n.text("Calendar seasons must start in a valid day within the month");
        }
        if (mEndMonth < 1 || mEndMonth > cal.mMonths.size()) {
            return I18n.text("Calendar seasons must end in a valid month");
        }
        if (mEndDay < 1 || mEndDay > cal.mMonths.get(mEndMonth - 1).mDays) {
            return I18n.text("Calendar seasons must end in a valid day within the month");
        }
        return null;
    }
}
