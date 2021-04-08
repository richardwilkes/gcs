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

package com.trollworks.gcs.menu;

import javax.swing.Action;
import javax.swing.JMenu;

public final class MenuHelpers {
    private MenuHelpers() {
    }

    public static JMenu createSubMenu(String name, Action... actions) {
        JMenu menu = new JMenu(name);
        for (Action action : actions) {
            if (action != null) {
                menu.add(new DynamicMenuItem(action));
            } else {
                menu.addSeparator();
            }
        }
        DynamicMenuEnabler.add(menu);
        return menu;
    }
}
