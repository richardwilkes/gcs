/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
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
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.GraphicsUtilities;
import com.trollworks.toolkit.ui.RetinaIcon;
import com.trollworks.toolkit.ui.border.TitledBorder;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.scale.Scale;
import com.trollworks.toolkit.ui.widget.WindowUtils;
import com.trollworks.toolkit.utility.Localization;
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
    @Localize("Select A Portrait")
    @Localize(locale = "de", value = "Wähle ein Charakterild")
    @Localize(locale = "ru", value = "Выберите изображение")
    @Localize(locale = "es", value = "Selecciona un retrato")
    private static String SELECT_PORTRAIT;
    @Localize("Portrait")
    @Localize(locale = "de", value = "Charakterbild")
    @Localize(locale = "ru", value = "Изображение")
    @Localize(locale = "es", value = "Retrato")
    private static String PORTRAIT;
    @Localize("<html><body><b>Double-click</b> to set a character portrait.<br><br>The dimensions of the chosen picture should be in a ratio of<br><b>3 pixels wide for every 4 pixels tall</b> to scale without distortion.<br><br>Dimensions of <b>{0}x{1}</b> are ideal.</body></html>")
    @Localize(locale = "de", value = "<html><body><b>Doppelklicken</b>, um ein Charakterbild anzugeben.<br><br>Das gewählte Bild sollte ein <b>Seitenverhältnis von 3:4</b><br> aufweisen, um unverzerrt dargestellt zu werden.<br><br>Eine Größe von <b>{0}x{1} Pixel</b> ist ideal.</body></html>")
    @Localize(locale = "ru", value = "<html><body><b>Дважды щёлкните</b> чтобы изменить изображение персонажа.<br><br>Для масштабирования без искажений, размер картинки должен быть<br>в пропорции <b>3 пикселя в ширину на 4 пикселя в высоту</b>.<br><br>Размер <b>{0}x{1}</b> будет идеальным.</body></html>")
    @Localize(locale = "es", value = "<html><body><b>Dobleclic</b> para establecer el retarto del personaje.<br><br>Las dimensiones de la imagen seleccionada deben mantener un ratio de<br><b>3 pixels de ancho por cada 4 pixels de alto</b> para mostrarse sin distorsión.<br><br> <b>{0}x{1}</b> es la dimensión ideal.</body></html>")
    private static String PORTRAIT_TOOLTIP;
    @Localize("Unable to load\n{0}.")
    @Localize(locale = "de", value = "Kann Datei {0} nicht laden.")
    @Localize(locale = "ru", value = "Невозможно загрузить\n{0}.")
    @Localize(locale = "es", value = "No puede cargarse\n{0}.")
    private static String BAD_IMAGE;

    static {
        Localization.initialize();
    }

    private CharacterSheet mSheet;

    /**
     * Creates a new character portrait.
     *
     * @param sheet The owning sheet.
     */
    public PortraitPanel(CharacterSheet sheet) {
        super(null, true);
        setBorder(new TitledBorder(UIManager.getFont(GCSFonts.KEY_LABEL), PORTRAIT));
        mSheet = sheet;
        setToolTipText(Text.wrapPlainTextForToolTip(MessageFormat.format(PORTRAIT_TOOLTIP, new Integer(Profile.PORTRAIT_WIDTH * 2), new Integer(Profile.PORTRAIT_HEIGHT * 2))));
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
                WindowUtils.showError(this, MessageFormat.format(BAD_IMAGE, PathUtils.getFullPath(file)));
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
        Scale scale = Scale.get(this);
        Insets insets = getInsets();
        return new Dimension(insets.left + scale.scale(Profile.PORTRAIT_WIDTH) + insets.right, insets.top + scale.scale(Profile.PORTRAIT_HEIGHT) + insets.bottom);
    }

    @Override
    public Dimension getMaximumSize() {
        return getPreferredSize();
    }
}
