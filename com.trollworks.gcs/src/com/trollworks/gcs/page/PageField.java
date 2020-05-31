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

package com.trollworks.gcs.page;

import com.trollworks.gcs.character.Armor;
import com.trollworks.gcs.character.CharacterSheet;
import com.trollworks.gcs.character.Encumbrance;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.Profile;
import com.trollworks.gcs.character.Settings;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.utility.Platform;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.DateTimeFormatter;
import com.trollworks.gcs.utility.text.DiceFormatter;
import com.trollworks.gcs.utility.text.DoubleFormatter;
import com.trollworks.gcs.utility.text.HeightFormatter;
import com.trollworks.gcs.utility.text.IntegerFormatter;
import com.trollworks.gcs.utility.text.Text;
import com.trollworks.gcs.utility.text.WeightFormatter;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Rectangle;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.FocusEvent;
import java.beans.PropertyChangeEvent;
import java.beans.PropertyChangeListener;
import java.text.ParseException;
import java.util.HashMap;
import javax.swing.JFormattedTextField;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.plaf.basic.BasicTextFieldUI;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** A generic field for a page. */
public class PageField extends JFormattedTextField implements NotifierTarget, PropertyChangeListener, ActionListener, Commitable {
    private CharacterSheet mSheet;
    private String         mConsumedType;

    /**
     * Creates a new, left-aligned, text input field.
     *
     * @param sheet        The sheet to listen to.
     * @param consumedType The field to listen to.
     */
    public PageField(CharacterSheet sheet, String consumedType) {
        this(sheet, consumedType, SwingConstants.LEFT, true, null);
    }

    /**
     * Creates a new, left-aligned, text input field.
     *
     * @param sheet        The sheet to listen to.
     * @param consumedType The field to listen to.
     * @param tooltip      The tooltip to set.
     */
    public PageField(CharacterSheet sheet, String consumedType, String tooltip) {
        this(sheet, consumedType, SwingConstants.LEFT, true, tooltip);
    }

    /**
     * Creates a new text input field.
     *
     * @param sheet        The sheet to listen to.
     * @param consumedType The field to listen to.
     * @param alignment    The alignment of the field.
     * @param tooltip      The tooltip to set.
     */
    public PageField(CharacterSheet sheet, String consumedType, int alignment, String tooltip) {
        this(sheet, consumedType, alignment, true, tooltip);
    }

    /**
     * Creates a new text input field.
     *
     * @param sheet        The sheet to listen to.
     * @param consumedType The field to listen to.
     * @param alignment    The alignment of the field.
     * @param editable     Whether or not the user can edit this field.
     * @param tooltip      The tooltip to set.
     */
    public PageField(CharacterSheet sheet, String consumedType, int alignment, boolean editable, String tooltip) {
        super(getFormatterFactoryForType(sheet.getCharacter(), consumedType), sheet.getCharacter().getValueForID(consumedType));
        if (Platform.isLinux()) {
            // I override the UI here since the GTK UI on Linux has no way to turn off the border
            // around text fields.
            setUI(new BasicTextFieldUI());
        }
        mSheet = sheet;
        mConsumedType = consumedType;
        setFont(sheet.getScale().scale(UIManager.getFont(Fonts.KEY_FIELD_PRIMARY)));
        setBorder(null);
        setOpaque(false);
        // Just setting opaque to false isn't enough for some reason, so I'm also setting the
        // background color to a 100% transparent value.
        setBackground(new Color(255, 255, 255, 0));
        setHorizontalAlignment(alignment);
        setEditable(editable);
        setEnabled(editable);
        if (editable) {
            setForeground(new Color(0, 0, 192));
        } else {
            setDisabledTextColor(Color.BLACK);
        }
        setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        mSheet.getCharacter().addTarget(this, mConsumedType);
        addPropertyChangeListener("value", this);
        addActionListener(this);
        setFocusLostBehavior(COMMIT_OR_REVERT);
    }

    @Override
    public Dimension getPreferredSize() {
        Dimension size = super.getPreferredSize();
        // Don't know why this is needed, but it seems to be. Without it, text is being truncated by
        // about 2 pixels.
        size.width += mSheet.getScale().scale(2);
        return size;
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        if (name.startsWith(Settings.PREFIX)) {
            setValue(mSheet.getCharacter().getSettings().optionsCode());
        } else {
            setValue(data);
        }
        invalidate();
        repaint();
    }

    @Override
    protected void processFocusEvent(FocusEvent event) {
        super.processFocusEvent(event);
        if (event.getID() == FocusEvent.FOCUS_GAINED) {
            selectAll();
        }
    }

    @Override
    protected void paintComponent(Graphics gc) {
        if (isEditable()) {
            Rectangle bounds = getBounds();
            bounds.x = 0;
            bounds.y = 0;
            gc.setColor(Color.lightGray);
            int height = mSheet.getScale().scale(1);
            gc.fillRect(bounds.x, bounds.y + bounds.height - height, bounds.width, height);
            gc.setColor(getForeground());
        }
        super.paintComponent(GraphicsUtilities.prepare(gc));
    }

    /** @return The consumed type. */
    public String getConsumedType() {
        return mConsumedType;
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        if (isEditable()) {
            mSheet.getCharacter().setValueForID(mConsumedType, getValue());
        }
    }

