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

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.CollectedOutlinesDockable;
import com.trollworks.gcs.character.CollectedOutlines;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.dock.Dock;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.notification.Notifier;
import com.trollworks.gcs.utility.undo.StdUndoManager;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Component;
import javax.swing.JScrollPane;

/** A list of advantages and disadvantages from a library. */
public class TemplateDockable extends CollectedOutlinesDockable {
    private static TemplateDockable LAST_ACTIVATED;
    private        TemplateSheet    mTemplate;

    /** Creates a new {@link TemplateDockable}. */
    public TemplateDockable(Template template) {
        super(template);
        Template dataFile = getDataFile();
        mTemplate = new TemplateSheet(dataFile);
        createToolbar();
        JScrollPane scroller = new JScrollPane(mTemplate);
        scroller.setBorder(null);
        scroller.getViewport().setBackground(Color.LIGHT_GRAY);
        add(scroller, BorderLayout.CENTER);
        dataFile.setModified(false);
        StdUndoManager undoManager = getUndoManager();
        undoManager.discardAllEdits();
        dataFile.setUndoManager(undoManager);
        Preferences.getInstance().getNotifier().add(this, Fonts.FONT_NOTIFICATION_KEY, Preferences.KEY_USE_MULTIPLICATIVE_MODIFIERS);
    }

    @Override
    public CollectedOutlines getCollectedOutlines() {
        return mTemplate;
    }

    @Override
    public boolean attemptClose() {
        boolean closed = super.attemptClose();
        if (closed) {
            Notifier notifier = Preferences.getInstance().getNotifier();
            notifier.remove(mTemplate);
            notifier.remove(this);
        }
        return closed;
    }

    @Override
    public Component getRetargetedFocus() {
        return mTemplate;
    }

    /** @return The last activated {@link TemplateDockable}. */
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
        return I18n.Text("Untitled Template");
    }

    @Override
    public PrintProxy getPrintProxy() {
        return null;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        if (Fonts.FONT_NOTIFICATION_KEY.equals(name)) {
            mTemplate.updateRowHeights();
            mTemplate.revalidate();
        } else {
            getDataFile().notifySingle(Advantage.ID_LIST_CHANGED, null);
        }
    }
}
