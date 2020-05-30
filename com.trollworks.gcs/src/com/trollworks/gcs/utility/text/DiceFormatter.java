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

package com.trollworks.gcs.utility.text;

import com.trollworks.gcs.character.GURPSCharacter;
import com.trollworks.gcs.utility.Dice;

import java.text.ParseException;
import javax.swing.JFormattedTextField;

/** Provides dice field conversion. */
public class DiceFormatter extends JFormattedTextField.AbstractFormatter {
    private GURPSCharacter mCharacter;

    public DiceFormatter(GURPSCharacter character) {
        mCharacter = character;
    }

    @Override
    public Object stringToValue(String text) throws ParseException {
        return new Dice(text);
    }

    @Override
    public String valueToString(Object value) throws ParseException {
        return value instanceof Dice ? ((Dice) value).toString(mCharacter.getSettings().useModifyingDicePlusAdds()) : "";
    }
}
