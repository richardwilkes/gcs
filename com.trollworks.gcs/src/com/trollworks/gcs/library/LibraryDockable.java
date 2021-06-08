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

import com.trollworks.gcs.datafile.DataChangeListener;
import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.menu.RetargetableFocus;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.gcs.settings.Settings;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.FontAwesomeButton;
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
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Font;
import java.awt.KeyboardFocusManager;
import java.awt.dnd.DropTarget;
import javax.swing.DefaultListCellRenderer;
import javax.swing.JComboBox;
import javax.swing.JList;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** A list from a library. */
public abstract class LibraryDockable extends DataFileDockable implements RowFilter, DocumentListener, JumpToSearchTarget, RetargetableFocus, DataChangeListener, Runnable {
    private Toolbar           mToolbar;
    private JComboBox<Scales> mScaleCombo;
    private JTextField        mFilterField;
    private JComboBox<String> mCategoryCombo;
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
        mScaleCombo = new JComboBox<>(Scales.values());
        mScaleCombo.setSelectedItem(prefs.getInitialUIScale());
        mScaleCombo.addActionListener((event) -> {
            Scales scales = (Scales) mScaleCombo.getSelectedItem();
            if (scales == null) {
                scales = Scales.ACTUAL_SIZE;
            }
            Scale scale = scales.getScale();
            header.setScale(scale);
            content.setScale(scale);
        });
        mToolbar.add(mScaleCombo);
        createFilterField();
        createCategoryCombo();
        add(mToolbar, BorderLayout.NORTH);
        add(new ScrollPanel(header, content), BorderLayout.CENTER);
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
        mFilterField = new JTextField(10);
        mFilterField.getDocument().addDocumentListener(this);
        mFilterField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.text("Enter text here to narrow the list to only those rows containing matching items")));
        // This client property is specific to Mac OS X
        mFilterField.putClientProperty("JTextField.variant", "search");
        mFilterField.setMinimumSize(new Dimension(60, mFilterField.getPreferredSize().height));
        mToolbar.add(mFilterField, Toolbar.LAYOUT_FILL);
    }

    private void createCategoryCombo() {
        mCategoryCombo = new JComboBox<>();
        // Next two client properties are specific to Mac OS X
        mCategoryCombo.putClientProperty("JComponent.sizeVariant", "small");
        mCategoryCombo.putClientProperty("JComboBox.isPopDown", Boolean.TRUE);
        mCategoryCombo.setMaximumRowCount(20);
        mCategoryCombo.setRenderer(new CategoryCellRenderer());
        mCategoryCombo.addActionListener((event) -> {
            mCategoryCombo.setFont(mCategoryCombo.getFont().deriveFont(mCategoryCombo.getSelectedIndex() == 0 ? Font.ITALIC : Font.PLAIN));
            mCategoryCombo.revalidate();
            mCategoryCombo.repaint();
            if (mOutline != null) {
                mOutline.reapplyRowFilter();
            }
        });
        mToolbar.add(mCategoryCombo);
        adjustCategoryCombo();
    }

    private void adjustCategoryCombo() {
        mCategoryCombo.removeAllItems();
        mCategoryCombo.addItem(I18n.text("Any Category"));
        for (String category : getDataFile().getCategories()) {
            mCategoryCombo.addItem(category);
        }
        mCategoryCombo.setPreferredSize(null);
        mCategoryCombo.setFont(mCategoryCombo.getFont().deriveFont(Font.ITALIC));
        mCategoryCombo.setMinimumSize(new Dimension(36, mCategoryCombo.getPreferredSize().height));
        mCategoryCombo.revalidate();
        mCategoryCombo.repaint();
    }

    @Override
    public boolean isRowFiltered(Row row) {
        boolean filtered = false;
        if (row instanceof ListRow) {
            ListRow listRow = (ListRow) row;
            if (mCategoryCombo.getSelectedIndex() != 0) {
                String selectedItem = (String) mCategoryCombo.getSelectedItem();
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

    private static class CategoryCellRenderer extends DefaultListCellRenderer {
        @Override
        public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
            Component comp = super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
            setFont(getFont().deriveFont(index == 0 ? Font.ITALIC : Font.PLAIN));
            return comp;
        }
    }
}
