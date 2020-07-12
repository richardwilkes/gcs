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

package com.trollworks.gcs.library;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.datafile.DataFileDockable;
import com.trollworks.gcs.datafile.ListFile;
import com.trollworks.gcs.menu.RetargetableFocus;
import com.trollworks.gcs.menu.edit.JumpToSearchTarget;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.image.Images;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.IconButton;
import com.trollworks.gcs.ui.widget.Toolbar;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.ui.widget.outline.ListOutline;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowFilter;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.notification.BatchNotifierTarget;
import com.trollworks.gcs.utility.text.Text;

import java.awt.BorderLayout;
import java.awt.Component;
import java.awt.Dimension;
import java.awt.Font;
import java.awt.KeyboardFocusManager;
import java.awt.dnd.DropTarget;
import javax.swing.DefaultListCellRenderer;
import javax.swing.JComboBox;
import javax.swing.JList;
import javax.swing.JScrollPane;
import javax.swing.JTextField;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;

/** A list from a library. */
public abstract class LibraryDockable extends DataFileDockable implements RowFilter, DocumentListener, BatchNotifierTarget, JumpToSearchTarget, RetargetableFocus {
    private Toolbar           mToolbar;
    private JComboBox<Scales> mScaleCombo;
    private JTextField        mFilterField;
    private JComboBox<String> mCategoryCombo;
    private IconButton        mLockButton;
    private JScrollPane       mScroller;
    private ListOutline       mOutline;

    /** Creates a new {@link LibraryDockable}. */
    public LibraryDockable(ListFile file) {
        super(file);
        mOutline = createOutline();
        mOutline.setDynamicRowHeight(true);
        OutlineModel outlineModel = mOutline.getModel();
        outlineModel.applySortConfig(outlineModel.getSortConfig());
        outlineModel.setRowFilter(this);
        LibraryContent content = new LibraryContent(mOutline);
        LibraryHeader  header  = new LibraryHeader(mOutline.getHeaderPanel());
        Preferences    prefs   = Preferences.getInstance();
        mToolbar = new Toolbar();
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
        mLockButton = new IconButton(outlineModel.isLocked() ? Images.LOCKED : Images.UNLOCKED, I18n.Text("Switches between allowing editing and not"), () -> {
            OutlineModel model = mOutline.getModel();
            model.setLocked(!model.isLocked());
            mLockButton.setIcon(model.isLocked() ? Images.LOCKED : Images.UNLOCKED);
        });
        mToolbar.add(mLockButton);
        mToolbar.add(new IconButton(Images.TOGGLE_OPEN, I18n.Text("Opens/closes all hierarchical rows"), () -> mOutline.getModel().toggleRowOpenState()));
        mToolbar.add(new IconButton(Images.SIZE_TO_FIT, I18n.Text("Sets the width of each column to exactly fit its contents"), () -> mOutline.sizeColumnsToFit()));
        add(mToolbar, BorderLayout.NORTH);
        mScroller = new JScrollPane(content);
        mScroller.setBorder(null);
        mScroller.setColumnHeaderView(header);
        add(mScroller, BorderLayout.CENTER);
        prefs.getNotifier().add(this, Fonts.FONT_NOTIFICATION_KEY);
        setDropTarget(new DropTarget(mOutline, mOutline));
    }

    @Override
    public boolean attemptClose() {
        boolean closed = super.attemptClose();
        if (closed) {
            Preferences.getInstance().getNotifier().remove(this);
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

    /** @return The {@link JScrollPane}. */
    public JScrollPane getScroller() {
        return mScroller;
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
        mFilterField.setToolTipText(Text.wrapPlainTextForToolTip(I18n.Text("Enter text here to narrow the list to only those rows containing matching items")));
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
        mCategoryCombo.setRenderer(new DefaultListCellRenderer() {
            @Override
            public Component getListCellRendererComponent(JList<?> list, Object value, int index, boolean isSelected, boolean cellHasFocus) {
                Component comp = super.getListCellRendererComponent(list, value, index, isSelected, cellHasFocus);
                setFont(getFont().deriveFont(index == 0 ? Font.ITALIC : Font.PLAIN));
                return comp;
            }
        });
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
        mCategoryCombo.addItem(I18n.Text("Any Category"));
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
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public void enterBatchMode() {
        // Not needed.
    }

    @Override
    public void leaveBatchMode() {
        // Not needed.
    }

    @Override
    public void handleNotification(Object producer, String name, Object data) {
        if (Fonts.FONT_NOTIFICATION_KEY.equals(name)) {
            mOutline.updateRowHeights();
        } else if (Advantage.ID_TYPE.equals(name)) {
            getOutline().repaint();
        } else {
            adjustCategoryCombo();
        }
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
