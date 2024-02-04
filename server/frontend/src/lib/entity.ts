/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

import { get, writable } from 'svelte/store';
import { apiPrefix } from '$lib/dev.ts';
import { session } from '$lib/session.ts';

export const pc = writable<Entity | undefined>();

export async function loadEntity(file: File) {
	return new Promise<Entity>((resolve, reject) => {
		const reader = new FileReader();
		reader.onload = () => resolve(parsePC(reader.result as string));
		reader.onerror = () => reject;
		reader.readAsText(file);
	});
}

export async function fetchSheet(path :string) {
	const rsp = await fetch(apiPrefix(`/sheet/${path}`), {
		method: 'GET',
		headers: { 'X-Session': get(session)?.ID ?? '' },
		cache: 'no-store'
	});
	if (!rsp.ok) {
		return undefined;
	}
	return validateSheet(await rsp.json());
}

function parsePC(data: string) {
	const json = JSON.parse(data);
	const newPC = json as Entity;
	validateSheet(newPC);
	return newPC;
}

export function validateSheet(entity: Entity) {
	// Convert old format to new format
	if (!entity.traits && entity.advantages) {
		entity.traits = entity.advantages;
		delete entity.advantages;
	}
	for (const equipment of entity.equipment ?? []) {
		if (!equipment.tags && equipment.categories) {
			equipment.tags = equipment.categories;
			delete equipment.categories;
		}
		for (const modifier of equipment.modifiers ?? []) {
			if (modifier.cost_type === undefined) {
				modifier.cost_type = EMCostType.Original;
			}
			if (modifier.weight_type === undefined) {
				modifier.weight_type = EMWeightType.Original;
			}
		}
	}
	if (entity.settings) {
		if (!entity.settings.body_type && entity.settings.hit_locations) {
			entity.settings.body_type = entity.settings.hit_locations;
			delete entity.settings.hit_locations;
		}
		if (!entity.settings.show_trait_modifier_adj && entity.settings.show_advantage_modifier_adj) {
			entity.settings.show_trait_modifier_adj = entity.settings.show_advantage_modifier_adj;
			delete entity.settings.show_advantage_modifier_adj;
		}
	}
	return entity;
}

export interface Entity {
	type: EntityType;
	version: number;
	id: string;
	total_points: number;
	points_record?: PointsRecord[];
	profile?: Profile;
	settings?: SheetSettings;
	attributes?: Attribute[];
	traits?: Trait[];
	advantages?: Trait[]; // Old key for traits
	skills?: Skill[];
	spells?: Spell[];
	equipment?: Equipment[];
	other_equipment?: Equipment[];
	notes?: Note[];
	created_date: string;
	modified_date: string;
	calc: EntityCalc;
	third_party?: { [key: string]: unknown };
}

export enum EntityType {
	Character = 'character'
}

export interface PointsRecord {
	when: string;
	points: number;
	reason?: string;
}

export interface Profile {
	name?: string;
	age?: string;
	birthday?: string;
	eyes?: string;
	hair?: string;
	skin?: string;
	handedness?: string;
	gender?: string;
	height?: string;
	weight?: string;
	player_name?: string;
	title?: string;
	organization?: string;
	religion?: string;
	tech_level?: string;
	portrait?: string;
	SM?: number;
}

export interface EntityCalc {
	swing: string;
	thrust: string;
	basic_lift: string;
	lifting_st_bonus?: number;
	striking_st_bonus?: number;
	throwing_st_bonus?: number;
	dodge_bonus?: number;
	parry_bonus?: number;
	block_bonus?: number;
	move: number[];
	dodge: number[];
}

export interface SheetSettings {
	page?: PageSettings;
	block_layout?: string[];
	attributes?: AttributeDef[];
	body_type?: Body;
	hit_locations?: Body; // Old key for body_type
	damage_progression: DamageProgression;
	default_length_units: string;
	default_weight_units: string;
	user_description_display: DisplayOption;
	modifiers_display: DisplayOption;
	notes_display: DisplayOption;
	skill_level_adj_display: DisplayOption;
	use_multiplicative_modifiers?: boolean;
	use_modifying_dice_plus_adds?: boolean;
	use_half_stat_defaults?: boolean;
	show_trait_modifier_adj?: boolean;
	show_advantage_modifier_adj?: boolean; // Old key for show_trait_modifier_adj
	show_equipment_modifier_adj?: boolean;
	show_spell_adj?: boolean;
	use_title_in_footer?: boolean;
	exclude_unspent_points_from_total: boolean;
}

export interface PageSettings {
	paper_size: string;
	orientation: string;
	top_margin: string;
	left_margin: string;
	bottom_margin: string;
	right_margin: string;
}

export interface AttributeDefs {
	type: string;
	version: number;
	rows: AttributeDef[];
	attributes: AttributeDef[]; // Old key for rows
	attribute_settings: AttributeDef[]; // Old key for attributes
}

