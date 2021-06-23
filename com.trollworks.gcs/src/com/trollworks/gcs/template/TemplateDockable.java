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

package com.trollworks.gcs.template;

import com.trollworks.gcs.character.CollectedOutlines;
import com.trollworks.gcs.character.CollectedOutlinesDockable;
import com.trollworks.gcs.ui.Colors;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.dock.Dock;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Component;

/** A list of advantages and disadvantages from a library. */
public class TemplateDockable extends CollectedOutlinesDockable {
    private static TemplateDockable LAST_ACTIVATED;
    private        TemplateSheet    mTemplate;

    /** Creates a new TemplateDockable. */
    public TemplateDockable(Template template) {
        super(template);
        Template dataFile = getDataFile();
        mTemplate = new TemplateSheet(dataFile);
        createToolbar();
        ScrollPanel scroller = new ScrollPanel(mTemplate);
        scroller.getViewport().setBackground(Colors.PAGE_VOID);
        add(scroller, BorderLayout.CENTER);
        dataFile.setModified(false);
        StdUndoManager undoManager = getUndoManager();
        undoManager.discardAllEdits();
        dataFile.setUndoManager(undoManager);
    }

    @Override
    public CollectedOutlines getCollectedOutlines() {
        return mTemplate;
    }

    @Override
    public boolean attemptClose() {
        boolean closed = super.attemptClose();
        if (closed) {
            mTemplate.dispose();
        }
        return closed;
    }

    @Override
    public Component getRetargetedFocus() {
        return mTemplate;
    }

    /** @return The last activated TemplateDockable. */
    public static TemplateDockable getLastActivated() {
        if (LAST_ACTIVATED != null) {
            Dock dock = UIUtilities.getAncestorOfType(LAST_ACTIVATED, Dock.class);
            if (dock == null) {
                LAST_ACTIVATED = null;
            }
        }
        return LAST_ACTIVATED;
    }

    @Override
    public void activated() {
        super.activated();
        LAST_ACTIVATED = this;
    }

    @Override
    public Template getDataFile() {
        return (Template) super.getDataFile();
    }

    /** @return The {@link TemplateSheet}. */
    public TemplateSheet getTemplate() {
        return mTemplate;
    }

    @Override
    protected String getUntitledBaseName() {
        return I18n.text("Untitled Template");
    }

    @Override
    public PrintProxy getPrintProxy() {
        return null;
    }
}
