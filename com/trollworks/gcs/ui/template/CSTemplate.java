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
 * 2005-2007 the Initial Developer. All Rights Reserved.
 *
 * Contributor(s):
 *
 * ***** END LICENSE BLOCK ***** */

package com.trollworks.gcs.ui.template;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.CMTemplate;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.ui.advantage.CSAdvantageOutline;
import com.trollworks.gcs.ui.common.CSOutlineSyncer;
import com.trollworks.gcs.ui.equipment.CSEquipmentOutline;
import com.trollworks.gcs.ui.sheet.CSNotesPanel;
import com.trollworks.gcs.ui.sheet.CSTextEditor;
import com.trollworks.gcs.ui.skills.CSSkillOutline;
import com.trollworks.gcs.ui.spell.CSSpellOutline;
import com.trollworks.toolkit.notification.TKBatchNotifierTarget;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKCompoundBorder;
import com.trollworks.toolkit.widget.border.TKEmptyBorder;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutlineHeader;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKRowSelection;
import com.trollworks.toolkit.widget.scroll.TKScrollPanel;
import com.trollworks.toolkit.widget.scroll.TKScrollable;

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

/** The character sheet. */
public class CSTemplate extends TKPanel implements TKScrollable, TKBatchNotifierTarget, DropTargetListener, ActionListener {
	private static final TKEmptyBorder	NORMAL_BORDER	= new TKEmptyBorder(5);
	private CMTemplate					mTemplate;
	private boolean						mBatchMode;
	private CSAdvantageOutline			mAdvantageOutline;
	private CSSkillOutline				mSkillOutline;
	private CSSpellOutline				mSpellOutline;
	private CSEquipmentOutline			mEquipmentOutline;
	private CSNotesPanel				mNotesPanel;
	/** Used to determine whether an edit cell is pending. */
	protected boolean					mStartEditingPending;
	/** Used to determine whether a resize action is pending. */
	protected boolean					mSizePending;

