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
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Platform;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;

public enum PDFViewer {
    ACROBAT {
        @Override
        public String toString() {
            return I18n.text("Acrobat");
        }

        @Override
        public boolean available() {
            return Platform.isWindows();
        }

        @Override
        public String installFrom() {
            return "https://acrobat.adobe.com";
        }

        @Override
        public void open(Path path, int page) {
            slashAPageOpenForWindows("Adobe\\Acrobat DC\\Acrobat", "Acrobat", path, page, true);
        }
    },
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
            String exe = findExecutable("evince");
            if (exe != null) {
                ProcessBuilder pb = new ProcessBuilder("evince", "-i", Integer.toString(page),
                        path.normalize().toAbsolutePath().toString());
                try {
                    pb.start();
                } catch (IOException ioe) {
                    Log.error(ioe);
                    Modal.showError(null, ioe.getMessage());
                }
            }
        }
    },
    FOXIT {
        @Override
        public String toString() {
            return I18n.text("Foxit PDF");
        }

        @Override
        public boolean available() {
            return Platform.isWindows();
        }

        @Override
        public String installFrom() {
            return "https://www.foxit.com/pdf-reader";
        }

        @Override
        public void open(Path path, int page) {
            slashAPageOpenForWindows("Foxit Software\\Foxit PDF Reader", "FoxitPDFReader", path, page, false);
        }
    },
    PDF_XCHANGE {
        @Override
        public String toString() {
            return I18n.text("PDF-XChange");
        }

        @Override
        public boolean available() {
            return Platform.isWindows();
        }

        @Override
        public String installFrom() {
            return "https://www.tracker-software.com";
        }

        @Override
        public void open(Path path, int page) {
            slashAPageOpenForWindows("Tracker Software\\PDF Editor", "PDFXEdit", path, page, true);
        }
    },
    PDF_XCVIEW {
        @Override
        public String toString() {
            return I18n.text("PDF-XChange Viewer");
        }

        @Override
        public boolean available() {
            return Platform.isWindows();
        }

        @Override
        public String installFrom() {
            return "https://www.tracker-software.com";
        }

        @Override
        public void open(Path path, int page) {
            slashAPageOpenForWindows("Tracker Software\\PDF-Viewer", "PDFXCview", path, page, true);
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
            if (!Files.exists(Paths.get("/Applications/Skim.app"))) {
                showUnableToLocateMsg();
                return;
            }
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
            return I18n.text("SumatraPDF");
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
            String exeName = "SumatraPDF";
            String exe     = findWindowsExecutable(exeName, exeName);
            if (exe != null) {
                ProcessBuilder pb = new ProcessBuilder(exe, "-page", Integer.toString(page),
                        path.normalize().toAbsolutePath().toString());
                try {
                    pb.start();
                } catch (IOException ioe) {
                    Log.error(ioe);
                    Modal.showError(null, ioe.getMessage());
                }
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

    void showUnableToLocateMsg() {
        Modal.showError(null, String.format(I18n.text("""
                Unable to locate %s.
                                        
                You may need to install it from %s
                and/or add it to your system PATH variable."""), this, installFrom()));
    }

    String findExecutable(String exeName) {
        String exe = PathUtils.searchPathForExecutable(exeName);
        if (exe == null) {
            showUnableToLocateMsg();
        }
        return exe;
    }

    String findExecutable(String paths, String exeName) {
        String exe = PathUtils.searchPathsForExecutable(paths, exeName);
        if (exe == null) {
            showUnableToLocateMsg();
        }
        return exe;
    }

    String findWindowsExecutable(String partialPath, String exeName) {
        String paths = appendToPaths(System.getenv("PATH"), System.getenv("PROGRAMFILES"), partialPath);
        paths = appendToPaths(paths, System.getenv("ProgramFiles(x86)"), partialPath);
        paths = appendToPaths(paths, System.getenv("LOCALAPPDATA"), partialPath);
        paths = appendToPaths(paths, System.getenv("APPDATA"), partialPath);
        return findExecutable(paths, exeName);
    }

    String appendToPaths(String paths, String path, String additional) {
        if (path == null) {
            return paths;
        }
        if (additional != null) {
            path += File.separator + additional;
        }
        if (paths == null) {
            return path;
        }
        return paths + File.pathSeparator + path;
    }

    void slashAPageOpenForWindows(String partialPath, String exeName, Path pdfPath, int page, boolean before) {
        String exe = findWindowsExecutable(partialPath, exeName);
        if (exe != null) {
            List<String> args = new ArrayList<>();
            args.add(exe);
            if (before) {
                addSlashAPageArg(args, page);
            }
            args.add(pdfPath.normalize().toAbsolutePath().toString());
            if (!before) {
                addSlashAPageArg(args, page);
            }
            ProcessBuilder pb = new ProcessBuilder(args);
            try {
                pb.start();
            } catch (IOException ioe) {
                Log.error(ioe);
                Modal.showError(null, ioe.getMessage());
            }
        }
    }

    void addSlashAPageArg(List<String> args, int page) {
        args.add("/A");
        args.add("page=" + page);
    }
}
