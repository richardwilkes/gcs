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

package com.trollworks.gcs.character;

import com.lowagie.text.pdf.DefaultFontMapper;
import com.lowagie.text.pdf.PdfContentByte;
import com.lowagie.text.pdf.PdfTemplate;
import com.lowagie.text.pdf.PdfWriter;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageColumn;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.app.Main;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentColumn;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.preferences.SheetPreferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillColumn;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellColumn;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.utility.Debug;
import com.trollworks.gcs.utility.Fonts;
import com.trollworks.gcs.utility.StdUndoManager;
import com.trollworks.gcs.utility.collections.FilteredIterator;
import com.trollworks.gcs.utility.io.Images;
import com.trollworks.gcs.utility.io.LocalizedMessages;
import com.trollworks.gcs.utility.io.Path;
import com.trollworks.gcs.utility.io.Preferences;
import com.trollworks.gcs.utility.io.print.PrintManager;
import com.trollworks.gcs.utility.io.xml.XMLWriter;
import com.trollworks.gcs.utility.notification.BatchNotifierTarget;
import com.trollworks.gcs.utility.text.NumberUtils;
import com.trollworks.gcs.utility.units.WeightUnits;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponDisplayRow;
import com.trollworks.gcs.weapon.WeaponStats;
import com.trollworks.gcs.widgets.AppWindow;
import com.trollworks.gcs.widgets.GraphicsUtilities;
import com.trollworks.gcs.widgets.UIUtilities;
import com.trollworks.gcs.widgets.layout.ColumnLayout;
import com.trollworks.gcs.widgets.layout.RowDistribution;
import com.trollworks.gcs.widgets.outline.Column;
import com.trollworks.gcs.widgets.outline.ListRow;
import com.trollworks.gcs.widgets.outline.Outline;
import com.trollworks.gcs.widgets.outline.OutlineHeader;
import com.trollworks.gcs.widgets.outline.OutlineModel;
import com.trollworks.gcs.widgets.outline.OutlineSyncer;
import com.trollworks.gcs.widgets.outline.Row;
import com.trollworks.gcs.widgets.outline.RowSelection;
import com.trollworks.gcs.widgets.outline.Selection;

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
import java.awt.print.Paper;
import java.awt.print.Printable;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.FileWriter;
import java.io.IOException;
import java.io.InputStreamReader;
import java.text.DateFormat;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Date;
import java.util.HashMap;
import java.util.HashSet;

import javax.swing.JPanel;
import javax.swing.RepaintManager;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

/** The character sheet. */
public class CharacterSheet extends JPanel implements ChangeListener, Scrollable, BatchNotifierTarget, PageOwner, Printable, ActionListener, Runnable, DropTargetListener {
	private static String		MSG_PAGE_NUMBER;
	private static String		MSG_ADVERTISEMENT;
	private static String		MSG_MELEE_WEAPONS;
	private static String		MSG_RANGED_WEAPONS;
	private static String		MSG_ADVANTAGES;
	private static String		MSG_SKILLS;
	private static String		MSG_SPELLS;
	private static String		MSG_CARRIED_EQUIPMENT;
	private static String		MSG_OTHER_EQUIPMENT;
	private static String		MSG_CONTINUED;
	private static String		MSG_NATURAL;
	private static String		MSG_PUNCH;
	private static String		MSG_KICK;
	private static String		MSG_BOOTS;
	private static String		MSG_UNIDENTIFIED_KEY;
	private static String		MSG_NOTES;
	private static final String	BOXING_SKILL_NAME	= "Boxing";	//$NON-NLS-1$
	private static final String	KARATE_SKILL_NAME	= "Karate";	//$NON-NLS-1$
	private static final String	BRAWLING_SKILL_NAME	= "Brawling";	//$NON-NLS-1$
	private GURPSCharacter		mCharacter;
	private int					mLastPage;
	private boolean				mBatchMode;
	private AdvantageOutline	mAdvantageOutline;
	private SkillOutline		mSkillOutline;
	private SpellOutline		mSpellOutline;
	private EquipmentOutline	mCarriedEquipmentOutline;
	private EquipmentOutline	mOtherEquipmentOutline;
	private Outline				mMeleeWeaponOutline;
	private Outline				mRangedWeaponOutline;
	private boolean				mRebuildPending;
	private HashSet<Outline>	mRootsToSync;
	private boolean				mSyncWeapons;
	private boolean				mDisposed;

	static {
		LocalizedMessages.initialize(CharacterSheet.class);
	}

	/**
	 * Creates a new character sheet display. {@link #rebuild()} must be called prior to the first
	 * display of this panel.
	 * 
	 * @param character The character to display the data for.
	 */
	public CharacterSheet(GURPSCharacter character) {
		super(new CharacterSheetLayout());
		setOpaque(false);
		mCharacter = character;
		mLastPage = -1;
		mRootsToSync = new HashSet<Outline>();
		if (!GraphicsUtilities.inHeadlessPrintMode()) {
			setDropTarget(new DropTarget(this, this));
		}
		Preferences.getInstance().getNotifier().add(this, SheetPreferences.OPTIONAL_DICE_RULES_PREF_KEY);
	}

