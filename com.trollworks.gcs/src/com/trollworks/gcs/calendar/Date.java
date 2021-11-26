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

import com.trollworks.gcs.utility.Log;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Date holds a calendar date. This is the number of days since 1/1/1 in the calendar. Note that the
 * value -1 refers to the last day of the year -1, not year 0, as there is no year 0.
 */
public class Date {
    public static final  String   FULL_FORMAT      = "%W, %M %D, %Y";
    public static final  String   LONG_FORMAT      = "%M %D, %Y";
    public static final  String   MEDIUM_FORMAT    = "%m %D, %Y";
    public static final  String   SHORT_FORMAT     = "%N/%D/%Y";
    // "9/22/2017" or "9/22/2017 AD"
    private static final Pattern  regexMMDDYYY     = Pattern.compile("(\\d+)/(\\d+)/(-?\\d+) *([A-Za-z]+)?");
    // "September 22, 2017 AD", "September 22, 2017", "Sep 22, 2017 AD", or "Sep 22, 2017"
    private static final Pattern  regexMonthDDYYYY = Pattern.compile("([A-Za-z]+) *(\\d+), *(-?\\d+) *([A-Za-z]+)?");
    private              Calendar mCalendar;
    private              int      mDays;

    /** Creates a new date from a number of days, with 0 representing the date 1/1/1. */
    public Date(Calendar calendar, int days) {
        mCalendar = calendar;
        mDays = days;
    }

    /** Creates a new date from the specified month, day and year. */
    public Date(Calendar calendar, int month, int day, int year) throws CalendarException {
        mCalendar = calendar;
        computeDays(month, day, year);
    }

    /** Creates a new date from the specified text. */
    public Date(Calendar calendar, String in) throws CalendarException {
        mCalendar = calendar;
        Matcher matcher = regexMMDDYYY.matcher(in);
        if (matcher.find()) {
            int month;
            try {
                month = Integer.parseInt(matcher.group(1));
            } catch (NumberFormatException nfe) {
                throw new CalendarException(String.format("invalid month text '%s'", matcher.group(1)), nfe);
            }
            parseDate(month, matcher.group(2), matcher.group(3), matcher.group(4));
        } else {
            matcher = regexMonthDDYYYY.matcher(in);
            if (matcher.find()) {
                String month = matcher.group(1);
                int    count = calendar.mMonths.size();
                for (int i = 0; i < count; i++) {
                    String monthName = calendar.mMonths.get(i).mName;
                    if (monthName.equalsIgnoreCase(month) ||
                            (monthName.length() > 3 && monthName.substring(0, 3).equalsIgnoreCase(month))) {
                        parseDate(i + 1, matcher.group(2), matcher.group(3), matcher.group(4));
                        return;
                    }
                }
                throw new CalendarException(String.format("invalid month text '%s'", month));
            } else {
                throw new CalendarException(String.format("invalid date text '%s'", in));
            }
        }
    }

    private void parseDate(int month, String dayText, String yearText, String eraText) throws CalendarException {
        int year;
        try {
            year = Integer.parseInt(yearText);
        } catch (NumberFormatException nfe) {
            throw new CalendarException(String.format("invalid year text '%s'", yearText), nfe);
        }
        int day;
        try {
            day = Integer.parseInt(dayText);
        } catch (NumberFormatException nfe) {
            throw new CalendarException(String.format("invalid day text '%s'", dayText), nfe);
        }
        if (!mCalendar.mPreviousEra.isBlank() && !mCalendar.mPreviousEra.equals(mCalendar.mEra) &&
                mCalendar.mPreviousEra.equalsIgnoreCase(eraText)) {
            year = -year;
        }
        computeDays(month, day, year);
    }

    @SuppressWarnings("AutoBoxing")
    private void computeDays(int month, int day, int year) throws CalendarException {
        if (year == 0) {
            throw new CalendarException("year 0 is invalid");
        }
        if (month < 1 || month > mCalendar.mMonths.size()) {
            throw new CalendarException(String.format("month %d is invalid", month));
        }
        mDays = mCalendar.mMonths.get(month - 1).mDays;
        if (mCalendar.isLeapMonth(month) && mCalendar.isLeapYear(year)) {
            mDays++;
        }
        if (day < 1 || day > mDays) {
            throw new CalendarException(String.format("day %d is invalid", day));
        }
        mDays = yearToDays(year) + day - 1;
        for (int i = 1; i < month; i++) {
            mDays += mCalendar.mMonths.get(i - 1).mDays;
        }
        if (mCalendar.isLeapYear(year) && mCalendar.mLeapYear.mMonth < month) {
            mDays++;
        }
    }

