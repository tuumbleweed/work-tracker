# Project Layout

Layout that can be used with golang.

## Activity

#### Accuracy
```
The way it works right now is that if we input anything at the very end of the
tick then the whole tick will get near 100% activity.
that's why we need to keep tick size small right now (500ms will do).
UI and activity share same ticker at the moment.
later we can implement a separate tick that would sample activity in shorter periods
calculate active time for the last tick
```