	/**
	 * Creates a new template display.
	 * 
	 * @param template The template to display the data for.
	 */
	public CSTemplate(CMTemplate template) {
		super(new TKColumnLayout(1, 0, 5));
		setBorder(NORMAL_BORDER);
		mTemplate = template;

		// Make sure our primary outlines exist
		mAdvantageOutline = new CSAdvantageOutline(mTemplate);
		mSkillOutline = new CSSkillOutline(mTemplate);
		mSpellOutline = new CSSpellOutline(mTemplate);
		mEquipmentOutline = new CSEquipmentOutline(mTemplate, true);
		mNotesPanel = new CSNotesPanel(template.getNotes(), false);
		add(new CSTemplateOutlinePanel(mAdvantageOutline, Msgs.ADVANTAGES));
		add(new CSTemplateOutlinePanel(mSkillOutline, Msgs.SKILLS));
		add(new CSTemplateOutlinePanel(mSpellOutline, Msgs.SPELLS));
		add(new CSTemplateOutlinePanel(mEquipmentOutline, Msgs.EQUIPMENT));
		add(mNotesPanel);
		mAdvantageOutline.addActionListener(this);
		mSkillOutline.addActionListener(this);
		mSpellOutline.addActionListener(this);
		mEquipmentOutline.addActionListener(this);
		mNotesPanel.addActionListener(this);

		// Ensure everything is laid out and register for notification
		revalidate();
		mTemplate.addTarget(this, CMTemplate.TEMPLATE_PREFIX, CMCharacter.CHARACTER_PREFIX);

		setDropTarget(new DropTarget(this, this));
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (TKOutline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
			adjustSize();
		} else if (CSNotesPanel.CMD_EDIT_NOTES.equals(command)) {
			String notes = CSTextEditor.edit(Msgs.NOTES, mTemplate.getNotes());

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
	public static void prepOutline(TKOutline outline) {
		TKOutlineHeader header = outline.getHeaderPanel();

		outline.setDynamicRowHeight(true);
		outline.setAllowColumnDrag(false);
		outline.setAllowColumnResize(false);
		outline.setAllowColumnContextMenu(false);
		header.setIgnoreResizeOK(true);
		header.setBackground(Color.black);
	}

	/** @return The outline containing the Advantages, Disadvantages & Quirks. */
	public CSAdvantageOutline getAdvantageOutline() {
		return mAdvantageOutline;
	}

	/** @return The outline containing the skills. */
	public CSSkillOutline getSkillOutline() {
		return mSkillOutline;
	}

	/** @return The outline containing the spells. */
	public CSSpellOutline getSpellOutline() {
		return mSpellOutline;
	}

	/** @return The outline containing the carried equipment. */
	public CSEquipmentOutline getEquipmentOutline() {
		return mEquipmentOutline;
	}

	/**
	 * {@inheritDoc}
	 */
	public void enterBatchMode() {
		mBatchMode = true;
	}

	/**
	 * {@inheritDoc}
	 */
	public void leaveBatchMode() {
		mBatchMode = false;
		validate();
	}

	public void handleNotification(Object producer, String type, Object data) {
		if (type.startsWith(CMAdvantage.PREFIX)) {
			CSOutlineSyncer.add(mAdvantageOutline);
		} else if (type.startsWith(CMSkill.PREFIX)) {
			CSOutlineSyncer.add(mSkillOutline);
		} else if (type.startsWith(CMSpell.PREFIX)) {
			CSOutlineSyncer.add(mSpellOutline);
		} else if (type.startsWith(CMEquipment.PREFIX)) {
			CSOutlineSyncer.add(mEquipmentOutline);
		}

		if (!mBatchMode) {
			validate();
		}
	}

	public int getBlockScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		return (upLeftDirection ? -1 : 1) * (vertical ? visibleBounds.height : visibleBounds.width);
	}

	public Dimension getPreferredViewportSize() {
		return getPreferredSize();
	}

	public int getUnitScrollIncrement(Rectangle visibleBounds, boolean vertical, boolean upLeftDirection) {
		return upLeftDirection ? -10 : 10;
	}

	public boolean shouldTrackViewportHeight() {
		return TKScrollPanel.shouldTrackViewportHeight(this);
	}

	public boolean shouldTrackViewportWidth() {
		TKScrollPanel scroller = (TKScrollPanel) getAncestorOfType(TKScrollPanel.class);
		int width = getPreferredSize().width;
		int delta = getOutlineSizeDelta(mAdvantageOutline);
		int tmp = getOutlineSizeDelta(mSkillOutline);

		if (delta < tmp) {
			delta = tmp;
		}
		tmp = getOutlineSizeDelta(mSpellOutline);
		if (delta < tmp) {
			delta = tmp;
		}
		tmp = getOutlineSizeDelta(mEquipmentOutline);
		if (delta < tmp) {
			delta = tmp;
		}
		width += delta;

		return (scroller != null ? scroller.getContentViewSize(false) : width) > width;
	}

	private int getOutlineSizeDelta(TKOutline outline) {
		TKColumn column = outline.getModel().getColumnAtIndex(0);

		return column.getPreferredWidth(outline) - column.getWidth();
	}

	private boolean				mDragWasAcceptable;
	private ArrayList<TKRow>	mDragRows;

	public void dragEnter(DropTargetDragEvent dtde) {
		mDragWasAcceptable = false;

		try {
			if (dtde.isDataFlavorSupported(TKRowSelection.DATA_FLAVOR)) {
				TKRow[] rows = (TKRow[]) dtde.getTransferable().getTransferData(TKRowSelection.DATA_FLAVOR);

				if (rows != null && rows.length > 0) {
					mDragRows = new ArrayList<TKRow>(rows.length);

					for (TKRow element : rows) {
						if (element instanceof CMRow) {
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
			assert false : TKDebug.throwableToString(exception);
		}

		if (!mDragWasAcceptable) {
			dtde.rejectDrag();
		}
	}

	public void dragOver(DropTargetDragEvent dtde) {
		if (mDragWasAcceptable) {
			setBorder(new TKCompoundBorder(new TKLineBorder(TKColor.HIGHLIGHT, 2, TKLineBorder.ALL_EDGES, false), NORMAL_BORDER), true, false);
			dtde.acceptDrag(DnDConstants.ACTION_MOVE);
		} else {
			dtde.rejectDrag();
		}
	}

	public void dropActionChanged(DropTargetDragEvent dtde) {
		if (mDragWasAcceptable) {
			dtde.acceptDrag(DnDConstants.ACTION_MOVE);
		} else {
			dtde.rejectDrag();
		}
	}

	public void drop(DropTargetDropEvent dtde) {
		dtde.acceptDrop(dtde.getDropAction());
		((CSTemplateWindow) getBaseWindow()).addRows(mDragRows);
		mDragRows = null;
		setBorder(NORMAL_BORDER, true, false);
		dtde.dropComplete(true);
	}

	public void dragExit(DropTargetEvent dte) {
		mDragRows = null;
		setBorder(NORMAL_BORDER, true, false);
	}
}
