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

package com.trollworks.gcs.preferences;

import com.trollworks.gcs.character.Profile;
import com.trollworks.toolkit.annotation.Localize;
import com.trollworks.toolkit.ui.UIUtilities;
import com.trollworks.toolkit.ui.image.StdImage;
import com.trollworks.toolkit.ui.layout.Alignment;
import com.trollworks.toolkit.ui.layout.FlexColumn;
import com.trollworks.toolkit.ui.layout.FlexComponent;
import com.trollworks.toolkit.ui.layout.FlexGrid;
import com.trollworks.toolkit.ui.layout.FlexRow;
import com.trollworks.toolkit.ui.layout.FlexSpacer;
import com.trollworks.toolkit.ui.preferences.PreferencePanel;
import com.trollworks.toolkit.ui.preferences.PreferencesWindow;
import com.trollworks.toolkit.ui.scale.Scales;
import com.trollworks.toolkit.ui.widget.StdFileDialog;
import com.trollworks.toolkit.utility.Dice;
import com.trollworks.toolkit.utility.FileType;
import com.trollworks.toolkit.utility.Localization;
import com.trollworks.toolkit.utility.PathUtils;
import com.trollworks.toolkit.utility.Preferences;
import com.trollworks.toolkit.utility.text.Enums;
import com.trollworks.toolkit.utility.text.Numbers;
import com.trollworks.toolkit.utility.text.Text;
import com.trollworks.toolkit.utility.units.LengthUnits;
import com.trollworks.toolkit.utility.units.WeightUnits;

import java.awt.Dimension;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.event.ItemEvent;
import java.awt.event.ItemListener;
import java.io.File;
import java.util.ArrayList;
import java.util.List;

import javax.swing.JCheckBox;
import javax.swing.JComboBox;
import javax.swing.JTextField;
import javax.swing.SwingConstants;
import javax.swing.event.DocumentEvent;
import javax.swing.event.DocumentListener;
import javax.swing.filechooser.FileNameExtensionFilter;
import javax.swing.text.Document;

