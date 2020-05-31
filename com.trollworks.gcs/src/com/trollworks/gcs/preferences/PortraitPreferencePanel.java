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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.RetinaIcon;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.TitledBorder;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.widget.ActionPanel;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.Text;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.text.MessageFormat;
import javax.swing.UIManager;

/** The character portrait. */
public class PortraitPreferencePanel extends ActionPanel {
    private RetinaIcon mPortrait;

    /**
     * Creates a new character portrait.
     *
     * @param image The image to display.
     */
    public PortraitPreferencePanel(Img image) {
        mPortrait = Profile.createPortrait(image);
        setBorder(new TitledBorder(UIManager.getFont(Fonts.KEY_LABEL_PRIMARY), I18n.Text("Portrait")));
        Insets insets = getInsets();
        UIUtilities.setOnlySize(this, new Dimension(insets.left + insets.right + Profile.PORTRAIT_WIDTH, insets.top + insets.bottom + Profile.PORTRAIT_HEIGHT));
        setToolTipText(Text.wrapPlainTextForToolTip(MessageFormat.format(I18n.Text("<html><body>The portrait to use when a new character sheet is created.<br><br>Ideal original portrait size is {0} pixels wide by {1} pixels tall,<br>although the image will be automatically scaled to these<br>dimensions, if necessary.</body></html>"), Integer.valueOf(Profile.PORTRAIT_WIDTH * 2), Integer.valueOf(Profile.PORTRAIT_HEIGHT * 2))));
        addMouseListener(new MouseAdapter() {
            @Override
            public void mouseClicked(MouseEvent event) {
                if (event.getClickCount() == 2) {
                    notifyActionListeners();
                }
            }
        });
    }

    /** @param image The new portrait. */
    public void setPortrait(Img image) {
        mPortrait = Profile.createPortrait(image);
        repaint();
    }

    @Override
    protected void paintComponent(Graphics gc) {
        Insets    insets = getInsets();
        Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
        gc.setColor(Color.white);
        gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
        if (mPortrait != null) {
            mPortrait.paintIcon(this, gc, insets.left, insets.top);
        }
    }
}
