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

package com.trollworks.gcs.utility;

import java.util.Calendar;

/** Provides support for various forms of version numbers. */
public class Version {
    public static final int  MAJOR_BITS             = 8;
    public static final int  MINOR_BITS             = 8;
    public static final int  BUGFIX_BITS            = 9;
    public static final int  QUALIFIER_YEAR_BITS    = 12;
    public static final int  QUALIFIER_MONTH_BITS   = 4;
    public static final int  QUALIFIER_DAY_BITS     = 5;
    public static final int  QUALIFIER_HOUR_BITS    = 5;
    public static final int  QUALIFIER_MINUTE_BITS  = 6;
    public static final int  QUALIFIER_SECOND_BITS  = 6;
    public static final int  QUALIFIER_BITS         = QUALIFIER_YEAR_BITS + QUALIFIER_MONTH_BITS + QUALIFIER_DAY_BITS + QUALIFIER_HOUR_BITS + QUALIFIER_MINUTE_BITS + QUALIFIER_SECOND_BITS;
    public static final int  QUALIFIER_SECOND_SHIFT = 0;
    public static final long QUALIFIER_SECOND_MASK  = (1L << QUALIFIER_SECOND_BITS) - 1 << QUALIFIER_SECOND_SHIFT;
    public static final int  QUALIFIER_MINUTE_SHIFT = QUALIFIER_SECOND_BITS + QUALIFIER_SECOND_SHIFT;
    public static final long QUALIFIER_MINUTE_MASK  = (1L << QUALIFIER_MINUTE_BITS) - 1 << QUALIFIER_MINUTE_SHIFT;
    public static final int  QUALIFIER_HOUR_SHIFT   = QUALIFIER_MINUTE_BITS + QUALIFIER_MINUTE_SHIFT;
    public static final long QUALIFIER_HOUR_MASK    = (1L << QUALIFIER_HOUR_BITS) - 1 << QUALIFIER_HOUR_SHIFT;
    public static final int  QUALIFIER_DAY_SHIFT    = QUALIFIER_HOUR_BITS + QUALIFIER_HOUR_SHIFT;
    public static final long QUALIFIER_DAY_MASK     = (1L << QUALIFIER_DAY_BITS) - 1 << QUALIFIER_DAY_SHIFT;
    public static final int  QUALIFIER_MONTH_SHIFT  = QUALIFIER_DAY_BITS + QUALIFIER_DAY_SHIFT;
    public static final long QUALIFIER_MONTH_MASK   = (1L << QUALIFIER_MONTH_BITS) - 1 << QUALIFIER_MONTH_SHIFT;
    public static final int  QUALIFIER_YEAR_SHIFT   = QUALIFIER_MONTH_BITS + QUALIFIER_MONTH_SHIFT;
    public static final long QUALIFIER_YEAR_MASK    = (1L << QUALIFIER_YEAR_BITS) - 1 << QUALIFIER_YEAR_SHIFT;
    public static final int  QUALIFIER_SHIFT        = 0;
    public static final long QUALIFIER_MASK         = (1L << QUALIFIER_BITS) - 1 << QUALIFIER_SHIFT;
    public static final int  BUGFIX_SHIFT           = QUALIFIER_BITS + QUALIFIER_SHIFT;
    public static final long BUGFIX_MASK            = (1L << BUGFIX_BITS) - 1 << BUGFIX_SHIFT;
    public static final int  MINOR_SHIFT            = BUGFIX_BITS + BUGFIX_SHIFT;
    public static final long MINOR_MASK             = (1L << MINOR_BITS) - 1 << MINOR_SHIFT;
    public static final int  MAJOR_SHIFT            = MINOR_BITS + MINOR_SHIFT;
    public static final long MAJOR_MASK             = (1L << MAJOR_BITS) - 1 << MAJOR_SHIFT;

    public static int getMajor(long version) {
        return (int) ((version & MAJOR_MASK) >>> MAJOR_SHIFT);
    }

    public static int getMinor(long version) {
        return (int) ((version & MINOR_MASK) >>> MINOR_SHIFT);
    }

    public static int getBugFix(long version) {
        return (int) ((version & BUGFIX_MASK) >>> BUGFIX_SHIFT);
    }

    public static long getQualifier(long version) {
        return (version & QUALIFIER_MASK) >>> QUALIFIER_SHIFT;
    }

