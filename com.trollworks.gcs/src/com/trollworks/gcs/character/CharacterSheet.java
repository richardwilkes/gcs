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

import com.lowagie.text.pdf.DefaultFontMapper;
import com.lowagie.text.pdf.PdfContentByte;
import com.lowagie.text.pdf.PdfTemplate;
import com.lowagie.text.pdf.PdfWriter;
import com.trollworks.gcs.GCS;
import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.advantage.AdvantageOutline;
import com.trollworks.gcs.equipment.Equipment;
import com.trollworks.gcs.equipment.EquipmentColumn;
import com.trollworks.gcs.equipment.EquipmentOutline;
import com.trollworks.gcs.modifier.EquipmentModifier;
import com.trollworks.gcs.notes.Note;
import com.trollworks.gcs.notes.NoteOutline;
import com.trollworks.gcs.page.Page;
import com.trollworks.gcs.page.PageField;
import com.trollworks.gcs.page.PageOwner;
import com.trollworks.gcs.preferences.DisplayPreferences;
import com.trollworks.gcs.preferences.OutputPreferences;
import com.trollworks.gcs.skill.Skill;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.SkillDefaultType;
import com.trollworks.gcs.skill.SkillOutline;
import com.trollworks.gcs.spell.Spell;
import com.trollworks.gcs.spell.SpellOutline;
import com.trollworks.gcs.ui.Fonts;
import com.trollworks.gcs.ui.GraphicsUtilities;
import com.trollworks.gcs.ui.Selection;
import com.trollworks.gcs.ui.UIUtilities;
import com.trollworks.gcs.ui.image.Img;
import com.trollworks.gcs.ui.layout.ColumnLayout;
import com.trollworks.gcs.ui.layout.RowDistribution;
import com.trollworks.gcs.ui.print.PrintManager;
import com.trollworks.gcs.ui.scale.Scale;
import com.trollworks.gcs.ui.scale.ScaleRoot;
import com.trollworks.gcs.ui.scale.Scales;
import com.trollworks.gcs.ui.widget.Wrapper;
import com.trollworks.gcs.ui.widget.dock.Dockable;
import com.trollworks.gcs.ui.widget.outline.Column;
import com.trollworks.gcs.ui.widget.outline.ListRow;
import com.trollworks.gcs.ui.widget.outline.Outline;
import com.trollworks.gcs.ui.widget.outline.OutlineHeader;
import com.trollworks.gcs.ui.widget.outline.OutlineModel;
import com.trollworks.gcs.ui.widget.outline.OutlineSyncer;
import com.trollworks.gcs.ui.widget.outline.Row;
import com.trollworks.gcs.ui.widget.outline.RowIterator;
import com.trollworks.gcs.ui.widget.outline.RowSelection;
import com.trollworks.gcs.utility.Dice;
import com.trollworks.gcs.utility.FileType;
import com.trollworks.gcs.utility.I18n;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.PathUtils;
import com.trollworks.gcs.utility.Preferences;
import com.trollworks.gcs.utility.PrintProxy;
import com.trollworks.gcs.utility.notification.BatchNotifierTarget;
import com.trollworks.gcs.utility.notification.NotifierTarget;
import com.trollworks.gcs.utility.text.DateTimeFormatter;
import com.trollworks.gcs.utility.text.Numbers;
import com.trollworks.gcs.utility.undo.StdUndoManager;
import com.trollworks.gcs.weapon.MeleeWeaponStats;
import com.trollworks.gcs.weapon.RangedWeaponStats;
import com.trollworks.gcs.weapon.WeaponDamage;
import com.trollworks.gcs.weapon.WeaponDisplayRow;
import com.trollworks.gcs.weapon.WeaponSTDamage;
import com.trollworks.gcs.weapon.WeaponStats;

import java.awt.Color;
import java.awt.Component;
import java.awt.Container;
import java.awt.Dimension;
import java.awt.EventQueue;
import java.awt.Font;
import java.awt.FontMetrics;
import java.awt.Frame;
import java.awt.Graphics;
import java.awt.Graphics2D;
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
import java.awt.print.PageFormat;
import java.awt.print.Paper;
import java.io.File;
import java.io.FileOutputStream;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;
import javax.imageio.ImageIO;
import javax.swing.JPanel;
import javax.swing.RepaintManager;
import javax.swing.Scrollable;
import javax.swing.SwingConstants;
import javax.swing.UIManager;
import javax.swing.event.ChangeEvent;
import javax.swing.event.ChangeListener;

