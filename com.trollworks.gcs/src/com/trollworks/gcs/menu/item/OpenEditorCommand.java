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

import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.OutlineProxy;
import com.trollworks.gcs.utility.I18n;

import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Open Detail Editor" command. */
public class OpenEditorCommand extends Command {
    /** The action command this command will issue. */
    public static final String            CMD_OPEN_EDITOR = "OpenEditor";
    /** The singleton {@link OpenEditorCommand}. */
    public static final OpenEditorCommand INSTANCE        = new OpenEditorCommand();

    private OpenEditorCommand() {
        super(I18n.Text("Open Detail Editor"), CMD_OPEN_EDITOR, KeyEvent.VK_I);
    }

    /**
     * Creates a new {@link OpenEditorCommand}.
     *
     * @param outline The outline to work against.
     */
    public OpenEditorCommand(ListOutline outline) {
        super(I18n.Text("Open Detail Editor"), CMD_OPEN_EDITOR);
    }

    @Override
    public void adjust() {
        ListOutline outline = getOutline();
        if (outline != null) {
            setEnabled(outline.getModel().hasSelection());
        } else {
            setEnabled(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline outline = getOutline();
        if (outline != null) {
            ((ListOutline) outline.getRealOutline()).openDetailEditor(false);
        }
    }

    private ListOutline getOutline() {
        Component comp = getFocusOwner();
        if (comp instanceof OutlineProxy) {
            comp = ((OutlineProxy) comp).getRealOutline();
        }
        if (comp instanceof ListOutline) {
            return (ListOutline) comp;
        }
        return null;
    }
}