    public static int getQualifierYear(long version) {
        return (int) ((version & QUALIFIER_YEAR_MASK) >>> QUALIFIER_YEAR_SHIFT);
    }

    public static int getQualifierMonth(long version) {
        return (int) ((version & QUALIFIER_MONTH_MASK) >>> QUALIFIER_MONTH_SHIFT);
    }

    public static int getQualifierDay(long version) {
        return (int) ((version & QUALIFIER_DAY_MASK) >>> QUALIFIER_DAY_SHIFT);
    }

    public static int getQualifierHour(long version) {
        return (int) ((version & QUALIFIER_HOUR_MASK) >>> QUALIFIER_HOUR_SHIFT);
    }

    public static int getQualifierMinute(long version) {
        return (int) ((version & QUALIFIER_MINUTE_MASK) >>> QUALIFIER_MINUTE_SHIFT);
    }

    public static int getQualifierSecond(long version) {
        return (int) ((version & QUALIFIER_SECOND_MASK) >>> QUALIFIER_SECOND_SHIFT);
    }

    public static Calendar getQualifierAsCalendar(long version) {
        Calendar cal = Calendar.getInstance();
        cal.clear();
        cal.set(getQualifierYear(version), getQualifierMonth(version) - 1, getQualifierDay(version), getQualifierHour(version), getQualifierMinute(version), getQualifierSecond(version));
        return cal;
    }

    public static String toBuildTimestamp(long version) {
        return String.format(I18n.Text("Built on %1$tB %1$te, %1$tY at %1$tr"), getQualifierAsCalendar(version));
    }

    public static String toString(long version, boolean includeQualifier) {
        StringBuilder buffer = new StringBuilder();
        buffer.append(getMajor(version));
        buffer.append('.');
        buffer.append(getMinor(version));
        int  bugFix    = getBugFix(version);
        long qualifier = getQualifier(version);
        if (bugFix != 0 || includeQualifier && qualifier != 0) {
            buffer.append('.');
            buffer.append(bugFix);
            if (includeQualifier && qualifier != 0) {
                buffer.append('.');
                buffer.append(getQualifierYear(version));
                pad2(getQualifierMonth(version), buffer);
                pad2(getQualifierDay(version), buffer);
                pad2(getQualifierHour(version), buffer);
                pad2(getQualifierMinute(version), buffer);
                pad2(getQualifierSecond(version), buffer);
            }
        }
        return buffer.toString();
    }

    private static void pad2(int value, StringBuilder buffer) {
        if (value < 10) {
            buffer.append('0');
        }
        buffer.append(value);
    }

    public static long toVersion(int major, int minor, int bugfix, int year, int month, int day, int hour, int minute, int second) {
        return (long) major << MAJOR_SHIFT | (long) minor << MINOR_SHIFT | (long) bugfix << BUGFIX_SHIFT | (long) year << QUALIFIER_YEAR_SHIFT | (long) month << QUALIFIER_MONTH_SHIFT | (long) day << QUALIFIER_DAY_SHIFT | (long) hour << QUALIFIER_HOUR_SHIFT | (long) minute << QUALIFIER_MINUTE_SHIFT | (long) second << QUALIFIER_SECOND_SHIFT;
    }

    public static long toVersion(int major, int minor, int bugfix, long dateTimeMillis) {
        return (long) major << MAJOR_SHIFT | (long) minor << MINOR_SHIFT | (long) bugfix << BUGFIX_SHIFT | dateTimeMillis / 1000 << QUALIFIER_SHIFT;
    }

    public static long toVersion(int major, int minor, int bugfix) {
        return toVersion(major, minor, bugfix, 0);
    }

    public static long extract(String version, long def) {
        try {
            return extract(version);
        } catch (NumberFormatException nfe) {
            return def;
        }
    }

