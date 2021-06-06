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

package com.trollworks.gcs.menu.settings;

import com.trollworks.gcs.character.SheetDockable;
import com.trollworks.gcs.menu.Command;
import com.trollworks.gcs.settings.AttributeSettingsWindow;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.utility.I18n;

import java.awt.event.ActionEvent;

public final class AttributeSettingsCommand extends Command {
    public static final AttributeSettingsCommand PER_SHEET = new AttributeSettingsCommand(false);
    public static final AttributeSettingsCommand DEFAULTS  = new AttributeSettingsCommand(true);
    private             boolean                  mForDefaults;

    private AttributeSettingsCommand(boolean defaults) {
        super(defaults ? I18n.text("Default Attributes…") : I18n.text("Attributes…"),
                defaults ? "default_attribute_settings" : "attribute_settings");
        mForDefaults = defaults;
    }

    @Override
    public void adjust() {
        boolean shouldEnable = !UIUtilities.inModalState();
        if (shouldEnable && !mForDefaults) {
            shouldEnable = getTarget(SheetDockable.class) != null;
        }
        setEnabled(shouldEnable);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        if (!UIUtilities.inModalState()) {
            if (mForDefaults) {
                AttributeSettingsWindow.display(null);
            } else {
                SheetDockable target = getTarget(SheetDockable.class);
                if (target != null) {
                    AttributeSettingsWindow.display(target.getDataFile());
                }
            }
        }
    }
}
