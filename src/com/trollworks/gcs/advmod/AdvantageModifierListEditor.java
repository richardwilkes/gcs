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

package com.trollworks.gcs.advmod;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.common.ListFile;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.RowEditor;
import com.trollworks.toolkit.collections.FilteredIterator;
import com.trollworks.toolkit.collections.FilteredList;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.widget.ActionPanel;
import com.trollworks.toolkit.ui.widget.IconButton;
import com.trollworks.toolkit.ui.widget.outline.Outline;
import com.trollworks.toolkit.ui.widget.outline.OutlineModel;
import com.trollworks.toolkit.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JScrollPane;
import javax.swing.ScrollPaneConstants;

/** Editor for {@link AdvantageModifierList}s. */
public class AdvantageModifierListEditor extends ActionPanel implements ActionListener {
    private DataFile mOwner;
    private Outline  mOutline;
    IconButton mAddButton;
    boolean    mModified;

    /**
     * @param advantage The {@link Advantage} to edit.
     * @return An instance of {@link AdvantageModifierListEditor}.
     */
    public static AdvantageModifierListEditor createEditor(Advantage advantage) {
        return new AdvantageModifierListEditor(advantage);
    }

    /**
     * Creates a new {@link AdvantageModifierListEditor} editor.
     *
     * @param owner             The owning row.
     * @param readOnlyModifiers The list of {@link AdvantageModifier}s from parents, which are not
     *                          to be modified.
     * @param modifiers         The list of {@link AdvantageModifier}s to modify.
     */
    public AdvantageModifierListEditor(DataFile owner, List<AdvantageModifier> readOnlyModifiers, List<AdvantageModifier> modifiers) {
        super(new BorderLayout());
        mOwner = owner;
        add(createOutline(readOnlyModifiers, modifiers), BorderLayout.CENTER);
        setName(toString());
    }

    /**
     * Creates a new {@link AdvantageModifierListEditor}.
     *
     * @param advantage Associated advantage
     */
    public AdvantageModifierListEditor(Advantage advantage) {
        this(advantage.getDataFile(), advantage.getParent() != null ? ((Advantage) advantage.getParent()).getAllModifiers() : null, advantage.getModifiers());
    }

    /** @return Whether a {@link AdvantageModifier} was modified. */
    public boolean wasModified() {
        return mModified;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (mOutline == source) {
            handleOutline(event.getActionCommand());
        }
    }

    private void handleOutline(String cmd) {
        if (Outline.CMD_OPEN_SELECTION.equals(cmd)) {
            openDetailEditor();
        }
    }

    private Component createOutline(List<AdvantageModifier> readOnlyModifiers, List<AdvantageModifier> modifiers) {
        JScrollPane  scroller;
        OutlineModel model;

        mAddButton = new IconButton(StdImage.ADD, I18n.Text("Add a modifier"), () -> addModifier());

        mOutline = new ModifierOutline();
        model = mOutline.getModel();
        AdvantageModifierColumnID.addColumns(mOutline, true);

        if (readOnlyModifiers != null) {
            for (AdvantageModifier modifier : readOnlyModifiers) {
                if (modifier.isEnabled()) {
                    AdvantageModifier romod = modifier.cloneModifier();
                    romod.setReadOnly(true);
                    model.addRow(romod);
                }
            }
        }
        for (AdvantageModifier modifier : modifiers) {
            model.addRow(modifier.cloneModifier());
        }
        mOutline.addActionListener(this);

        scroller = new JScrollPane(mOutline, ScrollPaneConstants.VERTICAL_SCROLLBAR_ALWAYS, ScrollPaneConstants.HORIZONTAL_SCROLLBAR_ALWAYS);
        scroller.setColumnHeaderView(mOutline.getHeaderPanel());
        scroller.setCorner(ScrollPaneConstants.UPPER_RIGHT_CORNER, mAddButton);
        return scroller;
    }

    private void openDetailEditor() {
        List<ListRow> rows = new ArrayList<>();
        for (AdvantageModifier row : new FilteredIterator<>(mOutline.getModel().getSelectionAsList(), AdvantageModifier.class)) {
            if (!row.isReadOnly()) {
                rows.add(row);
            }
        }
        if (!rows.isEmpty()) {
            mOutline.getModel().setLocked(!mAddButton.isEnabled());
            if (RowEditor.edit(mOutline, rows)) {
                mModified = true;
                for (ListRow row : rows) {
                    row.update();
                }
                mOutline.updateRowHeights(rows);
                mOutline.sizeColumnsToFit();
                notifyActionListeners();
            }
        }
    }

    private void addModifier() {
        AdvantageModifier modifier = new AdvantageModifier(mOwner);
        OutlineModel      model    = mOutline.getModel();

        if (mOwner instanceof ListFile) {
            modifier.setEnabled(false);
        }
        model.addRow(modifier);
        mOutline.sizeColumnsToFit();
        model.select(modifier, false);
        mOutline.revalidate();
        mOutline.scrollSelectionIntoView();
        mOutline.requestFocus();
        mModified = true;
        openDetailEditor();
    }

    /** @return Modifiers edited by this editor */
    public List<AdvantageModifier> getModifiers() {
        List<AdvantageModifier> modifiers = new ArrayList<>();
        for (AdvantageModifier modifier : new FilteredIterator<>(mOutline.getModel().getRows(), AdvantageModifier.class)) {
            if (!modifier.isReadOnly()) {
                modifiers.add(modifier);
            }
        }
        return modifiers;
    }

    /** @return Modifiers edited by this editor plus inherited Modifiers */
    public List<AdvantageModifier> getAllModifiers() {
        return new FilteredList<>(mOutline.getModel().getRows(), AdvantageModifier.class);
    }

    @Override
    public String toString() {
        return I18n.Text("Modifiers");
    }

    class ModifierOutline extends Outline {
        ModifierOutline() {
            super(false);
            setAllowColumnDrag(false);
            setAllowColumnResize(false);
            setAllowRowDrag(false);
        }

        @Override
        public boolean canDeleteSelection() {
            OutlineModel model = getModel();
            boolean      can   = mAddButton.isEnabled() && model.hasSelection();
            if (can) {
                for (AdvantageModifier row : new FilteredIterator<>(model.getSelectionAsList(), AdvantageModifier.class)) {
                    if (row.isReadOnly()) {
                        return false;
                    }
                }
            }
            return can;
        }

        @Override
        public void deleteSelection() {
            if (canDeleteSelection()) {
                getModel().removeSelection();
                sizeColumnsToFit();
                mModified = true;
                notifyActionListeners();
            }
        }
    }
}
