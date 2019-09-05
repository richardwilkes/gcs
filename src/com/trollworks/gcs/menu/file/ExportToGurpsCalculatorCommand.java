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

package com.trollworks.gcs.menu.file;

import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.GURPSCalculator;
import com.trollworks.toolkit.ui.menu.Command;
import com.trollworks.toolkit.utility.I18n;

import java.awt.event.ActionEvent;
import java.awt.event.KeyEvent;

public class ExportToGurpsCalculatorCommand extends Command {
    /** The action command this command will issue. */
    public static final String                         CMD_EXPORT_TO_GUPRS_CALCULATOR = "ExportToGurpsCalculator";
    /** The singleton {@link ExportToGurpsCalculatorCommand}. */
    public static final ExportToGurpsCalculatorCommand INSTANCE                       = new ExportToGurpsCalculatorCommand();

    private ExportToGurpsCalculatorCommand() {
        super(I18n.Text("Export to GURPS Calculator\u2026"), CMD_EXPORT_TO_GUPRS_CALCULATOR, KeyEvent.VK_L);
    }

    @Override
    public void adjust() {
        setEnabled(getTarget(CharacterSheet.class) != null);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        GURPSCalculator.export(getTarget(CharacterSheet.class));
    }
}
