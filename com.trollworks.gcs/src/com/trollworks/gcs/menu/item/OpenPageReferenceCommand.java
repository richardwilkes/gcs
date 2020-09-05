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

package com.trollworks.gcs.menu.item;

import com.trollworks.gcs.datafile.PageRefCell;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.pdfview.PDFRef;
import com.trollworks.gcs.pdfview.PDFServer;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.Selection;
import com.trollworks.gcs.ui.widget.StdFileDialog;
import com.trollworks.gcs.ui.widget.WindowUtils;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.ReverseListIterator;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;

/** Provides the "Open Page Reference" command. */
public class OpenPageReferenceCommand extends Command {
    /** The singleton {@link OpenPageReferenceCommand} for opening a single page reference. */
    public static final OpenPageReferenceCommand OPEN_ONE_INSTANCE  = new OpenPageReferenceCommand(true, COMMAND_MODIFIER);
    /** The singleton {@link OpenPageReferenceCommand} for opening all page references. */
    public static final OpenPageReferenceCommand OPEN_EACH_INSTANCE = new OpenPageReferenceCommand(false, SHIFTED_COMMAND_MODIFIER);
    private             ListOutline              mOutline;

    private OpenPageReferenceCommand(boolean one, int modifiers) {
        super(getTitle(one), getCmd(one), KeyEvent.VK_G, modifiers);
    }

    /**
     * Creates a new {@link OpenPageReferenceCommand}.
     *
     * @param outline The outline to work against.
     * @param one     Whether to open just the first page reference, or all of them.
     */
    public OpenPageReferenceCommand(ListOutline outline, boolean one) {
        super(getTitle(one), getCmd(one));
        mOutline = outline;
    }

    private static String getTitle(boolean one) {
        return one ? I18n.Text("Open Page Reference") : I18n.Text("Open Each Page Reference");
    }

    private static String getCmd(boolean one) {
        return one ? "OpenPageReference" : "OpenEachPageReferences";
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
            int    page;
            String id = reference.substring(0, i);
            try {
                page = Integer.parseInt(reference.substring(i));
            } catch (NumberFormatException nfex) {
                return; // Has no page number, so bail
            }
            Preferences prefs = Preferences.getInstance();
            PDFRef      ref   = prefs.lookupPdfRef(id, true);
            if (ref == null) {
                Path path = StdFileDialog.showOpenDialog(getFocusOwner(), String.format(I18n.Text("Locate the PDF file for the prefix \"%s\""), id), FileType.PDF.getFilter());
                if (path != null) {
                    ref = new PDFRef(id, path, 0);
                    prefs.putPdfRef(ref);
                }
            }
            if (ref != null) {
                try {
                    PDFServer.showPDF(ref.getPath(), page + ref.getPageToIndexOffset());
                } catch (Exception exception) {
                    WindowUtils.showError(null, exception.getMessage());
                }
            }
        }
    }

    private HasSourceReference getTarget() {
        HasSourceReference ref     = null;
        ListOutline        outline = mOutline;
        if (outline == null) {
            Component comp = getFocusOwner();
            if (comp instanceof OutlineProxy) {
                comp = ((OutlineProxy) comp).getRealOutline();
            }
            if (comp instanceof ListOutline) {
                outline = (ListOutline) comp;
            }
        }
        if (outline != null) {
            OutlineModel model = outline.getModel();
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
            String[] refs = PageRefCell.SEPARATORS_PATTERN.split(ref.getReference());
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
