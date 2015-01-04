/*
 * Copyright (c) 1998-2015 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License,
 * version 2.0. If a copy of the MPL was not distributed with this file, You
 * can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined
 * by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.gcs.character.Profile;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.border.BoxedDropShadowBorder;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.ActionPanel;
import com.trollworks.toolkit.utility.Localization;

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
	@Localize("Portrait")
	@Localize(locale = "de", value = "Charakterbild")
	@Localize(locale = "ru", value = "Изображение")
	private static String	PORTRAIT;
	@Localize("<html><body>The portrait to use when a new character sheet is created.<br><br>Ideal original portrait size is {0} pixels wide by {1} pixels tall,<br>although the image will be automatically scaled to these<br>dimensions, if necessary.</body></html>")
	@Localize(locale = "de", value = "<html><body>Das zu verwendende Charakterbild, wenn ein neues Charakterblatt erstellt wird.<br><br>Ideale Größe für das Bild ist {0} Pixel breit und {1} Pixel hoch,<br>aber das Bild wird automatisch auf diese Maße<br>skaliert, wenn dies notwendig sein sollte.</body></html>")
	@Localize(locale = "ru", value = "<html><body>Изображение, используемое при создании нового листа персонажа.<br><br>Наилучший размер изображения - {0} пикселей в ширину и {1} пикселей в высоту, но при необходимости, <br>изображение будет автоматически смасштабировано в этипропорции.</body></html>")
	private static String	PORTRAIT_TOOLTIP;

	static {
		Localization.initialize();
	}

	private StdImage		mImage;

	/**
	 * Creates a new character portrait.
	 *
	 * @param image The image to display.
	 */
	public PortraitPreferencePanel(StdImage image) {
		super();
		mImage = image;
		setBorder(new BoxedDropShadowBorder(UIManager.getFont(GCSFonts.KEY_LABEL), PORTRAIT));
		Insets insets = getInsets();
		UIUtilities.setOnlySize(this, new Dimension(insets.left + insets.right + Profile.PORTRAIT_WIDTH, insets.top + insets.bottom + Profile.PORTRAIT_HEIGHT));
		setToolTipText(MessageFormat.format(PORTRAIT_TOOLTIP, new Integer(Profile.PORTRAIT_WIDTH * 2), new Integer(Profile.PORTRAIT_HEIGHT * 2)));
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
	public void setPortrait(StdImage image) {
		mImage = image;
		repaint();
	}

	@Override
	protected void paintComponent(Graphics gc) {
		Insets insets = getInsets();
		Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
		gc.setColor(Color.white);
		gc.fillRect(bounds.x, bounds.y, bounds.width, bounds.height);
		if (mImage != null) {
			gc.drawImage(mImage, bounds.x, bounds.y, null);
		}
	}
}