/** The character sheet. */
public class CharacterSheet extends JPanel implements ChangeListener, Scrollable, BatchNotifierTarget, PageOwner, PrintProxy, ActionListener, Runnable, DropTargetListener, ScaleRoot {
    private static final String           BOXING_SKILL_NAME   = "Boxing";
    private static final String           KARATE_SKILL_NAME   = "Karate";
    private static final String           BRAWLING_SKILL_NAME = "Brawling";
    private static final int              GAP                 = 2;
    private static final String           MELEE_KEY           = "melee";
    private static final String           RANGED_KEY          = "ranged";
    private static final String           ADVANTAGES_KEY      = "advantages";
    private static final String           SKILLS_KEY          = "skills";
    private static final String           SPELLS_KEY          = "spells";
    private static final String           EQUIPMENT_KEY       = "equipment";
    private static final String           OTHER_EQUIPMENT_KEY = "other_equipment";
    private static final String           NOTES_KEY           = "notes";
    private static final String[]         ALL_KEYS            = {MELEE_KEY, RANGED_KEY, ADVANTAGES_KEY, SKILLS_KEY, SPELLS_KEY, EQUIPMENT_KEY, OTHER_EQUIPMENT_KEY, NOTES_KEY};
    private              Scale            mScale;
    private              GURPSCharacter   mCharacter;
    private              int              mLastPage;
    private              boolean          mBatchMode;
    private              AdvantageOutline mAdvantageOutline;
    private              SkillOutline     mSkillOutline;
    private              SpellOutline     mSpellOutline;
    private              EquipmentOutline mEquipmentOutline;
    private              EquipmentOutline mOtherEquipmentOutline;
    private              NoteOutline      mNoteOutline;
    private              Outline          mMeleeWeaponOutline;
    private              Outline          mRangedWeaponOutline;
    private              boolean          mRebuildPending;
    private              Set<Outline>     mRootsToSync;
    private              PrintManager     mPrintManager;
    private              Scale            mSavedScale;
    private              boolean          mOkToPaint          = true;
    private              boolean          mIsPrinting;
    private              boolean          mSyncWeapons;
    private              boolean          mDisposed;

