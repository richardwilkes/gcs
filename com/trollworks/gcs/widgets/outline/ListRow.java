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

package com.trollworks.gcs.widgets.outline;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.common.DataFile;
import com.trollworks.gcs.feature.AttributeBonus;
import com.trollworks.gcs.feature.CostReduction;
import com.trollworks.gcs.feature.DRBonus;
import com.trollworks.gcs.feature.Feature;
import com.trollworks.gcs.feature.SkillBonus;
import com.trollworks.gcs.feature.SpellBonus;
import com.trollworks.gcs.feature.WeaponBonus;
import com.trollworks.gcs.prereq.PrereqList;
import com.trollworks.gcs.skill.SkillDefault;
import com.trollworks.gcs.skill.Technique;
import com.trollworks.gcs.template.Template;
import com.trollworks.gcs.utility.io.xml.XMLNodeType;
import com.trollworks.gcs.utility.io.xml.XMLReader;
import com.trollworks.gcs.utility.io.xml.XMLWriter;

import java.awt.image.BufferedImage;
import java.io.IOException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;

/** A common row super-class for the model. */
public abstract class ListRow extends Row {
	private static final String		ATTRIBUTE_OPEN	= "open";	//$NON-NLS-1$
	private static final String		TAG_NOTES		= "notes";	//$NON-NLS-1$
	/** The data file the row is associated with. */
	protected DataFile				mDataFile;
	private ArrayList<Feature>		mFeatures;
	private PrereqList				mPrereqList;
	private ArrayList<SkillDefault>	mDefaults;
	private boolean					mIsSatisfied;
	private String					mUnsatisfiedReason;
	private String					mNotes;

	/**
	 * Extracts any "nameable" portions of the buffer and puts their keys into the provided set.
	 * 
	 * @param set The set to add the nameable keys to.
	 * @param buffer The text to check for nameable portions.
	 */
	public static void extractNameables(HashSet<String> set, String buffer) {
		int first = buffer.indexOf('@');
		int last = buffer.indexOf('@', first + 1);

		while (first != -1 && last != -1) {
			set.add(buffer.substring(first + 1, last));
			first = buffer.indexOf('@', last + 1);
			last = buffer.indexOf('@', first + 1);
		}
	}

	/**
	 * Names any "nameable" portions of the data and returns the resulting string.
	 * 
	 * @param map The map of nameable keys to names.
	 * @param data The data to change.
	 * @return The revised string.
	 */
	public static String nameNameables(HashMap<String, String> map, String data) {
		int first = data.indexOf('@');
		int last = data.indexOf('@', first + 1);
		StringBuilder buffer = new StringBuilder();

		while (first != -1 && last != -1) {
			String key = data.substring(first + 1, last);
			String replacement = map.get(key);

			buffer.setLength(0);
			if (first != 0) {
				buffer.append(data.substring(0, first));
			}
			if (replacement != null) {
				buffer.append(replacement);
			} else {
				buffer.append('@');
				buffer.append(key);
				buffer.append('@');
			}
			if (last + 1 != data.length()) {
				data = data.substring(last + 1);
			} else {
				data = ""; //$NON-NLS-1$
			}
			first = data.indexOf('@');
			last = data.indexOf('@', first + 1);
		}
		buffer.append(data);
		return buffer.toString();
	}

	/**
	 * Creates a new data row.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param isContainer Whether or not this row allows children.
	 */
	public ListRow(DataFile dataFile, boolean isContainer) {
		super();
		setCanHaveChildren(isContainer);
		setOpen(isContainer);
		mDataFile = dataFile;
		mFeatures = new ArrayList<Feature>();
		mPrereqList = new PrereqList(null, true);
		mDefaults = new ArrayList<SkillDefault>();
		mIsSatisfied = true;
		mNotes = ""; //$NON-NLS-1$
	}

