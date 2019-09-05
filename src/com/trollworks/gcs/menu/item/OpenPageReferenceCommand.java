/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.common.HasSourceReference;
import com.trollworks.gcs.library.LibraryExplorerDockable;
import com.trollworks.gcs.pdfview.PdfDockable;
import com.trollworks.gcs.pdfview.PdfRef;
import com.trollworks.toolkit.collections.ReverseListIterator;
import com.trollworks.toolkit.ui.Selection;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.ui.widget.StdFileDialog;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.ui.widget.outline.Row;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.io.File;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

import javax.swing.filechooser.FileNameExtensionFilter;

/** Provides the "Open Page Reference" command. */
public class OpenPageReferenceCommand extends Command {
    /** The singleton {@link OpenPageReferenceCommand} for opening a single page reference. */
    public static final OpenPageReferenceCommand OPEN_ONE_INSTANCE  = new OpenPageReferenceCommand(I18n.Text("Open Page Reference"), "OpenPageReference", KeyEvent.VK_G, COMMAND_MODIFIER);
    /** The singleton {@link OpenPageReferenceCommand} for opening all page references. */
    public static final OpenPageReferenceCommand OPEN_EACH_INSTANCE = new OpenPageReferenceCommand(I18n.Text("Open Each Page Reference"), "OpenEachPageReferences", KeyEvent.VK_G, SHIFTED_COMMAND_MODIFIER);

    private OpenPageReferenceCommand(String title, String cmd, int key, int modifiers) {
        super(title, cmd, key, modifiers);
    }

    @Override
    public void adjust() {
        setEnabled(!getReferences(getTarget()).isEmpty());
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        HasSourceReference target = getTarget();
        if (target != null) {
            List<String> references = getReferences(target);
            if (!references.isEmpty()) {
                String highlight = target.getReferenceHighlight();
                if (this == OPEN_ONE_INSTANCE) {
                    openReference(references.get(0), highlight);
                } else {
                    for (String one : new ReverseListIterator<>(references)) {
                        openReference(one, highlight);
                    }
                }
            }
        }
    }

    public static void openReference(String reference, String highlight) {
        int i = reference.length() - 1;
        while (i >= 0) {
            char ch = reference.charAt(i);
            if (ch >= '0' && ch <= '9') {
                i--;
            } else {
                i++;
                break;
            }
        }
        if (i > 0) {
            String id = reference.substring(0, i);
            try {
                int    page = Integer.parseInt(reference.substring(i));
                PdfRef ref  = PdfRef.lookup(id, true);
                if (ref == null) {
                    File file = StdFileDialog.showOpenDialog(getFocusOwner(), String.format(I18n.Text("Locate the PDF file for the prefix \"%s\""), id), new FileNameExtensionFilter(I18n.Text("PDF File"), FileType.PDF_EXTENSION));
                    if (file != null) {
                        ref = new PdfRef(id, file, 0);
                        ref.save();
                    }
                }
                if (ref != null) {
                    Path                    path     = ref.getFile().toPath();
                    LibraryExplorerDockable library  = LibraryExplorerDockable.get();
                    PdfDockable             dockable = (PdfDockable) library.getDockableFor(path);
                    if (dockable != null) {
                        dockable.goToPage(ref, page, highlight);
                        dockable.getDockContainer().setCurrentDockable(dockable);
                    } else {
                        dockable = new PdfDockable(ref, page, highlight);
                        library.dockPdf(dockable);
                        library.open(path);
                    }
                }
            } catch (NumberFormatException nfex) {
                // Ignore
            }
        }
    }

    private static HasSourceReference getTarget() {
        HasSourceReference ref  = null;
        Component          comp = getFocusOwner();
        if (comp instanceof Outline) {
            OutlineModel model = ((Outline) comp).getModel();
            if (model.hasSelection()) {
                Selection selection = model.getSelection();
                if (selection.getCount() == 1) {
                    Row row = model.getFirstSelectedRow();
                    if (row instanceof HasSourceReference) {
                        ref = (HasSourceReference) row;
                    }
                }
            }
        }
        return ref;
    }

    private static List<String> getReferences(HasSourceReference ref) {
        List<String> list = new ArrayList<>();
        if (ref != null) {
            String[] refs = ref.getReference().split("[,;]");
            if (refs.length > 0) {
                for (String one : refs) {
                    String trimmed = one.trim();
                    if (!trimmed.isEmpty()) {
                        list.add(trimmed);
                    }
                }
            }
        }
        return list;
    }
}
