/*
 * Copyright (c) 1998-2017 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package com.trollworks.gcs.advantage;

import com.trollworks.gcs.common.LoadState;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.io.xml.XMLReader;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.text.NumberFilter;

import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JTextField;

/** The types of {@link Advantage} containers. */
public enum AdvantageContainerType {
    /** The standard grouping container type. */
    GROUP {
        @Override
        public String toString() {
            return GROUP_TITLE;
        }

        @Override
        ContainerTypeEditor addControls(AdvantageEditor editor) {
            return SummativeContainerTypeEditor.getInstance();
        }

        @Override
        AdvantageContainer loadAttributes(XMLReader reader, LoadState state) {
            return SummativeAdvantageContainer.getInstance();
        }
    },
    /**
     * The meta-trait grouping container type. Acts as one normal trait, listed as an advantage if
     * its point total is positive, or a disadvantage if it is negative.
     */
    META_TRAIT {
        @Override
        public String toString() {
            return META_TRAIT_TITLE;
        }

        @Override
        ContainerTypeEditor addControls(AdvantageEditor editor) {
            return SummativeContainerTypeEditor.getInstance();
        }

        @Override
        AdvantageContainer loadAttributes(XMLReader reader, LoadState state) {
            return SummativeAdvantageContainer.getInstance();
        }
    },
    /**
     * The race grouping container type. Its point cost is tracked separately from normal advantages
     * and disadvantages.
     */
    RACE {
        @Override
        public String toString() {
            return RACE_TITLE;
        }

        @Override
        ContainerTypeEditor addControls(AdvantageEditor editor) {
            return SummativeContainerTypeEditor.getInstance();
        }

        @Override
        AdvantageContainer loadAttributes(XMLReader reader, LoadState state) {
            return SummativeAdvantageContainer.getInstance();
        }
    },
    /**
     * The alternative abilities grouping container type. It behaves similar to a
     * {@link #META_TRAIT} , but applies the rules for alternative abilities (see B61 and P11) to
     * its immediate children.
     */
    ALTERNATIVE_ABILITIES {
        @Override
        public String toString() {
            return ALTERNATIVE_ABILITIES_TITLE;
        }

        @Override
        ContainerTypeEditor addControls(AdvantageEditor editor) {
            return new AlternativeAbilitiesContainerTypeEditor();
        }

        @Override
        AdvantageContainer loadAttributes(XMLReader reader, LoadState state) {
            return new AlternativeAbilitiesAdvantageContainer();
        }

    },
    /**
     * The alternate forms grouping container type. It behaves similar to a {@link #GROUP}, but
     * applies the rules for alternate forms (see B83, P18 and P74-75) to immediate children of the
     * {@link #ALTERNATE_FORM} type, costing only the most expensive alternate form, or the cost of
     * the base form plus only 90% of the cost of the single alternate form.
     */
    ALTERNATE_FORMS {
        @Override
        public String toString() {
            return ALTERNATE_FORMS_TITLE;
        }

        @SuppressWarnings("unused")
        @Override
        ContainerTypeEditor addControls(AdvantageEditor editor) {
            JPanel typePanel = editor.getContainerTypePanel();

            JTextField cost = new JTextField(5);
            new NumberFilter(cost, false, true, true, 5);
            typePanel.add(cost);

            JLabel desc = new JLabel(ALTERNATE_FORMS_TEXT);
            typePanel.add(desc);

            JTextField base = new JTextField(10);
            typePanel.add(base);

            if (editor.getContainerType() == this && editor.getAdvantageContainer() instanceof AlternateFormsAdvantageContainer) {
                AlternateFormsAdvantageContainer container = (AlternateFormsAdvantageContainer) editor.getAdvantageContainer();
                cost.setText(Integer.toString(container.mCostModPercent));
                base.setText(container.mBaseForm);
            } else {
                cost.setText(Integer.toString(0));
                base.setText(""); //$NON-NLS-1$
            }

            return new AlternateFormsContainerTypeEditor(cost, base);
        }

        @Override
        AdvantageContainer loadAttributes(XMLReader reader, LoadState state) {
            int costModPercent = reader.getAttributeAsInteger(AlternateFormsAdvantageContainer.TAG_COST_MOD_PERCENT, 0);
            String baseForm = reader.getAttribute(AlternateFormsAdvantageContainer.TAG_BASE_FORM, ""); //$NON-NLS-1$
            return new AlternateFormsAdvantageContainer(costModPercent, baseForm);
        }
    },
    /**
     * The alternate form grouping container type. It behaves similar to a {@link #META_TRAIT}, but
     * applies the rules for a single alternate form (see B84) to the other child Alternate Form of
     * the parent Alternate Forms, costing only 90% of the difference in cost if it costs more.
     */
    ALTERNATE_FORM {
        @Override
        public String toString() {
            return ALTERNATE_FORM_TITLE;
        }

        @Override
        ContainerTypeEditor addControls(AdvantageEditor editor) {
            return new AlternateFormContainerTypeEditor();
        }

        @Override
        AdvantageContainer loadAttributes(XMLReader reader, LoadState state) {
            return AlternateFormAdvantageContainer.getInstance();
        }
    };

    abstract ContainerTypeEditor addControls(AdvantageEditor editor);

    abstract AdvantageContainer loadAttributes(XMLReader reader, LoadState state);

    @Localize("Group")
    @Localize(locale = "de", value = "Gruppe")
    @Localize(locale = "ru", value = "Группа")
    @Localize(locale = "es", value = "Grupo")
    static String GROUP_TITLE;
    @Localize("Meta-Trait")
    @Localize(locale = "de", value = "Meta-Eigenschaft")
    @Localize(locale = "ru", value = "Мета-черта")
    static String META_TRAIT_TITLE;
    @Localize("Race")
    @Localize(locale = "de", value = "Rasse")
    @Localize(locale = "ru", value = "Раса")
    @Localize(locale = "es", value = "Raza")
    static String RACE_TITLE;
    @Localize("Alternative Abilities")
    @Localize(locale = "de", value = "Alternative Fähigkeiten")
    @Localize(locale = "ru", value = "Альтернативные способности")
    @Localize(locale = "es", value = "Habilidades Alternativas")
    static String ALTERNATIVE_ABILITIES_TITLE;
    @Localize("Alternate Forms")
    static String ALTERNATE_FORMS_TITLE;
    @Localize("% to the cost over base form named")
    static String ALTERNATE_FORMS_TEXT;
    @Localize("Alternate Form")
    static String ALTERNATE_FORM_TITLE;

    static {
        Localization.initialize();
    }

}