    /**
     * Creates a new character sheet display. {@link #rebuild()} must be called prior to the first
     * display of this panel.
     *
     * @param character The character to display the data for.
     */
    public CharacterSheet(GURPSCharacter character) {
        setLayout(new CharacterSheetLayout(this));
        setOpaque(false);
        mScale = DisplayPreferences.initialUIScale().getScale();
        mCharacter = character;
        mLastPage = -1;
        mRootsToSync = new HashSet<>();
        if (!GraphicsUtilities.inHeadlessPrintMode()) {
            setDropTarget(new DropTarget(this, this));
        }
        mCharacter.addTarget(this, FEATURES_AND_PREREQS_NOTIFICATIONS.toArray(new String[0]));
        mCharacter.addTarget(this, Settings.PREFIX);
        Preferences.getInstance().getNotifier().add(this, Fonts.FONT_NOTIFICATION_KEY);
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

    @Override
    public Scale getScale() {
        return mScale;
    }

    @Override
    public void setScale(Scale scale) {
        if (mScale.getScale() != scale.getScale()) {
            mScale = scale;
            markForRebuild();
        }
    }

    /** @return Whether a rebuild is pending. */
    public boolean isRebuildPending() {
        return mRebuildPending;
    }

    /** Synchronizes the display with the underlying model. */
    public void rebuild() {
        KeyboardFocusManager focusMgr = KeyboardFocusManager.getCurrentKeyboardFocusManager();
        Component            focus    = focusMgr.getPermanentFocusOwner();
        int                  firstRow = 0;
        String               focusKey = null;
        PageAssembler        pageAssembler;

        if (UIUtilities.getSelfOrAncestorOfType(focus, CharacterSheet.class) == this) {
            if (focus instanceof PageField) {
                focusKey = ((PageField) focus).getConsumedType();
                focus = null;
            } else if (focus instanceof Outline) {
                Outline   outline   = (Outline) focus;
                Selection selection = outline.getModel().getSelection();

                firstRow = outline.getFirstRowToDisplay();
                int selRow = selection.nextSelectedIndex(firstRow);
                if (selRow >= 0) {
                    firstRow = selRow;
                }
                focus = outline.getRealOutline();
            }
            focusMgr.clearFocusOwner();
        } else {
            focus = null;
        }

        // Make sure our primary outlines exist
        createAdvantageOutline();
        createSkillOutline();
        createSpellOutline();
        createMeleeWeaponOutline();
        createRangedWeaponOutline();
        createEquipmentOutline();
        createOtherEquipmentOutline();
        createNoteOutline();

        // Clear out the old pages
        removeAll();
        List<NotifierTarget> targets = new ArrayList<>();
        targets.add(this);
        SheetDockable sheetDockable = UIUtilities.getAncestorOfType(this, SheetDockable.class);
        if (sheetDockable != null) {
            targets.add(sheetDockable);
        }
        mCharacter.resetNotifier(targets.toArray(new NotifierTarget[0]));

        // Create the first page, which holds stuff that has a fixed vertical size.
        pageAssembler = new PageAssembler(this);
        pageAssembler.addToContent(hwrap(new PortraitPanel(this), vwrap(hwrap(new IdentityPanel(this), new MiscPanel(this)), new DescriptionPanel(this)), new PointsPanel(this)), null, null);
        pageAssembler.addToContent(hwrap(new AttributesPanel(this), vwrap(new EncumbrancePanel(this), new LiftPanel(this)), new HitLocationPanel(this), new HitPointsPanel(this)), null, null);

        // Add the various outline blocks, based on the layout preference.
        Set<String> remaining   = prepBlockLayoutRemaining();
        String      blockLayout = mCharacter.getSettings().blockLayout().toLowerCase().trim().replaceAll("\n+", "\n").replaceAll(" +", " ");
        for (String line : blockLayout.split("\n")) {
            String[] parts = line.trim().split(" ");
            if (!parts[0].isEmpty() && remaining.contains(parts[0])) {
                Outline o1 = getOutlineForKey(parts[0]);
                if (o1 != null) {
                    String t1 = getOutlineTitleForKey(parts[0]);
                    remaining.remove(parts[0]);
                    if (parts.length > 1 && remaining.contains(parts[1])) {
                        Outline o2 = getOutlineForKey(parts[1]);
                        if (o2 != null) {
                            String t2 = getOutlineTitleForKey(parts[1]);
                            remaining.remove(parts[1]);
                            if (o1.getModel().getRowCount() > 0 && o2.getModel().getRowCount() > 0) {
                                addOutline(pageAssembler, o1, t1, o2, t2);
                            } else {
                                addOutline(pageAssembler, o1, t1);
                                addOutline(pageAssembler, o2, t2);
                            }
                            continue;
                        }
                    }
                    addOutline(pageAssembler, o1, t1);
                }
            }
        }
        for (String one : ALL_KEYS) {
            if (remaining.contains(one)) {
                Outline outline = getOutlineForKey(one);
                if (outline != null) {
                    addOutline(pageAssembler, outline, getOutlineTitleForKey(one));
                }
            }
        }
        pageAssembler.finish();

        // Ensure everything is laid out and register for notification
        validate();
        OutlineSyncer.remove(mMeleeWeaponOutline);
        OutlineSyncer.remove(mRangedWeaponOutline);
        OutlineSyncer.remove(mAdvantageOutline);
        OutlineSyncer.remove(mSkillOutline);
        OutlineSyncer.remove(mSpellOutline);
        OutlineSyncer.remove(mEquipmentOutline);
        OutlineSyncer.remove(mOtherEquipmentOutline);
        OutlineSyncer.remove(mNoteOutline);
        mCharacter.addTarget(this, GURPSCharacter.CHARACTER_PREFIX);
        mCharacter.calculateWeightAndWealthCarried(true);
        mCharacter.calculateWealthNotCarried(true);
        if (focusKey != null) {
            restoreFocusToKey(focusKey, this);
        } else if (focus instanceof Outline) {
            ((Outline) focus).getBestOutlineForRowIndex(firstRow).requestFocusInWindow();
        } else if (focus != null) {
            focus.requestFocusInWindow();
        }
        setSize(getPreferredSize());
        repaint();
    }

    private static Set<String> prepBlockLayoutRemaining() {
        Set<String> remaining = new HashSet<>();
        remaining.add(MELEE_KEY);
        remaining.add(RANGED_KEY);
        remaining.add(ADVANTAGES_KEY);
        remaining.add(SKILLS_KEY);
        remaining.add(SPELLS_KEY);
        remaining.add(EQUIPMENT_KEY);
        remaining.add(OTHER_EQUIPMENT_KEY);
        remaining.add(NOTES_KEY);
        return remaining;
    }

    public String getHTMLGridTemplate() {
        Set<String>   remaining   = prepBlockLayoutRemaining();
        String        blockLayout = mCharacter.getSettings().blockLayout().toLowerCase().trim().replaceAll("\n+", "\n").replaceAll(" +", " ");
        StringBuilder buffer      = new StringBuilder();
        for (String line : blockLayout.split("\n")) {
            String[] parts = line.trim().split(" ");
            if (!parts[0].isEmpty() && remaining.contains(parts[0])) {
                remaining.remove(parts[0]);
                if (parts.length > 1 && remaining.contains(parts[1])) {
                    remaining.remove(parts[1]);
                    appendToGridTemplate(buffer, parts[0], parts[1]);
                    continue;
                }
                appendToGridTemplate(buffer, parts[0], parts[0]);
            }
        }
        for (String one : ALL_KEYS) {
            if (remaining.contains(one)) {
                appendToGridTemplate(buffer, one, one);
            }
        }
        return buffer.toString();
    }

    private static void appendToGridTemplate(StringBuilder buffer, String col1, String col2) {
        buffer.append('"');
        buffer.append(col1);
        buffer.append(' ');
        buffer.append(col2);
        buffer.append('"');
        buffer.append('\n');
    }

    private static String getOutlineTitleForKey(String key) {
        switch (key) {
        case MELEE_KEY:
            return I18n.Text("Melee Weapons");
        case RANGED_KEY:
            return I18n.Text("Ranged Weapons");
        case ADVANTAGES_KEY:
            return I18n.Text("Advantages, Disadvantages & Quirks");
        case SKILLS_KEY:
            return I18n.Text("Skills");
        case SPELLS_KEY:
            return I18n.Text("Spells");
        case EQUIPMENT_KEY:
            return I18n.Text("Equipment");
        case OTHER_EQUIPMENT_KEY:
            return I18n.Text("Other Equipment");
        case NOTES_KEY:
            return I18n.Text("Notes");
        default:
            return "";
        }
    }

    private Outline getOutlineForKey(String key) {
        switch (key) {
        case MELEE_KEY:
            return mMeleeWeaponOutline;
        case RANGED_KEY:
            return mRangedWeaponOutline;
        case ADVANTAGES_KEY:
            return mAdvantageOutline;
        case SKILLS_KEY:
            return mSkillOutline;
        case SPELLS_KEY:
            return mSpellOutline;
        case EQUIPMENT_KEY:
            return mEquipmentOutline;
        case OTHER_EQUIPMENT_KEY:
            return mOtherEquipmentOutline;
        case NOTES_KEY:
            return mNoteOutline;
        default:
            return null;
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
            OutlineInfo info     = new OutlineInfo(outline, pageAssembler.getContentWidth());
            boolean     useProxy = false;

            while (pageAssembler.addToContent(new SingleOutlinePanel(mScale, outline, title, useProxy), info, null)) {
                if (!useProxy) {
                    title = MessageFormat.format(I18n.Text("{0} (continued)"), title);
                    useProxy = true;
                }
            }
        }
    }

    private void addOutline(PageAssembler pageAssembler, Outline leftOutline, String leftTitle, Outline rightOutline, String rightTitle) {
        int         width     = pageAssembler.getContentWidth() / 2 - 1;
        OutlineInfo infoLeft  = new OutlineInfo(leftOutline, width);
        OutlineInfo infoRight = new OutlineInfo(rightOutline, width);
        boolean     useProxy  = false;

        while (pageAssembler.addToContent(new DoubleOutlinePanel(mScale, leftOutline, leftTitle, rightOutline, rightTitle, useProxy), infoLeft, infoRight)) {
            if (!useProxy) {
                leftTitle = MessageFormat.format(I18n.Text("{0} (continued)"), leftTitle);
                rightTitle = MessageFormat.format(I18n.Text("{0} (continued)"), rightTitle);
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
        header.setTopDividerColor(Color.black);
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

    /** @return The outline containing the notes. */
    public NoteOutline getNoteOutline() {
        return mNoteOutline;
    }

    private void createNoteOutline() {
        if (mNoteOutline == null) {
            mNoteOutline = new NoteOutline(mCharacter);
            initOutline(mNoteOutline);
        } else {
            resetOutline(mNoteOutline);
        }
    }

    /** @return The outline containing the equipment. */
    public EquipmentOutline getEquipmentOutline() {
        return mEquipmentOutline;
    }

    private void createEquipmentOutline() {
        if (mEquipmentOutline == null) {
            mEquipmentOutline = new EquipmentOutline(mCharacter, mCharacter.getEquipmentRoot());
            initOutline(mEquipmentOutline);
        } else {
            resetOutline(mEquipmentOutline);
        }
    }

    /** @return The outline containing the other equipment. */
    public EquipmentOutline getOtherEquipmentOutline() {
        return mOtherEquipmentOutline;
    }

    private void createOtherEquipmentOutline() {
        if (mOtherEquipmentOutline == null) {
            mOtherEquipmentOutline = new EquipmentOutline(mCharacter, mCharacter.getOtherEquipmentRoot());
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
            String       sortConfig;

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
            String       sortConfig;

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

    private void addBuiltInWeapons(Class<? extends WeaponStats> weaponClass, Map<HashedWeapon, WeaponDisplayRow> map) {
        if (weaponClass == MeleeWeaponStats.class) {
            boolean            savedModified = mCharacter.isModified();
            long savedModifiedOn = mCharacter.getModifiedOn();
            List<SkillDefault> defaults      = new ArrayList<>();
            Advantage          phantom;
            MeleeWeaponStats   weapon;

            StdUndoManager mgr = mCharacter.getUndoManager();
            mCharacter.setUndoManager(new StdUndoManager());

            phantom = new Advantage(mCharacter, false);
            phantom.setName(I18n.Text("Natural"));

            if (mCharacter.includePunch()) {
                defaults.add(new SkillDefault(SkillDefaultType.DX, null, null, 0));
                defaults.add(new SkillDefault(SkillDefaultType.Skill, BOXING_SKILL_NAME, null, 0));
                defaults.add(new SkillDefault(SkillDefaultType.Skill, BRAWLING_SKILL_NAME, null, 0));
                defaults.add(new SkillDefault(SkillDefaultType.Skill, KARATE_SKILL_NAME, null, 0));
                weapon = new MeleeWeaponStats(phantom);
                weapon.setUsage(I18n.Text("Punch"));
                weapon.setDefaults(defaults);
                WeaponDamage damage = new WeaponDamage(weapon);
                damage.setWeaponSTDamage(WeaponSTDamage.THR);
                damage.setBase(new Dice(0, -1));
                damage.setType("cr");
                weapon.setDamage(damage); // thr-1 cr
                weapon.setReach("C");
                weapon.setParry("0");
                map.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
                defaults.clear();
            }

            defaults.add(new SkillDefault(SkillDefaultType.DX, null, null, -2));
            defaults.add(new SkillDefault(SkillDefaultType.Skill, BRAWLING_SKILL_NAME, null, -2));
            defaults.add(new SkillDefault(SkillDefaultType.Skill, KARATE_SKILL_NAME, null, -2));

            if (mCharacter.includeKick()) {
                weapon = new MeleeWeaponStats(phantom);
                weapon.setUsage(I18n.Text("Kick"));
                weapon.setDefaults(defaults);
                WeaponDamage damage = new WeaponDamage(weapon);
                damage.setWeaponSTDamage(WeaponSTDamage.THR);
                damage.setType("cr");
                weapon.setDamage(damage); // thr cr
                weapon.setReach("C,1");
                weapon.setParry("No");
                map.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
            }

            if (mCharacter.includeKickBoots()) {
                weapon = new MeleeWeaponStats(phantom);
                weapon.setUsage(I18n.Text("Kick w/Boots"));
                weapon.setDefaults(defaults);
                WeaponDamage damage = new WeaponDamage(weapon);
                damage.setWeaponSTDamage(WeaponSTDamage.THR);
                damage.setBase(new Dice(0, 1));
                damage.setType("cr");
                weapon.setDamage(damage); // thr+1 cr
                weapon.setReach("C,1");
                weapon.setParry("No");
                map.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
            }

            mCharacter.setUndoManager(mgr);
            mCharacter.setModifiedOn(savedModifiedOn);
            mCharacter.setModified(savedModified);
        }
    }

    private List<WeaponDisplayRow> collectWeapons(Class<? extends WeaponStats> weaponClass) {
        Map<HashedWeapon, WeaponDisplayRow> weaponMap = new HashMap<>();
        List<WeaponDisplayRow>              weaponList;

        addBuiltInWeapons(weaponClass, weaponMap);

        for (Advantage advantage : mCharacter.getAdvantagesIterator(false)) {
            for (WeaponStats weapon : advantage.getWeapons()) {
                if (weaponClass.isInstance(weapon)) {
                    weaponMap.put(new HashedWeapon(weapon), new WeaponDisplayRow(weapon));
                }
            }
        }

        for (Equipment equipment : mCharacter.getEquipmentIterator()) {
            if (equipment.getQuantity() > 0 && equipment.isEquipped()) {
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

        weaponList = new ArrayList<>(weaponMap.values());
        return weaponList;
    }

    private void initOutline(Outline outline) {
        outline.addActionListener(this);
    }

    private static void resetOutline(Outline outline) {
        outline.clearProxies();
    }

    private static Container hwrap(Component left, Component right) {
        Wrapper wrapper = new Wrapper(new ColumnLayout(2, GAP, GAP));
        wrapper.add(left);
        wrapper.add(right);
        wrapper.setAlignmentY(-1.0f);
        return wrapper;
    }

    private static Container hwrap(Component left, Component center, Component right) {
        Wrapper wrapper = new Wrapper(new ColumnLayout(3, GAP, GAP));
        wrapper.add(left);
        wrapper.add(center);
        wrapper.add(right);
        wrapper.setAlignmentY(-1.0f);
        return wrapper;
    }

    private static Container hwrap(Component left, Component center, Component center2, Component right) {
        Wrapper wrapper = new Wrapper(new ColumnLayout(4, GAP, GAP));
        wrapper.add(left);
        wrapper.add(center);
        wrapper.add(center2);
        wrapper.add(right);
        wrapper.setAlignmentY(-1.0f);
        return wrapper;
    }

    private static Container vwrap(Component top, Component bottom) {
        Wrapper wrapper = new Wrapper(new ColumnLayout(1, GAP, GAP, RowDistribution.GIVE_EXCESS_TO_LAST));
        wrapper.add(top);
        wrapper.add(bottom);
        wrapper.setAlignmentY(-1.0f);
        return wrapper;
    }

    /** @return The number of pages in this character sheet. */
    public int getPageCount() {
        return getComponentCount();
    }

    @Override
    public void paint(Graphics g) {
        if (mOkToPaint) {
            super.paint(g);
        }
    }

    @Override
    public int print(Graphics graphics, PageFormat pageFormat, int pageIndex) {
        if (pageIndex >= getComponentCount()) {
            mLastPage = -1;
            return NO_SUCH_PAGE;
        }

        // We do the following trick to avoid going through the work twice,
        // as we are called twice for each page, the first of which doesn't
        // seem to be used.
        if (mLastPage == pageIndex) {
            Component      comp  = getComponent(pageIndex);
            RepaintManager mgr   = RepaintManager.currentManager(comp);
            boolean        saved = mgr.isDoubleBufferingEnabled();
            mgr.setDoubleBufferingEnabled(false);
            mOkToPaint = true;
            comp.print(graphics);
            mOkToPaint = false;
            mgr.setDoubleBufferingEnabled(saved);
        } else {
            mLastPage = pageIndex;
        }
        return PAGE_EXISTS;
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

    private static Set<String> MARK_FOR_REBUILD_NOTIFICATIONS        = new HashSet<>();
    private static Set<String> MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS = new HashSet<>();
    private static Set<String> FEATURES_AND_PREREQS_NOTIFICATIONS    = new HashSet<>();

    static {
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_BLOCK_LAYOUT);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_DEFAULT_WEIGHT_UNITS);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Fonts.FONT_NOTIFICATION_KEY);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Profile.ID_BODY_TYPE);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_SIMPLE_METRIC_CONVERSIONS);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_MODIFYING_DICE_PLUS_ADDS);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_REDUCED_SWING);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_KNOW_YOUR_OWN_STRENGTH);
        MARK_FOR_REBUILD_NOTIFICATIONS.add(Settings.ID_USE_THRUST_EQUALS_SWING_MINUS_2);

        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Advantage.ID_DISABLED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Advantage.ID_WEAPON_STATUS_CHANGED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Equipment.ID_EQUIPPED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Equipment.ID_QUANTITY);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Equipment.ID_WEAPON_STATUS_CHANGED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(GURPSCharacter.ID_INCLUDE_BOOTS);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(GURPSCharacter.ID_INCLUDE_KICK);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(GURPSCharacter.ID_INCLUDE_PUNCH);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Skill.ID_WEAPON_STATUS_CHANGED);
        MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.add(Spell.ID_WEAPON_STATUS_CHANGED);

        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Advantage.ID_LEVELS);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Advantage.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Advantage.ID_NAME);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_EQUIPPED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_EXTENDED_WEIGHT);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Equipment.ID_QUANTITY);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(EquipmentModifier.ID_COST_ADJ);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(EquipmentModifier.ID_WEIGHT_ADJ);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_DEXTERITY);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_HEALTH);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_INTELLIGENCE);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_PERCEPTION);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_STRENGTH);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(GURPSCharacter.ID_WILL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Profile.ID_TECH_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_ENCUMBRANCE_PENALTY);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_NAME);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_POINTS);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_RELATIVE_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_SPECIALIZATION);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Skill.ID_TECH_LEVEL);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_COLLEGE);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_LIST_CHANGED);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_NAME);
        FEATURES_AND_PREREQS_NOTIFICATIONS.add(Spell.ID_POINTS);
    }

    @Override
    public void handleNotification(Object producer, String type, Object data) {
        if (MARK_FOR_REBUILD_NOTIFICATIONS.contains(type)) {
            markForRebuild();
        } else {
            if (type.startsWith(Advantage.PREFIX)) {
                OutlineSyncer.add(mAdvantageOutline);
            } else if (Settings.ID_USER_DESCRIPTION_DISPLAY.equals(type) || Settings.ID_MODIFIERS_DISPLAY.equals(type) || Settings.ID_NOTES_DISPLAY.equals(type)) {
                OutlineSyncer.add(mAdvantageOutline);
                OutlineSyncer.add(mSkillOutline);
                OutlineSyncer.add(mSpellOutline);
                OutlineSyncer.add(mEquipmentOutline);
                OutlineSyncer.add(mOtherEquipmentOutline);
                mSyncWeapons = true;
                markForRebuild();
            } else if (type.startsWith(Skill.PREFIX)) {
                OutlineSyncer.add(mSkillOutline);
            } else if (type.startsWith(Spell.PREFIX)) {
                OutlineSyncer.add(mSpellOutline);
            } else if (type.startsWith(Equipment.PREFIX)) {
                OutlineSyncer.add(mEquipmentOutline);
                OutlineSyncer.add(mOtherEquipmentOutline);
                mSyncWeapons = true;
                markForRebuild();
            } else if (type.startsWith(Note.PREFIX)) {
                OutlineSyncer.add(mNoteOutline);
            }

            if (MARK_FOR_WEAPON_REBUILD_NOTIFICATIONS.contains(type)) {
                mSyncWeapons = true;
                markForRebuild();
            } else if (GURPSCharacter.ID_PARRY_BONUS.equals(type) || Skill.ID_LEVEL.equals(type)) {
                OutlineSyncer.add(mMeleeWeaponOutline);
                OutlineSyncer.add(mRangedWeaponOutline);
            } else if (GURPSCharacter.ID_CARRIED_WEIGHT.equals(type) || GURPSCharacter.ID_CARRIED_WEALTH.equals(type)) {
                Column column = mEquipmentOutline.getModel().getColumnWithID(EquipmentColumn.DESCRIPTION.ordinal());
                column.setName(EquipmentColumn.DESCRIPTION.toString(mCharacter, true));
            } else if (GURPSCharacter.ID_NOT_CARRIED_WEALTH.equals(type)) {
                Column column = mOtherEquipmentOutline.getModel().getColumnWithID(EquipmentColumn.DESCRIPTION.ordinal());
                column.setName(EquipmentColumn.DESCRIPTION.toString(mCharacter, false));
            } else if (Settings.ID_BASE_WILL_AND_PER_ON_10.equals(type)) {
                mCharacter.updateWillAndPerceptionDueToOptionalIQRuleUseChange();
            } else if (Settings.ID_USE_MULTIPLICATIVE_MODIFIERS.equals(type)) {
                mCharacter.notifySingle(Advantage.ID_LIST_CHANGED, null);
            } else if (Settings.ID_USE_KNOW_YOUR_OWN_STRENGTH.equals(type)) {
                mCharacter.notifySingle(type, data);
            } else if (Settings.ID_USE_THRUST_EQUALS_SWING_MINUS_2.equals(type)) {
                mCharacter.notifySingle(type, data);
            } else if (GURPSCharacter.ID_MODIFIED.equals(type)) {
                int count = getComponentCount();
                for (int i = 0; i < count; i++) {
                    Page      page   = (Page) getComponent(i);
                    Rectangle bounds = page.getBounds();
                    Insets    insets = page.getInsets();
                    bounds.y = bounds.y + bounds.height - insets.bottom;
                    bounds.height = insets.bottom;
                    repaint(bounds);
                }
            }
        }
        if (FEATURES_AND_PREREQS_NOTIFICATIONS.contains(type)) {
            if (mCharacter.processFeaturesAndPrereqs()) {
                repaint();
            }
        }
        if (!mBatchMode && !mRebuildPending) {
            validate();
        }
    }

    @Override
    public void drawPageAdornments(Page page, Graphics gc) {
        Rectangle bounds = page.getBounds();
        Insets    insets = page.getInsets();
        bounds.width -= insets.left + insets.right;
        bounds.height -= insets.top + insets.bottom;
        bounds.x = insets.left;
        bounds.y = insets.top;
        int         pageNumber = 1 + UIUtilities.getIndexOf(this, page);
        String      pageString = MessageFormat.format(I18n.Text("Page {0} of {1}"), Numbers.format(pageNumber), Numbers.format(getPageCount()));
        Font        font1      = mScale.scale(UIManager.getFont(Fonts.KEY_SECONDARY_FOOTER));
        Font        font2      = mScale.scale(UIManager.getFont(Fonts.KEY_PRIMARY_FOOTER));
        FontMetrics fm1        = gc.getFontMetrics(font1);
        FontMetrics fm2        = gc.getFontMetrics(font2);
        int         y          = bounds.y + bounds.height + fm2.getAscent();
        String      modified   = String.format("Modified %s", DateTimeFormatter.getFormattedDateTime(mCharacter.getModifiedOn()));
        String      left;
        String      right;

        if ((pageNumber & 1) == 1) {
            left = GCS.COPYRIGHT_FOOTER;
            right = modified;
        } else {
            left = modified;
            right = GCS.COPYRIGHT_FOOTER;
        }

        Font savedFont = gc.getFont();
        gc.setColor(Color.BLACK);
        gc.setFont(font1);
        gc.drawString(left, bounds.x, y);
        gc.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, gc).getWidth(), y);
        gc.setFont(font2);
        String center = mCharacter.getProfile().getName();
        gc.drawString(center, bounds.x + (bounds.width - (int) fm2.getStringBounds(center, gc).getWidth()) / 2, y);

        String allRightsReserved = I18n.Text("All rights reserved");
        if ((pageNumber & 1) == 1) {
            left = allRightsReserved;
            right = pageString;
        } else {
            left = pageString;
            right = allRightsReserved;
        }

        y += fm2.getDescent() + fm1.getAscent();

        gc.setFont(font1);
        gc.drawString(left, bounds.x, y);
        gc.drawString(right, bounds.x + bounds.width - (int) fm1.getStringBounds(right, gc).getWidth(), y);
        // Trim off the leading URI scheme and authority path component. (http://, https://, ...)
        String webSite = GCS.WEB_SITE.replaceAll(".*://", "");
        gc.drawString(webSite, bounds.x + (bounds.width - (int) fm1.getStringBounds(webSite, gc).getWidth()) / 2, y);

        gc.setFont(savedFont);
    }

    @Override
    public Insets getPageAdornmentsInsets(Page page) {
        FontMetrics fm1 = Fonts.getFontMetrics(UIManager.getFont(Fonts.KEY_SECONDARY_FOOTER));
        FontMetrics fm2 = Fonts.getFontMetrics(UIManager.getFont(Fonts.KEY_PRIMARY_FOOTER));
        return new Insets(0, 0, fm1.getAscent() + fm1.getDescent() + fm2.getAscent() + fm2.getDescent(), 0);
    }

    @Override
    public PrintManager getPageSettings() {
        return mCharacter.getPageSettings();
    }

    /** @return The character being displayed. */
    public GURPSCharacter getCharacter() {
        return mCharacter;
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        String command = event.getActionCommand();
        if (Outline.CMD_POTENTIAL_CONTENT_SIZE_CHANGE.equals(command)) {
            mRootsToSync.add(((Outline) event.getSource()).getRealOutline());
            markForRebuild();
        }
    }

    /** Marks the sheet for a rebuild in the near future. */
    public void markForRebuild() {
        if (!mRebuildPending) {
            mRebuildPending = true;
            EventQueue.invokeLater(this);
        }
    }

    @Override
    public void run() {
        syncRoots();
        rebuild();
        mRebuildPending = false;
    }

    private void syncRoots() {
        if (mSyncWeapons || mRootsToSync.contains(mEquipmentOutline) || mRootsToSync.contains(mAdvantageOutline) || mRootsToSync.contains(mSpellOutline) || mRootsToSync.contains(mSkillOutline)) {
            OutlineModel outlineModel = mMeleeWeaponOutline.getModel();
            String       sortConfig   = outlineModel.getSortConfig();

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

    @Override
    public void stateChanged(ChangeEvent event) {
        Dimension size = getLayout().preferredLayoutSize(this);
        if (!getSize().equals(size)) {
            invalidate();
            repaint();
            setSize(size);
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
        UIUtilities.getAncestorOfType(this, SheetDockable.class).addRows(mDragRows);
        mDragRows = null;
        dtde.dropComplete(true);
    }

    @Override
    public void dragExit(DropTargetEvent dte) {
        mDragRows = null;
    }

    private static PageFormat createDefaultPageFormat() {
        Paper      paper  = new Paper();
        PageFormat format = new PageFormat();
        format.setOrientation(PageFormat.PORTRAIT);
        paper.setSize(8.5 * 72.0, 11.0 * 72.0);
        paper.setImageableArea(0.5 * 72.0, 0.5 * 72.0, 7.5 * 72.0, 10 * 72.0);
        format.setPaper(paper);
        return format;
    }

    private Set<Row> expandAllContainers() {
        Set<Row> changed = new HashSet<>();
        expandAllContainers(mCharacter.getAdvantagesIterator(true), changed);
        expandAllContainers(mCharacter.getSkillsIterator(), changed);
        expandAllContainers(mCharacter.getSpellsIterator(), changed);
        expandAllContainers(mCharacter.getEquipmentIterator(), changed);
        expandAllContainers(mCharacter.getOtherEquipmentIterator(), changed);
        if (mRebuildPending) {
            syncRoots();
            rebuild();
        }
        return changed;
    }

    private static void expandAllContainers(RowIterator<? extends Row> iterator, Set<Row> changed) {
        for (Row row : iterator) {
            if (!row.isOpen()) {
                row.setOpen(true);
                changed.add(row);
            }
        }
    }

    private void closeContainers(Set<Row> rows) {
        for (Row row : rows) {
            row.setOpen(false);
        }
        if (mRebuildPending) {
            syncRoots();
            rebuild();
        }
    }

    /**
     * @param file The file to save to.
     * @return {@code true} on success.
     */
    public boolean saveAsPDF(File file) {
        Set<Row> changed = expandAllContainers();
        try {
            PrintManager settings = mCharacter.getPageSettings();
            PageFormat   format   = settings != null ? settings.createPageFormat() : createDefaultPageFormat();
            Paper        paper    = format.getPaper();
            float        width    = (float) paper.getWidth();
            float        height   = (float) paper.getHeight();

            adjustToPageSetupChanges(true);
            setPrinting(true);

            com.lowagie.text.Document pdfDoc = new com.lowagie.text.Document(new com.lowagie.text.Rectangle(width, height));
            try (FileOutputStream out = new FileOutputStream(file)) {
                PdfWriter      writer  = PdfWriter.getInstance(pdfDoc, out);
                int            pageNum = 0;
                PdfContentByte cb;

                pdfDoc.open();
                cb = writer.getDirectContent();
                while (true) {
                    PdfTemplate template = cb.createTemplate(width, height);
                    Graphics2D  g2d      = template.createGraphics(width, height, new DefaultFontMapper());

                    if (print(g2d, format, pageNum) == NO_SUCH_PAGE) {
                        g2d.dispose();
                        break;
                    }
                    if (pageNum != 0) {
                        pdfDoc.newPage();
                    }
                    g2d.setClip(0, 0, (int) width, (int) height);
                    print(g2d, format, pageNum++);
                    g2d.dispose();
                    cb.addTemplate(template, 0, 0);
                }
                pdfDoc.close();
            }
            return true;
        } catch (Exception exception) {
            Log.error(exception);
            return false;
        } finally {
            setPrinting(false);
            closeContainers(changed);
        }
    }

    /**
     * @param file         The file to save to.
     * @param createdFiles The files that were created.
     * @return {@code true} on success.
     */
    public boolean saveAsPNG(File file, List<File> createdFiles) {
        Set<Row> changed = expandAllContainers();
        try {
            int          dpi      = OutputPreferences.getPNGResolution();
            PrintManager settings = mCharacter.getPageSettings();
            PageFormat   format   = settings != null ? settings.createPageFormat() : createDefaultPageFormat();
            Paper        paper    = format.getPaper();
            int          width    = (int) (paper.getWidth() / 72.0 * dpi);
            int          height   = (int) (paper.getHeight() / 72.0 * dpi);
            Img          buffer   = Img.create(width, height, Transparency.OPAQUE);
            int          pageNum  = 0;
            String       name     = PathUtils.getLeafName(file.getName(), false);

            file = file.getParentFile();

            adjustToPageSetupChanges(true);
            setPrinting(true);

            while (true) {
                File pngFile;

                Graphics2D gc = buffer.getGraphics();
                if (print(gc, format, pageNum) == NO_SUCH_PAGE) {
                    gc.dispose();
                    break;
                }
                gc.setClip(0, 0, width, height);
                gc.setBackground(Color.WHITE);
                gc.clearRect(0, 0, width, height);
                gc.scale(dpi / 72.0, dpi / 72.0);
                print(gc, format, pageNum++);
                gc.dispose();
                pngFile = new File(file, PathUtils.enforceExtension(name + (pageNum > 1 ? " " + pageNum : ""), FileType.PNG.getExtension()));
                ImageIO.write(buffer, "png", pngFile);
                createdFiles.add(pngFile);
            }
            return true;
        } catch (Exception exception) {
            Log.error(exception);
            return false;
        } finally {
            setPrinting(false);
            closeContainers(changed);
        }
    }

    @Override
    public int getNotificationPriority() {
        return 0;
    }

    @Override
    public PrintManager getPrintManager() {
        if (mPrintManager == null) {
            try {
                mPrintManager = mCharacter.getPageSettings();
            } catch (Exception exception) {
                // Ignore
            }
        }
        return mPrintManager;
    }

    @Override
    public String getPrintJobTitle() {
        Dockable dockable = UIUtilities.getAncestorOfType(this, Dockable.class);
        if (dockable != null) {
            return dockable.getTitle();
        }
        Frame frame = UIUtilities.getAncestorOfType(this, Frame.class);
        if (frame != null) {
            return frame.getTitle();
        }
        return mCharacter.getProfile().getName();
    }

    @Override
    public void adjustToPageSetupChanges(boolean willPrint) {
        OutputPreferences.setDefaultPageSettings(getPrintManager());
        if (willPrint) {
            mSavedScale = mScale;
            mScale = Scales.ACTUAL_SIZE.getScale();
            mOkToPaint = false;
        }
        rebuild();
    }

    @Override
    public boolean isPrinting() {
        return mIsPrinting;
    }

    @Override
    public void setPrinting(boolean printing) {
        mIsPrinting = printing;
        if (!printing) {
            mOkToPaint = true;
            if (mSavedScale != null && mSavedScale.getScale() != mScale.getScale()) {
                mScale = mSavedScale;
                rebuild();
            } else {
                repaint();
            }
        }
    }
}