/** The sheet preferences panel. */
public class SheetPreferences extends PreferencePanel implements ActionListener, DocumentListener, ItemListener {
    @Localize("Sheet")
    @Localize(locale = "de", value = "Charakterblatt")
    @Localize(locale = "ru", value = "Лист")
    @Localize(locale = "es", value = "Hoja de Personaje")
    private static String SHEET;
    @Localize("Player")
    @Localize(locale = "de", value = "Spieler")
    @Localize(locale = "ru", value = "Игрок")
    @Localize(locale = "es", value = "Jugador")
    private static String PLAYER;
    @Localize("The player name to use when a new character sheet is created")
    @Localize(locale = "de",
              value = "Der Spielername, der verwendet wird, wenn ein neues Charakterblatt erstellt wird.")
    @Localize(locale = "ru", value = "Имя игрока при создании нового листа персонажа")
    @Localize(locale = "es",
              value = "Nombre del jugador que se usará cuando se cree una nueva hoja de personaje")
    private static String PLAYER_TOOLTIP;
    @Localize("Campaign")
    @Localize(locale = "de", value = "Kampagne")
    @Localize(locale = "ru", value = "Компания")
    @Localize(locale = "es", value = "Campaña")
    private static String CAMPAIGN;
    @Localize("The campaign to use when a new character sheet is created")
    @Localize(locale = "de",
              value = "Die Kampagne, die verwendet wird, wenn ein neues Charakterblatt erstellt wird.")
    @Localize(locale = "ru", value = "Название компании при создании нового листа персонаж")
    @Localize(locale = "es",
              value = "Nombre de la campaña que se usará cuando se cree una nueva hoja de personaje")
    private static String CAMPAIGN_TOOLTIP;
    @Localize("Tech Level")
    @Localize(locale = "de", value = "Techlevel")
    @Localize(locale = "ru", value = "Технологический уровень")
    @Localize(locale = "es", value = "Nivel Tecnológico")
    private static String TECH_LEVEL;
    @Localize("<html><body>TL0: Stone Age (Prehistory and later)<br>TL1: Bronze Age (3500 B.C.+)<br>TL2: Iron Age (1200 B.C.+)<br>TL3: Medieval (600 A.D.+)<br>TL4: Age of Sail (1450+)<br>TL5: Industrial Revolution (1730+)<br>TL6: Mechanized Age (1880+)<br>TL7: Nuclear Age (1940+)<br>TL8: Digital Age (1980+)<br>TL9: Microtech Age (2025+?)<br>TL10: Robotic Age (2070+?)<br>TL11: Age of Exotic Matter<br>TL12: Anything Goes</body></html>")
    @Localize(locale = "de",
              value = "<html><body>TL0: Steinzeit (Vorzeit und älter)<br>TL1: Bronzezeit (ab 3500 v.Chr.)<br>TL2: Eisenzeit (ab 1200 v. Chr.)<br>TL3: Mittelalter(ab 600 n. Chr.)<br>TL4: Renaissance / Kolonialzeitalter (ab 1450)<br>TL5: Industrielle Revolution (ab 1730)<br>TL6: Zeitalter der Mechanik (ab 1880)<br>TL7: Ära der Kernkraft (ab 1940)<br>TL8: Ära der Digitalisierung (ab 1980)<br>TL9: Ära der Mikrotechnik (ab 2025?)<br>TL10: Ära der Robotik (ab 2070?)<br>TL11: Ära der Exotischen Materie<br>TL12: Was ihr wollt!</body></html>")
    @Localize(locale = "ru",
              value = "<html><body>ТУ0: Каменный век<br>ТУ1: Бронзовый век<br>ТУ2: Железный век<br>ТУ3: Средневековье<br>ТУ4: Эпоха парусов<br>ТУ5: Промышленный переворот<br>ТУ6: Эпоха механики<br>ТУ7: Атомная эпоха<br>ТУ8: Цифровая эпоха<br>ТУ9: Эпоха микротехники<br>ТУ10: Эпоха роботизации<br>ТУ11: Эпоха экзотических материалов<br>ТУ12: Всё, что угодно</body></html>")
    @Localize(locale = "es",
              value = "<html><body>TL0: Edad de Piedra (Prehistoria y anterior)<br>TL1: Edad de Bronce (3500 a. C.+)<br>TL2: Edad de Hierro (1200 a. C.+)<br>TL3: Edad Media (600 a. C.+)<br>TL4: Era del Renacimiento (1450+)<br>TL5: Revolución Industrial(1730+)<br>TL6: Era de la Mecanización(1880+)<br>TL7: Era Nuclear(1940+)<br>TL8: Era Digital(1980+)<br>TL9: Era de la Microtecnología(¿2025+?)<br>TL10: Era de la Robótica(¿2070+?)<br>TL11: Era del Dominio de la Materia<br>TL12: Más allá</body></html>")
    private static String TECH_LEVEL_TOOLTIP;
    @Localize("Initial Points")
    @Localize(locale = "de", value = "Charakterpunkte")
    @Localize(locale = "ru", value = "Начальные очки")
    @Localize(locale = "es", value = "Puntos Iniciales")
    private static String INITIAL_POINTS;
    @Localize("The initial number of character points to start with")
    @Localize(locale = "de",
              value = "Die Zahl der Charakterpunkte, mit denen der Charakter startet.")
    @Localize(locale = "ru", value = "Первоначальное количество очков персонажа")
    @Localize(locale = "es", value = "Cantidad inicial de puntos de parsonaje")
    private static String INITIAL_POINTS_TOOLTIP;
    @Localize("Select A Portrait")
    @Localize(locale = "de", value = "Wähle ein Charakterbild")
    @Localize(locale = "ru", value = "Выберите изображение")
    @Localize(locale = "es", value = "Selecciona un retrato")
    private static String SELECT_PORTRAIT;
    @Localize("Use optional (house) rule: Will and Perception are not based upon IQ")
    @Localize(locale = "de",
              value = "Verwende die optionale (Haus-) Regel: Wille und Wahrnehmung basieren nicht auf IQ")
    @Localize(locale = "ru",
              value = "Использовать опциональное (домашнее) правило: Воля и Восприятие не основаны на интеллекте (ИН)")
    @Localize(locale = "es",
              value = "Usar regla opcional (no oficial): Voluntad y Percepción no se basan en Inteligencia (CI)")
    private static String OPTIONAL_IQ_RULES;
    @Localize("Use optional rule \"Multiplicative Modifiers\" from PW102 (note: changes point value)")
    @Localize(locale = "de",
              value = "Verwende die optionale Regel \"Multiplicative Modifiers\" aus PW102 (Hinweis: Verändert den Punktwert)")
    @Localize(locale = "ru",
              value = "Использовать необязательное правило \"Накопительные модификаторы\" из PW102 (прим.: изменяет количество очков)")
    @Localize(locale = "es",
              value = "Usar regla opcional: \"Modificadores Multiplicadores\", véase PW102 (Nota: cambia el valor en puntos)")
    private static String OPTIONAL_MODIFIER_RULES;
    @Localize("Use optional rule \"Modifying Dice + Adds\" from B269")
    @Localize(locale = "de",
              value = "Verwende die optionale Regel \"Verändere Würfel und Zugabe\" aus B269 (\"Modifying Dice + Adds\")")
    @Localize(locale = "ru",
              value = "Использовать необязательное правило \"Замена модификаторов кубиками\" из B269")
    @Localize(locale = "es",
              value = "Usar regla opcional: \"Modificando dado + incrementos\", véase B269")
    private static String OPTIONAL_DICE_RULES;
    @Localize("Use optional strength rules from the \"Knowing Your Own Strength\" Pyramid article")
    private static String OPTIONAL_STRENGTH_RULES;
    @Localize("Use optional Thrust damage (Thrust = Swing -2)")
    private static String OPTIONAL_THRUST_DAMAGE;
    @Localize("Use optional Reduced Swing rules from the  \"Adjusting Swing Damage in Dungeon Fantasy\" No School Grognard article")
    private static String OPTIONAL_REDUCED_SWING;
    @Localize("for the initial scale when opening character sheets, templates and lists")
    private static String UI_SCALE_POST;
    @Localize("HTML Template Override")
    @Localize(locale = "de", value = "Alternative HTML-Vorlage")