    private int yearToDays(int year) {
        int days = (year > 1 ? (year - 1) : year) * mCalendar.minDaysPerYear();
        if (mCalendar.mLeapYear != null) {
            int leaps = mCalendar.mLeapYear.since(year);
            if (year > 1) {
                days += leaps;
            } else {
                days -= leaps;
                if (mCalendar.mLeapYear.is(year)) {
                    days--;
                }
            }
        }
        return days;
    }

    /** @return the number of days since 1/1/1. */
    public int days() {
        return mDays;
    }

    /** @return the year of the date. */
    public int year() {
        int estimate = mDays / mCalendar.minDaysPerYear();
        if (mDays < 0) {
            estimate--;
            while (mDays >= yearToDays(estimate + 1)) {
                estimate++;
            }
        } else {
            estimate++;
            while (mDays < yearToDays(estimate)) {
                estimate--;
            }
        }
        return estimate;
    }

    /** @return the month of the date. Note that the first month is represented by 1, not 0. */
    public int month() {
        boolean isLeapYear = mCalendar.isLeapYear(year());
        int     days       = dayInYear();
        int     count      = mCalendar.mMonths.size();
        for (int i = 0; i < count; i++) {
            int amt = mCalendar.mMonths.get(i).mDays;
            if (isLeapYear && mCalendar.isLeapMonth(i + 1)) {
                amt++;
            }
            if (days <= amt) {
                return i + 1;
            }
            days -= amt;
        }
        // If this is reached, the algorithm is wrong.
        Log.error("Unable to determine month");
        return 1;
    }

    /** @return the name of the month of the date. */
    public String monthName() {
        return mCalendar.mMonths.get(month() - 1).mName;
    }

    /**
     * @return the day within the year of the date. Note that the first day is represented by a 1,
     *         not 0.
     */
    public int dayInYear() {
        return 1 + mDays - yearToDays(year());
    }

    /**
     * @return the day within the month of the date. Note that the first day is represented by a 1,
     *         not 0.
     */
    public int dayInMonth() {
        boolean isLeapYear = mCalendar.isLeapYear(year());
        int     days       = dayInYear();
        int     count      = mCalendar.mMonths.size();
        for (int i = 0; i < count; i++) {
            int amt = mCalendar.mMonths.get(i).mDays;
            if (isLeapYear && mCalendar.isLeapMonth(i + 1)) {
                amt++;
            }
            if (days <= amt) {
                return days;
            }
            days -= amt;
        }
        // If this is reached, the algorithm is wrong.
        Log.error("Unable to determine day in month");
        return 1;
    }

    /** @return the number of days in the month of the date. */
    public int daysInMonth() {
        return mCalendar.mMonths.get(month() - 1).mDays;
    }

    /** @return the weekday of the date. */
    public int weekDay() {
        int weekdayCount = mCalendar.mWeekDays.size();
        int weekday      = mDays % weekdayCount;
        if (mDays < 0) {
            weekday += weekdayCount;
        }
        return (weekday + mCalendar.mDayZeroWeekDay) % weekdayCount;
    }

    /** @return the name of the weekday of the date. */
    public String weekDayName() {
        return mCalendar.mWeekDays.get(weekDay());
    }

    /** @return the era suffix for the year. */
    public String era() {
        return year() < 0 ? mCalendar.mPreviousEra : mCalendar.mEra;
    }

    @Override
    public String toString() {
        return format(SHORT_FORMAT);
    }

