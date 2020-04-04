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
    requires org.apache.fontbox;

    requires transitive com.trollworks.toolkit;
    requires transitive java.desktop;
    requires transitive org.apache.pdfbox;

    exports com.trollworks.gcs.advantage;
    exports com.trollworks.gcs.app;
    exports com.trollworks.gcs.character;
    exports com.trollworks.gcs.character.names;
    exports com.trollworks.gcs.common;
    exports com.trollworks.gcs.criteria;
    exports com.trollworks.gcs.equipment;
    exports com.trollworks.gcs.feature;
    exports com.trollworks.gcs.library;
    exports com.trollworks.gcs.menu;
    exports com.trollworks.gcs.menu.edit;
    exports com.trollworks.gcs.menu.file;
    exports com.trollworks.gcs.menu.item;
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
    exports com.trollworks.gcs.weapon;
    exports com.trollworks.gcs.widgets.outline;
}