	/**
	 * Creates a clone of an existing data row and associates it with the specified data file.
	 * 
	 * @param dataFile The data file to associate it with.
	 * @param rowToClone The data row to clone.
	 */
	public ListRow(DataFile dataFile, ListRow rowToClone) {
		this(dataFile, rowToClone.canHaveChildren());
		setOpen(rowToClone.isOpen());
		mNotes = rowToClone.mNotes;

		for (Feature feature : rowToClone.mFeatures) {
			mFeatures.add(feature.cloneFeature());
		}
		mPrereqList = new PrereqList(null, rowToClone.getPrereqs());
		mDefaults = new ArrayList<SkillDefault>();
		for (SkillDefault skillDefault : rowToClone.mDefaults) {
			mDefaults.add(new SkillDefault(skillDefault));
		}
	}

	/** @return Creates a detailed editor for this row. */
	public abstract RowEditor<? extends ListRow> createEditor();

	/** @return The localized name for this row object. */
	public abstract String getLocalizedName();

	@Override public boolean addChild(Row row) {
		boolean result = super.addChild(row);

		if (result) {
			notifySingle(getListChangedID());
		}
		return result;
	}

	/** @return The ID for the "list changed" notification. */
	public abstract String getListChangedID();

	/** @return The XML root container tag name for this particular row. */
	public abstract String getXMLTagName();

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

	/** @return The reason {@link #isSatisfied()} is returning <code>false</code>. */
	public String getReasonForUnsatisfied() {
		return mUnsatisfiedReason;
	}

	/** @param reason The reason {@link #isSatisfied()} is returning <code>false</code>. */
	public void setReasonForUnsatisfied(String reason) {
		mUnsatisfiedReason = reason;
	}

	/**
	 * Loads this row's contents.
	 * 
	 * @param reader The XML reader to load from.
	 * @param forUndo Whether this is being called to load undo state.
	 * @throws IOException
	 */
	public final void load(XMLReader reader, boolean forUndo) throws IOException {
		String marker = reader.getMarker();

		prepareForLoad(forUndo);
		loadAttributes(reader, forUndo);
		do {
			if (reader.next() == XMLNodeType.START_TAG) {
				String name = reader.getName();

				if (AttributeBonus.TAG_ROOT.equals(name)) {
					mFeatures.add(new AttributeBonus(reader));
				} else if (DRBonus.TAG_ROOT.equals(name)) {
					mFeatures.add(new DRBonus(reader));
				} else if (SkillBonus.TAG_ROOT.equals(name)) {
					mFeatures.add(new SkillBonus(reader));
				} else if (SpellBonus.TAG_ROOT.equals(name)) {
					mFeatures.add(new SpellBonus(reader));
				} else if (WeaponBonus.TAG_ROOT.equals(name)) {
					mFeatures.add(new WeaponBonus(reader));
				} else if (CostReduction.TAG_ROOT.equals(name)) {
					mFeatures.add(new CostReduction(reader));
				} else if (PrereqList.TAG_ROOT.equals(name)) {
					mPrereqList = new PrereqList(null, reader);
				} else if (!(this instanceof Technique) && SkillDefault.TAG_ROOT.equals(name)) {
					mDefaults.add(new SkillDefault(reader));
				} else if (TAG_NOTES.equals(name)) {
					mNotes = reader.readText();
				} else {
					loadSubElement(reader, forUndo);
				}
			}
		} while (reader.withinMarker(marker));
		finishedLoading();
	}

	/**
	 * Called to prepare the row for loading.
	 * 
	 * @param forUndo Whether this is being called to load undo state.
	 */
	protected void prepareForLoad(@SuppressWarnings("unused") boolean forUndo) {
		mNotes = ""; //$NON-NLS-1$
		mFeatures.clear();
		mDefaults.clear();
		mPrereqList = new PrereqList(null, true);
	}

	/**
	 * Loads this row's custom attributes from the specified element.
	 * 
	 * @param reader The XML reader to load from.
	 * @param forUndo Whether this is being called to load undo state.
	 */
	protected void loadAttributes(XMLReader reader, @SuppressWarnings("unused") boolean forUndo) {
		if (canHaveChildren()) {
			setOpen(reader.isAttributeSet(ATTRIBUTE_OPEN));
		}
	}

	/**
	 * Loads this row's custom data from the specified element.
	 * 
	 * @param reader The XML reader to load from.
	 * @param forUndo Whether this is being called to load undo state.
	 * @throws IOException
	 */
	protected void loadSubElement(XMLReader reader, @SuppressWarnings("unused") boolean forUndo) throws IOException {
		reader.skipTag(reader.getName());

	}