    @Localize(locale = "ru", value = "Переопределить HTML-шаблон")
    @Localize(locale = "es", value = "Ignorar la plantilla HTML")
    private static String HTML_TEMPLATE_OVERRIDE;
    @Localize("Choose\u2026")
    @Localize(locale = "de", value = "wählen\u2026")
    @Localize(locale = "ru", value = "Выбрать\u2026")
    @Localize(locale = "es", value = "Elegir\u2026")
    private static String HTML_TEMPLATE_PICKER;
    @Localize("Specify a file to use as the template when exporting to HTML")
    @Localize(locale = "de",
              value = "Wähle die Datei, die als Vorlage für den HTML-Export verwendet werden soll.")
    @Localize(locale = "ru", value = "Использовать указанный файл шаблона при экспорте в HTML")
    @Localize(locale = "es",
              value = "Especifica que archivose usará como plantilla cuando se exporta a formato HTML")
    private static String HTML_TEMPLATE_OVERRIDE_TOOLTIP;
    @Localize("Select A HTML Template")
    @Localize(locale = "de", value = "Wähle eine HTML-Vorlage")
    @Localize(locale = "ru", value = "Выберите HTML-шаблон")
    @Localize(locale = "es", value = "Selecionar una plantilla HTML")
    private static String SELECT_HTML_TEMPLATE;
    @Localize("Use platform native print dialogs (settings cannot be saved)")
    @Localize(locale = "de",
              value = "Verwende Druckdialoge des Betriebssystems (Einstellungen können nicht gespeichert werden)")
    @Localize(locale = "ru",
              value = "Использовать диалоги печати родные для ОС (в этом случае не сохраняются настройки диалогов)")
    @Localize(locale = "es",
              value = "Usar los diálogos de impresión del sistema operativo (No pueden guardarse las preferencias)")
    private static String NATIVE_PRINTER;
    @Localize("<html><body>Whether or not the native print dialogs should be used.<br>Choosing this option will prevent the program from saving<br>and restoring print settings with the document.</body></html>")
    @Localize(locale = "de",
              value = "<html><body>Ob die Druckdialoge des Betriebssystems verwendet werden sollen.<br>Das Auswählen dieser Option wird das Programm daran hindern,<br>die Druckeinstellungen im Dokument zu speichern und werderherzustellen.</body></html>")
    @Localize(locale = "ru",
              value = "<html><body>Использовать родные диалоги печати ОС.<br>При выборе этого параметра программа не будет сохранять<br>настройки печати документа.</body></html>")
    @Localize(locale = "es",
              value = "<html><body>Indica si se usan o no los diálogos de impresión del sistema operativo.<br>Si se selecciona esta opción, el programa no podrá salvar<br>y restaurar configuración del documento.</body></html>")
    private static String NATIVE_PRINTER_TOOLTIP;
    @Localize("Automatically name new characters")
    @Localize(locale = "de", value = "Neue Charaktere automatisch benennen")
    @Localize(locale = "ru", value = "Автоматически называть новых персонажей")
    @Localize(locale = "es", value = "Dar nombre automáticamente a los personajes nuevos")
    private static String AUTO_NAME;
    @Localize("The units to use for display of generated lengths")
    @Localize(locale = "de", value = "Die Einheit, in der die berechnete Länge angezeigt wird.")
    @Localize(locale = "ru", value = "Единицы измерения создаваемых длин")
    @Localize(locale = "es", value = "Unidades en las que mostrar longitudes")
    private static String LENGTH_UNITS_TOOLTIP;
    @Localize("The units to use for display of generated weights")
    @Localize(locale = "de", value = "Die Einheit, in der das berechnete Gewicht angezeigt wird.")
    @Localize(locale = "ru", value = "Единицы измерения создаваемых весов")
    @Localize(locale = "es", value = "Unidades en las que mostrar pesos")
    private static String WEIGHT_UNITS_TOOLTIP;
    @Localize("Use")
    @Localize(locale = "de", value = "Verwende")
    @Localize(locale = "ru", value = "Использовать")
    @Localize(locale = "es", value = "usar")
    private static String USE;
    @Localize("and")
    @Localize(locale = "de", value = "und")
    @Localize(locale = "ru", value = "и")
    @Localize(locale = "es", value = "y")
    private static String AND;
    @Localize("for display of generated units")
    @Localize(locale = "de", value = "zur Anzeige berechneter Werte")
    @Localize(locale = "ru", value = "для отображения созданных единиц измерения")
    @Localize(locale = "es", value = "para mostrar las unidades generadas")
    private static String FOR_UNIT_DISPLAY;
    @Localize("Character point total display includes unspent points")
    @Localize(locale = "de",
              value = "Die Gesamtpunktzahl des Charakters enthält die noch nicht ausgegebenen Punkte")
    @Localize(locale = "ru",
              value = "Показывать в общих очках персонажа неизрасходованные (заработанные)")
    @Localize(locale = "es",
              value = "Total de puntos de personaje, que incluye los puntos no consumidos")
    private static String TOTAL_POINTS_INCLUDES_UNSPENT_POINTS;
    @Localize("Use GURPS Metric rules for height, weight, encumbrance and lifting things")
    @Localize(locale = "de",
              value = "Benutze GURPS Metrisch-Regeln für Höhe, Gewicht, Belastung und Heben von Gegenständen")
    @Localize(locale = "ru",
              value = "Использовать метрическую систему по правилам GURPS для роста, веса, нагрузки и грузоподъёмности")
    @Localize(locale = "es",
              value = "Usar las reglas de métrica de GURPS para altura, peso, carga y levantar objetos")
    private static String USE_METRIC_RULES;
    @Localize("All Readable Image Files")
    private static String ALL_READABLE_IMAGE_FILES;
    @Localize("JPEG Files")
    private static String JPEG_FILES;
    @Localize("GIF Files")
    private static String GIF_FILES;
    @Localize("GURPS Calculator Key")
    private static String GURPS_CALCULATOR_KEY;
    @Localize("Find mine")
    private static String WHERE_OBTAIN;
    @Localize("Unable to open {0}")
    private static String UNABLE_TO_OPEN_URL;
    @Localize("Show Advantage User Description as a Tooltip")
    @Localize(locale = "de", value = "Vorteilsbenutzerbeschreibung als QuickInfo anzeigen")
    @Localize(locale = "ru",
              value = "Показать описание пользователя Advantage в виде всплывающей подсказки")
    @Localize(locale = "es",
              value = "Mostrar la descripción del usuario de Advantage como información sobre herramientas")
    private static String SHOW_USER_DESC_AS_TOOL_TIP;

