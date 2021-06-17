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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.page.PageSettings;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.Modal;
import com.trollworks.gcs.utility.FileProxy;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;

import java.awt.EventQueue;
import java.awt.desktop.PrintFilesEvent;
import java.awt.desktop.PrintFilesHandler;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.awt.print.PrinterException;
import java.awt.print.PrinterJob;
import java.io.File;
import java.nio.file.Path;
import java.util.concurrent.TimeUnit;
import javax.print.attribute.HashPrintRequestAttributeSet;
import javax.print.attribute.standard.DialogTypeSelection;

/** Provides the "Print..." command. */
public final class PrintCommand extends Command implements PrintFilesHandler {
    /** The action command this command will issue. */
    public static final String CMD_PRINT = "Print";

    /** The singleton {@link PrintCommand}. */
    public static final PrintCommand INSTANCE = new PrintCommand();

    private PrintCommand() {
        super(I18n.text("Print…"), CMD_PRINT, KeyEvent.VK_P);
    }

    @Override
    public void adjust() {
        setEnabled(!UIUtilities.inModalState() && getTarget(PrintProxy.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        PrintProxy proxy = getTarget(PrintProxy.class);
        if (proxy != null) {
            print(proxy);
        }
    }

    /** @param proxy The {@link PrintProxy} to print. */
    public static void print(PrintProxy proxy) {
        if (proxy != null) {
            PrinterJob job = PrinterJob.getPrinterJob();
            job.setJobName(proxy.getPrintJobTitle());
            PageSettings                 pageSettings = proxy.getPageSettings();
            HashPrintRequestAttributeSet set          = new HashPrintRequestAttributeSet();
            set.add(pageSettings.getPaperSize().getMediaSizeName());
            set.add(DialogTypeSelection.NATIVE);
            if (job.printDialog()) {
                try {
                    proxy.setPrinting(true);
                    job.setPrintable(proxy, pageSettings.createPageFormat());
                    job.print();
                } catch (PrinterException exception) {
                    Modal.showError(UIUtilities.getComponentForDialog(proxy), I18n.text("Printing failed!"));
                } finally {
                    proxy.setPrinting(false);
                }
            }
        }
    }

    @Override
    public void printFiles(PrintFilesEvent event) {
        for (File file : event.getFiles()) {
            EventQueue.invokeLater(new DeferredPrint(file.toPath()));
        }
    }

    static class DeferredPrint implements Runnable {
        private long mStart;
        private Path mPath;

        DeferredPrint(Path path) {
            mPath = path;
            OpenDataFileCommand.open(path);
            mStart = System.currentTimeMillis();
        }

        @Override
        public void run() {
            LibraryExplorerDockable library = LibraryExplorerDockable.get();
            FileProxy               proxy   = library == null ? null : (FileProxy) library.getDockableFor(mPath);
            if (proxy != null) {
                print(proxy.getPrintProxy());
            } else if (System.currentTimeMillis() - mStart < TimeUnit.MILLISECONDS.convert(2, TimeUnit.MINUTES)) {
                EventQueue.invokeLater(this);
            }
        }
    }
}