    /**
     * @return a formatted version of the date. The layout is parsed for directives and anything
     *         that is not a directive is passed through unchanged. Valid directives:
     *         <br><br>
     *         <table>
     *         <tr><td width="30px">%W</td><td>Full weekday, e.g. 'Friday'</td></tr>
     *         <tr><td>%w</td><td>Short weekday, e.g. 'Fri'</td></tr>
     *         <tr><td>%M</td><td>Full month name, e.g. 'September'</td></tr>
     *         <tr><td>%m</td><td>Short month name, e.g. 'Sep'</td></tr>
     *         <tr><td>%N</td><td>Month, e.g. '9'</td></tr>
     *         <tr><td>%n</td><td>Month padded with zeroes, e.g. '09'</td></tr>
     *         <tr><td>%D</td><td>Day, e.g. '2'</td></tr>
     *         <tr><td>%d</td><td>Day padded with zeroes, e.g. '02'</td></tr>
     *         <tr><td>%Y</td><td>Year, e.g. '2017' if positive, '2017 BC' if negative; however, if the
     *             eras aren't empty and match each other, then this will behave the
     *             same as %y</td></tr>
     *         <tr><td>%y</td><td>Year with era, e.g. '2017 AD'; however, if the eras are empty or they
     *             match each other, then negative years will result in '-2017 AD'</td></tr>
     *         <tr><td>%z</td><td>Year without the era, e.g. '2017' or '-2017'</td></tr>
     *         <tr><td>%%</td><td>%</td></tr>
     *         </table>
     */
    @SuppressWarnings("AutoBoxing")
    public String format(String layout) {
        StringBuilder buffer = new StringBuilder();
        boolean       cmd    = false;
        int           count  = layout.length();
        for (int i = 0; i < count; i++) {
            char ch = layout.charAt(i);
            if (cmd) {
                cmd = false;
                switch (ch) {
                    case 'W' -> buffer.append(weekDayName());
                    case 'w' -> {
                        String weekDayName = weekDayName();
                        if (weekDayName.length() > 3) {
                            weekDayName = weekDayName.substring(0, 3);
                        }
                        buffer.append(weekDayName);
                    }
                    case 'M' -> buffer.append(monthName());
                    case 'm' -> {
                        String monthName = monthName();
                        if (monthName.length() > 3) {
                            monthName = monthName.substring(0, 3);
                        }
                        buffer.append(monthName);
                    }
                    case 'N' -> buffer.append(month());
                    case 'n' -> buffer.append(String.format(String.format("%%0%d", widthNeeded(mCalendar.mMonths.size())), month()));
                    case 'D' -> buffer.append(dayInMonth());
                    case 'd' -> buffer.append(String.format(String.format("%%0%d", widthNeeded(mCalendar.mMonths.get(month()).mDays)), dayInMonth()));
                    case 'Y' -> {
                        int year = year();
                        if (mCalendar.mPreviousEra.isBlank()) {
                            buffer.append(year);
                        } else if (mCalendar.mEra.equals(mCalendar.mPreviousEra)) {
                            buffer.append(String.format("%d %s", year, mCalendar.mPreviousEra));
                        } else if (year < 0) {
                            buffer.append(String.format("%d %s", -year, mCalendar.mPreviousEra));
                        } else {
                            buffer.append(year);
                        }
                    }
                    case 'y' -> {
                        String era  = era();
                        int    year = year();
                        if (year < 0 && !era.isBlank() && !mCalendar.mEra.equals(mCalendar.mPreviousEra)) {
                            year = -year;
                        }
                        if (era.isBlank()) {
                            buffer.append(year);
                        } else {
                            buffer.append(String.format("%d %s", year, era));
                        }
                    }
                    case 'z' -> buffer.append(year());
                    case '%' -> buffer.append('%');
                    default -> {
                    }
                }
            } else if (ch == '%') {
                cmd = true;
            } else {
                buffer.append(ch);
            }
        }
        return buffer.toString();
    }

    private static int widthNeeded(int count) {
        int needed = 1;
        while (count > 9) {
            count /= 10;
            needed++;
        }
        return needed;
    }

    /** @return a text representation of the month. */
    @SuppressWarnings("AutoBoxing")
    public String textCalendarMonth() throws CalendarException {
        StringBuilder buffer   = new StringBuilder();
        int           mostDays = 0;
        for (Month month : mCalendar.mMonths) {
            if (mostDays < month.mDays) {
                mostDays = month.mDays;
            }
        }
        int month = month();
        buffer.append(month);
        buffer.append(": ");
        buffer.append(monthName());
        int count         = mCalendar.mWeekDays.size();
        int lastDayOfWeek = count - 1;
        int width         = Integer.toString(mostDays).length();
        for (int i = 0; i < count; i++) {
            buffer.append(i == 0 ? '\n' : ' ');
            buffer.append(" ".repeat(Math.max(0, width - 1)));
            buffer.append(mCalendar.mWeekDays.get(i).charAt(0));
        }
        count = daysInMonth();
        int    year   = year();
        String numFmt = String.format("%%%dd", width);
        for (int i = 1; i <= count; i++) {
            Date d       = new Date(mCalendar, month, i, year);
            int  weekDay = d.weekDay();
            if (i == 1 || weekDay == 0) {
                buffer.append('\n');
            }
            if (i == 1 && weekDay != 0) {
                buffer.append(" ".repeat(Math.max(0, weekDay * (width + 1))));
            }
            buffer.append(String.format(numFmt, i));
            if (weekDay != lastDayOfWeek) {
                buffer.append(' ');
            }
        }
        buffer.append('\n');
        return buffer.toString();
    }
}
