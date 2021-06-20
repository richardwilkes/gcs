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

package com.trollworks.gcs.skill;

import com.trollworks.gcs.attribute.AttributeChoice;
import com.trollworks.gcs.attribute.AttributeDef;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.layout.Alignment;
import com.trollworks.gcs.ui.layout.FlexGrid;
import com.trollworks.gcs.ui.layout.FlexRow;
import com.trollworks.gcs.ui.layout.FlexSpacer;
import com.trollworks.gcs.ui.widget.Commitable;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.EditorPanel;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.text.IntegerFormatter;

import java.awt.Insets;
import java.util.List;
import java.util.Map;
import javax.swing.JComponent;
import javax.swing.SwingConstants;
import javax.swing.text.DefaultFormatter;
import javax.swing.text.DefaultFormatterFactory;

/** A skill default editor panel. */
public class SkillDefaultEditor extends EditorPanel {
    private static String       LAST_ITEM_TYPE;
    private        DataFile     mDataFile;
    private        SkillDefault mDefault;
    private        EditorField  mSkillNameField;
    private        EditorField  mSpecializationField;
    private        EditorField  mModifierField;

    public static synchronized String getLastItemType(DataFile dataFile) {
        if (LAST_ITEM_TYPE == null) {
            Map<String, AttributeDef> defs    = ((dataFile != null) ? dataFile.getSheetSettings() : Settings.getInstance().getSheetSettings()).getAttributes();
            List<AttributeDef>        ordered = AttributeDef.getOrdered(defs);
            LAST_ITEM_TYPE = ordered.isEmpty() ? "st" : ordered.get(0).getID();
        }
        return LAST_ITEM_TYPE;
    }

    /** @param type The last item type created or switched to. */
    public static synchronized void setLastItemType(String type) {
        LAST_ITEM_TYPE = type;
    }

    /** Creates a new placeholder SkillDefaultEditor. */
    public SkillDefaultEditor(DataFile dataFile) {
        this(dataFile, null);
    }

    /**
     * Creates a new skill default editor panel.
     *
     * @param skillDefault The skill default to edit.
     */
    public SkillDefaultEditor(DataFile dataFile, SkillDefault skillDefault) {
        mDataFile = dataFile;
        mDefault = skillDefault;
        rebuild();
    }

    /** Rebuilds the contents of this panel with the current feature settings. */
    protected void rebuild() {
        removeAll();

        if (mDefault != null) {
            FlexGrid grid = new FlexGrid();

            OrLabel or = new OrLabel(this);
            add(or);
            grid.add(or, 0, 0);

            FlexRow row     = new FlexRow();
            String  current = mDefault.getType();
            row.add(SkillDefaultType.createPopup(this, mDataFile, current,
                    (p) -> {
                        AttributeChoice selectedItem = p.getSelectedItem();
                        if (selectedItem != null) {
                            String value = selectedItem.getAttribute();
                            if (!mDefault.getType().equals(value)) {
                                Commitable.sendCommitToFocusOwner();
                                mDefault.setType(value);
                                rebuild();
                                notifyActionListeners();
                            }
                            setLastItemType(value);
                        }
                    }, true));
            grid.add(row, 0, 1);

            mModifierField = new EditorField(new DefaultFormatterFactory(new IntegerFormatter(-99, 99, true)), this, SwingConstants.LEFT, Integer.valueOf(mDefault.getModifier()), Integer.valueOf(99), null);
            UIUtilities.setToPreferredSizeOnly(mModifierField);
            add(mModifierField);

            if ("skill".equalsIgnoreCase(current)) {
                row.add(new FlexSpacer(0, 0, true, false));
                row = new FlexRow();
                row.setInsets(new Insets(0, 20, 0, 0));
                DefaultFormatter formatter = new DefaultFormatter();
                formatter.setOverwriteMode(false);
                mSkillNameField = new EditorField(new DefaultFormatterFactory(formatter), this, SwingConstants.LEFT, mDefault.getName(), null);
                add(mSkillNameField);
                row.add(mSkillNameField);
                String optionalSpecialization = I18n.text("Optional Specialization");
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
            FontAwesomeButton button = new FontAwesomeButton("\uf1f8", I18n.text("Remove this default"), this::removeDefault);
            add(button);
            row.add(button);
            button = new FontAwesomeButton("\uf055", I18n.text("Add a default"), this::addDefault);
            add(button);
            row.add(button);
            grid.add(row, 0, 2);
            grid.apply(this);
        } else {
            FlexRow row = new FlexRow();
            row.setHorizontalAlignment(Alignment.RIGHT_BOTTOM);
            row.add(new FlexSpacer(0, 0, true, false));
            FontAwesomeButton button = new FontAwesomeButton("\uf055", I18n.text("Add a default"), this::addDefault);
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
        String       lastItemType = getLastItemType(mDataFile);
        SkillDefault skillDefault = new SkillDefault(lastItemType, SkillDefaultType.isSkillBased(lastItemType) ? "" : null, null, 0); //$NON-NLS-1$
        JComponent   parent       = (JComponent) getParent();
        parent.add(new SkillDefaultEditor(mDataFile, skillDefault));
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
            parent.add(new SkillDefaultEditor(mDataFile));
        }
        parent.revalidate();
        parent.repaint();
        notifyActionListeners();
    }

    @Override
    public void editorFieldChanged(EditorField field) {
        if (field == mSkillNameField) {
            mDefault.setName((String) mSkillNameField.getValue());
            notifyActionListeners();
        } else if (field == mSpecializationField) {
            mDefault.setSpecialization((String) mSpecializationField.getValue());
            notifyActionListeners();
        } else if (field == mModifierField) {
            mDefault.setModifier(((Integer) mModifierField.getValue()).intValue());
            notifyActionListeners();
        } else {
            super.editorFieldChanged(field);
        }
    }
}
