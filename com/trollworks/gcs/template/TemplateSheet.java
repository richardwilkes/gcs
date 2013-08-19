/* ***** BEGIN LICENSE BLOCK *****
 * Version: MPL 1.1
 *
 * The contents of this file are subject to the Mozilla Public License Version
 * 1.1 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 * http://www.mozilla.org/MPL/
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * The Original Code is GURPS Character Sheet.
 *
 * The Initial Developer of the Original Code is Richard A. Wilkes.
 * Portions created by the Initial Developer are Copyright (C) 1998-2002,
 * 2005-2009 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.template;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.character.NotesPanel;
import com.trollworks.gcs.character.TextEditor;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.ttk.layout.ColumnLayout;
import com.trollworks.ttk.notification.BatchNotifierTarget;
import com.trollworks.ttk.utility.Debug;
import com.trollworks.ttk.utility.LocalizedMessages;
import com.trollworks.ttk.widgets.outline.Outline;
import com.trollworks.ttk.widgets.outline.OutlineHeader;
import com.trollworks.ttk.widgets.outline.OutlineSyncer;
import com.trollworks.ttk.widgets.outline.Row;
import com.trollworks.ttk.widgets.outline.RowSelection;

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

import javax.swing.JPanel;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;
import javax.swing.border.EmptyBorder;

/** The template sheet. */
public class TemplateSheet extends JPanel implements Scrollable, BatchNotifierTarget, DropTargetListener, ActionListener {
	private static String				MSG_ADVANTAGES;
	private static String				MSG_SKILLS;
	private static String				MSG_SPELLS;
	private static String				MSG_EQUIPMENT;
	private static String				MSG_NOTES;
	private static final EmptyBorder	NORMAL_BORDER	= new EmptyBorder(5, 5, 5, 5);
	private Template					mTemplate;
	private boolean						mBatchMode;
	private AdvantageOutline			mAdvantageOutline;
	private SkillOutline				mSkillOutline;
	private SpellOutline				mSpellOutline;
	private EquipmentOutline			mEquipmentOutline;
	private NotesPanel					mNotesPanel;
	/** Used to determine whether an edit cell is pending. */
	protected boolean					mStartEditingPending;
	/** Used to determine whether a resize action is pending. */
	protected boolean					mSizePending;

	static {
		LocalizedMessages.initialize(TemplateSheet.class);
	}

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
		mTemplate = template;

		// Make sure our primary outlines exist
		mAdvantageOutline = new AdvantageOutline(mTemplate);
		mSkillOutline = new SkillOutline(mTemplate);
		mSpellOutline = new SpellOutline(mTemplate);
		mEquipmentOutline = new EquipmentOutline(mTemplate);
		mNotesPanel = new NotesPanel(template.getNotes(), false);
		add(new TemplateOutlinePanel(mAdvantageOutline, MSG_ADVANTAGES));
		add(new TemplateOutlinePanel(mSkillOutline, MSG_SKILLS));
		add(new TemplateOutlinePanel(mSpellOutline, MSG_SPELLS));
		add(new TemplateOutlinePanel(mEquipmentOutline, MSG_EQUIPMENT));
		add(mNotesPanel);
		mAdvantageOutline.addActionListener(this);
		mSkillOutline.addActionListener(this);
		mSpellOutline.addActionListener(this);
		mEquipmentOutline.addActionListener(this);
		mNotesPanel.addActionListener(this);

		// Ensure everything is laid out and register for notification
		revalidate();
		mTemplate.addTarget(this, Template.TEMPLATE_PREFIX, GURPSCharacter.CHARACTER_PREFIX);

		setDropTarget(new DropTarget(this, this));
	}

	@Override
	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
			adjustSize();
		} else if (NotesPanel.CMD_EDIT_NOTES.equals(command)) {
			String notes = TextEditor.edit(MSG_NOTES, mTemplate.getNotes());

			if (notes != null) {
				mTemplate.setNotes(notes);
				mNotesPanel.setNotes(notes);
			}
		}
	}

	private void adjustSize() {
		if (!mSizePending) {
			mSizePending = true;
			EventQueue.invokeLater(new Runnable() {
				@Override
				public void run() {
					mSizePending = false;
					setSize(getPreferredSize());
				}
			});
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

	/** @return The outline containing the carried equipment. */
	public EquipmentOutline getEquipmentOutline() {
		return mEquipmentOutline;
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

	private boolean			mDragWasAcceptable;
	private ArrayList<Row>	mDragRows;

	@Override
	public void dragEnter(DropTargetDragEvent dtde) {
		mDragWasAcceptable = false;

		try {
			if (dtde.isDataFlavorSupported(RowSelection.DATA_FLAVOR)) {
				Row[] rows = (Row[]) dtde.getTransferable().getTransferData(RowSelection.DATA_FLAVOR);

				if (rows != null && rows.length > 0) {
					mDragRows = new ArrayList<Row>(rows.length);

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
			assert false : Debug.throwableToString(exception);
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
		((TemplateWindow) getTopLevelAncestor()).addRows(mDragRows);
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
