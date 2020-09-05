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

package com.trollworks.gcs.character;

import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.notes.NoteOutline;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.ThemeColor;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.ScaleRoot;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineHeader;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowSelection;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.notification.BatchNotifierTarget;

import java.awt.Dimension;
import java.awt.Rectangle;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JPanel;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;

public abstract class CollectedOutlines extends JPanel implements ActionListener, ScaleRoot, Scrollable, BatchNotifierTarget, DropTargetListener {
    private Scale            mScale;
    private AdvantageOutline mAdvantageOutline;
    private SkillOutline     mSkillOutline;
    private SpellOutline     mSpellOutline;
    private EquipmentOutline mEquipmentOutline;
    private EquipmentOutline mOtherEquipmentOutline;
    private NoteOutline      mNoteOutline;
    private boolean          mDragWasAcceptable;
    private List<Row>        mDragRows;
    private boolean          mBatchMode;

    public CollectedOutlines() {
        mScale = Preferences.getInstance().getInitialUIScale().getScale();
    }

    protected void createOutlines(CollectedModels models) {
        if (mAdvantageOutline == null) {
            mAdvantageOutline = new AdvantageOutline(models);
            initOutline(mAdvantageOutline);
        }
        resetOutline(mAdvantageOutline);

        if (mSkillOutline == null) {
            mSkillOutline = new SkillOutline(models);
            initOutline(mSkillOutline);
        }
        resetOutline(mSkillOutline);

        if (mSpellOutline == null) {
            mSpellOutline = new SpellOutline(models);
            initOutline(mSpellOutline);
        }
        resetOutline(mSpellOutline);

        if (mEquipmentOutline == null) {
            mEquipmentOutline = new EquipmentOutline(models, models.getEquipmentModel());
            initOutline(mEquipmentOutline);
        }
        resetOutline(mEquipmentOutline);

        if (mOtherEquipmentOutline == null) {
            mOtherEquipmentOutline = new EquipmentOutline(models, models.getOtherEquipmentModel());
            initOutline(mOtherEquipmentOutline);
        }
        resetOutline(mOtherEquipmentOutline);

        if (mNoteOutline == null) {
            mNoteOutline = new NoteOutline(models);
            initOutline(mNoteOutline);
        }
        resetOutline(mNoteOutline);
    }

    protected void initOutline(Outline outline) {
        outline.addActionListener(this);
    }

    protected static void resetOutline(Outline outline) {
        outline.clearProxies();
        for (Column column : outline.getModel().getColumns()) {
            column.setWidth(outline, -1);
        }
    }

    /**
     * Prepares the specified outline for embedding in the sheet.
     *
     * @param outline The outline to prepare.
     */
    public static void prepOutline(Outline outline) {
        OutlineHeader header = outline.getHeaderPanel();
        outline.setDynamicRowHeight(true);
        outline.setAllowColumnResize(false);
        header.setIgnoreResizeOK(true);
        header.setBackground(ThemeColor.HEADER);
        header.setTopDividerColor(ThemeColor.HEADER);
    }

    @Override
    public Scale getScale() {
        return mScale;
    }

    @Override
    public void setScale(Scale scale) {
        if (mScale.getScale() != scale.getScale()) {
            mScale = scale;
            scaleChanged();
        }
    }

    protected abstract void scaleChanged();

    /** @return The outline containing the Advantages, Disadvantages & Quirks. */
    public AdvantageOutline getAdvantageOutline() {
        return mAdvantageOutline;
    }

    /** @return The outline containing the skills. */
    public SkillOutline getSkillOutline() {
        return mSkillOutline;
    }

    /** @return The outline containing the spells. */
    public SpellOutline getSpellOutline() {
        return mSpellOutline;
    }

    /** @return The outline containing the equipment. */
    public EquipmentOutline getEquipmentOutline() {
        return mEquipmentOutline;
    }

    /** @return The outline containing the other equipment. */
    public EquipmentOutline getOtherEquipmentOutline() {
        return mOtherEquipmentOutline;
    }

    /** @return The outline containing the notes. */
    public NoteOutline getNoteOutline() {
        return mNoteOutline;
    }

    /** Update the row heights of each outline. */
    public void updateRowHeights() {
        mAdvantageOutline.updateRowHeights();
        mSkillOutline.updateRowHeights();
        mSpellOutline.updateRowHeights();
        mEquipmentOutline.updateRowHeights();
        mOtherEquipmentOutline.updateRowHeights();
        mNoteOutline.updateRowHeights();
    }

    @Override
    public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
        return orientation == SwingConstants.VERTICAL ? visibleRect.height : visibleRect.width;
    }

    @Override
    public Dimension getPreferredScrollableViewportSize() {
        return getPreferredSize();
    }

    @Override
    public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
        return 10;
    }

    @Override
    public boolean getScrollableTracksViewportHeight() {
        return false;
    }

    @Override
    public boolean getScrollableTracksViewportWidth() {
        return false;
    }

    /** @return {@code true} if in batch mode. */
    public boolean inBatchMode() {
        return mBatchMode;
    }

    @Override
    public void enterBatchMode() {
        mBatchMode = true;
    }

    @Override
    public void leaveBatchMode() {
        mBatchMode = false;
        validate();
    }

    @Override
    public void dragEnter(DropTargetDragEvent dtde) {
        mDragWasAcceptable = false;
        try {
            if (dtde.isDataFlavorSupported(RowSelection.DATA_FLAVOR)) {
                Row[] rows = (Row[]) dtde.getTransferable().getTransferData(RowSelection.DATA_FLAVOR);
                if (rows.length > 0) {
                    mDragRows = new ArrayList<>(rows.length);
                    for (Row element : rows) {
                        if (element instanceof ListRow) {
                            mDragRows.add(element);
                        }
                    }
                    if (!mDragRows.isEmpty()) {
                        mDragWasAcceptable = true;
                        dtde.acceptDrag(DnDConstants.ACTION_MOVE);
                    }
                }
            }
        } catch (Exception exception) {
            Log.error(exception);
        }
        if (!mDragWasAcceptable) {
            dtde.rejectDrag();
        }
    }

    @Override
    public void dragOver(DropTargetDragEvent dtde) {
        if (mDragWasAcceptable) {
            dtde.acceptDrag(DnDConstants.ACTION_MOVE);
        } else {
            dtde.rejectDrag();
        }
    }

    @Override
    public void dropActionChanged(DropTargetDragEvent dtde) {
        if (mDragWasAcceptable) {
            dtde.acceptDrag(DnDConstants.ACTION_MOVE);
        } else {
            dtde.rejectDrag();
        }
    }

    @Override
    public void drop(DropTargetDropEvent dtde) {
        dtde.acceptDrop(dtde.getDropAction());
        UIUtilities.getAncestorOfType(this, CollectedOutlinesDockable.class).addRows(mDragRows);
        mDragRows = null;
        dtde.dropComplete(true);
    }

    @Override
    public void dragExit(DropTargetEvent dte) {
        mDragRows = null;
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }
}
