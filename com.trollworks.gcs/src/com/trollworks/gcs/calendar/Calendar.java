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

import java.util.List;

public class Calendar {
    public int          mDayZeroWeekDay;
    public List<String> mWeekDays;
    public List<Month>  mMonths;
    public List<Season> mSeasons;
    public String       mEra;
    public String       mPreviousEra;
    public LeapYear     mLeapYear;

    /** Creates a default Gregorian calendar. */
    public Calendar() {
        mDayZeroWeekDay = 1;
        mWeekDays = List.of("Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday");
        mMonths = List.of(
                new Month("January", 31),
                new Month("Feburary", 28),
                new Month("March", 31),
                new Month("April", 30),
                new Month("May", 31),
                new Month("June", 30),
                new Month("July", 31),
                new Month("August", 31),
                new Month("September", 30),
                new Month("October", 31),
                new Month("November", 30),
                new Month("December", 31)
        );
        mSeasons = List.of(
                new Season("Winter", 11, 1, 2, 28),
                new Season("Spring", 3, 1, 5, 31),
                new Season("Summer", 6, 1, 8, 31),
                new Season("Fall", 9, 1, 10, 31)
        );
        mEra = "AD";
        mPreviousEra = "BC";
        mLeapYear = new LeapYear(2, 4, 100, 400);
    }

    /** @return null if the calendar data is usable. */
    public String checkValidity() {
        if (mWeekDays == null || mWeekDays.isEmpty()) {
            return I18n.text("Calendar must have at least one week day");
        }
        if (mMonths == null || mMonths.isEmpty()) {
            return I18n.text("Calendar must have at least one month");
        }
        if (mSeasons == null || mSeasons.isEmpty()) {
            return I18n.text("Calendar must have at least one season");
        }
        if (mDayZeroWeekDay < 0 || mDayZeroWeekDay >= mWeekDays.size()) {
            return I18n.text("Calendar's first week day of the first year must be a valid week day");
        }
        for (String weekday : mWeekDays) {
            if (weekday.isBlank()) {
                return I18n.text("Calendar week day names must not be empty");
            }
        }
        for (Month month : mMonths) {
            String check = month.checkValidity();
            if (check != null) {
                return check;
            }
        }
        for (Season season : mSeasons) {
            String check = season.checkValidity(this);
            if (check != null) {
                return check;
            }
        }
        return mLeapYear != null ? mLeapYear.checkValidity(this) : null;
    }

    /** @return the minimum number of days in a year. */
    public int minDaysPerYear() {
        int days = 0;
        for (Month month : mMonths) {
            days += month.mDays;
        }
        return days;
    }

    /** @return the number of days contained in a specific year. */
    public int days(int year) {
        int days = minDaysPerYear();
        return isLeapYear(year) ? days + 1 : days;
    }

    /** @return true if the year is a leap year. */
    public boolean isLeapYear(int year) {
        return mLeapYear != null && mLeapYear.is(year);
    }

    /** @return true if the month is the leap month. */
    public boolean isLeapMonth(int month) {
        return mLeapYear != null && mLeapYear.mMonth == month;
    }

    /** @return a text representation of the year. */
    public String text(int year) throws CalendarException {
        StringBuilder buffer = new StringBuilder();
        Date          date   = new Date(this, 1, 1, year);
        buffer.append(date.format("Year %Y\n"));
        int count = mMonths.size();
        for (int i = 1; i <= count; i++) {
            buffer.append('\n');
            date = new Date(this, i, 1, year);
            buffer.append(date.textCalendarMonth());
        }
        buffer.append("\nSeasons:\n");
        for (Season season : mSeasons) {
            buffer.append("  ");
            buffer.append(season);
            buffer.append('\n');
        }
        buffer.append("\nWeek Days:\n");
        count = mWeekDays.size();
        for (int i = 0; i < count; i++) {
            buffer.append("  ");
            buffer.append(i + 1);
            buffer.append(": (");
            String weekday = mWeekDays.get(i);
            buffer.append(weekday.charAt(0));
            buffer.append(") ");
            buffer.append(weekday);
            buffer.append('\n');
        }
        return buffer.toString();
    }
}
