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

const MULTIPLIER = 10000;

// Int holds a fixed-point value that contains 4 decimal places. Values are truncated, not rounded.
export class Fixed {
	private value = 0;

	constructor(input?: number | string | Fixed | undefined) {
		switch (typeof input) {
			case 'object':
				this.value = input.value;
				break;
			case 'number':
				this.value = Math.trunc(input * MULTIPLIER);
				break;
			case 'string':
				{
					const parts = input.split('.', 2);
					if (parts.length > 0) {
						this.value = parseInt(parts[0]) * MULTIPLIER;
						if (parts.length > 1) {
							let buf = '1' + parts[1];
							while (buf.length < 5) {
								buf += '0';
							}
							this.value += parseInt(buf.slice(0, 5)) - 10000;
						}
					}
				}
				break;
		}
	}

	// add adds this value to the passed-in value, returning a new value.
	add(input: Fixed): Fixed {
		const result = new Fixed();
		result.value = this.value + input.value;
		return result;
	}

	// sub subtracts the passed-in value from this value, returning a new value.
	sub(input: Fixed): Fixed {
		const result = new Fixed();
		result.value = this.value - input.value;
		return result;
	}

	// mul multiplies this value by the passed-in value, returning a new value.
	mul(input: Fixed): Fixed {
		const result = new Fixed();
		result.value = Math.trunc((this.value * input.value) / MULTIPLIER);
		return result;
	}

	// div divides this value by the passed-in value, returning a new value.
	div(input: Fixed): Fixed {
		const result = new Fixed();
		result.value = Math.trunc((this.value * MULTIPLIER) / input.value);
		return result;
	}

	// mod returns the remainder after subtracting all full multiples of the passed-in value.
	mod(input: Fixed): Fixed {
		const result = new Fixed();
		result.value = this.value - input.mul(this.div(input).trunc()).value;
		return result;
	}

	// abs returns the absolute value of this value.
	abs(): Fixed {
		const result = new Fixed();
		result.value = Math.abs(this.value);
		return result;
	}

	// trunc returns a new value which has everything to the right of the decimal place truncated.
	trunc(): Fixed {
		const result = new Fixed();
		result.value = Math.trunc(this.value / MULTIPLIER) * MULTIPLIER;
		return result;
	}

	// ceil returns the value rounded up to the nearest whole number.
	ceil(): Fixed {
		const result = this.trunc();
		if (this.value > 0 && this.value !== result.value) {
			result.value += MULTIPLIER;
		}
		return result;
	}

	// round returns the nearest integer, rounding half away from zero.
	round(): Fixed {
		const result = this.trunc();
		const remainder = this.value - result.value;
		if (remainder >= MULTIPLIER / 2) {
			result.value += MULTIPLIER;
		} else if (remainder < -MULTIPLIER / 2) {
			result.value -= MULTIPLIER;
		}
		return result;
	}

	// min returns the minimum of this value or its argument.
	min(input: Fixed): Fixed {
		const result = new Fixed();
		result.value = Math.min(this.value, input.value);
		return result;
	}

	// max returns the maximum of this value or its argument.
	max(input: Fixed): Fixed {
		const result = new Fixed();
		result.value = Math.max(this.value, input.value);
		return result;
	}

	// inc returns the value incremented by 1.
	inc(): Fixed {
		const result = new Fixed();
		result.value = this.value + MULTIPLIER;
		return result;
	}

	// dec returns the value decremented by 1.
	dec(): Fixed {
		const result = new Fixed();
		result.value = this.value - MULTIPLIER;
		return result;
	}

	// asFloat returns the value as a floating-point number.
	asFloat() {
		return this.value / MULTIPLIER;
	}

	// asInteger returns the value as an integer.
	asInteger() {
		return Math.trunc(this.value / MULTIPLIER);
	}

	eq(input: Fixed) {
		return this.value === input.value;
	}

	gt(input: Fixed) {
		return this.value > input.value;
	}

	lt(input: Fixed) {
		return this.value < input.value;
	}

	gte(input: Fixed) {
		return this.value >= input.value;
	}

	lte(input: Fixed) {
		return this.value <= input.value;
	}

	// comma returns the value as a string with commas separating the thousands.
	comma(withSign?: boolean) {
		let result = '';
		let str = this.toString(withSign);
		if (str.startsWith('-') || str.startsWith('+')) {
			result = str[0];
			str = str.slice(1);
		}
		const parts = str.split('.', 2);
		const left = parts[0];
		let i = 0;
		let needComma = false;
		if (left.length % 3 !== 0) {
			i += left.length % 3;
			result += left.slice(0, i);
			needComma = true;
		}
		for (; i < left.length; i += 3) {
			if (needComma) {
				result += ',';
			} else {
				needComma = true;
			}
			result += left.slice(i, i + 3);
		}
		if (parts.length > 1) {
			result += '.' + parts[1];
		}
		return result;
	}

	// toString returns the value as a string.
	toString(withSign?: boolean) {
		const integer = this.asInteger();
		let fraction = this.value % MULTIPLIER;
		let result: string;
		if (fraction === 0) {
			result = integer.toString();
		} else {
			if (fraction < 0) {
				fraction = -fraction;
			}
			fraction += MULTIPLIER;
			result = fraction.toString();
			for (let i = result.length - 1; i > 0; i--) {
				if (result[i] !== '0') {
					result = result.slice(1, i + 1);
					break;
				}
			}
			result = (integer === 0 && this.value < 0 ? '-' : '') + integer.toString() + '.' + result;
		}
		if (withSign && this.value >= 0) {
			return '+' + result;
		}
		return result;
	}

