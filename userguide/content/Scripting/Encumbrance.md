# Scripting: Encumbrance

> [Return to the Scripting Guide.](../Scripting%20Guide)

This object holds encumbrance-related data for an [Entity](Entity).

| Property                 | Type   | Value                                                                                       | Availability |
|--------------------------|--------|---------------------------------------------------------------------------------------------|--------------|
| levelName                | string | The name of the current encumbrance level.                                                  | GCS v5.36+   |
| level                    | number | The current encumbrance level, a value from 0 to 4.                                         | GCS v5.36+   |
| levelForSkills           | number | The current encumbrance level with regards to skill usage, a value from 0 to 4.             | GCS v5.36+   |
| moveFactor               | number | A multiplier to use on the character's Move due to load.                                    | GCS v5.36+   |
| weightCarried            | number | The number of pounds being carried.                                                         | GCS v5.36+   |
| maximumCarry             | number | The maximum number of pounds that may be carried.                                           | GCS v5.36+   |
| basicLift                | number | The number of pounds that can be lifted overhead with one hand in one second.               | GCS v5.36+   |
| oneHandedLift            | number | The number of pounds that can be lifted overhead with one hand in two seconds.              | GCS v5.36+   |
| twoHandedLift            | number | The number of pounds that can be lifted overhead with both hands in four seconds.           | GCS v5.36+   |
| shoveAndKnockOver        | number | The number of pounds of an object that can be shoved and knocked over.                      | GCS v5.36+   |
| runningShoveAndKnockOver | number | The number of pounds of an object that can be shoved and knocked over with a running start. | GCS v5.36+   |
| carryOnBack              | number | The number of pounds that can be carried slung across the back.                             | GCS v5.36+   |
| shiftSlightly            | number | The number of pounds that can be shifted slightly on a floor.                               | GCS v5.36+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