	/** Call when the sheet is no longer in use. */
	public void dispose() {
		Preferences.getInstance().getNotifier().remove(this);
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
		int firstRow = 0;
		String focusKey = null;
		PageAssembler pageAssembler;

		if (focus instanceof PageField) {
			focusKey = ((PageField) focus).getConsumedType();
			focus = null;
		}
		if (focus instanceof Outline) {
			Outline outline = (Outline) focus;
			Selection selection = outline.getModel().getSelection();

			firstRow = outline.getFirstRowToDisplay();
			int selRow = selection.nextSelectedIndex(firstRow);
			if (selRow >= 0) {
				firstRow = selRow;
			}
			focus = outline.getRealOutline();
		}

		Container window = getTopLevelAncestor();
		if (window != null) {
			((SheetWindow) window).jumpToSearchField();
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
		removeAll();
		mCharacter.resetNotifier(PrerequisitesThread.getThread(mCharacter));

		// Create the first page, which holds stuff that has a fixed vertical size.
		pageAssembler = new PageAssembler(this);
		pageAssembler.addToContent(hwrap(new PortraitPanel(mCharacter), vwrap(hwrap(new IdentityPanel(mCharacter), new PlayerInfoPanel(mCharacter)), new DescriptionPanel(mCharacter), RowDistribution.GIVE_EXCESS_TO_LAST), new PointsPanel(mCharacter)), null, null);
		pageAssembler.addToContent(hwrap(new AttributesPanel(mCharacter), vwrap(new EncumbrancePanel(mCharacter), new LiftPanel(mCharacter)), new HitLocationPanel(mCharacter), new HitPointsPanel(mCharacter)), null, null);

		// Add our outlines
		if (mAdvantageOutline.getModel().getRowCount() > 0 && mSkillOutline.getModel().getRowCount() > 0) {
			addOutline(pageAssembler, mAdvantageOutline, MSG_ADVANTAGES, mSkillOutline, MSG_SKILLS);
		} else {
			addOutline(pageAssembler, mAdvantageOutline, MSG_ADVANTAGES);
			addOutline(pageAssembler, mSkillOutline, MSG_SKILLS);
		}
		addOutline(pageAssembler, mSpellOutline, MSG_SPELLS);
		addOutline(pageAssembler, mMeleeWeaponOutline, MSG_MELEE_WEAPONS);
		addOutline(pageAssembler, mRangedWeaponOutline, MSG_RANGED_WEAPONS);
		addOutline(pageAssembler, mCarriedEquipmentOutline, MSG_CARRIED_EQUIPMENT);
		addOutline(pageAssembler, mOtherEquipmentOutline, MSG_OTHER_EQUIPMENT);

		pageAssembler.addNotes();

		// Ensure everything is laid out and register for notification
		validate();
		OutlineSyncer.remove(mAdvantageOutline);
		OutlineSyncer.remove(mSkillOutline);
		OutlineSyncer.remove(mSpellOutline);
		OutlineSyncer.remove(mCarriedEquipmentOutline);
		OutlineSyncer.remove(mOtherEquipmentOutline);
		mCharacter.addTarget(this, GURPSCharacter.CHARACTER_PREFIX);
		if (focusKey != null) {
			restoreFocusToKey(focusKey, this);
		} else if (focus instanceof Outline) {
			((Outline) focus).getBestOutlineForRowIndex(firstRow).requestFocusInWindow();
		}
	}

	private boolean restoreFocusToKey(String key, Component panel) {
		if (key != null) {
			if (panel instanceof PageField) {
				if (key.equals(((PageField) panel).getConsumedType())) {
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

	private void addOutline(PageAssembler pageAssembler, Outline outline, String title) {
		if (outline.getModel().getRowCount() > 0) {
			OutlineInfo info = new OutlineInfo(outline);
			boolean useProxy = false;

			while (pageAssembler.addToContent(new SingleOutlinePanel(outline, title, useProxy), info, null)) {
				if (!useProxy) {
					title = MessageFormat.format(MSG_CONTINUED, title);
					useProxy = true;
				}
			}
		}
	}

	private void addOutline(PageAssembler pageAssembler, Outline leftOutline, String leftTitle, Outline rightOutline, String rightTitle) {
		OutlineInfo infoLeft = new OutlineInfo(leftOutline);
		OutlineInfo infoRight = new OutlineInfo(rightOutline);
		boolean useProxy = false;

		while (pageAssembler.addToContent(new DoubleOutlinePanel(leftOutline, leftTitle, rightOutline, rightTitle, useProxy), infoLeft, infoRight)) {
			if (!useProxy) {
				leftTitle = MessageFormat.format(MSG_CONTINUED, leftTitle);
				rightTitle = MessageFormat.format(MSG_CONTINUED, rightTitle);
				useProxy = true;
			}
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
	}

	/** @return The outline containing the Advantages, Disadvantages & Quirks. */
	public AdvantageOutline getAdvantageOutline() {
		return mAdvantageOutline;
	}

	private void createAdvantageOutline() {
		if (mAdvantageOutline == null) {
			mAdvantageOutline = new AdvantageOutline(mCharacter);
			initOutline(mAdvantageOutline);
		} else {
			resetOutline(mAdvantageOutline);
		}
	}

	/** @return The outline containing the skills. */
	public SkillOutline getSkillOutline() {
		return mSkillOutline;
	}

	private void createSkillOutline() {
		if (mSkillOutline == null) {
			mSkillOutline = new SkillOutline(mCharacter);
			initOutline(mSkillOutline);
		} else {
			resetOutline(mSkillOutline);
		}
	}

	/** @return The outline containing the spells. */
	public SpellOutline getSpellOutline() {
		return mSpellOutline;
	}

	private void createSpellOutline() {
		if (mSpellOutline == null) {
			mSpellOutline = new SpellOutline(mCharacter);
			initOutline(mSpellOutline);
		} else {
			resetOutline(mSpellOutline);
		}
	}

	/** @return The outline containing the carried equipment. */
	public EquipmentOutline getCarriedEquipmentOutline() {
		return mCarriedEquipmentOutline;
	}

	private void createCarriedEquipmentOutline() {
		if (mCarriedEquipmentOutline == null) {
			mCarriedEquipmentOutline = new EquipmentOutline(mCharacter, true);
			initOutline(mCarriedEquipmentOutline);
		} else {
			resetOutline(mCarriedEquipmentOutline);
		}
	}

	/** @return The outline containing the other equipment. */
	public EquipmentOutline getOtherEquipmentOutline() {
		return mOtherEquipmentOutline;
	}

	private void createOtherEquipmentOutline() {
		if (mOtherEquipmentOutline == null) {
			mOtherEquipmentOutline = new EquipmentOutline(mCharacter, false);
			initOutline(mOtherEquipmentOutline);
		} else {
			resetOutline(mOtherEquipmentOutline);
		}
	}

	/** @return The outline containing the melee weapons. */
	public Outline getMeleeWeaponOutline() {
		return mMeleeWeaponOutline;
	}

	private void createMeleeWeaponOutline() {
		if (mMeleeWeaponOutline == null) {
			OutlineModel outlineModel;
			String sortConfig;

			mMeleeWeaponOutline = new WeaponOutline(MeleeWeaponStats.class);
			outlineModel = mMeleeWeaponOutline.getModel();
			sortConfig = outlineModel.getSortConfig();
			for (WeaponDisplayRow row : collectWeapons(MeleeWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);
			initOutline(mMeleeWeaponOutline);
		} else {
			resetOutline(mMeleeWeaponOutline);
		}
	}

	/** @return The outline containing the ranged weapons. */
	public Outline getRangedWeaponOutline() {
		return mRangedWeaponOutline;
	}

	private void createRangedWeaponOutline() {
		if (mRangedWeaponOutline == null) {
			OutlineModel outlineModel;
			String sortConfig;

			mRangedWeaponOutline = new WeaponOutline(RangedWeaponStats.class);
			outlineModel = mRangedWeaponOutline.getModel();
			sortConfig = outlineModel.getSortConfig();
			for (WeaponDisplayRow row : collectWeapons(RangedWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);
			initOutline(mRangedWeaponOutline);
		} else {
			resetOutline(mRangedWeaponOutline);
		}
	}

	private void addBuiltInWeapons(Class<? extends WeaponStats> weaponClass, HashMap<HashedWeapon, WeaponDisplayRow> map) {
		if (weaponClass == MeleeWeaponStats.class) {
			boolean savedModified = mCharacter.isModified();
			ArrayList<SkillDefault> defaults = new ArrayList<SkillDefault>();
			Advantage phantom;
			MeleeWeaponStats weapon;

			StdUndoManager mgr = mCharacter.getUndoManager();
			mCharacter.setUndoManager(null);

			phantom = new Advantage(mCharacter, false);
			phantom.setName(MSG_NATURAL);

			if (mCharacter.includePunch()) {
				defaults.add(new SkillDefault(SkillDefaultType.DX, null, null, 0));
				defaults.add(new SkillDefault(SkillDefaultType.Skill, BOXING_SKILL_NAME, null, 0));
				defaults.add(new SkillDefault(SkillDefaultType.Skill, BRAWLING_SKILL_NAME, null, 0));
				defaults.add(new SkillDefault(SkillDefaultType.Skill, KARATE_SKILL_NAME, null, 0));
				weapon = new MeleeWeaponStats(phantom);
				weapon.setUsage(MSG_PUNCH);
				weapon.setDefaults(defaults);
				weapon.setDamage("thr-1 cr"); //$NON-NLS-1$
				weapon.setReach("C"); //$NON-NLS-1$
				weapon.setParry("0"); //$NON-NLS-1$
				map.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
				defaults.clear();
			}

			defaults.add(new SkillDefault(SkillDefaultType.DX, null, null, -2));
			defaults.add(new SkillDefault(SkillDefaultType.Skill, BRAWLING_SKILL_NAME, null, -2));
			defaults.add(new SkillDefault(SkillDefaultType.Skill, KARATE_SKILL_NAME, null, -2));

			if (mCharacter.includeKick()) {
				weapon = new MeleeWeaponStats(phantom);
				weapon.setUsage(MSG_KICK);
				weapon.setDefaults(defaults);
				weapon.setDamage("thr cr"); //$NON-NLS-1$
				weapon.setReach("C,1"); //$NON-NLS-1$
				weapon.setParry("No"); //$NON-NLS-1$
				map.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
			}

			if (mCharacter.includeKickBoots()) {
				weapon = new MeleeWeaponStats(phantom);
				weapon.setUsage(MSG_BOOTS);
				weapon.setDefaults(defaults);
				weapon.setDamage("thr+1 cr"); //$NON-NLS-1$
				weapon.setReach("C,1"); //$NON-NLS-1$
				weapon.setParry("No"); //$NON-NLS-1$
				map.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
			}

			mCharacter.setUndoManager(mgr);
			mCharacter.setModified(savedModified);
		}
	}

	private ArrayList<WeaponDisplayRow> collectWeapons(Class<? extends WeaponStats> weaponClass) {
		HashMap<HashedWeapon, WeaponDisplayRow> weaponMap = new HashMap<HashedWeapon, WeaponDisplayRow>();
		ArrayList<WeaponDisplayRow> weaponList;

		addBuiltInWeapons(weaponClass, weaponMap);

		for (Advantage advantage : mCharacter.getAdvantagesIterator()) {
			for (WeaponStats weapon : advantage.getWeapons()) {
				if (weaponClass.isInstance(weapon)) {
					weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
				}
			}
		}

		for (Equipment equipment : mCharacter.getCarriedEquipmentIterator()) {
			if (equipment.getQuantity() > 0) {
				for (WeaponStats weapon : equipment.getWeapons()) {
					if (weaponClass.isInstance(weapon)) {
						weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
					}
				}
			}
		}

		for (Spell spell : mCharacter.getSpellsIterator()) {
			for (WeaponStats weapon : spell.getWeapons()) {
				if (weaponClass.isInstance(weapon)) {
					weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
				}
			}
		}

		for (Skill skill : mCharacter.getSkillsIterator()) {
			for (WeaponStats weapon : skill.getWeapons()) {
				if (weaponClass.isInstance(weapon)) {
					weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
				}
			}
		}

		weaponList = new ArrayList<WeaponDisplayRow>(weaponMap.values());
		return weaponList;
	}

	private void initOutline(Outline outline) {
		outline.addActionListener(this);
	}

	private void resetOutline(Outline outline) {
		outline.clearProxies();
	}

	private Container hwrap(Component left, Component right) {
		JPanel wrapper = new JPanel(new ColumnLayout(2, 2, 2));

		wrapper.add(left);
		wrapper.add(right);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private Container hwrap(Component left, Component center, Component right) {
		JPanel wrapper = new JPanel(new ColumnLayout(3, 2, 2));

		wrapper.add(left);
		wrapper.add(center);
		wrapper.add(right);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private Container hwrap(Component left, Component center, Component center2, Component right) {
		JPanel wrapper = new JPanel(new ColumnLayout(4, 2, 2));

		wrapper.add(left);
		wrapper.add(center);
		wrapper.add(center2);
		wrapper.add(right);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private Container vwrap(Component top, Component bottom) {
		JPanel wrapper = new JPanel(new ColumnLayout(1, 2, 2));

		wrapper.add(top);
		wrapper.add(bottom);
		wrapper.setAlignmentY(-1f);
		return wrapper;
	}

	private Container vwrap(Component top, Component bottom, RowDistribution distribution) {
		JPanel wrapper = new JPanel(new ColumnLayout(1, 2, 2, distribution));

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
			Component comp = getComponent(pageIndex);
			RepaintManager mgr = RepaintManager.currentManager(comp);
			boolean saved = mgr.isDoubleBufferingEnabled();
			mgr.setDoubleBufferingEnabled(false);
			comp.print(graphics);
			mgr.setDoubleBufferingEnabled(saved);
		}
		return PAGE_EXISTS;
	}

	public void enterBatchMode() {
		mBatchMode = true;
	}

	public void leaveBatchMode() {
		mBatchMode = false;
		validate();
	}

	public void handleNotification(Object producer, String type, Object data) {
		if (SheetPreferences.OPTIONAL_DICE_RULES_PREF_KEY.equals(type)) {
			markForRebuild();
		} else {
			if (type.startsWith(Advantage.PREFIX)) {
				OutlineSyncer.add(mAdvantageOutline);
			} else if (type.startsWith(Skill.PREFIX)) {
				OutlineSyncer.add(mSkillOutline);
			} else if (type.startsWith(Spell.PREFIX)) {
				OutlineSyncer.add(mSpellOutline);
			} else if (type.startsWith(Equipment.PREFIX)) {
				OutlineSyncer.add(mCarriedEquipmentOutline);
				OutlineSyncer.add(mOtherEquipmentOutline);
			}

			if (GURPSCharacter.ID_LAST_MODIFIED.equals(type)) {
				int count = getComponentCount();

				for (int i = 0; i < count; i++) {
					Page page = (Page) getComponent(i);
					Rectangle bounds = page.getBounds();
					Insets insets = page.getInsets();

					bounds.y = bounds.y + bounds.height - insets.bottom;
					bounds.height = insets.bottom;
					repaint(bounds);
				}
			} else if (Equipment.ID_QUANTITY.equals(type) || Equipment.ID_WEAPON_STATUS_CHANGED.equals(type) || Advantage.ID_WEAPON_STATUS_CHANGED.equals(type) || Spell.ID_WEAPON_STATUS_CHANGED.equals(type) || Skill.ID_WEAPON_STATUS_CHANGED.equals(type) || GURPSCharacter.ID_INCLUDE_PUNCH.equals(type) || GURPSCharacter.ID_INCLUDE_KICK.equals(type) || GURPSCharacter.ID_INCLUDE_BOOTS.equals(type)) {
				mSyncWeapons = true;
				markForRebuild();
			} else if (GURPSCharacter.ID_PARRY_BONUS.equals(type) || Skill.ID_LEVEL.equals(type)) {
				OutlineSyncer.add(mMeleeWeaponOutline);
				OutlineSyncer.add(mRangedWeaponOutline);
			} else if (GURPSCharacter.ID_CARRIED_WEIGHT.equals(type) || GURPSCharacter.ID_CARRIED_WEALTH.equals(type)) {
				Column column = mCarriedEquipmentOutline.getModel().getColumnWithID(EquipmentColumn.DESCRIPTION.ordinal());

				column.setName(EquipmentColumn.DESCRIPTION.toString(mCharacter, true));
			} else if (!mBatchMode) {
				validate();
			}
		}
	}

	public void drawPageAdornments(Page page, Graphics gc) {
		Rectangle bounds = page.getBounds();
		Insets insets = page.getInsets();
		bounds.width -= insets.left + insets.right;
		bounds.height -= insets.top + insets.bottom;
		bounds.x = insets.left;
		bounds.y = insets.top;
		int pageNumber = 1 + UIUtilities.getIndexOf(this, page);
		String pageString = MessageFormat.format(MSG_PAGE_NUMBER, NumberUtils.format(pageNumber), NumberUtils.format(getPageCount()));
		String copyright1 = Main.getCopyrightBanner(true);
		String copyright2 = copyright1.substring(copyright1.indexOf('\n') + 1);
		Font font1 = UIManager.getFont(Fonts.KEY_SECONDARY_FOOTER);
		Font font2 = UIManager.getFont(Fonts.KEY_PRIMARY_FOOTER);
		FontMetrics fm1 = gc.getFontMetrics(font1);
		FontMetrics fm2 = gc.getFontMetrics(font2);
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

		Font savedFont = gc.getFont();
		gc.setFont(font1);
		gc.drawString(left, bounds.x, y);
		gc.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, gc).getWidth(), y);
		gc.setFont(font2);
		center = mCharacter.getDescription().getName();
		gc.drawString(center, bounds.x + (bounds.width - (int) fm2.getStringBounds(center, gc).getWidth()) / 2, y);

		if (pageNumber % 2 == 1) {
			left = copyright2;
			right = pageString;
		} else {
			left = pageString;
			right = copyright2;
		}

		y += fm2.getDescent() + fm1.getAscent();

		gc.setFont(font1);
		gc.drawString(left, bounds.x, y);
		gc.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, gc).getWidth(), y);
		gc.drawString(MSG_ADVERTISEMENT, bounds.x + (bounds.width - (int) fm1.getStringBounds(MSG_ADVERTISEMENT, gc).getWidth()) / 2, y);

		gc.setFont(savedFont);
	}

	public Insets getPageAdornmentsInsets(Page page) {
		FontMetrics fm1 = Fonts.getFontMetrics(UIManager.getFont(Fonts.KEY_SECONDARY_FOOTER));
		FontMetrics fm2 = Fonts.getFontMetrics(UIManager.getFont(Fonts.KEY_PRIMARY_FOOTER));

		return new Insets(0, 0, fm1.getAscent() + fm1.getDescent() + fm2.getAscent() + fm2.getDescent(), 0);
	}

	public PrintManager getPageSettings() {
		return mCharacter.getPageSettings();
	}

	/** @return The character being displayed. */
	public GURPSCharacter getCharacter() {
		return mCharacter;
	}

	public void actionPerformed(ActionEvent event) {
		String command = event.getActionCommand();
		if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
			mRootsToSync.add(((Outline) event.getSource()).getRealOutline());
			markForRebuild();
		} else if (NotesPanel.CMD_EDIT_NOTES.equals(command)) {
			Profile description = mCharacter.getDescription();
			String notes = TextEditor.edit(MSG_NOTES, description.getNotes());
			if (notes != null) {
				description.setNotes(notes);
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
			OutlineModel outlineModel = mMeleeWeaponOutline.getModel();
			String sortConfig = outlineModel.getSortConfig();

			outlineModel.removeAllRows();
			for (WeaponDisplayRow row : collectWeapons(MeleeWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);

			outlineModel = mRangedWeaponOutline.getModel();
			sortConfig = outlineModel.getSortConfig();

			outlineModel.removeAllRows();
			for (WeaponDisplayRow row : collectWeapons(RangedWeaponStats.class)) {
				outlineModel.addRow(row);
			}
			outlineModel.applySortConfig(sortConfig);
		}
		mSyncWeapons = false;
		mRootsToSync.clear();
	}

	public void stateChanged(ChangeEvent event) {
		Dimension size = getLayout().preferredLayoutSize(this);
		if (!getSize().equals(size)) {
			invalidate();
			setSize(size);
		}
	}

	public int getScrollableBlockIncrement(Rectangle visibleRect, int orientation, int direction) {
		return orientation == SwingConstants.VERTICAL ? visibleRect.height : visibleRect.width;
	}

	public Dimension getPreferredScrollableViewportSize() {
		return getPreferredSize();
	}

	public int getScrollableUnitIncrement(Rectangle visibleRect, int orientation, int direction) {
		return 10;
	}

	public boolean getScrollableTracksViewportHeight() {
		return false;
	}

	public boolean getScrollableTracksViewportWidth() {
		return false;
	}

	private boolean			mDragWasAcceptable;
	private ArrayList<Row>	mDragRows;

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

	public void dragOver(DropTargetDragEvent dtde) {
		if (mDragWasAcceptable) {
// setBorder(new LineBorder(Colors.HIGHLIGHT, 2, LineBorder.ALL_EDGES, false));
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
		((SheetWindow) getTopLevelAncestor()).addRows(mDragRows);
		mDragRows = null;
// setBorder(null);
		dtde.dropComplete(true);
	}

	public void dragExit(DropTargetEvent dte) {
		mDragRows = null;
// setBorder(null);
	}

	private PageFormat createDefaultPageFormat() {
		Paper paper = new Paper();
		PageFormat format = new PageFormat();
		format.setOrientation(PageFormat.PORTRAIT);
		paper.setSize(8.5 * 72.0, 11.0 * 72.0);
		paper.setImageableArea(0.5 * 72.0, 0.5 * 72.0, 7.5 * 72.0, 10 * 72.0);
		format.setPaper(paper);
		return format;
	}

	/**
	 * @param file The file to save to.
	 * @return <code>true</code> on success.
	 */
	public boolean saveAsPDF(File file) {
		try {
			PrintManager settings = mCharacter.getPageSettings();
			PageFormat format = settings != null ? settings.createPageFormat() : createDefaultPageFormat();
			Paper paper = format.getPaper();
			float width = (float) paper.getWidth();
			float height = (float) paper.getHeight();
			com.lowagie.text.Document pdfDoc = new com.lowagie.text.Document(new com.lowagie.text.Rectangle(width, height));
			PdfWriter writer = PdfWriter.getInstance(pdfDoc, new FileOutputStream(file));
			int pageNum = 0;
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
				AppWindow window = (AppWindow) getTopLevelAncestor();
				if (window != null) {
					window.setPrinting(true);
				}
				print(g2d, format, pageNum++);
				if (window != null) {
					window.setPrinting(false);
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
				template = new File(SheetPreferences.getHTMLTemplate());
				if (!template.isFile() || !template.canRead()) {
					template = new File(SheetPreferences.getDefaultHTMLTemplate());
				}
			}
			if (templateUsed != null) {
				templateUsed.append(Path.getFullPath(template));
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
		Profile description = mCharacter.getDescription();

		if (key.equals("PORTRAIT")) { //$NON-NLS-1$
			String fileName = Path.getLeafName(base.getName(), false) + SheetWindow.PNG_EXTENSION;
			Images.writePNG(new File(base.getParentFile(), fileName), description.getPortrait(true), 150);
			writeXMLData(out, fileName);
		} else if (key.equals("NAME")) { //$NON-NLS-1$
			writeXMLText(out, description.getName());
		} else if (key.equals("TITLE")) { //$NON-NLS-1$
			writeXMLText(out, description.getTitle());
		} else if (key.equals("RELIGION")) { //$NON-NLS-1$
			writeXMLText(out, description.getReligion());
		} else if (key.equals("PLAYER")) { //$NON-NLS-1$
			writeXMLText(out, description.getPlayerName());
		} else if (key.equals("CAMPAIGN")) { //$NON-NLS-1$
			writeXMLText(out, description.getCampaign());
		} else if (key.equals("CREATED_ON")) { //$NON-NLS-1$
			Date date = new Date(mCharacter.getCreatedOn());
			writeXMLText(out, DateFormat.getDateInstance(DateFormat.MEDIUM).format(date));
		} else if (key.equals("MODIFIED_ON")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getLastModified());
		} else if (key.equals("CAMPAIGN")) { //$NON-NLS-1$
			writeXMLText(out, description.getCampaign());
		} else if (key.equals("TOTAL_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getTotalPoints()));
		} else if (key.equals("ATTRIBUTE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getAttributePoints()));
		} else if (key.equals("ADVANTAGE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getAdvantagePoints()));
		} else if (key.equals("DISADVANTAGE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getDisadvantagePoints()));
		} else if (key.equals("QUIRK_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getQuirkPoints()));
		} else if (key.equals("SKILL_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getSkillPoints()));
		} else if (key.equals("SPELL_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getSpellPoints()));
		} else if (key.equals("RACE_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getRacePoints()));
		} else if (key.equals("EARNED_POINTS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getEarnedPoints()));
		} else if (key.equals("RACE")) { //$NON-NLS-1$
			writeXMLText(out, description.getRace());
		} else if (key.equals("HEIGHT")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.formatHeight(description.getHeight()));
		} else if (key.equals("HAIR")) { //$NON-NLS-1$
			writeXMLText(out, description.getHair());
		} else if (key.equals("GENDER")) { //$NON-NLS-1$
			writeXMLText(out, description.getGender());
		} else if (key.equals("WEIGHT")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(description.getWeight()));
		} else if (key.equals("EYES")) { //$NON-NLS-1$
			writeXMLText(out, description.getEyeColor());
		} else if (key.equals("AGE")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(description.getAge()));
		} else if (key.equals("SIZE")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(description.getSizeModifier(), true));
		} else if (key.equals("SKIN")) { //$NON-NLS-1$
			writeXMLText(out, description.getSkinColor());
		} else if (key.equals("BIRTHDAY")) { //$NON-NLS-1$
			writeXMLText(out, description.getBirthday());
		} else if (key.equals("TL")) { //$NON-NLS-1$
			writeXMLText(out, description.getTechLevel());
		} else if (key.equals("HAND")) { //$NON-NLS-1$
			writeXMLText(out, description.getHandedness());
		} else if (key.equals("ST")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getStrength()));
		} else if (key.equals("DX")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getDexterity()));
		} else if (key.equals("IQ")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getIntelligence()));
		} else if (key.equals("HT")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getHealth()));
		} else if (key.equals("WILL")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getWill()));
		} else if (key.equals("BASIC_SPEED")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getBasicSpeed()));
		} else if (key.equals("BASIC_MOVE")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getBasicMove()));
		} else if (key.equals("PERCEPTION")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getPerception()));
		} else if (key.equals("VISION")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getVision()));
		} else if (key.equals("HEARING")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getHearing()));
		} else if (key.equals("TASTE_SMELL")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getTasteAndSmell()));
		} else if (key.equals("TOUCH")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getTouch()));
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
			writeXMLText(out, NumberUtils.format(mCharacter.getFatiguePoints()));
		} else if (key.equals("TIRED")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getTiredFatiguePoints()));
		} else if (key.equals("FP_COLLAPSE")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getUnconsciousChecksFatiguePoints()));
		} else if (key.equals("UNCONSCIOUS")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getUnconsciousFatiguePoints()));
		} else if (key.equals("HP")) { //$NON-NLS-1$
			writeXMLText(out, mCharacter.getCurrentHitPoints());
		} else if (key.equals("BASIC_HP")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getHitPoints()));
		} else if (key.equals("REELING")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getReelingHitPoints()));
		} else if (key.equals("HP_COLLAPSE")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getUnconsciousChecksHitPoints()));
		} else if (key.equals("DEATH_CHECK_1")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getDeathCheck1HitPoints()));
		} else if (key.equals("DEATH_CHECK_2")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getDeathCheck2HitPoints()));
		} else if (key.equals("DEATH_CHECK_3")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getDeathCheck3HitPoints()));
		} else if (key.equals("DEATH_CHECK_4")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getDeathCheck4HitPoints()));
		} else if (key.equals("DEAD")) { //$NON-NLS-1$
			writeXMLText(out, NumberUtils.format(mCharacter.getDeadHitPoints()));
		} else if (key.equals("BASIC_LIFT")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getBasicLift()));
		} else if (key.equals("ONE_HANDED_LIFT")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getOneHandedLift()));
		} else if (key.equals("TWO_HANDED_LIFT")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getTwoHandedLift()));
		} else if (key.equals("SHOVE")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getShoveAndKnockOver()));
		} else if (key.equals("RUNNING_SHOVE")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getRunningShoveAndKnockOver()));
		} else if (key.equals("CARRY_ON_BACK")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getCarryOnBack()));
		} else if (key.equals("SHIFT_SLIGHTLY")) { //$NON-NLS-1$
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getShiftSlightly()));
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
			writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getWeightCarried()));
		} else if (key.equals("CARRIED_VALUE")) { //$NON-NLS-1$
			writeXMLText(out, "$" + NumberUtils.format(mCharacter.getWealthCarried())); //$NON-NLS-1$
		} else if (key.startsWith("CARRIED_EQUIPMENT_LOOP_START")) { //$NON-NLS-1$
			processEquipmentLoop(out, extractUpToMarker(in, "CARRIED_EQUIPMENT_LOOP_END"), true); //$NON-NLS-1$
		} else if (key.startsWith("OTHER_EQUIPMENT_LOOP_START")) { //$NON-NLS-1$
			processEquipmentLoop(out, extractUpToMarker(in, "OTHER_EQUIPMENT_LOOP_END"), false); //$NON-NLS-1$
		} else if (key.equals("NOTES")) { //$NON-NLS-1$
			writeXMLText(out, description.getNotes());
		} else {
			writeXMLText(out, MSG_UNIDENTIFIED_KEY);
		}
	}

	private void writeXMLText(BufferedWriter out, String text) throws IOException {
		out.write(XMLWriter.encodeData(text).replaceAll("&#10;", "<br>")); //$NON-NLS-1$ //$NON-NLS-2$
	}

	private void writeXMLData(BufferedWriter out, String text) throws IOException {
		out.write(XMLWriter.encodeData(text).replaceAll(" ", "%20")); //$NON-NLS-1$ //$NON-NLS-2$
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

		for (int level = GURPSCharacter.ENCUMBRANCE_NONE; level <= GURPSCharacter.ENCUMBRANCE_EXTRA_HEAVY; level++) {
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
								writeXMLText(out, MessageFormat.format(level == mCharacter.getEncumbranceLevel() ? EncumbrancePanel.MSG_CURRENT_ENCUMBRANCE_FORMAT : EncumbrancePanel.MSG_ENCUMBRANCE_FORMAT, EncumbrancePanel.ENCUMBRANCE_TITLES[level], NumberUtils.format(level)));
							} else if (key.equals("MAX_LOAD")) { //$NON-NLS-1$
								writeXMLText(out, WeightUnits.POUNDS.format(mCharacter.getMaximumCarry(level)));
							} else if (key.equals("MOVE")) { //$NON-NLS-1$
								writeXMLText(out, NumberUtils.format(mCharacter.getMove(level)));
							} else if (key.equals("DODGE")) { //$NON-NLS-1$
								writeXMLText(out, NumberUtils.format(mCharacter.getDodge(level)));
							} else {
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
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

		for (int which = 0; which < HitLocationPanel.DR_KEYS.length; which++) {
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
								writeXMLText(out, HitLocationPanel.ROLLS[which]);
							} else if (key.equals("WHERE")) { //$NON-NLS-1$
								writeXMLText(out, HitLocationPanel.LOCATIONS[which]);
							} else if (key.equals("PENALTY")) { //$NON-NLS-1$
								writeXMLText(out, HitLocationPanel.PENALTIES[which]);
							} else if (key.equals("DR")) { //$NON-NLS-1$
								writeXMLText(out, NumberUtils.format(((Integer) mCharacter.getValueForID(HitLocationPanel.DR_KEYS[which])).intValue()));
							} else {
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
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

		for (Advantage advantage : mCharacter.getAdvantagesIterator()) {
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
								writeXMLText(out, AdvantageColumn.POINTS.getDataAsText(advantage));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, AdvantageColumn.REFERENCE.getDataAsText(advantage));
							} else {
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
							}
						}
						break;
				}
			}
			odd = !odd;
		}
	}

	private void writeNote(BufferedWriter out, String notes) throws IOException {
		if (notes.length() > 0) {
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

		for (Skill skill : mCharacter.getSkillsIterator()) {
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
								writeXMLText(out, SkillColumn.LEVEL.getDataAsText(skill));
							} else if (key.equals("RSL")) { //$NON-NLS-1$
								writeXMLText(out, SkillColumn.RELATIVE_LEVEL.getDataAsText(skill));
							} else if (key.equals("POINTS")) { //$NON-NLS-1$
								writeXMLText(out, SkillColumn.POINTS.getDataAsText(skill));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, SkillColumn.REFERENCE.getDataAsText(skill));
							} else {
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
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

		for (Spell spell : mCharacter.getSpellsIterator()) {
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
								writeXMLText(out, SpellColumn.LEVEL.getDataAsText(spell));
							} else if (key.equals("RSL")) { //$NON-NLS-1$
								writeXMLText(out, SpellColumn.RELATIVE_LEVEL.getDataAsText(spell));
							} else if (key.equals("POINTS")) { //$NON-NLS-1$
								writeXMLText(out, SpellColumn.POINTS.getDataAsText(spell));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, SpellColumn.REFERENCE.getDataAsText(spell));
							} else {
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
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

		for (WeaponDisplayRow row : new FilteredIterator<WeaponDisplayRow>(getMeleeWeaponOutline().getModel().getRows(), WeaponDisplayRow.class)) {
			MeleeWeaponStats weapon = (MeleeWeaponStats) row.getWeapon();

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
								writeXMLText(out, NumberUtils.format(weapon.getSkillLevel()));
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
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
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

		for (WeaponDisplayRow row : new FilteredIterator<WeaponDisplayRow>(getRangedWeaponOutline().getModel().getRows(), WeaponDisplayRow.class)) {
			RangedWeaponStats weapon = (RangedWeaponStats) row.getWeapon();

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
								writeXMLText(out, NumberUtils.format(weapon.getSkillLevel()));
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
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
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

		for (Equipment equipment : carried ? mCharacter.getCarriedEquipmentIterator() : mCharacter.getOtherEquipmentIterator()) {
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
								writeXMLText(out, NumberUtils.format(equipment.getQuantity()));
							} else if (key.equals("COST")) { //$NON-NLS-1$
								writeXMLText(out, NumberUtils.format(equipment.getValue()));
							} else if (key.equals("WEIGHT")) { //$NON-NLS-1$
								writeXMLText(out, WeightUnits.POUNDS.format(equipment.getWeight()));
							} else if (key.equals("COST_SUMMARY")) { //$NON-NLS-1$
								writeXMLText(out, NumberUtils.format(equipment.getExtendedValue()));
							} else if (key.equals("WEIGHT_SUMMARY")) { //$NON-NLS-1$
								writeXMLText(out, WeightUnits.POUNDS.format(equipment.getExtendedWeight()));
							} else if (key.equals("REF")) { //$NON-NLS-1$
								writeXMLText(out, equipment.getReference());
							} else {
								writeXMLText(out, MSG_UNIDENTIFIED_KEY);
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
			Graphics2D g2d = GraphicsUtilities.getGraphics();
			GraphicsConfiguration gc = g2d.getDeviceConfiguration();
			int dpi = SheetPreferences.getPNGResolution();
			PrintManager settings = mCharacter.getPageSettings();
			PageFormat format = settings != null ? settings.createPageFormat() : createDefaultPageFormat();
			Paper paper = format.getPaper();
			int width = (int) (paper.getWidth() / 72.0 * dpi);
			int height = (int) (paper.getHeight() / 72.0 * dpi);
			BufferedImage buffer = gc.createCompatibleImage(width, height, Transparency.OPAQUE);
			int pageNum = 0;
			String name = Path.getLeafName(file.getName(), false);

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
				AppWindow window = (AppWindow) getTopLevelAncestor();
				if (window != null) {
					window.setPrinting(true);
				}
				print(g2d, format, pageNum++);
				if (window != null) {
					window.setPrinting(false);
				}
				g2d.dispose();
				pngFile = new File(file, name + (pageNum > 1 ? " " + pageNum : "") + SheetWindow.PNG_EXTENSION); //$NON-NLS-1$ //$NON-NLS-2$
				if (!Images.writePNG(pngFile, buffer, dpi)) {
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
