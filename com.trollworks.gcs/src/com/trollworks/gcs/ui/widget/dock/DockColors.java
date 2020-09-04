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

package com.trollworks.gcs.ui.widget.dock;

import com.trollworks.gcs.ui.Colors;

import java.awt.Color;
import javax.swing.UIManager;

/** Provides the colors used by the {@link Dock}. */
public final class DockColors {
    public static Color BACKGROUND             = UIManager.getColor("Panel.background");
    public static Color ACTIVE_TAB_BACKGROUND  = new Color(224, 212, 175);
    public static Color CURRENT_TAB_BACKGROUND = Colors.adjustBrightness(Colors.adjustSaturation(ACTIVE_TAB_BACKGROUND, -0.15f), -0.05f);
    public static Color HIGHLIGHT              = Colors.adjustBrightness(BACKGROUND, 0.2f);
    public static Color SHADOW                 = Colors.adjustBrightness(BACKGROUND, -0.2f);
    public static Color DROP_AREA_OUTER_BORDER = Color.BLUE;
    public static Color DROP_AREA_INNER_BORDER = Color.WHITE;
    public static Color DROP_AREA              = Colors.getWithAlpha(DROP_AREA_OUTER_BORDER, 64);

    private DockColors() {
    }
}
