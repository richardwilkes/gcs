/*
 * Copyright (c) 1998-2019 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.skill;

import com.trollworks.gcs.common.EditorPanel;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.Alignment;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.widget.Commitable;
import com.trollworks.toolkit.ui.widget.EditorField;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.utility.I18n;
import com.trollworks.toolkit.utility.text.IntegerFormatter;

import java.awt.Insets;
import java.awt.event.ActionEvent;
import java.beans.PropertyChangeEvent;

import javax.swing.JComboBox;
import javax.swing.JComponent;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** A skill default editor panel. */
public class SkillDefaultEditor extends EditorPanel {
    private static SkillDefaultType LAST_ITEM_TYPE = SkillDefaultType.DX;
    private SkillDefault            mDefault;
    private JComboBox<?>            mTypeCombo;
    private EditorField             mSkillNameField;
    private EditorField             mSpecializationField;
    private EditorField             mModifierField;

    /** @param type The last item type created or switched to. */
    public static void setLastItemType(SkillDefaultType type) {
        LAST_ITEM_TYPE = type;
    }

    /** Creates a new placeholder {@link SkillDefaultEditor}. */
    public SkillDefaultEditor() {
        this(null);
    }

    /**
     * Creates a new skill default editor panel.
     *
     * @param skillDefault The skill default to edit.
     */
    public SkillDefaultEditor(SkillDefault skillDefault) {
        super();
        mDefault = skillDefault;
        rebuild();
    }

    /** Rebuilds the contents of this panel with the current feature settings. */
    protected void rebuild() {
        removeAll();

        if (mDefault != null) {
            FlexGrid grid = new FlexGrid();

            OrLabel  or   = new OrLabel(this);
            add(or);
            grid.add(or, 0, 0);

            FlexRow          row     = new FlexRow();
            SkillDefaultType current = mDefault.getType();
            mTypeCombo = new JComboBox<Object>(SkillDefaultType.values());
            mTypeCombo.setOpaque(false);
            mTypeCombo.setSelectedItem(current);
            mTypeCombo.setActionCommand(SkillDefault.TAG_TYPE);
            mTypeCombo.addActionListener(this);
            UIUtilities.setOnlySize(mTypeCombo, mTypeCombo.getPreferredSize());
            add(mTypeCombo);
            row.add(mTypeCombo);
            grid.add(row, 0, 1);

            mModifierField = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(-99, 99, true)), this, SwingConstants.LEFT, Integer.valueOf(mDefault.getModifier()), Integer.valueOf(99), null);
            UIUtilities.setOnlySize(mModifierField, mModifierField.getPreferredSize());
            add(mModifierField);

            if (current.isSkillBased()) {
                row.add(new FlexSpacer(0, 0, true, false));
                row = new FlexRow();
                row.setInsets(new Insets(0, 20, 0, 0));
                DefaultFormatter formatter = new DefaultFormatter();
                formatter.setOverwriteMode(false);
                mSkillNameField = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, mDefault.getName(), null);
                add(mSkillNameField);
                row.add(mSkillNameField);
                String optionalSpecialization = I18n.Text("Optional Specialization");
                mSpecializationField = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, mDefault.getSpecialization(), optionalSpecialization);
                mSpecializationField.setHint(optionalSpecialization);
                add(mSpecializationField);
                row.add(mSpecializationField);
                row.add(mModifierField);
                grid.add(row, 1, 1);
            } else {
                row.add(mModifierField);
                row.add(new FlexSpacer(0, 0, true, false));
            }

            row = new FlexRow();
            row.setHorizontalAlignment(Alignment.RIGHT_BOTTOM);
            IconButton button = new IconButton(StdImage.REMOVE, I18n.Text("Remove this default"), () -> removeDefault());
            add(button);
            row.add(button);
            button = new IconButton(StdImage.ADD, I18n.Text("Add a default"), () -> addDefault());
            add(button);
            row.add(button);
            grid.add(row, 0, 2);
            grid.apply(this);
        } else {
            FlexRow row = new FlexRow();
            row.setHorizontalAlignment(Alignment.RIGHT_BOTTOM);
            row.add(new FlexSpacer(0, 0, true, false));
            IconButton button = new IconButton(StdImage.ADD, I18n.Text("Add a default"), () -> addDefault());
            add(button);
            row.add(button);
            row.apply(this);
        }

        revalidate();
        repaint();
    }

    /** @return The underlying skill default. */
    public SkillDefault getSkillDefault() {
        return mDefault;
    }

    private void addDefault() {
        SkillDefault skillDefault = new SkillDefault(LAST_ITEM_TYPE, LAST_ITEM_TYPE.isSkillBased() ? "" : null, null, 0); //$NON-NLS-1$
        JComponent   parent       = (JComponent) getParent();
        parent.add(new SkillDefaultEditor(skillDefault));
        if (mDefault == null) {
            parent.remove(this);
        }
        parent.revalidate();
        parent.repaint();
        notifyActionListeners();
    }

    private void removeDefault() {
        JComponent parent = (JComponent) getParent();
        parent.remove(this);
        if (parent.getComponentCount() == 0) {
            parent.add(new SkillDefaultEditor());
        }
        parent.revalidate();
        parent.repaint();
        notifyActionListeners();
    }

    @Override
    public void propertyChange(PropertyChangeEvent event) {
        Object src = event.getSource();
        if (src == mSkillNameField) {
            mDefault.setName((String) mSkillNameField.getValue());
            notifyActionListeners();
        } else if (src == mSpecializationField) {
            mDefault.setSpecialization((String) mSpecializationField.getValue());
            notifyActionListeners();
        } else if (src == mModifierField) {
            mDefault.setModifier(((Integer) mModifierField.getValue()).intValue());
            notifyActionListeners();
        } else {
            super.propertyChange(event);
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object src     = event.getSource();
        String command = event.getActionCommand();
        if (SkillDefault.TAG_TYPE.equals(command)) {
            SkillDefaultType current = mDefault.getType();
            SkillDefaultType value   = (SkillDefaultType) ((JComboBox<?>) src).getSelectedItem();
            if (!current.equals(value)) {
                Commitable.sendCommitToFocusOwner();
                mDefault.setType(value);
                rebuild();
                notifyActionListeners();
            }
            setLastItemType(value);
        } else {
            super.actionPerformed(event);
        }
    }
}
