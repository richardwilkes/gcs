/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.ui.widget.outline;

import com.trollworks.gcs.advantage.Advantage;
import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.datafile.DataFile;
import com.trollworks.gcs.datafile.LoadState;
import com.trollworks.gcs.feature.AttributeBonus;
import com.trollworks.gcs.feature.ConditionalModifier;
import com.trollworks.gcs.feature.ContainedWeightReduction;
import com.trollworks.gcs.feature.CostReduction;
import com.trollworks.gcs.feature.DRBonus;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.ReactionBonus;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.gcs.feature.SkillPointBonus;
import com.trollworks.gcs.feature.SpellBonus;
import com.trollworks.gcs.feature.SpellPointBonus;
import com.trollworks.gcs.feature.WeaponDamageBonus;
import com.trollworks.gcs.modifier.Modifier;
import com.trollworks.gcs.prereq.PrereqList;
import com.trollworks.gcs.settings.SheetSettings;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.utility.Filtered;
import com.trollworks.gcs.utility.Log;
import com.trollworks.gcs.utility.SaveType;
import com.trollworks.gcs.utility.json.JsonArray;
import com.trollworks.gcs.utility.json.JsonMap;
import com.trollworks.gcs.utility.json.JsonWriter;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.TreeSet;
import java.util.UUID;
import javax.swing.Icon;

/** A common row super-class for the model. */
public abstract class ListRow extends Row {
    private static final int    LAST_CUSTOMIZABLE_HIT_LOCATIONS_VERSION = 1; // Last version before customizable hit locations (v4.29.1 and earlier)
    private static final String KEY_ID                                  = "id";
    private static final String KEY_OPEN                                = "open";
    private static final String KEY_NOTES                               = "notes";
    private static final String KEY_VTT_NOTES                           = "vtt_notes";
    private static final String KEY_CATEGORIES                          = "categories";
    private static final String KEY_FEATURES                            = "features";
    private static final String KEY_DEFAULTS                            = "defaults";
    private static final String KEY_CHILDREN                            = "children";
    private static final String KEY_PREREQS                             = "prereqs";

    protected DataFile           mDataFile;
    private   UUID               mID;
    private   List<Feature>      mFeatures;
    private   PrereqList         mPrereqList;
    private   List<SkillDefault> mDefaults;
    private   boolean            mIsSatisfied;
    private   String             mUnsatisfiedReason;
    private   String             mNotes;
    private   String             mVTTNotes;
    private   TreeSet<String>    mCategories;

    public static void saveList(JsonWriter w, String key, List<?> list, SaveType saveType) throws IOException {
        List<ListRow> rows = Filtered.list(list, ListRow.class);
        if (!rows.isEmpty()) {
            w.key(key);
            w.startArray();
            for (ListRow row : rows) {
                row.save(w, saveType);
            }
            w.endArray();
        }
    }

    /**
     * Extracts any "nameable" portions of the buffer and puts their keys into the provided set.
     *
     * @param set    The set to add the nameable keys to.
     * @param buffer The text to check for nameable portions.
     */
    public static void extractNameables(Set<String> set, String buffer) {
        int first = buffer.indexOf('@');
        int last  = buffer.indexOf('@', first + 1);

        while (first != -1 && last != -1) {
            set.add(buffer.substring(first + 1, last));
            first = buffer.indexOf('@', last + 1);
            last = buffer.indexOf('@', first + 1);
        }
    }

    /**
     * Names any "nameable" portions of the data and returns the resulting string.
     *
     * @param map  The map of nameable keys to names.
     * @param data The data to change.
     * @return The revised string.
     */
    public static String nameNameables(Map<String, String> map, String data) {
        int           first  = data.indexOf('@');
        int           last   = data.indexOf('@', first + 1);
        StringBuilder buffer = new StringBuilder();

        while (first != -1 && last != -1) {
            String key         = data.substring(first + 1, last);
            String replacement = map.get(key);

            if (first != 0) {
                buffer.append(data, 0, first);
            }
            if (replacement != null) {
                buffer.append(replacement);
            } else {
                buffer.append('@');
                buffer.append(key);
                buffer.append('@');
            }
            data = last + 1 == data.length() ? "" : data.substring(last + 1);
            first = data.indexOf('@');
            last = data.indexOf('@', first + 1);
        }
        buffer.append(data);
        return buffer.toString();
    }