	/** Called when loading of this row is complete. Does nothing by default. */
	protected void finishedLoading() {
		// Nothing to do.
	}

	/**
	 * Saves the row.
	 * 
	 * @param out The XML writer to use.
	 * @param forUndo Whether this is being called to save undo state.
	 */
	public void save(XMLWriter out, boolean forUndo) {
		out.startTag(getXMLTagName());
		if (canHaveChildren()) {
			out.writeAttribute(ATTRIBUTE_OPEN, isOpen());
		}
		saveAttributes(out, forUndo);
		out.finishTagEOL();
		saveSelf(out, forUndo);
		out.simpleTagNotEmpty(TAG_NOTES, mNotes);

		if (!mFeatures.isEmpty()) {
			for (Feature feature : mFeatures) {
				feature.save(out);
			}
		}

		mPrereqList.save(out);

		if (!(this instanceof Technique) && !mDefaults.isEmpty()) {
			for (SkillDefault skillDefault : mDefaults) {
				skillDefault.save(out);
			}
		}

		if (!forUndo && canHaveChildren()) {
			for (Row row : getChildren()) {
				((ListRow) row).save(out, false);
			}
		}
		out.endTagEOL(getXMLTagName(), true);
	}

	/**
	 * Saves the row.
	 * 
	 * @param out The XML writer to use.
	 * @param forUndo Whether this is being called to save undo state.
	 */
	protected abstract void saveSelf(XMLWriter out, boolean forUndo);

	/**
	 * Saves extra attributes of the row, if any.
	 * 
	 * @param out The XML writer to use.
	 * @param forUndo Whether this is being called to save undo state.
	 */
	protected void saveAttributes(@SuppressWarnings("unused") XMLWriter out, @SuppressWarnings("unused") boolean forUndo) {
		// Does nothing by default.
	}

	/**
	 * Starts the notification process. Should be called before calling
	 * {@link #notify(String,Object)}.
	 */
	protected final void startNotify() {
		if (mDataFile != null) {
			mDataFile.startNotify();
		}
	}

	/**
	 * Sends a notification to all interested consumers.
	 * 
	 * @param type The notification type.
	 * @param data Extra data specific to this notification.
	 */
	public void notify(String type, @SuppressWarnings("unused") Object data) {
		if (mDataFile != null) {
			mDataFile.notify(type, this);
		}
	}

	/**
	 * Sends a notification to all interested consumers.
	 * 
	 * @param type The notification type.
	 */
	public final void notifySingle(String type) {
		if (mDataFile != null) {
			mDataFile.notifySingle(type, this);
		}
	}

	/**
	 * Ends the notification process. Must be called after calling {@link #notify(String,Object)}.
	 */
	public void endNotify() {
		if (mDataFile != null) {
			mDataFile.endNotify();
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
			mFeatures = new ArrayList<Feature>(features);
			return true;
		}
		return false;
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
			mDefaults = new ArrayList<SkillDefault>(defaults);
			return true;
		}
		return false;
	}

	@Override public final void setData(Column column, Object data) {
		// Not used.
	}

	/**
	 * @param text The text to search for.
	 * @param lowerCaseOnly The passed in text is all lowercase.
	 * @return <code>true</code> if this row contains the text.
	 */
	public abstract boolean contains(String text, boolean lowerCaseOnly);

	/**
	 * @param large Whether to return the small (16x16) or large (32x32) image.
	 * @return An image representative of this row.
	 */
	public abstract BufferedImage getImage(boolean large);

	/** @param set The nameable keys. */
	public void fillWithNameableKeys(HashSet<String> set) {
		extractNameables(set, mNotes);
		for (SkillDefault def : mDefaults) {
			def.fillWithNameableKeys(set);
		}
		for (Feature feature : mFeatures) {
			feature.fillWithNameableKeys(set);
		}
		mPrereqList.fillWithNameableKeys(set);
	}

	/** @param map The map of nameable keys to names to apply. */
	public void applyNameableKeys(HashMap<String, String> map) {
		mNotes = nameNameables(map, mNotes);
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
		return ""; //$NON-NLS-1$
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
}
