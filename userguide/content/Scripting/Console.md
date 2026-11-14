# Scripting: Console

> [Return to the Scripting Guide.](../Scripting%20Guide)

This object provides basic logging capabilities, similar to the standard Javascript console provided by web browsers.

| Method | Arguments                              | Result Type | Result                                                                                                                                                                                                                                            | Availability |
|--------|----------------------------------------|-------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------|
| log    | val1: any, [val2: any, ..., valN: any] | undefined   | Concatenates the string representation of each argument value, separated by a space, and emits it as an INFO-level message to the GCS log file (or console, if GCS was started with the `-console` option to route logs to the console instead).  | GCS v5.36+   |
| error  | val1: any, [val2: any, ..., valN: any] | undefined   | Concatenates the string representation of each argument value, separated by a space, and emits it as an ERROR-level message to the GCS log file (or console, if GCS was started with the `-console` option to route logs to the console instead). | GCS v5.41+   |

> [Return to the Scripting Guide.](../Scripting%20Guide)