    private static final HashMap<String, AbstractFormatterFactory> FACTORY_MAP = new HashMap<>();
    private static final AbstractFormatterFactory                  DEFAULT_FACTORY;

    static {
        DefaultFormatterFactory factory = new DefaultFormatterFactory(new WeightFormatter(true));
        FACTORY_MAP.put(Profile.ID_WEIGHT, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_BASIC_LIFT, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_ONE_HANDED_LIFT, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_TWO_HANDED_LIFT, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_SHOVE_AND_KNOCK_OVER, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_RUNNING_SHOVE_AND_KNOCK_OVER, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_CARRY_ON_BACK, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_SHIFT_SLIGHTLY, factory);
        for (Encumbrance encumbrance : Encumbrance.values()) {
            FACTORY_MAP.put(GURPSCharacter.MAXIMUM_CARRY_PREFIX + encumbrance.ordinal(), factory);
        }

        factory = new DefaultFormatterFactory(new IntegerFormatter(0, 99999, false));
        FACTORY_MAP.put(GURPSCharacter.ID_STRENGTH, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_DEXTERITY, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_INTELLIGENCE, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_HEALTH, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_WILL, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_FRIGHT_CHECK, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_BASIC_MOVE, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_PERCEPTION, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_VISION, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_HEARING, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_TASTE_AND_SMELL, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_TOUCH, factory);
        FACTORY_MAP.put(Armor.ID_EYES_DR, factory);
        FACTORY_MAP.put(Armor.ID_SKULL_DR, factory);
        FACTORY_MAP.put(Armor.ID_FACE_DR, factory);
        FACTORY_MAP.put(Armor.ID_LEG_DR, factory);
        FACTORY_MAP.put(Armor.ID_ARM_DR, factory);
        FACTORY_MAP.put(Armor.ID_TORSO_DR, factory);
        FACTORY_MAP.put(Armor.ID_GROIN_DR, factory);
        FACTORY_MAP.put(Armor.ID_HAND_DR, factory);
        FACTORY_MAP.put(Armor.ID_FOOT_DR, factory);
        FACTORY_MAP.put(Armor.ID_NECK_DR, factory);
        for (Encumbrance encumbrance : Encumbrance.values()) {
            int index = encumbrance.ordinal();
            FACTORY_MAP.put(GURPSCharacter.MOVE_PREFIX + index, factory);
            FACTORY_MAP.put(GURPSCharacter.DODGE_PREFIX + index, factory);
        }

        factory = new DefaultFormatterFactory(new IntegerFormatter(-999999, 999999, false));
        FACTORY_MAP.put(GURPSCharacter.ID_ATTRIBUTE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_ADVANTAGE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_DISADVANTAGE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_QUIRK_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_SKILL_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_SPELL_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_RACE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_UNSPENT_POINTS, factory);

        factory = new DefaultFormatterFactory(new IntegerFormatter(-9999999, 9999999, false));
        FACTORY_MAP.put(GURPSCharacter.ID_TIRED_FATIGUE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_UNCONSCIOUS_CHECKS_FATIGUE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_UNCONSCIOUS_FATIGUE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_REELING_HIT_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_UNCONSCIOUS_CHECKS_HIT_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_1_HIT_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_2_HIT_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_3_HIT_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_DEATH_CHECK_4_HIT_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_DEAD_HIT_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_CURRENT_HP, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_CURRENT_FP, factory);

        factory = new DefaultFormatterFactory(new IntegerFormatter(0, 999999, false));
        FACTORY_MAP.put(GURPSCharacter.ID_FATIGUE_POINTS, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_HIT_POINTS, factory);

        factory = new DefaultFormatterFactory(new DateTimeFormatter());
        FACTORY_MAP.put(GURPSCharacter.ID_CREATED, factory);
        FACTORY_MAP.put(GURPSCharacter.ID_MODIFIED, factory);

        FACTORY_MAP.put(Profile.ID_SIZE_MODIFIER, new DefaultFormatterFactory(new IntegerFormatter(-99, 9999, true)));
        FACTORY_MAP.put(Profile.ID_AGE, new DefaultFormatterFactory(new IntegerFormatter(0, Integer.MAX_VALUE, false, true)));
        FACTORY_MAP.put(Profile.ID_HEIGHT, new DefaultFormatterFactory(new HeightFormatter(true)));
        FACTORY_MAP.put(GURPSCharacter.ID_BASIC_SPEED, new DefaultFormatterFactory(new DoubleFormatter(0, 99999, false)));

        DefaultFormatter formatter = new DefaultFormatter();
        formatter.setOverwriteMode(false);
        DEFAULT_FACTORY = new DefaultFormatterFactory(formatter);
    }

    private static AbstractFormatterFactory getFormatterFactoryForType(GURPSCharacter character, String type) {
        if (GURPSCharacter.ID_BASIC_THRUST.equals(type) || GURPSCharacter.ID_BASIC_SWING.equals(type)) {
            return new DefaultFormatterFactory(new DiceFormatter(character));
        }
        AbstractFormatterFactory factory = FACTORY_MAP.get(type);
        return factory != null ? factory : DEFAULT_FACTORY;
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        attemptCommit();
    }

    @Override
    public void attemptCommit() {
        try {
            commitEdit();
        } catch (ParseException exception) {
            invalidEdit();
        }
    }
}