    public static Set<String> createCategoriesList(String categories) {
        return new TreeSet<>(createList(categories));
    }

    // This is the decompose method that works with the compose method (getCategoriesAsString())
    private static Collection<String> createList(String categories) {
        return Arrays.asList(categories.split(","));
    }

    /**
     * Creates a new data row.
     *
     * @param dataFile    The data file to associate it with.
     * @param isContainer Whether or not this row allows children.
     */
    protected ListRow(DataFile dataFile, boolean isContainer) {
        setCanHaveChildren(isContainer);
        setOpen(isContainer);
        mDataFile = dataFile;
        mID = UUID.randomUUID();
        mFeatures = new ArrayList<>();
        mPrereqList = new PrereqList(null, true);
        mDefaults = new ArrayList<>();
        mIsSatisfied = true;
        mNotes = "";
        mVTTNotes = "";
        mCategories = new TreeSet<>();
    }

    /**
     * Creates a clone of an existing data row and associates it with the specified data file.
     *
     * @param dataFile   The data file to associate it with.
     * @param rowToClone The data row to clone.
     */
    protected ListRow(DataFile dataFile, ListRow rowToClone) {
        this(dataFile, rowToClone.canHaveChildren());
        setOpen(rowToClone.isOpen());
        mNotes = rowToClone.mNotes;
        mVTTNotes = rowToClone.mVTTNotes;
        for (Feature feature : rowToClone.mFeatures) {
            mFeatures.add(feature.cloneFeature());
        }
        mPrereqList = new PrereqList(null, rowToClone.getPrereqs());
        mDefaults = new ArrayList<>();
        for (SkillDefault skillDefault : rowToClone.mDefaults) {
            mDefaults.add(new SkillDefault(skillDefault));
        }
        mCategories = new TreeSet<>(rowToClone.mCategories);
    }

    /**
     * Clones this row.
     *
     * @param dataFile The {@link DataFile} the new row should be owned by.
     * @param deep     {@code true} if children should be cloned, too.
     * @param forSheet {@code true} if this is for a character sheet.
     * @return The newly created row.
     */
    public abstract ListRow cloneRow(DataFile dataFile, boolean deep, boolean forSheet);

    public UUID getID() {
        return mID;
    }

    /**
     * @param obj The other object to compare against.
     * @return Whether or not this ListRow is equivalent.
     */
    public boolean isEquivalentTo(Object obj) {
        if (obj == this) {
            return true;
        }
        if (obj instanceof ListRow row) {
            if (!mNotes.equals(row.mNotes)) {
                return false;
            }
            if (!mVTTNotes.equals(row.mVTTNotes)) {
                return false;
            }
            if (!mCategories.equals(row.mCategories)) {
                return false;
            }
            if (!mDefaults.equals(row.mDefaults)) {
                return false;
            }
            if (!mPrereqList.equals(row.mPrereqList)) {
                return false;
            }
            if (!mFeatures.equals(row.mFeatures)) {
                return false;
            }
            int childCount = getChildCount();
            if (childCount == row.getChildCount()) {
                for (int i = 0; i < childCount; i++) {
                    if (!((ListRow) getChild(i)).isEquivalentTo(row.getChild(i))) {
                        return false;
                    }
                }
                return true;
            }
        }
        return false;
    }

    /** @return Creates a detailed editor for this row. */
    public abstract RowEditor<? extends ListRow> createEditor();

    /** @return The localized name for this row object. */
    public abstract String getLocalizedName();

    @Override
    public boolean addChild(Row row) {
        boolean result = super.addChild(row);
        if (result) {
            notifyOfChange();
        }
        return result;
    }

    /** @return The type name to use for this data. */
    public abstract String getJSONTypeName();

    /** @return The type of row. */
    public abstract String getRowType();

    /** @return Whether or not this row's prerequisites are currently satisfied. */
    public boolean isSatisfied() {
        return mIsSatisfied;
    }

