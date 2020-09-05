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

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantagesDockable;
import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.template.TemplateDockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

/** Provides the "New Advantage" command. */
public class NewAdvantageCommand extends Command {
    /** The action command this command will issue. */
    public static final String              CMD_NEW_ADVANTAGE           = "NewAdvantage";
    /** The action command this command will issue. */
    public static final String              CMD_NEW_ADVANTAGE_CONTAINER = "NewAdvantageContainer";
    /** The "New Advantage" command. */
    public static final NewAdvantageCommand INSTANCE                    = new NewAdvantageCommand(false, I18n.Text("New Advantage"), CMD_NEW_ADVANTAGE, COMMAND_MODIFIER);
    /** The "New Advantage Container" command. */
    public static final NewAdvantageCommand CONTAINER_INSTANCE          = new NewAdvantageCommand(true, I18n.Text("New Advantage Container"), CMD_NEW_ADVANTAGE_CONTAINER, SHIFTED_COMMAND_MODIFIER);
    private             boolean             mContainer;

    private NewAdvantageCommand(boolean container, String title, String cmd, int modifiers) {
        super(title, cmd, KeyEvent.VK_D, modifiers);
        mContainer = container;
    }

    @Override
    public void adjust() {
        AdvantagesDockable adq = getTarget(AdvantagesDockable.class);
        if (adq != null) {
            setEnabled(!adq.getOutline().getModel().isLocked());
        } else {
            if (getTarget(SheetDockable.class) != null) {
                setEnabled(true);
            } else {
                setEnabled(getTarget(TemplateDockable.class) != null);
            }
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        ListOutline        outline;
        DataFile           dataFile;
        AdvantagesDockable adq = getTarget(AdvantagesDockable.class);
        if (adq != null) {
            dataFile = adq.getDataFile();
            outline = adq.getOutline();
            if (outline.getModel().isLocked()) {
                return;
            }
        } else {
            SheetDockable sheet = getTarget(SheetDockable.class);
            if (sheet != null) {
                dataFile = sheet.getDataFile();
                outline = sheet.getSheet().getAdvantageOutline();
            } else {
                TemplateDockable template = getTarget(TemplateDockable.class);
                if (template != null) {
                    dataFile = template.getDataFile();
                    outline = template.getTemplate().getAdvantageOutline();
                } else {
                    return;
                }
            }
        }
        Advantage advantage = new Advantage(dataFile, mContainer);
        outline.addRow(advantage, getTitle(), false);
        outline.getModel().select(advantage, false);
        outline.scrollSelectionIntoView();
        outline.openDetailEditor(true);
    }
}
