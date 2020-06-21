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

import com.trollworks.gcs.library.LibraryDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "Copy to Template" command. */
public class CopyToTemplateCommand extends Command {
    /** The action command this command will issue. */
    public static final String                CMD_COPY_TO_TEMPLATE = "CopyToTemplate";
    /** The singleton {@link CopyToTemplateCommand} for the menu bar. */
    public static final CopyToTemplateCommand INSTANCE             = new CopyToTemplateCommand();
    private             ListOutline           mOutline;
    private             TemplateDockable      mTemplate;

    private CopyToTemplateCommand() {
        super(I18n.Text("Copy to Template"), CMD_COPY_TO_TEMPLATE, KeyEvent.VK_T, SHIFTED_COMMAND_MODIFIER);
    }

    /**
     * Creates a new {@link CopyToTemplateCommand}.
     *
     * @param outline  The outline to work against.
     * @param template The template to target.
     */
    public CopyToTemplateCommand(ListOutline outline, TemplateDockable template) {
        super(template.getTitle(), CMD_COPY_TO_TEMPLATE);
        mOutline = outline;
        mTemplate = template;
    }

    @Override
    public void adjust() {
        boolean     shouldEnable = false;
        ListOutline outline      = getOutline();
        if (outline != null && outline.getModel().hasSelection()) {
            TemplateDockable lastTemplate = getTemplate();
            shouldEnable = lastTemplate != null && lastTemplate != UIUtilities.getSelfOrAncestorOfType(outline, TemplateDockable.class);
        }
        setEnabled(shouldEnable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline outline = getOutline();
        if (outline != null) {
            OutlineModel outlineModel = outline.getModel();
            if (outlineModel.hasSelection()) {
                TemplateDockable template = TemplateDockable.getLastActivated();
                if (template != null) {
                    template.addRows(outlineModel.getSelectionAsList(true));
                }
            }
        }
    }

    private ListOutline getOutline() {
        ListOutline outline = mOutline;
        if (outline == null) {
            LibraryDockable library = getTarget(LibraryDockable.class);
            if (library != null) {
                outline = library.getOutline();
            }
        }
        return outline;
    }

    private TemplateDockable getTemplate() {
        TemplateDockable template = mTemplate;
        if (template == null) {
            template = TemplateDockable.getLastActivated();
        }
        return template;
    }
}
