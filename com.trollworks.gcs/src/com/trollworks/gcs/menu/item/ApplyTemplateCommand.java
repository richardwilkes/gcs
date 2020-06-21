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

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.undo.MultipleUndo;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;
import java.util.ArrayList;
import java.util.List;

/** Provides the "Apply Template To Sheet" command. */
public class ApplyTemplateCommand extends Command {
    /** The action command this command will issue. */
    public static final String               CMD_APPLY_TEMPLATE = "ApplyTemplate";
    /** The singleton {@link ApplyTemplateCommand}. */
    public static final ApplyTemplateCommand INSTANCE           = new ApplyTemplateCommand();
    private             TemplateDockable     mTemplate;
    private             SheetDockable        mSheet;

    private ApplyTemplateCommand() {
        super(I18n.Text("Apply Template To Character Sheet"), CMD_APPLY_TEMPLATE, KeyEvent.VK_A, SHIFTED_COMMAND_MODIFIER);
    }

    /**
     * Creates a new {@link ApplyTemplateCommand}.
     *
     * @param sheet The sheet to target.
     */
    public ApplyTemplateCommand(TemplateDockable template, SheetDockable sheet) {
        super(sheet.getTitle(), CMD_APPLY_TEMPLATE);
        mTemplate = template;
        mSheet = sheet;
    }

    @Override
    public void adjust() {
        TemplateDockable template = mTemplate != null ? mTemplate : getTarget(TemplateDockable.class);
        if (template != null) {
            setEnabled((mSheet != null ? mSheet : SheetDockable.getLastActivated()) != null);
        } else {
            setEnabled(false);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        TemplateDockable templateDockable = mTemplate != null ? mTemplate : getTarget(TemplateDockable.class);
        if (templateDockable != null) {
            SheetDockable sheetDockable = mSheet != null ? mSheet : SheetDockable.getLastActivated();
            if (sheetDockable != null) {
                Template     template = templateDockable.getDataFile();
                MultipleUndo edit     = new MultipleUndo(I18n.Text("Apply Template"));
                List<Row>    rows     = new ArrayList<>();
                template.addEdit(edit);
                rows.addAll(template.getAdvantagesModel().getTopLevelRows());
                rows.addAll(template.getSkillsModel().getTopLevelRows());
                rows.addAll(template.getSpellsModel().getTopLevelRows());
                rows.addAll(template.getEquipmentModel().getTopLevelRows());
                rows.addAll(template.getOtherEquipmentModel().getTopLevelRows());
                rows.addAll(template.getNotesModel().getTopLevelRows());
                sheetDockable.addRows(rows);
                edit.end();
            }
        }
    }
}
