/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

open module com.trollworks.gcs {
    requires com.lowagie.text;
    requires java.datatransfer;
    requires java.management;
    requires java.prefs;
    requires org.apache.fontbox;

    requires transitive java.desktop;
    requires transitive java.xml;
    requires transitive org.apache.pdfbox;

    exports com.trollworks.gcs.advantage;
    exports com.trollworks.gcs.app;
    exports com.trollworks.gcs.character.names;
    exports com.trollworks.gcs.character;
    exports com.trollworks.gcs.collections;
    exports com.trollworks.gcs.datafile;
    exports com.trollworks.gcs.criteria;
    exports com.trollworks.gcs.equipment;
    exports com.trollworks.gcs.feature;
    exports com.trollworks.gcs.io.conduit;
    exports com.trollworks.gcs.io.json;
    exports com.trollworks.gcs.io.xml;
    exports com.trollworks.gcs.io;
    exports com.trollworks.gcs.library;
    exports com.trollworks.gcs.menu.edit;
    exports com.trollworks.gcs.menu.file;
    exports com.trollworks.gcs.menu.help;
    exports com.trollworks.gcs.menu.item;
    exports com.trollworks.gcs.menu;
    exports com.trollworks.gcs.modifier;
    exports com.trollworks.gcs.notes;
    exports com.trollworks.gcs.page;
    exports com.trollworks.gcs.pdfview;
    exports com.trollworks.gcs.preferences;
    exports com.trollworks.gcs.prereq;
    exports com.trollworks.gcs.services;
    exports com.trollworks.gcs.skill;
    exports com.trollworks.gcs.spell;
    exports com.trollworks.gcs.template;
    exports com.trollworks.gcs.ui.border;
    exports com.trollworks.gcs.ui.image;
    exports com.trollworks.gcs.ui.layout;
    exports com.trollworks.gcs.ui.print;
    exports com.trollworks.gcs.ui.scale;
    exports com.trollworks.gcs.ui.widget.dock;
    exports com.trollworks.gcs.ui.widget.outline;
    exports com.trollworks.gcs.ui.widget.search;
    exports com.trollworks.gcs.ui.widget.tree;
    exports com.trollworks.gcs.ui.widget;
    exports com.trollworks.gcs.ui;
    exports com.trollworks.gcs.utility.notification;
    exports com.trollworks.gcs.utility.task;
    exports com.trollworks.gcs.utility.text;
    exports com.trollworks.gcs.utility.undo;
    exports com.trollworks.gcs.utility.units;
    exports com.trollworks.gcs.utility;
    exports com.trollworks.gcs.weapon;
}