	// applyRounding rounds in the positive direction of roundDown is false, or in the negative
	// direction if roundDown is true.
	applyRounding(roundDown?: boolean) {
		const truncated = this.trunc();
		if (this.value !== truncated.value) {
			if (roundDown) {
				if (this.value < 0) {
					truncated.value -= MULTIPLIER;
				}
			} else if (this.value > 0) {
				truncated.value += MULTIPLIER;
			}
			return truncated;
		}
		return this;
	}

	// resetIfOutOfRange checks the value and if it is lower than minValue or greater than maxValue,
	// returns defValue, otherwise returns the original value.
	resetIfOutOfRange(minValue: Fixed, maxValue: Fixed, defValue: Fixed) {
		if (this.value < minValue.value || this.value > maxValue.value) {
			return defValue;
		}
		return this;
	}

	// Extract a leading value from a string. If a value is found, it is returned along with the
	// portion of the string that was unused. If a value is not found, then 0 is returned along with
	// the original string.
	static extract(input: string) {
		let last = 0;
		const maximum = input.length;
		if (last < maximum && input[last] === ' ') {
			last++;
		}
		if (last >= maximum) {
			return { value: new Fixed(), remainder: input };
		}
		let ch = input[last];
		let found = false;
		let decimal = false;
		const start = last;
		while ((start === last && (ch === '-' || ch === '+')) || (!decimal && ch === '.') || (ch >= '0' && ch <= '9')) {
			if (ch >= '0' && ch <= '9') {
				found = true;
			}
			if (ch === '.') {
				decimal = true;
			}
			last++;
			if (last >= maximum) {
				break;
			}
			ch = input[last];
		}
		if (!found) {
			return { value: new Fixed(), remainder: input };
		}
		return { value: new Fixed(input.slice(start, last)), remainder: input.slice(last) };
	}

	static readonly Min = new Fixed();
	static readonly NegPointEight = new Fixed('-0.8');
	static readonly Zero = new Fixed();
	static readonly Twentieth = new Fixed('0.05');
	static readonly PointZeroSix = new Fixed('0.06');
	static readonly PointZeroSeven = new Fixed('0.07');
	static readonly PointZeroEight = new Fixed('0.08');
	static readonly PointZeroNine = new Fixed('0.09');
	static readonly Tenth = new Fixed('0.1');
	static readonly PointOneTwo = new Fixed('0.12');
	static readonly Eighth = new Fixed('0.125');
	static readonly PointOneFive = new Fixed('0.15');
	static readonly Fifth = new Fixed('0.2');
	static readonly Quarter = new Fixed('0.25');
	static readonly ThreeTenths = new Fixed('0.3');
	static readonly TwoFifths = new Fixed('0.4');
	static readonly Half = new Fixed('0.5');
	static readonly ThreeFifths = new Fixed('0.6');
	static readonly SevenTenths = new Fixed('0.7');
	static readonly ThreeQuarters = new Fixed('0.75');
	static readonly FourFifths = new Fixed('0.8');
	static readonly One = new Fixed(1);
	static readonly OnePointOne = new Fixed('1.1');
	static readonly OnePointTwo = new Fixed('1.2');
	static readonly OneAndAQuarter = new Fixed('1.25');
	static readonly OneAndAHalf = new Fixed('1.5');
	static readonly Two = new Fixed(2);
	static readonly TwoAndAHalf = new Fixed('2.5');
	static readonly Three = new Fixed(3);
	static readonly ThreeAndAHalf = new Fixed('3.5');
	static readonly Four = new Fixed(4);
	static readonly Five = new Fixed(5);
	static readonly Six = new Fixed(6);
	static readonly Seven = new Fixed(7);
	static readonly Eight = new Fixed(8);
	static readonly Nine = new Fixed(9);
	static readonly Ten = new Fixed(10);
	static readonly Twelve = new Fixed(12);
	static readonly Fifteen = new Fixed(15);
	static readonly Nineteen = new Fixed(19);
	static readonly Twenty = new Fixed(20);
	static readonly TwentyFour = new Fixed(24);
	static readonly TwentyFive = new Fixed(25);
	static readonly ThirtySix = new Fixed(36);
	static readonly Thirty = new Fixed(30);
	static readonly Forty = new Fixed(40);
	static readonly Fifty = new Fixed(50);
	static readonly Seventy = new Fixed(70);
	static readonly Eighty = new Fixed(80);
	static readonly NinetyNine = new Fixed(99);
	static readonly Hundred = new Fixed(100);
	static readonly Thousand = new Fixed(1000);
	static readonly MillionMinusOne = new Fixed(999999);
	static readonly BillionMinusOne = new Fixed(999999999);
	static readonly MaxBasePoints = Fixed.MillionMinusOne;
	static readonly Max = new Fixed();

	static {
		Fixed.Min.value = Number.MIN_SAFE_INTEGER;
		Fixed.Max.value = Number.MAX_SAFE_INTEGER;
	}
}
