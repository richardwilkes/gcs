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

package com.trollworks.gcs.ui.sheet;

import com.lowagie.text.pdf.DefaultFontMapper;
import com.lowagie.text.pdf.PdfContentByte;
import com.lowagie.text.pdf.PdfTemplate;
import com.lowagie.text.pdf.PdfWriter;

import com.trollworks.gcs.model.CMCharacter;
import com.trollworks.gcs.model.CMRow;
import com.trollworks.gcs.model.advantage.CMAdvantage;
import com.trollworks.gcs.model.equipment.CMEquipment;
import com.trollworks.gcs.model.skill.CMSkill;
import com.trollworks.gcs.model.skill.CMSkillDefault;
import com.trollworks.gcs.model.skill.CMSkillDefaultType;
import com.trollworks.gcs.model.spell.CMSpell;
import com.trollworks.gcs.model.weapon.CMMeleeWeaponStats;
import com.trollworks.gcs.model.weapon.CMRangedWeaponStats;
import com.trollworks.gcs.model.weapon.CMWeaponDisplayRow;
import com.trollworks.gcs.model.weapon.CMWeaponStats;
import com.trollworks.gcs.ui.advantage.CSAdvantageColumnID;
import com.trollworks.gcs.ui.advantage.CSAdvantageOutline;
import com.trollworks.gcs.ui.common.CSFont;
import com.trollworks.gcs.ui.common.CSOutlineSyncer;
import com.trollworks.gcs.ui.common.CSWindow;
import com.trollworks.gcs.ui.equipment.CSEquipmentColumnID;
import com.trollworks.gcs.ui.equipment.CSEquipmentOutline;
import com.trollworks.gcs.ui.preferences.CSSheetPreferences;
import com.trollworks.gcs.ui.skills.CSSkillColumnID;
import com.trollworks.gcs.ui.skills.CSSkillOutline;
import com.trollworks.gcs.ui.spell.CSSpellColumnID;
import com.trollworks.gcs.ui.spell.CSSpellOutline;
import com.trollworks.toolkit.collections.TKFilteredIterator;
import com.trollworks.toolkit.io.TKImage;
import com.trollworks.toolkit.io.TKPath;
import com.trollworks.toolkit.io.TKPreferences;
import com.trollworks.toolkit.io.xml.TKXMLWriter;
import com.trollworks.toolkit.notification.TKBatchNotifierTarget;
import com.trollworks.toolkit.print.TKPrintManager;
import com.trollworks.toolkit.utility.TKApp;
import com.trollworks.toolkit.utility.TKColor;
import com.trollworks.toolkit.utility.TKDebug;
import com.trollworks.toolkit.utility.TKFont;
import com.trollworks.toolkit.utility.TKGraphics;
import com.trollworks.toolkit.utility.TKNumberUtils;
import com.trollworks.toolkit.utility.units.TKLengthUnits;
import com.trollworks.toolkit.utility.units.TKWeightUnits;
import com.trollworks.toolkit.widget.TKPanel;
import com.trollworks.toolkit.widget.border.TKLineBorder;
import com.trollworks.toolkit.widget.layout.TKColumnLayout;
import com.trollworks.toolkit.widget.layout.TKRowDistribution;
import com.trollworks.toolkit.widget.outline.TKColumn;
import com.trollworks.toolkit.widget.outline.TKOutline;
import com.trollworks.toolkit.widget.outline.TKOutlineHeader;
import com.trollworks.toolkit.widget.outline.TKOutlineModel;
import com.trollworks.toolkit.widget.outline.TKRow;
import com.trollworks.toolkit.widget.outline.TKRowSelection;
import com.trollworks.toolkit.widget.scroll.TKScrollable;
import com.trollworks.toolkit.window.TKBaseWindow;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.GraphicsConfiguration;
import java.awt.Insets;
import java.awt.KeyboardFocusManager;
import java.awt.Rectangle;
import java.awt.Transparency;
import java.awt.dnd.DnDConstants;
import java.awt.dnd.DropTarget;
import java.awt.dnd.DropTargetDragEvent;
import java.awt.dnd.DropTargetDropEvent;
import java.awt.dnd.DropTargetEvent;
import java.awt.dnd.DropTargetListener;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.awt.print.PageFormat;
import java.awt.print.Printable;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.FileWriter;
import java.io.IOException;
import java.io.InputStreamReader;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;

/** The character sheet. */
public class CSSheet extends TKPanel implements TKScrollable, TKBatchNotifierTarget, CSPageOwner, Printable, ActionListener, Runnable, DropTargetListener {
	private static final String	BOXING_SKILL_NAME	= "Boxing";	//$NON-NLS-1$
	private static final String	KARATE_SKILL_NAME	= "Karate";	//$NON-NLS-1$
	private static final String	JUDO_SKILL_NAME		= "Judo";		//$NON-NLS-1$
	private static final String	BRAWLING_SKILL_NAME	= "Brawling";	//$NON-NLS-1$
	private CMCharacter			mCharacter;
	private int					mLastPage;
	private boolean				mBatchMode;
	private CSAdvantageOutline	mAdvantageOutline;
	private CSSkillOutline		mSkillOutline;
	private CSSpellOutline		mSpellOutline;
	private CSEquipmentOutline	mCarriedEquipmentOutline;
	private CSEquipmentOutline	mOtherEquipmentOutline;
	private TKOutline			mMeleeWeaponOutline;
	private TKOutline			mRangedWeaponOutline;
	private boolean				mRebuildPending;
	private HashSet<TKOutline>	mRootsToSync;
	private boolean				mSyncWeapons;
	private boolean				mIgnoreScroll;
	private boolean				mDisposed;

	/**
	 * Creates a new character sheet display. {@link #rebuild()} must be called prior to the first
	 * display of this panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CSSheet(CMCharacter character) {
		super(new TKColumnLayout(1, 0, 1));
		mCharacter = character;
		mLastPage = -1;
		mRootsToSync = new HashSet<TKOutline>();
		if (!TKGraphics.inHeadlessPrintMode()) {
			setDropTarget(new DropTarget(this, this));
		}
		TKPreferences.getInstance().getNotifier().add(this, CSSheetPreferences.OPTIONAL_DICE_RULES_PREF_KEY);
	}

	/** Call when the sheet is no longer in use. */
	public void dispose() {
		TKPreferences.getInstance().getNotifier().remove(this);
		mCharacter.resetNotifier();
		mDisposed = true;
	}

	/** @return Whether the sheet has had {@link #dispose()} called on it. */
	public boolean hasBeenDisposed() {
		return mDisposed;
	}

	/** @return Whether a rebuild is pending. */
	public boolean isRebuildPending() {
		return mRebuildPending;
	}

