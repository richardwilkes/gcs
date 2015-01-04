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

package com.trollworks.gcs.character;

import com.trollworks.gcs.app.GCSFonts;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.border.BoxedDropShadowBorder;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.print.PrintUtilities;
import com.trollworks.toolkit.ui.widget.StdFileDialog;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.notification.NotifierTarget;

import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Image;
import java.awt.Insets;
import java.awt.Rectangle;
import java.awt.RenderingHints;
import java.awt.event.MouseAdapter;
import java.awt.event.MouseEvent;
import java.awt.geom.AffineTransform;
import java.io.File;
import java.text.MessageFormat;

import javax.swing.UIManager;

/** The character portrait. */
public class PortraitPanel extends DropPanel implements NotifierTarget {
	@Localize("Select A Portrait")
	@Localize(locale = "de", value = "Wähle ein Charakterild")
	@Localize(locale = "ru", value = "Выберите изображение")
	private static String	SELECT_PORTRAIT;
	@Localize("Portrait")
	@Localize(locale = "de", value = "Charakterbild")
	@Localize(locale = "ru", value = "Изображение")
	private static String	PORTRAIT;
	@Localize("<html><body><b>Double-click</b> to set a character portrait.<br><br>The dimensions of the chosen picture should be in a ratio of<br><b>3 pixels wide for every 4 pixels tall</b> to scale without distortion.<br><br>Dimensions of <b>{0}x{1}</b> are ideal.</body></html>")
	@Localize(locale = "de", value = "<html><body><b>Doppelklicken</b>, um ein Charakterbild anzugeben.<br><br>Das gewählte Bild sollte ein <b>Seitenverhältnis von 3:4</b><br> aufweisen, um unverzerrt dargestellt zu werden.<br><br>Eine Größe von <b>{0}x{1} Pixel</b> ist ideal.</body></html>")
	@Localize(locale = "ru", value = "<html><body><b>Дважды щёлкните</b> чтобы изменить изображение персонажа.<br><br>Для масштабирования без искажений, размер картинки должен быть<br>в пропорции <b>3 пикселя в ширину на 4 пикселя в высоту</b>.<br><br>Размер <b>{0}x{1}</b> будет идеальным.</body></html>")
	private static String	PORTRAIT_TOOLTIP;
	@Localize("Unable to load\n{0}.")
	@Localize(locale = "de", value = "Kann Datei {0} nicht laden.")
	@Localize(locale = "ru", value = "Невозможно загрузить\n{0}.")
	private static String	BAD_IMAGE;

	static {
		Localization.initialize();
	}

	private GURPSCharacter	mCharacter;

	/**
	 * Creates a new character portrait.
	 *
	 * @param character The owning character.
	 */
	public PortraitPanel(GURPSCharacter character) {
		super(null, true);
		setBorder(new BoxedDropShadowBorder(UIManager.getFont(GCSFonts.KEY_LABEL), PORTRAIT));
		mCharacter = character;
		Insets insets = getInsets();
		UIUtilities.setOnlySize(this, new Dimension(insets.left + insets.right + Profile.PORTRAIT_WIDTH, insets.top + insets.bottom + Profile.PORTRAIT_HEIGHT));
		setToolTipText(MessageFormat.format(PORTRAIT_TOOLTIP, new Integer(Profile.PORTRAIT_WIDTH * 2), new Integer(Profile.PORTRAIT_HEIGHT * 2)));
		mCharacter.addTarget(this, Profile.ID_PORTRAIT);
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
		File file = StdFileDialog.choose(this, true, SELECT_PORTRAIT, null, null, "png", "jpg", "gif", "jpeg"); //$NON-NLS-1$ //$NON-NLS-2$ //$NON-NLS-3$ //$NON-NLS-4$
		if (file != null) {
			try {
				mCharacter.getDescription().setPortrait(StdImage.loadImage(file));
			} catch (Exception exception) {
				WindowUtils.showError(this, MessageFormat.format(BAD_IMAGE, PathUtils.getFullPath(file)));
			}
		}
	}

	@Override
	protected void paintComponent(Graphics g) {
		Graphics2D gc = GraphicsUtilities.prepare(g);
		super.paintComponent(gc);

		boolean isPrinting = PrintUtilities.isPrinting(this);
		Image portrait = mCharacter.getDescription().getPortrait(isPrinting);

		if (portrait != null) {
			Insets insets = getInsets();
			Rectangle bounds = new Rectangle(insets.left, insets.top, getWidth() - (insets.left + insets.right), getHeight() - (insets.top + insets.bottom));
			AffineTransform transform = null;

			if (isPrinting) {
				transform = gc.getTransform();
				gc.scale(0.5, 0.5);
				bounds.x *= 2;
				bounds.y *= 2;
			}
			RenderingHints saved = GraphicsUtilities.setMaximumQualityForGraphics(gc);
			gc.drawImage(portrait, bounds.x, bounds.y, null);
			gc.setRenderingHints(saved);
			if (isPrinting) {
				gc.setTransform(transform);
			}
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
}