    /** @param satisfied Whether or not this row's prerequisites are currently satisfied. */
    public void setSatisfied(boolean satisfied) {
        mIsSatisfied = satisfied;
        if (satisfied) {
            mUnsatisfiedReason = null;
        }
    }

    /** @return The reason {@link #isSatisfied()} is returning {@code false}. */
    public String getReasonForUnsatisfied() {
        return mUnsatisfiedReason;
    }

    /** @param reason The reason {@link #isSatisfied()} is returning {@code false}. */
    public void setReasonForUnsatisfied(String reason) {
        mUnsatisfiedReason = reason;
    }

    /**
     * Loads this row's contents.
     *
     * @param dataFile The {@link DataFile} being loaded.
     * @param m        The {@link JsonMap} to load data from.
     * @param state    The {@link LoadState} to use.
     */
    public final void load(DataFile dataFile, JsonMap m, LoadState state) throws IOException {
        if (m.has(KEY_ID)) {
            try {
                mID = UUID.fromString(m.getString(KEY_ID));
            } catch (Exception exception) {
                mID = UUID.randomUUID();
            }
        }
        boolean isContainer = m.getString(DataFile.TYPE).endsWith("_container");
        setCanHaveChildren(isContainer);
        setOpen(isContainer);
        prepareForLoad(state);
        loadSelf(m, state);
        if (m.has(KEY_PREREQS)) {
            mPrereqList = new PrereqList(null, mDataFile.getSheetSettings().defaultWeightUnits(), m.getMap(KEY_PREREQS));
        }
        if (!(this instanceof Technique) && m.has(KEY_DEFAULTS)) {
            JsonArray a     = m.getArray(KEY_DEFAULTS);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mDefaults.add(new SkillDefault(a.getMap(i), false));
            }
        }
        if (m.has(KEY_FEATURES)) {
            JsonArray a     = m.getArray(KEY_FEATURES);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                JsonMap m1   = a.getMap(i);
                String  type = m1.getString(DataFile.TYPE);
                switch (type) {
                    case AttributeBonus.KEY_ROOT -> mFeatures.add(new AttributeBonus(dataFile, m1));
                    case DRBonus.KEY_ROOT -> loadDRBonus(dataFile, state, m1);
                    case ReactionBonus.KEY_ROOT -> mFeatures.add(new ReactionBonus(dataFile, m1));
                    case ConditionalModifier.KEY_ROOT -> mFeatures.add(new ConditionalModifier(dataFile, m1));
                    case SkillBonus.KEY_ROOT -> mFeatures.add(new SkillBonus(dataFile, m1));
                    case SkillPointBonus.KEY_ROOT -> mFeatures.add(new SkillPointBonus(dataFile, m1));
                    case SpellBonus.KEY_ROOT -> mFeatures.add(new SpellBonus(dataFile, m1));
                    case SpellPointBonus.KEY_ROOT -> mFeatures.add(new SpellPointBonus(dataFile, m1));
                    case WeaponDamageBonus.KEY_ROOT -> mFeatures.add(new WeaponDamageBonus(dataFile, m1));
                    case CostReduction.KEY_ROOT -> mFeatures.add(new CostReduction(m1));
                    case ContainedWeightReduction.KEY_ROOT -> mFeatures.add(new ContainedWeightReduction(m1));
                    default -> Log.warn("unknown feature type: " + type);
                }
            }
        }
        mNotes = m.getString(KEY_NOTES);
        mVTTNotes = m.getString(KEY_VTT_NOTES);
        if (m.has(KEY_CATEGORIES)) {
            JsonArray a     = m.getArray(KEY_CATEGORIES);
            int       count = a.size();
            for (int i = 0; i < count; i++) {
                mCategories.add(a.getString(i));
            }
        }
        if (canHaveChildren()) {
            setOpen(m.getBoolean(KEY_OPEN));
            if (m.has(KEY_CHILDREN)) {
                JsonArray a     = m.getArray(KEY_CHILDREN);
                int       count = a.size();
                for (int i = 0; i < count; i++) {
                    loadChild(a.getMap(i), state);
                }
            }
        }
        finishedLoading(state);
    }

    private void loadDRBonus(DataFile dataFile, LoadState state, JsonMap m) throws IOException {
        DRBonus bonus = new DRBonus(dataFile, m);
        if (state.mDataFileVersion <= LAST_CUSTOMIZABLE_HIT_LOCATIONS_VERSION) {
            switch (bonus.getLocation()) {
                case "eyes":
                    bonus.setLocation("eye");
                    break;
                case "arms":
                    bonus.setLocation("arm");
                    break;
                case "hands":
                    bonus.setLocation("hand");
                    break;
                case "legs":
                    bonus.setLocation("leg");
                    break;
                case "feet":
                    bonus.setLocation("foot");
                    break;
                case "wings":
                    bonus.setLocation("wing");
                    break;
                case "fins":
                    bonus.setLocation("fin");
                    break;
                case "torso":
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("vitals");
                    break;
                case "full_body":
                    bonus.setLocation("eye");
                    mFeatures.add(bonus);
                    // Intentional fall-through
                case "full_body_except_eyes":
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("skull");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("face");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("neck");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("torso");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("vitals");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("groin");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("arm");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("hand");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("leg");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("foot");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("tail");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("wing");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("fin");
                    mFeatures.add(bonus);
                    bonus = new DRBonus(bonus);
                    bonus.setLocation("brain");
                    break;
            }
        }
        mFeatures.add(bonus);
    }

    protected abstract void loadSelf(JsonMap m, LoadState state) throws IOException;

    protected abstract void loadChild(JsonMap m, LoadState state) throws IOException;

    /**
     * Called to prepare the row for loading.
     *
     * @param state The {@link LoadState} to use.
     */
    protected void prepareForLoad(LoadState state) {
        mNotes = "";
        mVTTNotes = "";
        mFeatures.clear();
        mDefaults.clear();
        mPrereqList = new PrereqList(null, true);
        mCategories.clear();
    }

    /**
     * Called when loading of this row is complete. Does nothing by default.
     *
     * @param state The {@link LoadState} to use.
     */
    protected void finishedLoading(LoadState state) {
        // Nothing to do.
    }

    /**
     * Saves the row.
     *
     * @param w        The {@link JsonWriter} to use.
     * @param saveType The type of save being performed.
     */
    public void save(JsonWriter w, SaveType saveType) throws IOException {
        w.startMap();
        w.keyValue(DataFile.TYPE, getJSONTypeName());
        w.keyValue(KEY_ID, mID.toString());
        saveSelf(w, saveType);
        if (!mPrereqList.isEmpty()) {
            w.key(KEY_PREREQS);
            mPrereqList.save(w);
        }
        if (!(this instanceof Technique) && !mDefaults.isEmpty()) {
            w.key(KEY_DEFAULTS);
            w.startArray();
            for (SkillDefault skillDefault : mDefaults) {
                skillDefault.save(w, false);
            }
            w.endArray();
        }
        if (!mFeatures.isEmpty()) {
            w.key(KEY_FEATURES);
            w.startArray();
            for (Feature feature : mFeatures) {
                feature.save(w);
            }
            w.endArray();
        }
        w.keyValueNot(KEY_NOTES, mNotes, "");
        w.keyValueNot(KEY_VTT_NOTES, mVTTNotes, "");
        if (!mCategories.isEmpty()) {
            w.key(KEY_CATEGORIES);
            w.startArray();
            for (String category : mCategories) {
                w.value(category);
            }
            w.endArray();
        }
        if (canHaveChildren()) {
            if (saveType != SaveType.HASH) {
                w.keyValue(KEY_OPEN, isOpen());
            }
            if (saveType != SaveType.UNDO) {
                saveList(w, KEY_CHILDREN, getChildren(), saveType);
            }
        }
        w.endMap();
    }

    /**
     * Saves the row.
     *
     * @param w        The {@link JsonWriter} to use.
     * @param saveType The type of save being performed.
     */
    protected abstract void saveSelf(JsonWriter w, SaveType saveType) throws IOException;

    public void notifyOfChange() {
        if (mDataFile != null) {
            mDataFile.notifyOfChange();
        }
    }

    /** Called to update any information that relies on children. */
    public void update() {
        // Do nothing by default.
    }

    /** @return The owning data file. */
    public DataFile getDataFile() {
        return mDataFile;
    }

    /** @return The owning template. */
    public Template getTemplate() {
        return mDataFile instanceof Template ? (Template) mDataFile : null;
    }

    /** @return The "secondary" text, the text display below the primary text, usually as a note. */
    protected String getSecondaryText() {
        StringBuilder builder  = new StringBuilder();
        SheetSettings settings = getDataFile().getSheetSettings();
        if (settings.modifiersDisplay().inline()) {
            String txt = getModifierNotes();
            if (!txt.isBlank()) {
                builder.append(txt);
            }
        }
        if (settings.notesDisplay().inline()) {
            String txt = getNotes();
            if (!txt.isBlank()) {
                if (!builder.isEmpty()) {
                    builder.append('\n');
                }
                builder.append(txt);
            }
        }
        return builder.toString();
    }

    /** @return The owning character. */
    public GURPSCharacter getCharacter() {
        return mDataFile instanceof GURPSCharacter ? (GURPSCharacter) mDataFile : null;
    }

    /** @return The features provided by this data row. */
    public List<Feature> getFeatures() {
        return Collections.unmodifiableList(mFeatures);
    }

    /**
     * @param features The new features of this data row.
     * @return Whether there was a change or not.
     */
    public boolean setFeatures(List<Feature> features) {
        if (!mFeatures.equals(features)) {
            mFeatures = new ArrayList<>(features);
            return true;
        }
        return false;
    }

    /** @return The categories this data row belongs to. */
    public Set<String> getCategories() {
        return Collections.unmodifiableSet(mCategories);
    }

    /** @return The categories this data row belongs to. */
    public String getCategoriesAsString() {
        StringBuilder buffer = new StringBuilder();
        for (String category : mCategories) {
            if (!buffer.isEmpty()) {
                buffer.append(",");
                buffer.append(" ");
            }
            buffer.append(category);
        }
        return buffer.toString();
    }

    /*
     * Does this belong to a category? Added the ability to check for compound categories like
     * "Money: US", "Concoctions:Potions" using "Money" or "Concoctions".
     */
    public boolean hasCategory(String cat) {
        for (String category : mCategories) {
            int indexOfColon = category.indexOf(':');
            if (indexOfColon > 0) {
                if (category.substring(0, indexOfColon).equalsIgnoreCase(cat)) {
                    return true;
                }
            }
            if (category.equalsIgnoreCase(cat)) {
                return true;
            }
        }
        return false;
    }

    /**
     * @param categories The categories this data row belongs to.
     * @return Whether there was a change or not.
     */
    public boolean setCategories(Collection<String> categories) {
        Set<String> old = mCategories;
        mCategories = new TreeSet<>();
        for (String category : categories) {
            category = category.trim();
            if (!category.isEmpty()) {
                mCategories.add(category);
            }
        }
        if (!old.equals(mCategories)) {
            notifyOfChange();
            return true;
        }
        return false;
    }

    /**
     * @param categories The categories this data row belongs to. Use commas to separate
     *                   categories.
     * @return Whether there was a change or not.
     */
    public final boolean setCategories(String categories) {
        return setCategories(createList(categories));
    }

    /** @return The prerequisites needed by this data row. */
    public PrereqList getPrereqs() {
        return mPrereqList;
    }

    /**
     * @param prereqs The new prerequisites needed by this data row.
     * @return Whether there was a change or not.
     */
    public boolean setPrereqs(PrereqList prereqs) {
        if (!mPrereqList.equals(prereqs)) {
            mPrereqList = (PrereqList) prereqs.clone(null);
            return true;
        }
        return false;
    }

    /** @return The defaults for this row. */
    public List<SkillDefault> getDefaults() {
        return Collections.unmodifiableList(mDefaults);
    }

    /**
     * @param defaults The new defaults for this row.
     * @return Whether there was a change or not.
     */
    public boolean setDefaults(List<SkillDefault> defaults) {
        if (!mDefaults.equals(defaults)) {
            mDefaults = new ArrayList<>(defaults);
            return true;
        }
        return false;
    }

    @Override
    public final void setData(Column column, Object data) {
        // Not used.
    }

    /**
     * @param text          The text to search for.
     * @param lowerCaseOnly The passed in text is all lowercase.
     * @return {@code true} if this row contains the text.
     */
    public boolean contains(String text, boolean lowerCaseOnly) {
        return false;
    }

    /** @return An image representative of this row. */
    public abstract Icon getIcon();

    /** @param set The nameable keys. */
    public void fillWithNameableKeys(Set<String> set) {
        extractNameables(set, mNotes);
        extractNameables(set, mVTTNotes);
        for (SkillDefault def : mDefaults) {
            def.fillWithNameableKeys(set);
        }
        for (Feature feature : mFeatures) {
            feature.fillWithNameableKeys(set);
        }
        mPrereqList.fillWithNameableKeys(set);
    }

    /** @param map The map of nameable keys to names to apply. */
    public void applyNameableKeys(Map<String, String> map) {
        mNotes = nameNameables(map, mNotes);
        mVTTNotes = nameNameables(map, mVTTNotes);
        for (SkillDefault def : mDefaults) {
            def.applyNameableKeys(map);
        }
        for (Feature feature : mFeatures) {
            feature.applyNameableKeys(map);
        }
        mPrereqList.applyNameableKeys(map);
    }

    /** @return The notes. */
    public String getNotes() {
        return mNotes;
    }

    /** @return The notes due to modifiers. */
    public String getModifierNotes() {
        return "";
    }

    /**
     * @param notes The notes to set.
     * @return Whether it was changed.
     */
    public boolean setNotes(String notes) {
        if (!mNotes.equals(notes)) {
            mNotes = notes;
            return true;
        }
        return false;
    }

    /** @return The VTT notes. */
    public String getVTTNotes() {
        return mVTTNotes;
    }

    /**
     * @param notes The VTT notes to set.
     * @return Whether it was changed.
     */
    public boolean setVTTNotes(String notes) {
        if (!mVTTNotes.equals(notes)) {
            mVTTNotes = notes;
            return true;
        }
        return false;
    }

    public String getDescriptionText() {
        StringBuilder builder = new StringBuilder();
        builder.append(this);
        SheetSettings settings = getDataFile().getSheetSettings();
        if (settings.userDescriptionDisplay().tooltip() && this instanceof Advantage) {
            String userDesc = ((Advantage) this).getUserDesc();
            if (!userDesc.isBlank()) {
                builder.append(" - ");
                builder.append(userDesc);
            }
        }
        if (settings.modifiersDisplay().inline()) {
            String modNotes = getModifierNotes();
            if (!modNotes.isBlank()) {
                builder.append(" - ");
                builder.append(modNotes);
            }
        }
        if (settings.notesDisplay().inline()) {
            String notes = getNotes();
            if (!notes.isBlank()) {
                if (this instanceof Modifier) {
                    builder.append(" (");
                    builder.append(notes);
                    builder.append(')');
                } else {
                    builder.append(" - ");
                    builder.append(notes);
                }
            }
        }
        return builder.toString();
    }

    public String getDescriptionToolTipText() {
        StringBuilder builder  = new StringBuilder();
        SheetSettings settings = getDataFile().getSheetSettings();
        if (settings.userDescriptionDisplay().tooltip() && this instanceof Advantage) {
            String userDesc = ((Advantage) this).getUserDesc();
            if (!userDesc.isBlank()) {
                builder.append(userDesc);
                builder.append('\n');
            }
        }
        if (settings.modifiersDisplay().tooltip()) {
            String modNotes = getModifierNotes();
            if (!modNotes.isBlank()) {
                builder.append(modNotes);
                builder.append('\n');
            }
        }
        if (settings.notesDisplay().tooltip()) {
            String notes = getNotes();
            if (!notes.isBlank()) {
                builder.append(notes);
                builder.append('\n');
            }
        }
        if (!builder.isEmpty()) {
            builder.setLength(builder.length() - 1); // Remove the last '\n'
        }
        return builder.isEmpty() ? null : builder.toString();
    }
}