	/** Synchronizes the display with the underlying model. */
	public void rebuild() {
		Component focus = KeyboardFocusManager.getCurrentKeyboardFocusManager().getPermanentFocusOwner();
		String focusKey = null;
		TKBaseWindow window = getBaseWindow();
		CSPageAssembler pageAssembler;

		if (focus instanceof CSField) {
			focusKey = ((CSField) focus).getConsumedType();
			focus = null;
		}
		if (window instanceof CSWindow) {
			((CSWindow) window).jumpToFindField();
		}

		// Make sure our primary outlines exist
		createAdvantageOutline();
		createSkillOutline();
		createSpellOutline();
		createMeleeWeaponOutline();
		createRangedWeaponOutline();
		createCarriedEquipmentOutline();
		createOtherEquipmentOutline();

		// Clear out the old pages
		mIgnoreScroll = true;
		removeAll();
		mCharacter.resetNotifier(CSPrerequisitesThread.getThread(mCharacter));

		// Create the first page, which holds stuff that has a fixed vertical size.
		pageAssembler = new CSPageAssembler(this);
		pageAssembler.addToContent(hwrap(new CSPortraitPanel(mCharacter), vwrap(hwrap(new CSIdentityPanel(mCharacter), new CSPlayerInfoPanel(mCharacter)), new CSDescriptionPanel(mCharacter), TKRowDistribution.GIVE_EXCESS_TO_LAST), new CSPointsPanel(mCharacter)), null, null);
		pageAssembler.addToContent(hwrap(new CSAttributesPanel(mCharacter), vwrap(new CSEncumbrancePanel(mCharacter), new CSLiftPanel(mCharacter)), new CSHitLocationPanel(mCharacter), new CSHitPointsPanel(mCharacter)), null, null);

		// Add our outlines
		if (mAdvantageOutline.getModel().getRowCount() > 0 && mSkillOutline.getModel().getRowCount() > 0) {
			addOutline(pageAssembler, mAdvantageOutline, Msgs.ADVANTAGES, mSkillOutline, Msgs.SKILLS);
		} else {
			addOutline(pageAssembler, mAdvantageOutline, Msgs.ADVANTAGES);
			addOutline(pageAssembler, mSkillOutline, Msgs.SKILLS);
		}
		addOutline(pageAssembler, mSpellOutline, Msgs.SPELLS);
		addOutline(pageAssembler, mMeleeWeaponOutline, Msgs.MELEE_WEAPONS);
		addOutline(pageAssembler, mRangedWeaponOutline, Msgs.RANGED_WEAPONS);
		addOutline(pageAssembler, mCarriedEquipmentOutline, Msgs.CARRIED_EQUIPMENT);
		addOutline(pageAssembler, mOtherEquipmentOutline, Msgs.OTHER_EQUIPMENT);

		pageAssembler.addNotes();

		// Ensure everything is laid out and register for notification
		revalidate();
		CSOutlineSyncer.remove(mAdvantageOutline);
		CSOutlineSyncer.remove(mSkillOutline);
		CSOutlineSyncer.remove(mSpellOutline);
		CSOutlineSyncer.remove(mCarriedEquipmentOutline);
		CSOutlineSyncer.remove(mOtherEquipmentOutline);
		mCharacter.addTarget(this, CMCharacter.CHARACTER_PREFIX);
		restoreFocusToKey(focusKey, this);
		mIgnoreScroll = false;
	}

	private boolean restoreFocusToKey(String key, Component panel) {
		if (key != null) {
			if (panel instanceof CSField) {
				if (key.equals(((CSField) panel).getConsumedType())) {
					panel.requestFocusInWindow();
					return true;
				}
			} else if (panel instanceof Container) {
				Container container = (Container) panel;

				if (container.getComponentCount() > 0) {
					for (Component child : container.getComponents()) {
						if (restoreFocusToKey(key, child)) {
							return true;
						}
					}
				}
			}
		}
		return false;
	}

	@Override public void scrollRectIntoView(Rectangle bounds) {
		if (!mIgnoreScroll) {
			super.scrollRectIntoView(bounds);
		}
	}

	private void addOutline(CSPageAssembler pageAssembler, TKOutline outline, String title) {
		if (outline.getModel().getRowCount() > 0) {
			CSOutlineInfo info = new CSOutlineInfo(outline);
			boolean useProxy = false;

			while (pageAssembler.addToContent(new CSSingleOutlinePanel(outline, title, useProxy), info, null)) {
				if (!useProxy) {
					title = MessageFormat.format(Msgs.CONTINUED, title);
					useProxy = true;
				}
			}
		}
	}