export interface AttributeDef {
	id: string;
	type: AttributeType;
	name: string;
	full_name?: string;
	attribute_base?: string;
	cost_per_point?: number;
	cost_adj_percent_per_sm?: number;
	thresholds?: PoolThreshold[];
}

export enum AttributeType {
	Integer = 'integer',
	IntegerRef = 'integer_ref',
	Decimal = 'decimal',
	DecimalRef = 'decimal_ref',
	Pool = 'pool',
	PrimarySeparator = 'primary_separator',
	SecondarySeparator = 'secondary_separator',
	PoolSeparator = 'pool_separator'
}

export interface PoolThreshold {
	state: string;
	expression: string;
	explanation?: string;
	ops?: ThresholdOp[];
}

export enum ThresholdOp {
	HalveDodge = 'halve_dodge',
	HalveMove = 'halve_move',
	HalveSt = 'halve_st'
}

export interface Body {
	name?: string;
	roll: string;
	locations?: HitLocation[];
}

export interface HitLocation {
	id: string;
	choice_name: string;
	table_name: string;
	slots?: number;
	hit_penalty?: number;
	dr_bonus?: number;
	description?: string;
	calc: HitLocationCalc;
	sub_table?: Body;
}

export interface HitLocationCalc {
	roll_range: string;
	dr?: { [key: string]: number };
}

export enum DamageProgression {
	BasicSet = 'basic_set',
	KnowingYourOwnStrength = 'knowing_your_own_strength',
	NoSchoolGrognardDamage = 'no_school_grognard_damage',
	ThrustEqualsSwingMinus2 = 'thrust_equals_swing_minus_2',
	SwingEqualsThrustPlus2 = 'swing_equals_thrust_plus_2',
	Tbone1 = 'tbone_1',
	Tbone1Clean = 'tbone_1_clean',
	Tbone2 = 'tbone_2',
	Tbone2Clean = 'tbone_2_clean',
	PhoenixFlameD3 = 'phoenix_flame_d3'
}

export enum DisplayOption {
	NotShownDisplayOption = 'not_shown',
	InlineDisplayOption = 'inline',
	TooltipDisplayOption = 'tooltip',
	InlineAndTooltipDisplayOption = 'inline_and_tooltip'
}

export interface Attribute {
	attr_id: string;
	adj: number;
	damage?: number;
	calc?: AttributeCalc;
}

export interface AttributeCalc {
	value: number;
	current?: number;
	points: number;
}

export interface Trait {
	// TODO: Define
}

export interface Skill {
	// TODO: Define
}

export interface Spell {
	// TODO: Define
}

export interface Note {
	// TODO: Define
}

export interface Equipment {
	id: string;
	type: string;
	description?: string;
	reference?: string;
	reference_highlight?: string;
	notes?: string;
	vtt_notes?: string;
	tech_level?: string;
	legality_class?: string;
	tags?: string[];
	categories?: string[]; // Old key for tags
	modifiers?: EquipmentModifier[];
	rated_strength?: number;
	quantity?: number;
	value?: number;
	weight?: string;
	max_uses?: number;
	uses?: number;
	prereqs?: PrereqList;
	weapons?: Weapon[];
	features?: Features;
	open?: boolean;
	equipped?: boolean;
	ignore_weight_for_skills?: boolean;
	children?: Equipment[];
	calc: EquipmentCalc;
	third_party?: { [key: string]: unknown };
}

export interface EquipmentCalc {
	extended_value: number;
	extended_weight: string;
	extended_weight_for_skills?: string;
	resolved_notes?: string;
	unsatisfied_reason?: string;
}

export interface EquipmentModifier {
	id: string;
	type: string;
	open?: boolean;
	children?: EquipmentModifier[];
	name?: string;
	reference?: string;
	reference_highlight?: string;
	notes?: string;
	vtt_notes?: string;
	tags?: string[];
	cost_type?: EMCostType;
	weight_type?: EMWeightType;
	disabled?: boolean;
	tech_level?: string;
	cost?: string;
	weight?: string;
	features?: Features;
	calc?: EquipmentModifierCalc;
	third_party?: { [key: string]: unknown };
}

export enum EMCostType {
	Original = 'to_original_cost',
	Base = 'to_base_cost',
	FinalBase = 'to_final_base_cost',
	Final = 'to_final_cost'
}

export enum EMWeightType {
	Original = 'to_original_weight',
	Base = 'to_base_weight',
	FinalBase = 'to_final_base_weight',
	Final = 'to_final_weight'
}

export interface EquipmentModifierCalc {
	resolved_notes?: string;
}

export interface Features {
	// TODO: Define
}

export interface PrereqList {
	// TODO: Define
}

export interface Weapon {
	// TODO: Define
}
