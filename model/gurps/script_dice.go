package gurps

import (
	"errors"

	"github.com/richardwilkes/rpgtools/dice"
)

type scriptDice struct{}

func (d scriptDice) From(count, sides, modifier, multiplier *int) string {
	var result dice.Dice
	// Account for the dice(sides) shorthand
	if count != nil && sides == nil && modifier == nil && multiplier == nil {
		sides = count
		count = nil
	}
	if sides == nil {
		return ""
	}
	result.Sides = *sides
	if count != nil {
		result.Count = *count
	} else {
		result.Count = 1
	}
	if modifier != nil {
		result.Modifier = *modifier
	}
	if multiplier != nil {
		result.Multiplier = *multiplier
	} else {
		result.Multiplier = 1
	}
	return result.String()
}

func (d scriptDice) Add(left, right string) (string, error) {
	d1 := dice.New(left)
	d2 := dice.New(right)
	if d1.Sides != d2.Sides {
		return "", errors.New("dice sides must match")
	}
	if d1.Multiplier > 0 {
		d1.Count *= d1.Multiplier
		d1.Modifier *= d1.Multiplier
		d1.Multiplier = 1
	}
	if d2.Multiplier > 0 {
		d2.Count *= d2.Multiplier
		d2.Modifier *= d2.Multiplier
		d2.Multiplier = 1
	}
	d1.Count += d2.Count
	d1.Modifier += d2.Modifier
	return d1.String(), nil
}

func (d scriptDice) Subtract(left, right string) (string, error) {
	d1 := dice.New(left)
	d2 := dice.New(right)
	if d1.Sides != d2.Sides {
		return "", errors.New("dice sides must match")
	}
	if d1.Multiplier > 0 {
		d1.Count *= d1.Multiplier
		d1.Modifier *= d1.Multiplier
		d1.Multiplier = 1
	}
	if d2.Multiplier > 0 {
		d2.Count *= d2.Multiplier
		d2.Modifier *= d2.Multiplier
		d2.Multiplier = 1
	}
	d1.Count -= d2.Count
	d1.Count = max(d1.Count, 0)
	d1.Modifier -= d2.Modifier
	return d1.String(), nil
}

func (d scriptDice) Count(diceSpec string) int {
	return dice.New(diceSpec).Count
}

func (d scriptDice) Sides(diceSpec string) int {
	return dice.New(diceSpec).Sides
}

func (d scriptDice) Modifier(diceSpec string) int {
	return dice.New(diceSpec).Modifier
}

func (d scriptDice) Multiplier(diceSpec string) int {
	return dice.New(diceSpec).Multiplier
}

func (d scriptDice) Roll(diceSpec string, extraDiceFromModifiers bool) int {
	return dice.Roll(diceSpec, extraDiceFromModifiers)
}
