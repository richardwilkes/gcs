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

package com.trollworks.gcs.settings;

import com.trollworks.gcs.calendar.Calendar;
import com.trollworks.gcs.utility.Dirs;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.NamedData;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.json.Json;
import com.trollworks.gcs.utility.text.NumericComparator;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class CalendarRef implements Comparable<CalendarRef> {
    private static Map<String, CalendarRef> REGISTERED_CALENDARS = new HashMap<>();
    static         CalendarRef              DEFAULT              = new CalendarRef();
    private        String                   mName;
    private        Calendar                 mCalendar;

    public static final CalendarRef current() {
        return get(Settings.getInstance().getGeneralSettings().calendarRef());
    }

    public static final Calendar currentCalendar() {
        return current().calendar();
    }

    public static final CalendarRef get(String name) {
        load();
        CalendarRef ref = REGISTERED_CALENDARS.get(name);
        return ref != null ? ref : DEFAULT;
    }

    public static final List<CalendarRef> choices() {
        load();
        List<CalendarRef> list = new ArrayList<>(REGISTERED_CALENDARS.values());
        Collections.sort(list);
        return list;
    }

    private static void load() {
        if (REGISTERED_CALENDARS.isEmpty()) {
            for (NamedData<List<NamedData<CalendarRef>>> list : NamedData.scanLibraries(FileType.CALENDAR_SETTINGS, Dirs.SETTINGS, CalendarRef::new)) {
                for (NamedData<CalendarRef> data : list.getData()) {
                    REGISTERED_CALENDARS.putIfAbsent(data.getName(), data.getData());
                }
            }
            REGISTERED_CALENDARS.putIfAbsent(DEFAULT.mName, DEFAULT);
        }
    }

    public CalendarRef() {
        mName = "Gregorian";
        mCalendar = new Calendar();
    }

    public CalendarRef(Path path) throws IOException {
        mName = PathUtils.getLeafName(path, false);
        try (BufferedReader r = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            mCalendar = new Calendar(Json.asMap(Json.parse(r)));
        }
    }

    public CalendarRef(String name) {
        mName = name;
    }

    public String name() {
        return mName;
    }

    public Calendar calendar() {
        return mCalendar;
    }

    @Override
    public int compareTo(CalendarRef other) {
        return NumericComparator.caselessCompareStrings(mName, other.mName);
    }

    @Override
    public String toString() {
        return mName;
    }
}