    static {
        Localization.initialize();
    }

    static final String              MODULE                             = "Sheet"; //$NON-NLS-1$
    private static final String      OPTIONAL_DICE_RULES_KEY            = "UseOptionDiceRules"; //$NON-NLS-1$
    /** The optional dice rules preference key. */
    public static final String       OPTIONAL_DICE_RULES_PREF_KEY       = Preferences.getModuleKey(MODULE, OPTIONAL_DICE_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_DICE_RULES        = false;
    private static final String      OPTIONAL_IQ_RULES_KEY              = "UseOptionIQRules"; //$NON-NLS-1$
    /** The optional IQ rules preference key. */
    public static final String       OPTIONAL_IQ_RULES_PREF_KEY         = Preferences.getModuleKey(MODULE, OPTIONAL_IQ_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_IQ_RULES          = false;
    private static final String      OPTIONAL_MODIFIER_RULES_KEY        = "UseOptionModifierRules"; //$NON-NLS-1$
    /** The optional modifier rules preference key. */
    public static final String       OPTIONAL_MODIFIER_RULES_PREF_KEY   = Preferences.getModuleKey(MODULE, OPTIONAL_MODIFIER_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_MODIFIER_RULES    = false;
    private static final String      OPTIONAL_STRENGTH_RULES_KEY        = "UseOptionalStrengthRules"; //$NON-NLS-1$
    /** The optional Strength rules preference key. */
    public static final String       OPTIONAL_STRENGTH_RULES_PREF_KEY   = Preferences.getModuleKey(MODULE, OPTIONAL_STRENGTH_RULES_KEY);
    private static final boolean     DEFAULT_OPTIONAL_STRENGTH_RULES    = false;
    private static final String      OPTIONAL_REDUCED_SWING_KEY         = "UseOptionalReducedSwing"; //$NON-NLS-1$
    /** The optional Reduced Swing rules preference key. */
    public static final String       OPTIONAL_REDUCED_SWING_PREF_KEY    = Preferences.getModuleKey(MODULE, OPTIONAL_REDUCED_SWING_KEY);
    private static final boolean     DEFAULT_OPTIONAL_REDUCED_SWING     = false;
    private static final String      AUTO_NAME_KEY                      = "AutoNameNewCharacters"; //$NON-NLS-1$
    /** The auto-naming preference key. */
    public static final String       AUTO_NAME_PREF_KEY                 = Preferences.getModuleKey(MODULE, AUTO_NAME_KEY);
    private static final boolean     DEFAULT_AUTO_NAME                  = true;
    /** The optional Thrust Damage rules preference key. */
    private static final String      OPTIONAL_THRUST_DAMAGE_KEY         = "UseOptionalThurstDamage"; //$NON-NLS-1$
    public static final String       OPTIONAL_THRUST_DAMAGE_PREF_KEY    = Preferences.getModuleKey(MODULE, OPTIONAL_THRUST_DAMAGE_KEY);
    private static final boolean     DEFAULT_OPTIONAL_THRUST_DAMAGE     = false;
    private static final String      LENGTH_UNITS_KEY                   = "LengthUnits"; //$NON-NLS-1$
    /** The default length units preference key. */
    public static final String       LENGTH_UNITS_PREF_KEY              = Preferences.getModuleKey(MODULE, LENGTH_UNITS_KEY);
    private static final LengthUnits DEFAULT_LENGTH_UNITS               = LengthUnits.FT_IN;
    private static final String      WEIGHT_UNITS_KEY                   = "WeightUnits"; //$NON-NLS-1$
    /** The default weight units preference key. */
    public static final String       WEIGHT_UNITS_PREF_KEY              = Preferences.getModuleKey(MODULE, WEIGHT_UNITS_KEY);
    private static final WeightUnits DEFAULT_WEIGHT_UNITS               = WeightUnits.LB;
    private static final String      TOTAL_POINTS_DISPLAY_KEY           = "TotalPointsIncludesUnspentPoints"; //$NON-NLS-1$
    /** The total points includes unspent points preference key. */
    public static final String       TOTAL_POINTS_DISPLAY_PREF_KEY      = Preferences.getModuleKey(MODULE, TOTAL_POINTS_DISPLAY_KEY);
    private static final boolean     DEFAULT_TOTAL_POINTS_DISPLAY       = true;
    private static final String      GURPS_METRIC_RULES_KEY             = "UseGurpsMetricRules"; //$NON-NLS-1$
    /** The GURPS Metric preference key. */
    public static final String       GURPS_METRIC_RULES_PREF_KEY        = Preferences.getModuleKey(MODULE, GURPS_METRIC_RULES_KEY);

    public static final String       SHOW_USER_DESC_AS_TOOL_TIP_KEY     = "ShowUserDescAsToolTip"; //$NON-NLS-1$
    private static final boolean     DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP = true;

    private static final boolean     DEFAULT_GURPS_METRIC_RULES         = true;
    private static final Scales      DEFAULT_SCALE                      = Scales.ACTUAL_SIZE;
    private static final String      SCALE_KEY                          = "UIScale"; //$NON-NLS-1$
    private static final String      INITIAL_POINTS_KEY                 = "InitialPoints"; //$NON-NLS-1$
    private static final int         DEFAULT_INITIAL_POINTS             = 100;
    private JTextField               mPlayerName;
    private JTextField               mCampaign;
    private JTextField               mTechLevel;
    private JTextField               mInitialPoints;
    private PortraitPreferencePanel  mPortrait;
    private JComboBox<Scales>        mUIScaleCombo;
    private JComboBox<String>        mLengthUnitsCombo;
    private JComboBox<String>        mWeightUnitsCombo;
    private JCheckBox                mUseOptionalDiceRules;
    private JCheckBox                mUseOptionalIQRules;
    private JCheckBox                mUseOptionalModifierRules;
    private JCheckBox                mUseOptionalStrengthRules;
    private JCheckBox                mUseOptionalReducedSwing;
    private JCheckBox                mIncludeUnspentPointsInTotal;
    private JCheckBox                mUseGurpsMetricRules;
    private JCheckBox                mAutoName;
    private JCheckBox                mShowUserDescAsToolTips;
    private JCheckBox                mUseOptionalThrustDamage;

    /** Initializes the services controlled by these preferences. */
    public static void initialize() {
        adjustOptionalDiceRulesProperty(areOptionalDiceRulesUsed());
    }

    /** @return The default length units to use. */
    public static LengthUnits getLengthUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, LENGTH_UNITS_KEY), LengthUnits.values(), DEFAULT_LENGTH_UNITS);
    }