	private void addOutline(CSPageAssembler pageAssembler, TKOutline leftOutline, String leftTitle, TKOutline rightOutline, String rightTitle) {
		CSOutlineInfo infoLeft = new CSOutlineInfo(leftOutline);
		CSOutlineInfo infoRight = new CSOutlineInfo(rightOutline);
		boolean useProxy = false;

		while (pageAssembler.addToContent(new CSDoubleOutlinePanel(leftOutline, leftTitle, rightOutline, rightTitle, useProxy), infoLeft, infoRight)) {
			if (!useProxy) {
				leftTitle = MessageFormat.format(Msgs.CONTINUED, leftTitle);
				rightTitle = MessageFormat.format(Msgs.CONTINUED, rightTitle);
				useProxy = true;
			}
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

	private void createAdvantageOutline() {
		if (mAdvantageOutline == null) {
			mAdvantageOutline = new CSAdvantageOutline(mCharacter);
			initOutline(mAdvantageOutline);
		} else {
			resetOutline(mAdvantageOutline);
		}
	}

	/** @return The outline containing the skills. */
	public CSSkillOutline getSkillOutline() {
		return mSkillOutline;
	}

	private void createSkillOutline() {
		if (mSkillOutline == null) {
			mSkillOutline = new CSSkillOutline(mCharacter);
			initOutline(mSkillOutline);
		} else {
			resetOutline(mSkillOutline);
		}
	}

	/** @return The outline containing the spells. */
	public CSSpellOutline getSpellOutline() {
		return mSpellOutline;
	}

	private void createSpellOutline() {
		if (mSpellOutline == null) {
			mSpellOutline = new CSSpellOutline(mCharacter);
			initOutline(mSpellOutline);
		} else {
			resetOutline(mSpellOutline);
		}
	}

	/** @return The outline containing the carried equipment. */
	public CSEquipmentOutline getCarriedEquipmentOutline() {
		return mCarriedEquipmentOutline;
	}

	private void createCarriedEquipmentOutline() {
		if (mCarriedEquipmentOutline == null) {
			mCarriedEquipmentOutline = new CSEquipmentOutline(mCharacter, true);
			initOutline(mCarriedEquipmentOutline);
		} else {
			resetOutline(mCarriedEquipmentOutline);
		}
	}

	/** @return The outline containing the other equipment. */
	public CSEquipmentOutline getOtherEquipmentOutline() {
		return mOtherEquipmentOutline;
	}

	private void createOtherEquipmentOutline() {
		if (mOtherEquipmentOutline == null) {
			mOtherEquipmentOutline = new CSEquipmentOutline(mCharacter, false);
			initOutline(mOtherEquipmentOutline);
		} else {
			resetOutline(mOtherEquipmentOutline);
		}
	}

	/** @return The outline containing the melee weapons. */
	public TKOutline getMeleeWeaponOutline() {
		return mMeleeWeaponOutline;
	}

	private void createMeleeWeaponOutline() {
		if (mMeleeWeaponOutline == null) {
			TKOutlineModel outlineModel;
			String sortConfig;

			mMeleeWeaponOutline = new CSWeaponOutline(CMMeleeWeaponStats.class);
			outlineModel = mMeleeWeaponOutline.getModel();
			sortConfig = outlineModel.getSortConfig();
			for (CMWeaponDisplayRow row : collectWeapons(CMMeleeWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);
			initOutline(mMeleeWeaponOutline);
		} else {
			resetOutline(mMeleeWeaponOutline);
		}
	}

	/** @return The outline containing the ranged weapons. */
	public TKOutline getRangedWeaponOutline() {
		return mRangedWeaponOutline;
	}

	private void createRangedWeaponOutline() {
		if (mRangedWeaponOutline == null) {
			TKOutlineModel outlineModel;
			String sortConfig;

			mRangedWeaponOutline = new CSWeaponOutline(CMRangedWeaponStats.class);
			outlineModel = mRangedWeaponOutline.getModel();
			sortConfig = outlineModel.getSortConfig();
			for (CMWeaponDisplayRow row : collectWeapons(CMRangedWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);
			initOutline(mRangedWeaponOutline);
		} else {
			resetOutline(mRangedWeaponOutline);
		}
	}

	private void addBuiltInWeapons(Class<? extends CMWeaponStats> weaponClass, HashMap<CSHashedWeapon, CMWeaponDisplayRow> map) {
		if (weaponClass == CMMeleeWeaponStats.class) {
			boolean savedModified = mCharacter.isModified();
			ArrayList<CMSkillDefault> defaults = new ArrayList<CMSkillDefault>();
			CMAdvantage phantom;
			CMMeleeWeaponStats weapon;

			mCharacter.suspendUndo(true);
			phantom = new CMAdvantage(mCharacter, false);
			phantom.setName(Msgs.NATURAL);

			if (mCharacter.includePunch()) {
				defaults.add(new CMSkillDefault(CMSkillDefaultType.DX, null, null, 0));
				defaults.add(new CMSkillDefault(CMSkillDefaultType.Skill, BOXING_SKILL_NAME, null, 0));
				defaults.add(new CMSkillDefault(CMSkillDefaultType.Skill, BRAWLING_SKILL_NAME, null, 0));
				defaults.add(new CMSkillDefault(CMSkillDefaultType.Skill, KARATE_SKILL_NAME, null, 0));
				defaults.add(new CMSkillDefault(CMSkillDefaultType.Skill, JUDO_SKILL_NAME, null, 0));
				weapon = new CMMeleeWeaponStats(phantom);
				weapon.setUsage(Msgs.PUNCH);
				weapon.setDefaults(defaults);
				weapon.setDamage("thr-1 cr"); //$NON-NLS-1$
				weapon.setReach("C"); //$NON-NLS-1$
				weapon.setParry("0"); //$NON-NLS-1$
				map.put(new CSHashedWeapon(weapon), new CMWeaponDisplayRow(weapon));
				defaults.clear();
			}

			defaults.add(new CMSkillDefault(CMSkillDefaultType.DX, null, null, -2));
			defaults.add(new CMSkillDefault(CMSkillDefaultType.Skill, BRAWLING_SKILL_NAME, null, -2));
			defaults.add(new CMSkillDefault(CMSkillDefaultType.Skill, KARATE_SKILL_NAME, null, -2));

			if (mCharacter.includeKick()) {
				weapon = new CMMeleeWeaponStats(phantom);
				weapon.setUsage(Msgs.KICK);
				weapon.setDefaults(defaults);
				weapon.setDamage("thr cr"); //$NON-NLS-1$
				weapon.setReach("C,1"); //$NON-NLS-1$
				weapon.setParry("No"); //$NON-NLS-1$
				map.put(new CSHashedWeapon(weapon), new CMWeaponDisplayRow(weapon));
			}

			if (mCharacter.includeKickBoots()) {
				weapon = new CMMeleeWeaponStats(phantom);
				weapon.setUsage(Msgs.BOOTS);
				weapon.setDefaults(defaults);
				weapon.setDamage("thr+1 cr"); //$NON-NLS-1$
				weapon.setReach("C,1"); //$NON-NLS-1$
				weapon.setParry("No"); //$NON-NLS-1$
				map.put(new CSHashedWeapon(weapon), new CMWeaponDisplayRow(weapon));
			}

			mCharacter.suspendUndo(false);
			mCharacter.setModified(savedModified);
		}
	}

	private ArrayList<CMWeaponDisplayRow> collectWeapons(Class<? extends CMWeaponStats> weaponClass) {
		HashMap<CSHashedWeapon, CMWeaponDisplayRow> weaponMap = new HashMap<CSHashedWeapon, CMWeaponDisplayRow>();
		ArrayList<CMWeaponDisplayRow> weaponList;

		addBuiltInWeapons(weaponClass, weaponMap);

		for (CMAdvantage advantage : mCharacter.getAdvantagesIterator()) {
			for (CMWeaponStats weapon : advantage.getWeapons()) {
				if (weaponClass.isInstance(weapon)) {
					weaponMap.put(new CSHashedWeapon(weapon), new CMWeaponDisplayRow(weapon));
				}
			}
		}

		for (CMEquipment equipment : mCharacter.getCarriedEquipmentIterator()) {
			if (equipment.getQuantity() > 0) {
				for (CMWeaponStats weapon : equipment.getWeapons()) {
					if (weaponClass.isInstance(weapon)) {
						weaponMap.put(new CSHashedWeapon(weapon), new CMWeaponDisplayRow(weapon));
					}
				}
			}
		}

		for (CMSpell spell : mCharacter.getSpellsIterator()) {
			for (CMWeaponStats weapon : spell.getWeapons()) {
				if (weaponClass.isInstance(weapon)) {
					weaponMap.put(new CSHashedWeapon(weapon), new CMWeaponDisplayRow(weapon));
				}
			}
		}

		for (CMSkill skill : mCharacter.getSkillsIterator()) {
			for (CMWeaponStats weapon : skill.getWeapons()) {
				if (weaponClass.isInstance(weapon)) {
					weaponMap.put(new CSHashedWeapon(weapon), new CMWeaponDisplayRow(weapon));
				}
			}
		}

		weaponList = new ArrayList<CMWeaponDisplayRow>(weaponMap.values());
		return weaponList;
	}

	private void initOutline(TKOutline outline) {
		outline.addActionListener(this);
	}

	private void resetOutline(TKOutline outline) {
		outline.clearProxies();
	}

	private TKPanel hwrap(TKPanel left, TKPanel right) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(2, 2, 2));

		wrapper.add(left);
		wrapper.add(right);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private TKPanel hwrap(TKPanel left, TKPanel center, TKPanel right) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(3, 2, 2));

		wrapper.add(left);
		wrapper.add(center);
		wrapper.add(right);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private TKPanel hwrap(TKPanel left, TKPanel center, TKPanel center2, TKPanel right) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(4, 2, 2));

		wrapper.add(left);
		wrapper.add(center);
		wrapper.add(center2);
		wrapper.add(right);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private TKPanel vwrap(TKPanel top, TKPanel bottom) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(1, 2, 2));

