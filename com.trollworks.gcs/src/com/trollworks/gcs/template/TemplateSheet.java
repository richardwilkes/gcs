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

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.notes.NoteOutline;
import com.trollworks.gcs.preferences.Preferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.border.EmptyBorder;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.ScaleRoot;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineHeader;
import com.trollworks.gcs.ui.widget.outline.OutlineSyncer;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowSelection;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.notification.BatchNotifierTarget;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Rectangle;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.ArrayList;
import java.util.List;
import javax.swing.JPanel;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;

/** The template sheet. */
public class TemplateSheet extends JPanel implements Scrollable, BatchNotifierTarget, DropTargetListener, ActionListener, ScaleRoot {
    private static final EmptyBorder      NORMAL_BORDER = new EmptyBorder(5);
    private              Scale            mScale;
    private              boolean          mBatchMode;
    private              AdvantageOutline mAdvantageOutline;
    private              SkillOutline     mSkillOutline;
    private              SpellOutline     mSpellOutline;
    private              EquipmentOutline mEquipmentOutline;
    private              EquipmentOutline mOtherEquipmentOutline;
    private              NoteOutline      mNoteOutline;
    /** Used to determine whether an edit cell is pending. */
    protected            boolean          mStartEditingPending;
    /** Used to determine whether a resize action is pending. */
    protected            boolean          mSizePending;

    /**
     * Creates a new {@link TemplateSheet}.
     *
     * @param template The template to display the data for.
     */
    public TemplateSheet(Template template) {
        super(new ColumnLayout(1, 0, 5));
        setOpaque(true);
        setBackground(Color.WHITE);
        setBorder(NORMAL_BORDER);

        mScale = Preferences.getInstance().getInitialUIScale().getScale();

        // Make sure our primary outlines exist
        mAdvantageOutline = new AdvantageOutline(template);
        mSkillOutline = new SkillOutline(template);
        mSpellOutline = new SpellOutline(template);
        mEquipmentOutline = new EquipmentOutline(template, template.getEquipmentModel());
        mOtherEquipmentOutline = new EquipmentOutline(template, template.getOtherEquipmentModel());
        mNoteOutline = new NoteOutline(template);
        add(new TemplateOutlinePanel(mAdvantageOutline, I18n.Text("Advantages, Disadvantages & Quirks")));
        add(new TemplateOutlinePanel(mSkillOutline, I18n.Text("Skills")));
        add(new TemplateOutlinePanel(mSpellOutline, I18n.Text("Spells")));
        add(new TemplateOutlinePanel(mEquipmentOutline, I18n.Text("Equipment")));
        add(new TemplateOutlinePanel(mOtherEquipmentOutline, I18n.Text("Other Equipment")));
        add(new TemplateOutlinePanel(mNoteOutline, I18n.Text("Notes")));
        mAdvantageOutline.addActionListener(this);
        mSkillOutline.addActionListener(this);
        mSpellOutline.addActionListener(this);
        mEquipmentOutline.addActionListener(this);
        mOtherEquipmentOutline.addActionListener(this);
        mNoteOutline.addActionListener(this);

        // Ensure everything is laid out and register for notification
        revalidate();
        template.addTarget(this, Template.TEMPLATE_PREFIX, GURPSCharacter.CHARACTER_PREFIX);
        Preferences.getInstance().getNotifier().add(this, Preferences.KEY_SHOW_COLLEGE_IN_SHEET_SPELLS);

        setDropTarget(new DropTarget(this, this));

        runAdjustSize();
    }

    @Override
    public Scale getScale() {
        return mScale;
    }

    @Override
    public void setScale(Scale scale) {
        if (mScale.getScale() != scale.getScale()) {
            mScale = scale;
            revalidate();
            repaint();
        }
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
            adjustSize();
        }
    }

    private void adjustSize() {
        if (!mSizePending) {
            mSizePending = true;
            EventQueue.invokeLater(() -> runAdjustSize());
        }
    }

    void runAdjustSize() {
        Scale scale = Scale.get(this);
        mSizePending = false;
        Dimension size = getLayout().preferredLayoutSize(this);
        size.width = scale.scale(8 * 72);
        int min = scale.scale(4 * 72);
        if (size.height < min) {
            size.height = min;
        }
        UIUtilities.setOnlySize(this, size);
        setSize(size);
    }

    /**
     * Prepares the specified outline for embedding in the sheet.
     *
     * @param outline The outline to prepare.
     */
    public static void prepOutline(Outline outline) {
        OutlineHeader header = outline.getHeaderPanel();
        outline.setDynamicRowHeight(true);
        outline.setAllowColumnDrag(false);
        outline.setAllowColumnResize(false);
        outline.setAllowColumnContextMenu(false);
        header.setIgnoreResizeOK(true);
        header.setBackground(Color.black);
        header.setTopDividerColor(Color.black);
    }

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

    /** @return The outline containing the notes equipment. */
    public NoteOutline getNoteOutline() {
        return mNoteOutline;
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
    public void handleNotification(Object producer, String type, Object data) {
        if (type.startsWith(Advantage.PREFIX)) {
            OutlineSyncer.add(mAdvantageOutline);
        } else if (type.startsWith(Skill.PREFIX)) {
            OutlineSyncer.add(mSkillOutline);
        } else if (type.startsWith(Spell.PREFIX)) {
            OutlineSyncer.add(mSpellOutline);
        } else if (type.startsWith(Equipment.PREFIX)) {
            OutlineSyncer.add(mEquipmentOutline);
            OutlineSyncer.add(mOtherEquipmentOutline);
        } else if (type.startsWith(Note.PREFIX)) {
            OutlineSyncer.add(mNoteOutline);
        } else if (Preferences.KEY_SHOW_COLLEGE_IN_SHEET_SPELLS.equals(type)) {
            mSpellOutline.resetColumns();
            OutlineSyncer.add(mSpellOutline);
        }
        if (!mBatchMode) {
            validate();
        }
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

    private boolean   mDragWasAcceptable;
    private List<Row> mDragRows;

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
        UIUtilities.getAncestorOfType(this, TemplateDockable.class).addRows(mDragRows);
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