    public static long extract(String version) throws NumberFormatException {
        String[] parts = version.split("[\\._-]", 9);
        switch (parts.length) {
        case 2:
            long c2p1 = Long.parseLong(parts[0]);
            long c2p2 = Long.parseLong(parts[1]);
            if (c2p1 > (1L << MAJOR_BITS) - 1) {
                // Assume it is a qualifier in the form of YYYYMMDD-HHMMSS or YYYYMMDD-HHMM
                // since the first value is too large to fit into the major portion
                int year  = (int) (c2p1 / 10000);
                int month = (int) ((c2p1 - year * 10000) / 100);
                int day   = (int) (c2p1 - (year * 10000 + month * 100));
                int hour;
                int minute;
                int second;
                if (parts[1].length() > 4) {
                    hour = (int) (c2p2 / 10000);
                    minute = (int) ((c2p2 - hour * 10000) / 100);
                    second = (int) (c2p2 - (hour * 10000 + minute * 100));
                } else {
                    hour = (int) (c2p2 / 100);
                    minute = (int) (c2p2 - hour * 100);
                    second = 0;
                }
                return toVersion(1, 0, 0, check(year, QUALIFIER_YEAR_BITS), check(month, QUALIFIER_MONTH_BITS), check(day, QUALIFIER_DAY_BITS), check(hour, QUALIFIER_HOUR_BITS), check(minute, QUALIFIER_MINUTE_BITS), check(second, QUALIFIER_SECOND_BITS));
            }
            // Otherwise we assume it is MAJOR.MINOR
            return toVersion(check((int) c2p1, MAJOR_BITS), check((int) c2p2, MINOR_BITS), 0);
        case 3:
            return toVersion(check(Integer.parseInt(parts[0]), MAJOR_BITS), check(Integer.parseInt(parts[1]), MINOR_BITS), check(Integer.parseInt(parts[2]), BUGFIX_BITS));
        case 4:
            long c4p1 = Long.parseLong(parts[0]);
            long c4p2 = Long.parseLong(parts[1]);
            long c4p3 = Long.parseLong(parts[2]);
            long c4p4 = Long.parseLong(parts[3]);
            int hour;
            int minute;
            int second;
            if (c4p1 > (1L << MAJOR_BITS) - 1) {
                // Assume it is a qualifier in the form of YYYY-MM-DD-HHMMSS, or YYYY-MM-DD-HHMM
                // since the first value is too large to fit into the major portion
                if (parts[3].length() > 4) {
                    hour = (int) (c4p4 / 10000);
                    minute = (int) ((c4p4 - hour * 10000) / 100);
                    second = (int) (c4p4 - (hour * 10000 + minute * 100));
                } else {
                    hour = (int) (c4p4 / 100);
                    minute = (int) (c4p4 - hour * 100);
                    second = 0;
                }
                return toVersion(1, 0, 0, check((int) c4p1, QUALIFIER_YEAR_BITS), check((int) c4p2, QUALIFIER_MONTH_BITS), check((int) c4p3, QUALIFIER_DAY_BITS), check(hour, QUALIFIER_HOUR_BITS), check(minute, QUALIFIER_MINUTE_BITS), check(second, QUALIFIER_SECOND_BITS));
            }
            // Otherwise we assume it is MAJOR.MINOR.BUGFIX.QUALIFIER
            int year;
            int month;
            int day;
            if (parts[3].length() > 12) {
                year = (int) (c4p4 / 10000000000L);
                c4p4 -= year * 10000000000L;
                month = (int) (c4p4 / 100000000L);
                second = (int) (c4p4 - month * 100000000L);
                day = second / 1000000;
                second -= day * 1000000;
                hour = second / 10000;
                second -= hour * 10000;
                minute = second / 100;
                second -= minute * 100;
            } else {
                year = (int) (c4p4 / 100000000L);
                c4p4 -= year * 100000000L;
                month = (int) (c4p4 / 1000000L);
                minute = (int) (c4p4 - month * 1000000L);
                day = minute / 10000;
                minute -= day * 10000;
                hour = minute / 100;
                minute -= hour * 100;
                second = 0;
            }
            return toVersion(check((int) c4p1, MAJOR_BITS), check((int) c4p2, MINOR_BITS), check((int) c4p3, BUGFIX_BITS), check(year, QUALIFIER_YEAR_BITS), check(month, QUALIFIER_MONTH_BITS), check(day, QUALIFIER_DAY_BITS), check(hour, QUALIFIER_HOUR_BITS), check(minute, QUALIFIER_MINUTE_BITS), check(second, QUALIFIER_SECOND_BITS));
        default:
            throw new NumberFormatException(I18n.Text("Invalid version format"));
        }
    }

    private static int check(int value, int maxBits) throws NumberFormatException {
        if (value > (1 << maxBits) - 1) {
            throw new NumberFormatException(I18n.Text("Invalid version format"));
        }
        return value;
    }
}