		wrapper.add(top);
		wrapper.add(bottom);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private TKPanel vwrap(TKPanel top, TKPanel bottom, TKRowDistribution distribution) {
		TKPanel wrapper = new TKPanel(new TKColumnLayout(1, 2, 2, distribution));

		wrapper.add(top);
		wrapper.add(bottom);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	/** @return The number of pages in this character sheet. */
	public int getPageCount() {
		return getComponentCount();
	}

	public int print(Graphics graphics, PageFormat pageFormat, int pageIndex) {
		if (pageIndex >= getComponentCount()) {
			mLastPage = -1;
			return NO_SUCH_PAGE;
		}

		// We do the following trick to avoid going through the work twice,
		// as we are called twice for each page, the first of which doesn't
		// seem to be used.
		if (mLastPage != pageIndex) {
			mLastPage = pageIndex;
		} else {
			getComponent(pageIndex).print(graphics);
		}
		return PAGE_EXISTS;
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
		if (CSSheetPreferences.OPTIONAL_DICE_RULES_PREF_KEY.equals(type)) {
			markForRebuild();
		} else {
			if (type.startsWith(CMAdvantage.PREFIX)) {
				CSOutlineSyncer.add(mAdvantageOutline);
			} else if (type.startsWith(CMSkill.PREFIX)) {
				CSOutlineSyncer.add(mSkillOutline);
			} else if (type.startsWith(CMSpell.PREFIX)) {
				CSOutlineSyncer.add(mSpellOutline);
			} else if (type.startsWith(CMEquipment.PREFIX)) {
				CSOutlineSyncer.add(mCarriedEquipmentOutline);
				CSOutlineSyncer.add(mOtherEquipmentOutline);
			}

			if (CMCharacter.ID_LAST_MODIFIED.equals(type)) {
				int count = getComponentCount();

				for (int i = 0; i < count; i++) {
					CSPage page = (CSPage) getComponent(i);
					Rectangle bounds = page.getBounds();
					Insets insets = page.getInsets();

					bounds.y = bounds.y + bounds.height - insets.bottom;
					bounds.height = insets.bottom;
					repaint(bounds);
				}
			} else if (CMEquipment.ID_QUANTITY.equals(type) || CMEquipment.ID_WEAPON_STATUS_CHANGED.equals(type) || CMAdvantage.ID_WEAPON_STATUS_CHANGED.equals(type) || CMSpell.ID_WEAPON_STATUS_CHANGED.equals(type) || CMSkill.ID_WEAPON_STATUS_CHANGED.equals(type) || CMCharacter.ID_INCLUDE_PUNCH.equals(type) || CMCharacter.ID_INCLUDE_KICK.equals(type) || CMCharacter.ID_INCLUDE_BOOTS.equals(type)) {
				mSyncWeapons = true;
				markForRebuild();
			} else if (CMCharacter.ID_PARRY_BONUS.equals(type) || CMSkill.ID_LEVEL.equals(type)) {
				CSOutlineSyncer.add(mMeleeWeaponOutline);
				CSOutlineSyncer.add(mRangedWeaponOutline);
			} else if (CMCharacter.ID_CARRIED_WEIGHT.equals(type) || CMCharacter.ID_CARRIED_WEALTH.equals(type)) {
				TKColumn column = mCarriedEquipmentOutline.getModel().getColumnWithID(CSEquipmentColumnID.DESCRIPTION.ordinal());

				column.setName(CSEquipmentColumnID.DESCRIPTION.toString(mCharacter, true));
			} else if (!mBatchMode) {
				validate();
			}
		}
	}

	public void drawPageAdornments(CSPage page, Graphics2D g2d) {
		Rectangle bounds = page.getLocalInsetBounds();
		int pageNumber = 1 + getIndexOf(page);
		String pageString = MessageFormat.format(Msgs.PAGE_NUMBER, TKNumberUtils.format(pageNumber), TKNumberUtils.format(getPageCount()));
		String copyright1 = TKApp.getCopyrightBanner(true);
		String copyright2 = copyright1.substring(copyright1.indexOf('\n') + 1);
		Font font1 = TKFont.lookup(CSFont.KEY_SECONDARY_FOOTER);
		Font font2 = TKFont.lookup(CSFont.KEY_PRIMARY_FOOTER);
		FontMetrics fm1 = g2d.getFontMetrics(font1);
		FontMetrics fm2 = g2d.getFontMetrics(font2);
		int y;
		String left;
		String right;
		String center;

		copyright1 = copyright1.substring(0, copyright1.indexOf('\n'));
		y = bounds.y + bounds.height + fm2.getAscent();

		if (pageNumber % 2 == 1) {
			left = copyright1;
			right = mCharacter.getLastModified();
		} else {
			left = mCharacter.getLastModified();
			right = copyright1;
		}

		g2d.setFont(font1);
		g2d.drawString(left, bounds.x, y);
		g2d.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, g2d).getWidth(), y);
		g2d.setFont(font2);
		center = mCharacter.getName();
		g2d.drawString(center, bounds.x + (bounds.width - (int) fm2.getStringBounds(center, g2d).getWidth()) / 2, y);

		if (pageNumber % 2 == 1) {
			left = copyright2;
			right = pageString;
		} else {
			left = pageString;
			right = copyright2;
		}

		y += fm2.getDescent() + fm1.getAscent();

