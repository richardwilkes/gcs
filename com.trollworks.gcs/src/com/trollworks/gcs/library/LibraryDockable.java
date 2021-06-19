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

package com.trollworks.gcs.library;

import com.trollworks.gcs.character.FieldFactory;
import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.menu.RetargetableFocus;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.EditorField;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
import com.trollworks.gcs.ui.widget.PopupMenu;
import com.trollworks.gcs.ui.widget.ScrollPanel;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowFilter;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.text.Text;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.EventQueue;
import java.awt.KeyboardFocusManager;
import java.awt.dnd.DropTarget;
import java.util.ArrayList;
import java.util.List;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** A list from a library. */
public abstract class LibraryDockable extends DataFileDockable implements RowFilter, DocumentListener, JumpToSearchTarget, RetargetableFocus, DataChangeListener, Runnable {
    private Toolbar           mToolbar;
    private PopupMenu<Scales> mScalesPopup;
    private EditorField       mFilterField;
    private PopupMenu<String> mCategoryPopup;
    private FontAwesomeButton mLockButton;
    private ListOutline       mOutline;
    private boolean           mUpdatePending;

    /** Creates a new LibraryDockable. */
    protected LibraryDockable(ListFile file) {
        super(file);
        mOutline = createOutline();
        mOutline.setDynamicRowHeight(true);
        OutlineModel outlineModel = mOutline.getModel();
        outlineModel.applySortConfig(outlineModel.getSortConfig());
        outlineModel.setRowFilter(this);
        LibraryContent content = new LibraryContent(mOutline);
        LibraryHeader  header  = new LibraryHeader(mOutline.getHeaderPanel());
        Settings       prefs   = Settings.getInstance();
        mToolbar = new Toolbar();
        mLockButton = new FontAwesomeButton(outlineModel.isLocked() ? "\uf023" : "\uf13e", I18n.text("Switches between allowing editing and not"), () -> {
            OutlineModel model = mOutline.getModel();
            model.setLocked(!model.isLocked());
            mLockButton.setText(model.isLocked() ? "\uf023" : "\uf13e");
        });
        mToolbar.add(mLockButton);
        mToolbar.add(new FontAwesomeButton("\uf0e8", I18n.text("Opens/closes all hierarchical rows"), () -> mOutline.getModel().toggleRowOpenState()));
        mToolbar.add(new FontAwesomeButton("\uf337", I18n.text("Sets the width of each column to exactly fit its contents"), () -> mOutline.sizeColumnsToFit()));
        mScalesPopup = new PopupMenu<>(Scales.values(), (p) -> {
            Scales scales = mScalesPopup.getSelectedItem();
            if (scales == null) {
                scales = Scales.ACTUAL_SIZE;
            }
            Scale scale = scales.getScale();
            header.setScale(scale);
            content.setScale(scale);
        });
        mScalesPopup.setSelectedItem(prefs.getInitialUIScale(), false);
        mToolbar.add(mScalesPopup);
        createFilterField();
        createCategoryCombo();
        add(mToolbar, BorderLayout.NORTH);
        ScrollPanel scroller = new ScrollPanel(header, content);
        scroller.getViewport().setBackground(ThemeColor.PAGE_VOID);
        add(scroller, BorderLayout.CENTER);
        prefs.addChangeListener(this);
        getDataFile().addChangeListener(this);
        setDropTarget(new DropTarget(mOutline, mOutline));
    }

    @Override
    public boolean attemptClose() {
        boolean closed = super.attemptClose();
        if (closed) {
            Settings.getInstance().removeChangeListener(this);
            getDataFile().removeChangeListener(this);
        }
        return closed;
    }

    @Override
    public Component getRetargetedFocus() {
        return mOutline;
    }

    @Override
    public ListFile getDataFile() {
        return (ListFile) super.getDataFile();
    }

    @Override
    public PrintProxy getPrintProxy() {
        // Don't want to allow printing of the library files
        return null;
    }

    /** @return The {@link Toolbar}. */
    public Toolbar getToolbar() {
        return mToolbar;
    }

    /** @return The {@link ListOutline}. */
    public ListOutline getOutline() {
        return mOutline;
    }

    /**
     * Called to create the {@link ListOutline} for this {@link Dockable}.
     *
     * @return The newly created {@link ListOutline}.
     */
    protected abstract ListOutline createOutline();

    private void createFilterField() {
        mFilterField = new EditorField(FieldFactory.STRING, null, SwingConstants.LEFT, "",
                Text.makeFiller(10, 'M'), I18n.text("Enter text here to narrow the list to only those rows containing matching items"));
        mFilterField.setHint(I18n.text("Filter"));
        mFilterField.getDocument().addDocumentListener(this);
        mToolbar.add(mFilterField, Toolbar.LAYOUT_FILL);
    }

    private void createCategoryCombo() {
        mCategoryPopup = new PopupMenu<>(new ArrayList<>(), (p) -> {
            if (mOutline != null) {
                mOutline.reapplyRowFilter();
            }
        });
        adjustCategoryCombo();
        mToolbar.add(mCategoryPopup);
    }

    private void adjustCategoryCombo() {
        String last = mCategoryPopup.getSelectedItem();
        mCategoryPopup.clear();
        mCategoryPopup.addItem(I18n.text("Any Category"));
        List<String> categories = getDataFile().getCategories();
        String       selection  = null;
        if (!categories.isEmpty()) {
            mCategoryPopup.addSeparator();
            for (String category : categories) {
                mCategoryPopup.addItem(category);
                if (category.equals(last)) {
                    selection = category;
                }
            }
        }
        if (selection == null) {
            mCategoryPopup.setSelectedIndex(0, false);
        } else {
            mCategoryPopup.setSelectedItem(selection, false);
        }
        mCategoryPopup.revalidate();
        mCategoryPopup.repaint();
        if (mOutline != null) {
            mOutline.reapplyRowFilter();
        }
    }

    @Override
    public boolean isRowFiltered(Row row) {
        boolean filtered = false;
        if (row instanceof ListRow) {
            ListRow listRow = (ListRow) row;
            if (mCategoryPopup.getSelectedIndex() != 0) {
                String selectedItem = mCategoryPopup.getSelectedItem();
                if (selectedItem != null) {
                    filtered = !listRow.getCategories().contains(selectedItem);
                }
            }
            if (!filtered) {
                String filter = mFilterField.getText();
                if (!filter.isEmpty()) {
                    filtered = !listRow.contains(filter.toLowerCase(), true);
                }
            }
        }
        return filtered;
    }

    @Override
    public void dataWasChanged() {
        if (!mUpdatePending) {
            mUpdatePending = true;
            EventQueue.invokeLater(this);
        }
    }

    @Override
    public void run() {
        mOutline.updateRowHeights();
        adjustCategoryCombo();
        mUpdatePending = false;
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        documentChanged();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        documentChanged();
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        documentChanged();
    }

    private void documentChanged() {
        mOutline.reapplyRowFilter();
    }

    @Override
    public boolean isJumpToSearchAvailable() {
        return mFilterField.isEnabled() && mFilterField != KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
    }

    @Override
    public void jumpToSearchField() {
        mFilterField.requestFocus();
    }
}
