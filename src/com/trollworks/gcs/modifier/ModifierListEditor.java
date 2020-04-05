/*
 * Copyright (c) 1998-2020 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public License, version 2.0.
 * If a copy of the MPL was not distributed with this file, You can obtain one at
 * http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as defined by the
 * Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.modifier;

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
import java.awt.event.KeyEvent;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JScrollPane;
import javax.swing.ScrollPaneConstants;

public abstract class ModifierListEditor extends ActionPanel implements ActionListener {
    private DataFile        mOwner;
    private ModifierOutline mOutline;
    IconButton mAddButton;
    boolean    mModified;

    protected ModifierListEditor(DataFile owner, List<? extends Modifier> readOnlyModifiers, List<? extends Modifier> modifiers) {
        super(new BorderLayout());
        mOwner = owner;
        add(createOutline(readOnlyModifiers, modifiers), BorderLayout.CENTER);
        setName(toString());
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
        }
    }

    protected abstract void addColumns(ModifierOutline outline);

    private Component createOutline(List<? extends Modifier> readOnlyModifiers, List<? extends Modifier> modifiers) {
        JScrollPane  scroller;
        OutlineModel model;

        mAddButton = new IconButton(StdImage.ADD, I18n.Text("Add a modifier"), () -> addModifier());

        mOutline = new ModifierOutline(this);
        model = mOutline.getModel();
        addColumns(mOutline);

        if (readOnlyModifiers != null) {
            for (Modifier modifier : readOnlyModifiers) {
                if (modifier.isEnabled()) {
                    Modifier romod = modifier.cloneModifier();
                    romod.setReadOnly(true);
                    model.addRow(romod);
                }
            }
        }
        for (Modifier modifier : modifiers) {
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
        return new FilteredList<>(mOutline.getModel().getRows(), Modifier.class);
    }

    @Override
    public String toString() {
        return I18n.Text("Modifiers");
    }

    class ModifierOutline extends Outline {
        private ModifierListEditor mEditor;

        ModifierOutline(ModifierListEditor editor) {
            super(false);
            setAllowColumnDrag(false);
            setAllowColumnResize(false);
            setAllowRowDrag(false);
            mEditor = editor;
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
                    mEditor.notifyActionListeners();
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
                mEditor.notifyActionListeners();
            }
        }
    }
}
