# Calculators

The **Calculators** view in GCS provides quick tools for resolving common rules questions during play. Each calculator
has a tab of its own, and most can take their inputs from any open [character sheet](Character%20Sheet%20Overview) or
from values you type in, saving you from manual table lookups and calculations.

## Opening the calculators

To open the calculators:

- Select **Calculators** ![calculator](images/icons/icn-calculator.svg) from the **Character Sheet toolbar**, which
  starts the calculators on that character, or
- Go to **View > Calculators**.

Each tab has a **Source** dropdown that lists the open character sheets, plus **Manual** to type the inputs in
yourself. A tab's results appear in its **Results** box, along with the rolls to make and their consequences, and the
section headings link to the rules in the GURPS books.

> [!TIP]
> These calculators are **reference tools only**. They do not affect your sheet and are there for quick lookups during
> play.

## Explosions & Area Attacks

Works out the damage from an explosion, an area-effect attack, or a cone attack ([B413–B415](BX413)). Choose the
attack type, then:

- **Attack:** Take the blast damage from an explosive weapon on an open sheet, or type in the dice of damage, the
  damage type, and any fragmentation dice. Explosions ask for the environment (air, underwater, or vacuum) and whether
  the explosion is an airburst or has hot fragments; cones ask for their range and width; area attacks ask whether they
  dissipate with distance or are resisted by HT.
- **Target:** Take the target's SM, HP, and DR from an open sheet or type them in, then choose which hit locations are
  exposed, the target's posture, the situation (caught in the blast, struck directly, and so on), and the distance from
  the center of the blast.

The results show the collateral damage radius, the damage divisor at that distance, the blast damage, the damage taken
after DR, the wounding, and, for fragmentation, how many fragments hit and what they do.

## Scatter

Works out where a missed attack lands ([B414](BX414)). Enter why the attack missed (a failed roll, a squared miss, or
a dodge), the margin, and the distance to the target. The result is how far the attack scatters.

## Demolition

Works out the explosive needed for a blast, or the blast from a quantity of explosive ([B415](BX415)). Choose the
explosive from the list, or enter a custom relative explosive force (REF), then enter either the blast dice wanted or
the weight of explosive on hand.

## Collisions & Falls

Works out the damage from a fall onto a surface, a collision with an immovable object, a collision between two objects,
or the sudden stop of a vehicle ([B430–B432](BX430)). For each participant, take the HP, ST, SM, and DR from an open
sheet or type them in, and choose the object's shape and velocity. Falls also ask for the distance fallen, gravity,
terminal velocity, the surface landed on, and whether the surface can break; collisions ask for the angle; sudden
stops ask for the restraint. Successful Acrobatics and Swimming rolls can soften a fall.

The results give the velocity, the damage to each participant, and the damage to the surface.

## Jumping

Works out your character's jumping distance ([B352](BX352)). The jumper's Basic Move, Lifting ST, body weight,
encumbrance, Jumping skill, Enhanced Move, and Super Jump come from the chosen sheet. The Lifting ST includes any
halving from a [pool threshold](Point%20pools), such as _Tired_. Enter the **running start** and any **penalty for
extra effort** ([B356](BX356)). The results show:

- **High Jump** distance (vertical leap).
- **Broad Jump** distance (horizontal leap).

## Throwing

Works out your character's throwing distance ([B355](BX355)). The thrower's ST, striking ST, and Throwing and Throwing
Art skill levels come from the chosen sheet. The ST includes any halving of lifting ST from a [pool
threshold](Point%20pools), such as _Tired_. Enter the object’s **weight** and any **penalty for extra effort**. The
results show:

- **Distance** in yards.
- **Damage** (based on ST and object weight).

## Hiking

Works out your character's hiking distance ([B351](BX351), [HT55](HT55)) and the fatigue it costs ([B426](BX426)). The
hiker's Move, encumbrance, FP, fitness, and Recover Energy skill come from the chosen sheet. Then choose the
**terrain**, **weather**, **heat**, and **intensity** of the march, and check the modifiers that apply:

- **Roads are cleared** (depends on weather).
- **Using skis** or **skates**.
- **Made a successful Hiking roll** ([B200](B200)).
- **Hours of hiking per day**, with an optional rest halfway through, a decent meal, and any FP restored while resting.
- **Penalty taken for extra effort**.

The calculator then shows your **daily travel distance**, how long a given distance takes, the FP lost by the end of
the day, and an hour-by-hour table of distance, FP, and condition, including the slowdown and collapse of a tired
hiker and the effects of **Fit**, **Very Fit** ([B55](B55)), and **Recover Energy** ([B248](B248)).

---

[Return to Home](Home)
