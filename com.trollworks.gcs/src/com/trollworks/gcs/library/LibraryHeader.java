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

package com.trollworks.gcs.library;

import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.border.LineBorder;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.ScaleRoot;
import com.trollworks.gcs.ui.widget.Panel;
import com.trollworks.gcs.ui.widget.outline.OutlineHeader;

import java.awt.BorderLayout;

public class LibraryHeader extends Panel implements ScaleRoot {
    private Scale mScale;

    public LibraryHeader(OutlineHeader header) {
        super(new BorderLayout());
        header.setBorder(new LineBorder(ThemeColor.DIVIDER, 0, 0, 0, 1));
        add(header);
        mScale = Settings.getInstance().getInitialUIScale().getScale();
    }

    @Override
    public Scale getScale() {
        return mScale;
    }

    @Override
    public void setScale(Scale scale) {
        if (mScale.getScale() != scale.getScale()) {
            mScale = scale;
            revalidate();
            repaint();
        }
    }
}