		g2d.setFont(font1);
		g2d.drawString(left, bounds.x, y);
		g2d.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, g2d).getWidth(), y);
		g2d.drawString(Msgs.ADVERTISEMENT, bounds.x + (bounds.width - (int) fm1.getStringBounds(Msgs.ADVERTISEMENT, g2d).getWidth()) / 2, y);
	}

	public Insets getPageAdornmentsInsets(CSPage page) {
		FontMetrics fm1 = TKFont.getFontMetrics(TKFont.lookup(CSFont.KEY_SECONDARY_FOOTER));
		FontMetrics fm2 = TKFont.getFontMetrics(TKFont.lookup(CSFont.KEY_PRIMARY_FOOTER));

		return new Insets(0, 0, fm1.getAscent() + fm1.getDescent() + fm2.getAscent() + fm2.getDescent(), 0);
	}

	public TKPrintManager getPageSettings() {
		return mCharacter.getPageSettings();
	}

	/** @return The character being displayed. */
	public CMCharacter getCharacter() {
		return mCharacter;
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();

		if (TKOutline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
			mRootsToSync.add(((TKOutline) event.getSource()).getRealOutline());
			markForRebuild();
		} else if (CSNotesPanel.CMD_EDIT_NOTES.equals(command)) {
			String notes = CSTextEditor.edit(Msgs.NOTES, mCharacter.getNotes());

			if (notes != null) {
				mCharacter.setNotes(notes);
				rebuild();
			}
		}
	}

	/** Marks the sheet for a rebuild in the near future. */
	public void markForRebuild() {
		if (!mRebuildPending) {
			mRebuildPending = true;
			EventQueue.invokeLater(this);
		}
	}

	public void run() {
		syncRoots();
		rebuild();
		mRebuildPending = false;
	}

	private void syncRoots() {
		if (mSyncWeapons || mRootsToSync.contains(mCarriedEquipmentOutline) || mRootsToSync.contains(mAdvantageOutline) || mRootsToSync.contains(mSpellOutline) || mRootsToSync.contains(mSkillOutline)) {
			TKOutlineModel outlineModel = mMeleeWeaponOutline.getModel();
			String sortConfig = outlineModel.getSortConfig();

			outlineModel.removeAllRows();
			for (CMWeaponDisplayRow row : collectWeapons(CMMeleeWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);

			outlineModel = mRangedWeaponOutline.getModel();
			sortConfig = outlineModel.getSortConfig();

			outlineModel.removeAllRows();
			for (CMWeaponDisplayRow row : collectWeapons(CMRangedWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);
		}
		mSyncWeapons = false;
		mRootsToSync.clear();
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
		return false;
	}

	public boolean shouldTrackViewportWidth() {
		return false;
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
			setBorder(new TKLineBorder(TKColor.HIGHLIGHT, 2, TKLineBorder.ALL_EDGES, false), true, false);
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
		((CSSheetWindow) getBaseWindow()).addRows(mDragRows);
		mDragRows = null;
		setBorder(null, true, false);
		dtde.dropComplete(true);
	}

	public void dragExit(DropTargetEvent dte) {
		mDragRows = null;
		setBorder(null, true, false);
	}

	/**
	 * @param file The file to save to.
	 * @return <code>true</code> on success.
	 */
	public boolean saveAsPDF(File file) {
		try {
			TKBaseWindow window = getBaseWindow();
			TKPrintManager settings = mCharacter.getPageSettings();
			double[] size = settings.getPaperSize(TKLengthUnits.INCHES);
			float width = 72 * (float) size[0];
			float height = 72 * (float) size[1];
			com.lowagie.text.Document pdfDoc = new com.lowagie.text.Document(new com.lowagie.text.Rectangle(width, height));
			PdfWriter writer = PdfWriter.getInstance(pdfDoc, new FileOutputStream(file));
			int pageNum = 0;
			PageFormat format = settings.createPageFormat();
			PdfContentByte cb;

			pdfDoc.open();
			cb = writer.getDirectContent();
			while (true) {
				PdfTemplate template = cb.createTemplate(width, height);
				Graphics2D g2d = template.createGraphics(width, height, new DefaultFontMapper());

				if (print(g2d, format, pageNum) == NO_SUCH_PAGE) {
					g2d.dispose();
					break;
				}
				if (pageNum != 0) {
					pdfDoc.newPage();
				}
				g2d.setClip(0, 0, (int) width, (int) height);
				if (window != null) {
					window.getRepaintManager().setPrinting(true);
				}
				TKGraphics.configureGraphics(g2d);
				print(g2d, format, pageNum++);
				if (window != null) {
					window.getRepaintManager().setPrinting(false);
				}
				g2d.dispose();
				cb.addTemplate(template, 0, 0);
			}
			pdfDoc.close();
			return true;
		} catch (Exception exception) {
			return false;
		}
	}

	/**
	 * @param file The file to save to.
	 * @param template The template file to use.
	 * @param templateUsed A buffer to store the path actually used for the template. Use
	 *            <code>null</code> if this isn't wanted.
	 * @return <code>true</code> on success.
	 */
	public boolean saveAsHTML(File file, File template, StringBuilder templateUsed) {
		BufferedReader in = null;
		BufferedWriter out = null;

		try {
			char[] buffer = new char[1];
			int state = 0;
			StringBuilder keyBuffer = new StringBuilder();

			if (template == null || !template.isFile() || !template.canRead()) {
				template = new File(CSSheetPreferences.getHTMLTemplate());
				if (!template.isFile() || !template.canRead()) {
					template = new File(CSSheetPreferences.getDefaultHTMLTemplate());
				}
			}
			if (templateUsed != null) {
				templateUsed.append(TKPath.getFullPath(template));
			}
			in = new BufferedReader(new InputStreamReader(new FileInputStream(template)));
			out = new BufferedWriter(new FileWriter(file));

			while (in.read(buffer) != -1) {
				char ch = buffer[0];

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							in.mark(1);
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
							in.mark(1);
						} else {
							in.reset();
							emitHTMLKey(in, out, keyBuffer.toString(), file);
							keyBuffer.setLength(0);
							state = 0;
						}
						break;
				}
			}
			if (keyBuffer.length() != 0) {
				emitHTMLKey(in, out, keyBuffer.toString(), file);
			}
			return true;
		} catch (Exception exception) {
			return false;
		} finally {
			if (in != null) {
				try {
					in.close();
				} catch (IOException ioe) {
					// Ignore...
				}
			}
			if (out != null) {
				try {
					out.close();
				} catch (IOException ioe) {
					// Ignore...
				}
			}
		}
	}

	private void emitHTMLKey(BufferedReader in, BufferedWriter out, String key, File base) throws IOException {
		if (key.equals("PORTRAIT")) { //$NON-NLS-1$
			String fileName = TKPath.getLeafName(base.getName(), false) + CSSheetWindow.PNG_EXTENSION;

			TKImage.writePNG(new File(base.getParentFile(), fileName), mCharacter.getPortrait(true), 150);
			writeXMLData(out, fileName);
		} else if (key.equals("NAME")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getName());
		} else if (key.equals("TITLE")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getTitle());
		} else if (key.equals("RELIGION")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getReligion());
		} else if (key.equals("PLAYER")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getPlayerName());
		} else if (key.equals("CAMPAIGN")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getCampaign());
		} else if (key.equals("CREATED_ON")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getCreatedOn());
		} else if (key.equals("CAMPAIGN")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getCampaign());
		} else if (key.equals("ATTRIBUTE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getAttributePoints()));
		} else if (key.equals("ADVANTAGE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getAdvantagePoints()));
		} else if (key.equals("DISADVANTAGE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getDisadvantagePoints()));
		} else if (key.equals("QUIRK_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getQuirkPoints()));
		} else if (key.equals("SKILL_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getSkillPoints()));
		} else if (key.equals("SPELL_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getSpellPoints()));
		} else if (key.equals("RACE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getRacePoints()));
		} else if (key.equals("EARNED_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getEarnedPoints()));
		} else if (key.equals("RACE")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getRace());
		} else if (key.equals("HEIGHT")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.formatHeight(mCharacter.getHeight()));
		} else if (key.equals("HAIR")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getHair());
		} else if (key.equals("GENDER")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getGender());
		} else if (key.equals("WEIGHT")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getWeight()));
		} else if (key.equals("EYES")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getEyeColor());
		} else if (key.equals("AGE")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getAge()));
		} else if (key.equals("SIZE")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getSizeModifier(), true));
		} else if (key.equals("SKIN")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getSkinColor());
		} else if (key.equals("BIRTHDAY")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getBirthday());
		} else if (key.equals("TL")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getTechLevel());
		} else if (key.equals("HAND")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getHandedness());
		} else if (key.equals("ST")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getStrength()));
		} else if (key.equals("DX")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getDexterity()));
		} else if (key.equals("IQ")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getIntelligence()));
		} else if (key.equals("HT")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getHealth()));
		} else if (key.equals("WILL")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getWill()));
		} else if (key.equals("BASIC_SPEED")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getBasicSpeed()));
		} else if (key.equals("BASIC_MOVE")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getBasicMove()));
		} else if (key.equals("PERCEPTION")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getPerception()));
		} else if (key.equals("VISION")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getVision()));
		} else if (key.equals("HEARING")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getHearing()));
		} else if (key.equals("TASTE_SMELL")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getTasteAndSmell()));
		} else if (key.equals("TOUCH")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getTouch()));
		} else if (key.equals("THRUST")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getThrust().toString());
		} else if (key.equals("SWING")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getSwing().toString());
		} else if (key.startsWith("ENCUMBRANCE_LOOP_START")) { //$NON-NLS-1$
			processEncumbranceLoop(out, extractUpToMarker(in, "ENCUMBRANCE_LOOP_END")); //$NON-NLS-1$
		} else if (key.startsWith("HIT_LOCATION_LOOP_START")) { //$NON-NLS-1$
			processHitLocationLoop(out, extractUpToMarker(in, "HIT_LOCATION_LOOP_END")); //$NON-NLS-1$
		} else if (key.equals("FP")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getCurrentFatiguePoints());
		} else if (key.equals("BASIC_FP")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getFatiguePoints()));
		} else if (key.equals("TIRED")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getTiredFatiguePoints()));
		} else if (key.equals("FP_COLLAPSE")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getUnconsciousChecksFatiguePoints()));
		} else if (key.equals("UNCONSCIOUS")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getUnconsciousFatiguePoints()));
		} else if (key.equals("HP")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getCurrentHitPoints());
		} else if (key.equals("BASIC_HP")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getHitPoints()));
		} else if (key.equals("REELING")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getReelingHitPoints()));
		} else if (key.equals("HP_COLLAPSE")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getUnconsciousChecksHitPoints()));
		} else if (key.equals("DEATH_CHECK_1")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getDeathCheck1HitPoints()));
		} else if (key.equals("DEATH_CHECK_2")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getDeathCheck2HitPoints()));
		} else if (key.equals("DEATH_CHECK_3")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getDeathCheck3HitPoints()));
		} else if (key.equals("DEATH_CHECK_4")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getDeathCheck4HitPoints()));
		} else if (key.equals("DEAD")) { //$NON-NLS-1$
			writeXMLText(out, TKNumberUtils.format(mCharacter.getDeadHitPoints()));
		} else if (key.equals("BASIC_LIFT")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getBasicLift()));
		} else if (key.equals("ONE_HANDED_LIFT")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getOneHandedLift()));
		} else if (key.equals("TWO_HANDED_LIFT")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getTwoHandedLift()));
		} else if (key.equals("SHOVE")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getShoveAndKnockOver()));
		} else if (key.equals("RUNNING_SHOVE")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getRunningShoveAndKnockOver()));
		} else if (key.equals("CARRY_ON_BACK")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getCarryOnBack()));
		} else if (key.equals("SHIFT_SLIGHTLY")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getShiftSlightly()));
		} else if (key.startsWith("ADVANTAGES_LOOP_START")) { //$NON-NLS-1$
			processAdvantagesLoop(out, extractUpToMarker(in, "ADVANTAGES_LOOP_END")); //$NON-NLS-1$
		} else if (key.startsWith("SKILLS_LOOP_START")) { //$NON-NLS-1$
			processSkillsLoop(out, extractUpToMarker(in, "SKILLS_LOOP_END")); //$NON-NLS-1$
		} else if (key.startsWith("SPELLS_LOOP_START")) { //$NON-NLS-1$
			processSpellsLoop(out, extractUpToMarker(in, "SPELLS_LOOP_END")); //$NON-NLS-1$
		} else if (key.startsWith("MELEE_LOOP_START")) { //$NON-NLS-1$
			processMeleeLoop(out, extractUpToMarker(in, "MELEE_LOOP_END")); //$NON-NLS-1$
		} else if (key.startsWith("RANGED_LOOP_START")) { //$NON-NLS-1$
			processRangedLoop(out, extractUpToMarker(in, "RANGED_LOOP_END")); //$NON-NLS-1$
		} else if (key.equals("CARRIED_WEIGHT")) { //$NON-NLS-1$
			writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getWeightCarried()));
		} else if (key.equals("CARRIED_VALUE")) { //$NON-NLS-1$
			writeXMLText(out, "$" + TKNumberUtils.format(mCharacter.getWealthCarried())); //$NON-NLS-1$
		} else if (key.startsWith("CARRIED_EQUIPMENT_LOOP_START")) { //$NON-NLS-1$
			processEquipmentLoop(out, extractUpToMarker(in, "CARRIED_EQUIPMENT_LOOP_END"), true); //$NON-NLS-1$
		} else if (key.startsWith("OTHER_EQUIPMENT_LOOP_START")) { //$NON-NLS-1$
			processEquipmentLoop(out, extractUpToMarker(in, "OTHER_EQUIPMENT_LOOP_END"), false); //$NON-NLS-1$
		} else if (key.equals("NOTES")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getNotes());
		} else {
			writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
		}
	}

	private void writeXMLText(BufferedWriter out, String text) throws IOException {
		out.write(TKXMLWriter.encodeData(text).replaceAll("&#10;", "<br>")); //$NON-NLS-1$ //$NON-NLS-2$
	}

	private void writeXMLData(BufferedWriter out, String text) throws IOException {
		out.write(TKXMLWriter.encodeData(text).replaceAll(" ", "%20")); //$NON-NLS-1$ //$NON-NLS-2$
	}

	private String extractUpToMarker(BufferedReader in, String marker) throws IOException {
		char[] buffer = new char[1];
		StringBuilder keyBuffer = new StringBuilder();
		StringBuilder extraction = new StringBuilder();
		int state = 0;

		while (in.read(buffer) != -1) {
			char ch = buffer[0];

			switch (state) {
				case 0:
					if (ch == '@') {
						state = 1;
						in.mark(1);
						break;
					}
					extraction.append(ch);
					break;
				case 1:
					if (ch == '_' || Character.isLetterOrDigit(ch)) {
						keyBuffer.append(ch);
						in.mark(1);
					} else {
						String key = keyBuffer.toString();

						in.reset();
						if (key.equals(marker)) {
							return extraction.toString();
						}
						extraction.append('@');
						extraction.append(key);
						keyBuffer.setLength(0);
						state = 0;
					}
					break;
			}
		}
		return extraction.toString();
	}

	private void processEncumbranceLoop(BufferedWriter out, String contents) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;

		for (int level = CMCharacter.ENCUMBRANCE_NONE; level <= CMCharacter.ENCUMBRANCE_EXTRA_HEAVY; level++) {
			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("CURRENT_MARKER")) { //$NON-NLS-1$
								if (level == mCharacter.getEncumbranceLevel()) {
									out.write(" class=\"encumbrance\" "); //$NON-NLS-1$
								}
							} else if (key.equals("LEVEL")) { //$NON-NLS-1$
								writeXMLText(out, MessageFormat.format(level == mCharacter.getEncumbranceLevel() ? Msgs.CURRENT_ENCUMBRANCE_FORMAT : Msgs.ENCUMBRANCE_FORMAT, CSEncumbrancePanel.ENCUMBRANCE_TITLES[level], TKNumberUtils.format(level)));
							} else if (key.equals("MAX_LOAD")) { //$NON-NLS-1$
								writeXMLText(out, TKWeightUnits.POUNDS.format(mCharacter.getMaximumCarry(level)));
							} else if (key.equals("MOVE")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(mCharacter.getMove(level)));
							} else if (key.equals("DODGE")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(mCharacter.getDodge(level)));
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
		}
	}

	private void processHitLocationLoop(BufferedWriter out, String contents) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;

		for (int which = 0; which < CSHitLocationPanel.DR_KEYS.length; which++) {
			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("ROLL")) { //$NON-NLS-1$
								writeXMLText(out, CSHitLocationPanel.ROLLS[which]);
							} else if (key.equals("WHERE")) { //$NON-NLS-1$
								writeXMLText(out, CSHitLocationPanel.LOCATIONS[which]);
							} else if (key.equals("PENALTY")) { //$NON-NLS-1$
								writeXMLText(out, CSHitLocationPanel.PENALTIES[which]);
							} else if (key.equals("DR")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(((Integer) mCharacter.getValueForID(CSHitLocationPanel.DR_KEYS[which])).intValue()));
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
		}
	}

	private void processAdvantagesLoop(BufferedWriter out, String contents) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;
		boolean odd = true;

		for (CMAdvantage advantage : mCharacter.getAdvantagesIterator()) {
			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("EVEN_ODD")) { //$NON-NLS-1$
								out.write(odd ? "odd" : "even"); //$NON-NLS-1$ //$NON-NLS-2$
							} else if (key.equals("STYLE_INDENT_WARNING")) { //$NON-NLS-1$
								StringBuilder style = new StringBuilder();
								int depth = advantage.getDepth();

								if (depth > 0) {
									style.append(" style=\"padding-left: "); //$NON-NLS-1$
									style.append(depth * 12);
									style.append("px;"); //$NON-NLS-1$
								}
								if (!advantage.isSatisfied()) {
									if (style.length() == 0) {
										style.append(" style=\""); //$NON-NLS-1$
									}
									style.append(" color: red;"); //$NON-NLS-1$
								}
								if (style.length() > 0) {
									style.append("\" "); //$NON-NLS-1$
									out.write(style.toString());
								}
							} else if (key.equals("DESCRIPTION")) { //$NON-NLS-1$
								writeXMLText(out, advantage.toString());
								writeNote(out, advantage.getModifierNotes());
								writeNote(out, advantage.getNotes());
							} else if (key.equals("POINTS")) { //$NON-NLS-1$
								writeXMLText(out, CSAdvantageColumnID.POINTS.getDataAsText(advantage));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, CSAdvantageColumnID.REFERENCE.getDataAsText(advantage));
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
			odd = !odd;
		}
	}

	private void writeNote(BufferedWriter out, String notes) throws IOException {
		if (notes != null && notes.length() > 0) {
			out.write("<div class=\"note\">"); //$NON-NLS-1$
			writeXMLText(out, notes);
			out.write("</div>"); //$NON-NLS-1$
		}
	}

	private void processSkillsLoop(BufferedWriter out, String contents) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;
		boolean odd = true;

		for (CMSkill skill : mCharacter.getSkillsIterator()) {
			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("EVEN_ODD")) { //$NON-NLS-1$
								out.write(odd ? "odd" : "even"); //$NON-NLS-1$ //$NON-NLS-2$
							} else if (key.equals("STYLE_INDENT_WARNING")) { //$NON-NLS-1$
								StringBuilder style = new StringBuilder();
								int depth = skill.getDepth();

								if (depth > 0) {
									style.append(" style=\"padding-left: "); //$NON-NLS-1$
									style.append(depth * 12);
									style.append("px;"); //$NON-NLS-1$
								}
								if (!skill.isSatisfied()) {
									if (style.length() == 0) {
										style.append(" style=\""); //$NON-NLS-1$
									}
									style.append(" color: red;"); //$NON-NLS-1$
								}
								if (style.length() > 0) {
									style.append("\" "); //$NON-NLS-1$
									out.write(style.toString());
								}
							} else if (key.equals("DESCRIPTION")) { //$NON-NLS-1$
								writeXMLText(out, skill.toString());
								writeNote(out, skill.getNotes());
							} else if (key.equals("SL")) { //$NON-NLS-1$
								writeXMLText(out, CSSkillColumnID.LEVEL.getDataAsText(skill));
							} else if (key.equals("RSL")) { //$NON-NLS-1$
								writeXMLText(out, CSSkillColumnID.RELATIVE_LEVEL.getDataAsText(skill));
							} else if (key.equals("POINTS")) { //$NON-NLS-1$
								writeXMLText(out, CSSkillColumnID.POINTS.getDataAsText(skill));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, CSSkillColumnID.REFERENCE.getDataAsText(skill));
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
			odd = !odd;
		}
	}

	private void processSpellsLoop(BufferedWriter out, String contents) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;
		boolean odd = true;

		for (CMSpell spell : mCharacter.getSpellsIterator()) {
			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("EVEN_ODD")) { //$NON-NLS-1$
								out.write(odd ? "odd" : "even"); //$NON-NLS-1$ //$NON-NLS-2$
							} else if (key.equals("STYLE_INDENT_WARNING")) { //$NON-NLS-1$
								StringBuilder style = new StringBuilder();
								int depth = spell.getDepth();

								if (depth > 0) {
									style.append(" style=\"padding-left: "); //$NON-NLS-1$
									style.append(depth * 12);
									style.append("px;"); //$NON-NLS-1$
								}
								if (!spell.isSatisfied()) {
									if (style.length() == 0) {
										style.append(" style=\""); //$NON-NLS-1$
									}
									style.append(" color: red;"); //$NON-NLS-1$
								}
								if (style.length() > 0) {
									style.append("\" "); //$NON-NLS-1$
									out.write(style.toString());
								}
							} else if (key.equals("DESCRIPTION")) { //$NON-NLS-1$
								writeXMLText(out, spell.toString());
								writeNote(out, spell.getNotes());
							} else if (key.equals("CLASS")) { //$NON-NLS-1$
								writeXMLText(out, spell.getSpellClass());
							} else if (key.equals("COLLEGE")) { //$NON-NLS-1$
								writeXMLText(out, spell.getCollege());
							} else if (key.equals("MANA_CAST")) { //$NON-NLS-1$
								writeXMLText(out, spell.getCastingCost());
							} else if (key.equals("MANA_MAINTAIN")) { //$NON-NLS-1$
								writeXMLText(out, spell.getMaintenance());
							} else if (key.equals("TIME_CAST")) { //$NON-NLS-1$
								writeXMLText(out, spell.getCastingTime());
							} else if (key.equals("DURATION")) { //$NON-NLS-1$
								writeXMLText(out, spell.getDuration());
							} else if (key.equals("SL")) { //$NON-NLS-1$
								writeXMLText(out, CSSpellColumnID.LEVEL.getDataAsText(spell));
							} else if (key.equals("RSL")) { //$NON-NLS-1$
								writeXMLText(out, CSSpellColumnID.RELATIVE_LEVEL.getDataAsText(spell));
							} else if (key.equals("POINTS")) { //$NON-NLS-1$
								writeXMLText(out, CSSpellColumnID.POINTS.getDataAsText(spell));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, CSSpellColumnID.REFERENCE.getDataAsText(spell));
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
			odd = !odd;
		}
	}

	private void processMeleeLoop(BufferedWriter out, String contents) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;
		boolean odd = true;

		for (CMWeaponDisplayRow row : new TKFilteredIterator<CMWeaponDisplayRow>(getMeleeWeaponOutline().getModel().getRows(), CMWeaponDisplayRow.class)) {
			CMMeleeWeaponStats weapon = (CMMeleeWeaponStats) row.getWeapon();

			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("EVEN_ODD")) { //$NON-NLS-1$
								out.write(odd ? "odd" : "even"); //$NON-NLS-1$  //$NON-NLS-2$
							} else if (key.equals("DESCRIPTION")) { //$NON-NLS-1$
								writeXMLText(out, weapon.toString());
								writeNote(out, weapon.getNotes());
							} else if (key.equals("USAGE")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getUsage());
							} else if (key.equals("LEVEL")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(weapon.getSkillLevel()));
							} else if (key.equals("PARRY")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getResolvedParry());
							} else if (key.equals("BLOCK")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getResolvedBlock());
							} else if (key.equals("DAMAGE")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getResolvedDamage());
							} else if (key.equals("REACH")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getReach());
							} else if (key.equals("STRENGTH")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getStrength());
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
			odd = !odd;
		}
	}

	private void processRangedLoop(BufferedWriter out, String contents) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;
		boolean odd = true;

		for (CMWeaponDisplayRow row : new TKFilteredIterator<CMWeaponDisplayRow>(getRangedWeaponOutline().getModel().getRows(), CMWeaponDisplayRow.class)) {
			CMRangedWeaponStats weapon = (CMRangedWeaponStats) row.getWeapon();

			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("EVEN_ODD")) { //$NON-NLS-1$
								out.write(odd ? "odd" : "even"); //$NON-NLS-1$ //$NON-NLS-2$
							} else if (key.equals("DESCRIPTION")) { //$NON-NLS-1$
								writeXMLText(out, weapon.toString());
								writeNote(out, weapon.getNotes());
							} else if (key.equals("USAGE")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getUsage());
							} else if (key.equals("LEVEL")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(weapon.getSkillLevel()));
							} else if (key.equals("ACCURACY")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getAccuracy());
							} else if (key.equals("DAMAGE")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getResolvedDamage());
							} else if (key.equals("RANGE")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getResolvedRange());
							} else if (key.equals("ROF")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getRateOfFire());
							} else if (key.equals("SHOTS")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getShots());
							} else if (key.equals("BULK")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getBulk());
							} else if (key.equals("RECOIL")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getRecoil());
							} else if (key.equals("STRENGTH")) { //$NON-NLS-1$
								writeXMLText(out, weapon.getStrength());
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
			odd = !odd;
		}
	}

	private void processEquipmentLoop(BufferedWriter out, String contents, boolean carried) throws IOException {
		int length = contents.length();
		StringBuilder keyBuffer = new StringBuilder();
		int state = 0;
		boolean odd = true;

		for (CMEquipment equipment : carried ? mCharacter.getCarriedEquipmentIterator() : mCharacter.getOtherEquipmentIterator()) {
			for (int i = 0; i < length; i++) {
				char ch = contents.charAt(i);

				switch (state) {
					case 0:
						if (ch == '@') {
							state = 1;
							break;
						}
						out.append(ch);
						break;
					case 1:
						if (ch == '_' || Character.isLetterOrDigit(ch)) {
							keyBuffer.append(ch);
						} else {
							String key = keyBuffer.toString();

							i--;
							keyBuffer.setLength(0);
							state = 0;

							if (key.equals("EVEN_ODD")) { //$NON-NLS-1$
								out.write(odd ? "odd" : "even"); //$NON-NLS-1$ //$NON-NLS-2$
							} else if (key.equals("STYLE_INDENT_WARNING")) { //$NON-NLS-1$
								StringBuilder style = new StringBuilder();
								int depth = equipment.getDepth();

								if (depth > 0) {
									style.append(" style=\"padding-left: "); //$NON-NLS-1$
									style.append(depth * 12);
									style.append("px;"); //$NON-NLS-1$
								}
								if (!equipment.isSatisfied()) {
									if (style.length() == 0) {
										style.append(" style=\""); //$NON-NLS-1$
									}
									style.append(" color: red;"); //$NON-NLS-1$
								}
								if (style.length() > 0) {
									style.append("\" "); //$NON-NLS-1$
									out.write(style.toString());
								}
							} else if (key.equals("DESCRIPTION")) { //$NON-NLS-1$
								writeXMLText(out, equipment.toString());
								writeNote(out, equipment.getNotes());
							} else if (carried && key.equals("EQUIPPED")) { //$NON-NLS-1$
								if (equipment.isEquipped()) {
									out.write("&radic;"); //$NON-NLS-1$
								}
							} else if (key.equals("QTY")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(equipment.getQuantity()));
							} else if (key.equals("COST")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(equipment.getValue()));
							} else if (key.equals("WEIGHT")) { //$NON-NLS-1$
								writeXMLText(out, TKWeightUnits.POUNDS.format(equipment.getWeight()));
							} else if (key.equals("COST_SUMMARY")) { //$NON-NLS-1$
								writeXMLText(out, TKNumberUtils.format(equipment.getExtendedValue()));
							} else if (key.equals("WEIGHT_SUMMARY")) { //$NON-NLS-1$
								writeXMLText(out, TKWeightUnits.POUNDS.format(equipment.getExtendedWeight()));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, equipment.getReference());
							} else {
								writeXMLText(out, Msgs.UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
			odd = !odd;
		}
	}

	/**
	 * @param file The file to save to.
	 * @param createdFiles The files that were created.
	 * @return <code>true</code> on success.
	 */
	public boolean saveAsPNG(File file, ArrayList<File> createdFiles) {
		try {
			TKBaseWindow window = getBaseWindow();
			TKPrintManager settings = mCharacter.getPageSettings();
			Graphics2D g2d = TKGraphics.getGraphics();
			GraphicsConfiguration gc = g2d.getDeviceConfiguration();
			int dpi = CSSheetPreferences.getPNGResolution();
			double[] size = settings.getPaperSize(TKLengthUnits.INCHES);
			int width = (int) (dpi * size[0]);
			int height = (int) (dpi * size[1]);
			BufferedImage buffer = gc.createCompatibleImage(width, height, Transparency.OPAQUE);
			int pageNum = 0;
			String name = TKPath.getLeafName(file.getName(), false);
			PageFormat format = settings.createPageFormat();

			g2d.dispose();
			file = file.getParentFile();

			while (true) {
				File pngFile;

				g2d = (Graphics2D) buffer.getGraphics();
				if (print(g2d, format, pageNum) == NO_SUCH_PAGE) {
					g2d.dispose();
					break;
				}
				g2d.setClip(0, 0, width, height);
				g2d.setBackground(Color.WHITE);
				g2d.clearRect(0, 0, width, height);
				g2d.scale(dpi / 72.0, dpi / 72.0);
				if (window != null) {
					window.getRepaintManager().setPrinting(true);
				}
				TKGraphics.configureGraphics(g2d);
				print(g2d, format, pageNum++);
				if (window != null) {
					window.getRepaintManager().setPrinting(false);
				}
				g2d.dispose();
				pngFile = new File(file, name + (pageNum > 1 ? " " + pageNum : "") + CSSheetWindow.PNG_EXTENSION); //$NON-NLS-1$ //$NON-NLS-2$
				if (!TKImage.writePNG(pngFile, buffer, dpi)) {
					throw new IOException();
				}
				createdFiles.add(pngFile);
			}
			return true;
		} catch (Exception exception) {
			return false;
		}
	}
}