    /** @return The default weight units to use. */
    public static WeightUnits getWeightUnits() {
        return Enums.extract(Preferences.getInstance().getStringValue(MODULE, WEIGHT_UNITS_KEY), WeightUnits.values(), DEFAULT_WEIGHT_UNITS);
    }

    private static void adjustOptionalDiceRulesProperty(boolean use) {
        Dice.setConvertModifiersToExtraDice(use);
    }

    /** @return Whether the optional dice rules from B269 are in use. */
    public static boolean areOptionalDiceRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_DICE_RULES_KEY, DEFAULT_OPTIONAL_DICE_RULES);
    }

    /**
     * @return Whether the optional IQ rules (Will &amp; Perception are not based on IQ) are in use.
     */
    public static boolean areOptionalIQRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_IQ_RULES_KEY, DEFAULT_OPTIONAL_IQ_RULES);
    }

    /** @return Whether the optional modifier rules from PW102 are in use. */
    public static boolean areOptionalModifierRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_MODIFIER_RULES_KEY, DEFAULT_OPTIONAL_MODIFIER_RULES);
    }

    /** @return Whether the optional strength rules (KYOS) are in use. */
    public static boolean areOptionalStrengthRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_STRENGTH_RULES_KEY, DEFAULT_OPTIONAL_STRENGTH_RULES);
    }

    /** @return Whether the optional thrust damage (sw-2) rules are in use. */
    public static boolean areOptionalThrustDamageUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_THRUST_DAMAGE_KEY, DEFAULT_OPTIONAL_THRUST_DAMAGE);
    }

    /**
     * @return Whether the optional reduced swing rules are in use. Reduces KYOS damages if used
     *         together.
     */
    public static boolean areOptionalReducedSwingUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, OPTIONAL_REDUCED_SWING_KEY, DEFAULT_OPTIONAL_REDUCED_SWING);
    }

    /**
     * @return Whether the character's total points are displayed with or without including earned
     *         (but unspent) points.
     */
    public static boolean shouldIncludeUnspentPointsInTotalPointDisplay() {
        return Preferences.getInstance().getBooleanValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, DEFAULT_TOTAL_POINTS_DISPLAY);
    }

    /** @return Whether the GURPS Metrics rules are used for weight and height conversion. */
    public static boolean areGurpsMetricRulesUsed() {
        return Preferences.getInstance().getBooleanValue(MODULE, GURPS_METRIC_RULES_KEY, DEFAULT_GURPS_METRIC_RULES);
    }

    /** @return Whether a new character should be automatically named. */
    public static boolean isNewCharacterAutoNamed() {
        return Preferences.getInstance().getBooleanValue(MODULE, AUTO_NAME_KEY, DEFAULT_AUTO_NAME);
    }

    /** @return The {@link Scales} to use when opening a new scalable file. */
    public static Scales getInitialUIScale() {
        double value = Preferences.getInstance().getDoubleValue(MODULE, SCALE_KEY, DEFAULT_SCALE.getScale().getScale());
        for (Scales one : Scales.values()) {
            if (one.getScale().getScale() == value) {
                return one;
            }
        }
        return Scales.ACTUAL_SIZE;
    }

    /** @return The initial points to start a new character with. */
    public static int getInitialPoints() {
        return Preferences.getInstance().getIntValue(MODULE, INITIAL_POINTS_KEY, DEFAULT_INITIAL_POINTS);
    }

    public static boolean showUserDescAsTooltip() {
        return Preferences.getInstance().getBooleanValue(MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, DEFAULT_SHOW_USER_DESC_AS_TOOL_TIP);
    }

    /**
     * Creates a new {@link SheetPreferences}.
     *
     * @param owner The owning {@link PreferencesWindow}.
     */
    public SheetPreferences(PreferencesWindow owner) {
        super(SHEET, owner);
        FlexColumn column = new FlexColumn();

        FlexGrid   grid   = new FlexGrid();
        column.add(grid);

        int rowIndex = 0;
        mPortrait = createPortrait();
        FlexComponent comp = new FlexComponent(mPortrait, Alignment.LEFT_TOP, Alignment.LEFT_TOP);
        grid.add(comp, rowIndex, 0, 4, 1);

        grid.add(createFlexLabel(PLAYER, PLAYER_TOOLTIP), rowIndex, 1);
        mPlayerName = createTextField(PLAYER_TOOLTIP, Profile.getDefaultPlayerName());
        grid.add(mPlayerName, rowIndex++, 2);

        grid.add(createFlexLabel(CAMPAIGN, CAMPAIGN_TOOLTIP), rowIndex, 1);
        mCampaign = createTextField(CAMPAIGN_TOOLTIP, Profile.getDefaultCampaign());
        grid.add(mCampaign, rowIndex++, 2);

        grid.add(createFlexLabel(TECH_LEVEL, TECH_LEVEL_TOOLTIP), rowIndex, 1);
        mTechLevel = createTextField(TECH_LEVEL_TOOLTIP, Profile.getDefaultTechLevel());
        grid.add(mTechLevel, rowIndex++, 2);

        grid.add(createFlexLabel(INITIAL_POINTS, INITIAL_POINTS_TOOLTIP), rowIndex, 1);
        mInitialPoints = createTextField(INITIAL_POINTS_TOOLTIP, Integer.toString(getInitialPoints()));
        grid.add(mInitialPoints, rowIndex++, 2);

        grid.add(new FlexSpacer(0, 0, false, false), rowIndex, 1);
        grid.add(new FlexSpacer(0, 0, true, false), rowIndex, 2);

        FlexRow row = new FlexRow();
        row.add(createLabel(USE, null));
        mLengthUnitsCombo = createLengthUnitsPopup();
        row.add(mLengthUnitsCombo);
        row.add(createLabel(AND, null));
        mWeightUnitsCombo = createWeightUnitsPopup();
        row.add(mWeightUnitsCombo);
        row.add(createLabel(FOR_UNIT_DISPLAY, null));
        column.add(row);

        mAutoName = createCheckBox(AUTO_NAME, null, isNewCharacterAutoNamed());
        column.add(mAutoName);

        mUseOptionalIQRules = createCheckBox(OPTIONAL_IQ_RULES, null, areOptionalIQRulesUsed());
        column.add(mUseOptionalIQRules);

        mUseOptionalModifierRules = createCheckBox(OPTIONAL_MODIFIER_RULES, null, areOptionalModifierRulesUsed());
        column.add(mUseOptionalModifierRules);

        mUseOptionalDiceRules = createCheckBox(OPTIONAL_DICE_RULES, null, areOptionalDiceRulesUsed());
        column.add(mUseOptionalDiceRules);

        mUseOptionalStrengthRules = createCheckBox(OPTIONAL_STRENGTH_RULES, null, areOptionalStrengthRulesUsed());
        column.add(mUseOptionalStrengthRules);

        mUseOptionalReducedSwing = createCheckBox(OPTIONAL_REDUCED_SWING, null, areOptionalReducedSwingUsed());
        column.add(mUseOptionalReducedSwing);

        mUseOptionalThrustDamage = createCheckBox(OPTIONAL_THRUST_DAMAGE, null, areOptionalThrustDamageUsed());
        column.add(mUseOptionalThrustDamage);

        mIncludeUnspentPointsInTotal = createCheckBox(TOTAL_POINTS_INCLUDES_UNSPENT_POINTS, null, shouldIncludeUnspentPointsInTotalPointDisplay());
        column.add(mIncludeUnspentPointsInTotal);

        mUseGurpsMetricRules = createCheckBox(USE_METRIC_RULES, null, areGurpsMetricRulesUsed());
        column.add(mUseGurpsMetricRules);

        mShowUserDescAsToolTips = createCheckBox(SHOW_USER_DESC_AS_TOOL_TIP, SHOW_USER_DESC_AS_TOOL_TIP, showUserDescAsTooltip());
        column.add(mShowUserDescAsToolTips);

        row = new FlexRow();
        row.add(createLabel(USE, null));
        mUIScaleCombo = createUIScalePopup();
        row.add(mUIScaleCombo);
        row.add(createLabel(UI_SCALE_POST, null, SwingConstants.LEFT));
        column.add(row);

        column.add(new FlexSpacer(0, 0, false, true));

        column.apply(this);
    }

    private FlexComponent createFlexLabel(String title, String tooltip) {
        return new FlexComponent(createLabel(title, tooltip), Alignment.RIGHT_BOTTOM, Alignment.CENTER);
    }

    private PortraitPreferencePanel createPortrait() {
        PortraitPreferencePanel panel = new PortraitPreferencePanel(Profile.getPortraitFromPortraitPath(Profile.getDefaultPortraitPath()));
        panel.addActionListener(this);
        add(panel);
        return panel;
    }

    private JComboBox<Scales> createUIScalePopup() {
        JComboBox<Scales> combo = new JComboBox<>(Scales.values());
        setupCombo(combo, null);
        combo.setSelectedItem(getInitialUIScale());
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setOnlySize(combo, combo.getPreferredSize());
        return combo;
    }

    private JComboBox<String> createLengthUnitsPopup() {
        JComboBox<String> combo = new JComboBox<>();
        setupCombo(combo, LENGTH_UNITS_TOOLTIP);
        for (LengthUnits unit : LengthUnits.values()) {
            combo.addItem(unit.getDescription());
        }
        combo.setSelectedIndex(getLengthUnits().ordinal());
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setOnlySize(combo, combo.getPreferredSize());
        return combo;
    }

    private JComboBox<String> createWeightUnitsPopup() {
        JComboBox<String> combo = new JComboBox<>();
        setupCombo(combo, WEIGHT_UNITS_TOOLTIP);
        for (WeightUnits unit : WeightUnits.values()) {
            combo.addItem(unit.getDescription());
        }
        combo.setSelectedIndex(getWeightUnits().ordinal());
        combo.addActionListener(this);
        combo.setMaximumRowCount(combo.getItemCount());
        UIUtilities.setOnlySize(combo, combo.getPreferredSize());
        return combo;
    }

    private JTextField createTextField(String tooltip, String value) {
        JTextField field = new JTextField(value);
        field.setToolTipText(Text.wrapPlainTextForToolTip(tooltip));
        field.getDocument().addDocumentListener(this);
        Dimension size    = field.getPreferredSize();
        Dimension maxSize = field.getMaximumSize();
        maxSize.height = size.height;
        field.setMaximumSize(maxSize);
        add(field);
        return field;
    }

    public static File choosePortrait() {
        List<FileNameExtensionFilter> filters = new ArrayList<>();
        filters.add(new FileNameExtensionFilter(ALL_READABLE_IMAGE_FILES, FileType.PNG_EXTENSION, FileType.JPEG_EXTENSION, "jpeg", FileType.GIF_EXTENSION)); //$NON-NLS-1$
        filters.add(FileType.getPngFilter());
        filters.add(FileType.getJpegFilter());
        filters.add(FileType.getGifFilter());
        return StdFileDialog.showOpenDialog(null, SELECT_PORTRAIT, filters);
    }

    @Override
    public void actionPerformed(ActionEvent event) {
        Object source = event.getSource();
        if (source == mPortrait) {
            File file = choosePortrait();
            if (file != null) {
                setPortrait(PathUtils.getFullPath(file));
            }
        } else if (source == mUIScaleCombo) {
            Preferences.getInstance().setValue(MODULE, SCALE_KEY, ((Scales) mUIScaleCombo.getSelectedItem()).getScale().getScale());
        } else if (source == mLengthUnitsCombo) {
            Preferences.getInstance().setValue(MODULE, LENGTH_UNITS_KEY, Enums.toId(LengthUnits.values()[mLengthUnitsCombo.getSelectedIndex()]));
        } else if (source == mWeightUnitsCombo) {
            Preferences.getInstance().setValue(MODULE, WEIGHT_UNITS_KEY, Enums.toId(WeightUnits.values()[mWeightUnitsCombo.getSelectedIndex()]));
        }
        adjustResetButton();
    }

    @Override
    public void reset() {
        mPlayerName.setText(System.getProperty("user.name")); //$NON-NLS-1$
        mCampaign.setText(""); //$NON-NLS-1$
        mTechLevel.setText(Profile.DEFAULT_TECH_LEVEL);
        mInitialPoints.setText(Integer.toString(DEFAULT_INITIAL_POINTS));
        setPortrait(Profile.DEFAULT_PORTRAIT);
        mUIScaleCombo.setSelectedItem(DEFAULT_SCALE);
        mLengthUnitsCombo.setSelectedIndex(DEFAULT_LENGTH_UNITS.ordinal());
        mWeightUnitsCombo.setSelectedIndex(DEFAULT_WEIGHT_UNITS.ordinal());
        mAutoName.setSelected(DEFAULT_AUTO_NAME);
        mUseOptionalDiceRules.setSelected(DEFAULT_OPTIONAL_DICE_RULES);
        mUseOptionalIQRules.setSelected(DEFAULT_OPTIONAL_IQ_RULES);
        mUseOptionalModifierRules.setSelected(DEFAULT_OPTIONAL_MODIFIER_RULES);
        mUseOptionalStrengthRules.setSelected(DEFAULT_OPTIONAL_STRENGTH_RULES);
        mUseOptionalThrustDamage.setSelected(DEFAULT_OPTIONAL_THRUST_DAMAGE);
        mUseOptionalReducedSwing.setSelected(DEFAULT_OPTIONAL_REDUCED_SWING);
        mIncludeUnspentPointsInTotal.setSelected(DEFAULT_TOTAL_POINTS_DISPLAY);
        mUseGurpsMetricRules.setSelected(DEFAULT_GURPS_METRIC_RULES);
    }

    @Override
    public boolean isSetToDefaults() {
        return Profile.getDefaultPlayerName().equals(System.getProperty("user.name")) && Profile.getDefaultCampaign().equals("") && Profile.getDefaultPortraitPath().equals(Profile.DEFAULT_PORTRAIT) && Profile.getDefaultTechLevel().equals(Profile.DEFAULT_TECH_LEVEL) && getInitialPoints() == DEFAULT_INITIAL_POINTS && getInitialUIScale() == DEFAULT_SCALE && areOptionalDiceRulesUsed() == DEFAULT_OPTIONAL_DICE_RULES && areOptionalIQRulesUsed() == DEFAULT_OPTIONAL_IQ_RULES && areOptionalModifierRulesUsed() == DEFAULT_OPTIONAL_MODIFIER_RULES && areOptionalStrengthRulesUsed() == DEFAULT_OPTIONAL_STRENGTH_RULES && areOptionalReducedSwingUsed() == DEFAULT_OPTIONAL_REDUCED_SWING && isNewCharacterAutoNamed() == DEFAULT_AUTO_NAME; //$NON-NLS-1$ //$NON-NLS-2$
    }

    private void setPortrait(String path) {
        StdImage image = Profile.getPortraitFromPortraitPath(path);
        Profile.setDefaultPortraitPath(path);
        mPortrait.setPortrait(image);
    }

    @Override
    public void changedUpdate(DocumentEvent event) {
        Document document = event.getDocument();
        if (mPlayerName.getDocument() == document) {
            Profile.setDefaultPlayerName(mPlayerName.getText());
        } else if (mCampaign.getDocument() == document) {
            Profile.setDefaultCampaign(mCampaign.getText());
        } else if (mTechLevel.getDocument() == document) {
            Profile.setDefaultTechLevel(mTechLevel.getText());
        } else if (mInitialPoints.getDocument() == document) {
            Preferences.getInstance().setValue(MODULE, INITIAL_POINTS_KEY, Numbers.extractInteger(mInitialPoints.getText(), 0, true));
        }
        adjustResetButton();
    }

    @Override
    public void insertUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

    @Override
    public void removeUpdate(DocumentEvent event) {
        changedUpdate(event);
    }

    @Override
    public void itemStateChanged(ItemEvent event) {
        Object source = event.getSource();
        if (source == mUseOptionalDiceRules) {
            boolean checked = mUseOptionalDiceRules.isSelected();
            adjustOptionalDiceRulesProperty(checked);
            Preferences.getInstance().setValue(MODULE, OPTIONAL_DICE_RULES_KEY, checked);
        } else if (source == mUseOptionalIQRules) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_IQ_RULES_KEY, mUseOptionalIQRules.isSelected());
        } else if (source == mUseOptionalModifierRules) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_MODIFIER_RULES_KEY, mUseOptionalModifierRules.isSelected());
        } else if (source == mUseOptionalStrengthRules) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_STRENGTH_RULES_KEY, mUseOptionalStrengthRules.isSelected());
        } else if (source == mUseOptionalThrustDamage) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_THRUST_DAMAGE_KEY, mUseOptionalThrustDamage.isSelected());
        } else if (source == mUseOptionalReducedSwing) {
            Preferences.getInstance().setValue(MODULE, OPTIONAL_REDUCED_SWING_KEY, mUseOptionalReducedSwing.isSelected());
        } else if (source == mIncludeUnspentPointsInTotal) {
            Preferences.getInstance().setValue(MODULE, TOTAL_POINTS_DISPLAY_KEY, mIncludeUnspentPointsInTotal.isSelected());
        } else if (source == mUseGurpsMetricRules) {
            Preferences.getInstance().setValue(MODULE, GURPS_METRIC_RULES_KEY, mUseGurpsMetricRules.isSelected());
        } else if (source == mShowUserDescAsToolTips) {
            Preferences.getInstance().setValue(MODULE, SHOW_USER_DESC_AS_TOOL_TIP_KEY, mShowUserDescAsToolTips.isSelected());
        } else if (source == mAutoName) {
            Preferences.getInstance().setValue(MODULE, AUTO_NAME_KEY, mAutoName.isSelected());
        }
        adjustResetButton();
    }
}
