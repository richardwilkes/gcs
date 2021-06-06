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

package com.trollworks.gcs.modifier;

import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.border.TitledBorder;
import com.trollworks.gcs.ui.layout.PrecisionLayout;
import com.trollworks.gcs.ui.layout.PrecisionLayoutAlignment;
import com.trollworks.gcs.ui.layout.PrecisionLayoutData;
import com.trollworks.gcs.ui.widget.ActionPanel;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineHeader;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.RowEditor;
import com.trollworks.gcs.utility.Filtered;
import com.trollworks.gcs.utility.FilteredIterator;
import com.trollworks.gcs.utility.I18n;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.KeyEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JPanel;

public abstract class ModifierListEditor extends ActionPanel implements ActionListener {
    private DataFile        mOwner;
    private ModifierOutline mOutline;
    FontAwesomeButton mAddButton;
    private FontAwesomeButton mDeleteButton;
    boolean mModified;

    protected ModifierListEditor(DataFile owner, List<? extends Modifier> readOnlyModifiers, List<? extends Modifier> modifiers) {
        super(new PrecisionLayout().setMargins(0));
        mOwner = owner;

        mOutline = new ModifierOutline();
        OutlineModel model = mOutline.getModel();
        addColumns(mOutline);

        if (readOnlyModifiers != null) {
            for (Modifier modifier : readOnlyModifiers) {
                if (modifier.isEnabled()) {
                    Modifier romod = modifier.cloneModifier(false);
                    romod.setReadOnly(true);
                    model.addRow(romod);
                }
            }
        }
        for (Modifier modifier : modifiers) {
            model.addRow(modifier.cloneModifier(false));
        }
        mOutline.addActionListener(this);

        OutlineHeader header = mOutline.getHeaderPanel();
        header.setIgnoreResizeOK(true);
        JPanel panel = new JPanel(new BorderLayout());
        panel.add(header, BorderLayout.NORTH);
        panel.add(mOutline, BorderLayout.CENTER);

        mAddButton = new FontAwesomeButton("\uf055", I18n.text("Add a modifier"), this::addModifier);
        mDeleteButton = new FontAwesomeButton("\uf1f8", I18n.text("Remove the selected modifiers"), () -> mOutline.deleteSelection());
        mDeleteButton.setEnabled(false);

        JPanel right = new JPanel(new PrecisionLayout().setMargins(5));
        right.add(mAddButton);
        right.add(mDeleteButton);

        JPanel top = new JPanel(new PrecisionLayout().setMargins(0).setColumns(2).setHorizontalSpacing(1));
        top.add(panel, new PrecisionLayoutData().setFillAlignment().setGrabHorizontalSpace(true));
        top.add(right, new PrecisionLayoutData().setVerticalAlignment(PrecisionLayoutAlignment.BEGINNING));
        add(top, new PrecisionLayoutData().setFillHorizontalAlignment().setGrabHorizontalSpace(true));

        setBorder(new TitledBorder(Fonts.getDefaultSystemFont(), toString()));
    }

    /** @return Whether a {@link Modifier} was modified. */
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
        } else if (Outline.CMD_SELECTION_CHANGED.equals(cmd)) {
            mDeleteButton.setEnabled(mOutline.canDeleteSelection());
        }
    }

    protected abstract void addColumns(Outline outline);

    private void openDetailEditor() {
        List<ListRow> rows = new ArrayList<>();
        for (Modifier row : new FilteredIterator<>(mOutline.getModel().getSelectionAsList(), Modifier.class)) {
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

    protected abstract Modifier createModifier(DataFile owner);

    private void addModifier() {
        Modifier     modifier = createModifier(mOwner);
        OutlineModel model    = mOutline.getModel();
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
        notifyActionListeners();
        openDetailEditor();
    }

    /** @return Modifiers edited by this editor */
    public List<? extends Modifier> getModifiers() {
        List<Modifier> modifiers = new ArrayList<>();
        for (Modifier modifier : getAllModifiers()) {
            if (!modifier.isReadOnly()) {
                modifiers.add(modifier);
            }
        }
        return modifiers;
    }

    /** @return Modifiers edited by this editor plus inherited Modifiers */
    public List<? extends Modifier> getAllModifiers() {
        return Filtered.list(mOutline.getModel().getRows(), Modifier.class);
    }

    @Override
    public String toString() {
        return I18n.text("Modifiers");
    }

    class ModifierOutline extends Outline {
        ModifierOutline() {
            super(false);
            setAllowColumnResize(false);
            setAllowRowDrag(false);
            setBackground(Color.WHITE);
            setForeground(Color.BLACK);
        }

        @Override
        public void keyPressed(KeyEvent event) {
            super.keyPressed(event);
            if (!event.isConsumed() && (event.getModifiersEx() & getToolkit().getMenuShortcutKeyMaskEx()) == 0 && event.getKeyCode() == KeyEvent.VK_SPACE) {
                boolean      doNotify = false;
                OutlineModel model    = getModel();
                if (mAddButton.isEnabled() && model.hasSelection()) {
                    for (Modifier modifier : new FilteredIterator<>(model.getSelectionAsList(), Modifier.class)) {
                        if (!modifier.isReadOnly()) {
                            modifier.setEnabled(!modifier.isEnabled());
                            mModified = true;
                            doNotify = true;
                        }
                    }
                    repaintSelection();
                }
                event.consume();
                if (doNotify) {
                    ModifierListEditor.this.notifyActionListeners();
                }
            }
        }

        @Override
        public boolean canDeleteSelection() {
            OutlineModel model = getModel();
            boolean      can   = mAddButton.isEnabled() && model.hasSelection();
            if (can) {
                for (Modifier row : new FilteredIterator<>(model.getSelectionAsList(), Modifier.class)) {
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
                ModifierListEditor.this.notifyActionListeners();
            }
        }
    }
}
