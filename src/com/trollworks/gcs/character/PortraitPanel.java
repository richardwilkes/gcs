/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.character;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.gcs.page.DropPanel;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.RetinaIcon;
import com.trollworks.toolkit.ui.border.TitledBorder;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.scale.Scale;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.notification.NotifierTarget;
import com.trollworks.toolkit.utility.text.Text;

import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Insets;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.io.File;
import java.text.MessageFormat;
import javax.swing.UIManager;

/** The character portrait. */
public class PortraitPanel extends DropPanel implements NotifierTarget {
    private CharacterSheet mSheet;

    /**
     * Creates a new character portrait.
     *
     * @param sheet The owning sheet.
     */
    public PortraitPanel(CharacterSheet sheet) {
        super(null, true);
        setBorder(new TitledBorder(UIManager.getFont(GCSFonts.KEY_LABEL), I18n.Text("Portrait")));
        mSheet = sheet;
        setToolTipText(Text.wrapPlainTextForToolTip(MessageFormat.format(I18n.Text("<html><body><b>Double-click</b> to set a character portrait.<br><br>The dimensions of the chosen picture should be in a ratio of<br><b>3 pixels wide for every 4 pixels tall</b> to scale without distortion.<br><br>Dimensions of <b>{0}x{1}</b> are ideal.</body></html>"), Integer.valueOf(Profile.PORTRAIT_WIDTH * 2), Integer.valueOf(Profile.PORTRAIT_HEIGHT * 2))));
        sheet.getCharacter().addTarget(this, Profile.ID_PORTRAIT);
        addMouseListener(new MouseAdapter() {
            @Override
            public void mouseClicked(MouseEvent event) {
                if (event.getClickCount() == 2) {
                    choosePortrait();
                }
            }
        });
    }

    /** Allows the user to choose a portrait for their character. */
    public void choosePortrait() {
        File file = SheetPreferences.choosePortrait();
        if (file != null) {
            try {
                mSheet.getCharacter().getDescription().setPortrait(StdImage.loadImage(file));
            } catch (Exception exception) {
                WindowUtils.showError(this, MessageFormat.format(I18n.Text("Unable to load\n{0}."), PathUtils.getFullPath(file)));
            }
        }
    }

    @Override
    protected void paintComponent(Graphics g) {
        Graphics2D gc = GraphicsUtilities.prepare(g);
        super.paintComponent(gc);
        RetinaIcon portrait = mSheet.getCharacter().getDescription().getPortrait();
        if (portrait != null) {
            Insets insets = getInsets();
            portrait.paintIcon(this, gc, insets.left, insets.top);
        }
    }

    @Override
    public void handleNotification(Object producer, String type, Object data) {
        repaint();
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public Dimension getMinimumSize() {
        return getPreferredSize();
    }

    @Override
    public Dimension getPreferredSize() {
        Scale  scale  = Scale.get(this);
        Insets insets = getInsets();
        return new Dimension(insets.left + scale.scale(Profile.PORTRAIT_WIDTH) + insets.right, insets.top + scale.scale(Profile.PORTRAIT_HEIGHT) + insets.bottom);
    }

    @Override
    public Dimension getMaximumSize() {
        return getPreferredSize();
    }
}
