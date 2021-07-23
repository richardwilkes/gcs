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

package com.trollworks.gcs.pageref;

import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.Platform;

import java.io.IOException;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

public enum PDFViewer {
    EVINCE {
        @Override
        public String toString() {
            return I18n.text("Evince");
        }

        @Override
        public boolean available() {
            return Platform.isLinux();
        }

        @Override
        public String installFrom() {
            return "https://wiki.gnome.org/Apps/Evince";
        }

        @Override
        public void open(Path path, int page) {
            ProcessBuilder pb = new ProcessBuilder("evince", "-i", Integer.toString(page),
                    path.normalize().toAbsolutePath().toString());
            try {
                pb.start();
            } catch (IOException ioe) {
                Log.error(ioe);
                Modal.showError(null, ioe.getMessage());
            }
        }
    },
    SKIM {
        @Override
        public String toString() {
            return I18n.text("Skim");
        }

        @Override
        public boolean available() {
            return Platform.isMacintosh();
        }

        @Override
        public String installFrom() {
            return "https://skim-app.sourceforge.io";
        }

        @Override
        public void open(Path path, int page) {
            ProcessBuilder pb = new ProcessBuilder("open", "skim://" +
                    PDFServer.encodeQueryParam(path.normalize().toAbsolutePath().toString()) +
                    "#page=" + page);
            try {
                pb.start();
            } catch (IOException ioe) {
                Log.error(ioe);
                Modal.showError(null, ioe.getMessage());
            }
        }
    },
    SUMATRA {
        @Override
        public String toString() {
            return I18n.text("Sumatra");
        }

        @Override
        public boolean available() {
            return Platform.isWindows();
        }

        @Override
        public String installFrom() {
            return "https://www.sumatrapdfreader.org";
        }

        @Override
        public void open(Path path, int page) {
            ProcessBuilder pb = new ProcessBuilder("sumatra", "-page", Integer.toString(page),
                    path.normalize().toAbsolutePath().toString());
            try {
                pb.start();
            } catch (IOException ioe) {
                Log.error(ioe);
                Modal.showError(null, ioe.getMessage());
            }
        }
    },
    WEB_BROWSER {
        @Override
        public String toString() {
            return I18n.text("Web Browser");
        }

        @Override
        public boolean available() {
            return true;
        }

        @Override
        public String installFrom() {
            return "";
        }

        @Override
        public void open(Path path, int page) {
            try {
                PDFServer.showPDF(path, page);
            } catch (Exception ex) {
                Log.error(ex);
                Modal.showError(null, ex.getMessage());
            }
        }
    };

    public abstract boolean available();

    public abstract String installFrom();

    public abstract void open(Path path, int page);

    public static List<PDFViewer> valuesForPlatform() {
        List<PDFViewer> list = new ArrayList<>();
        for (PDFViewer one : values()) {
            if (one.available()) {
                list.add(one);
            }
        }
        return list;
    }
}
