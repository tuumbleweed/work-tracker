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

## Work Tracker
- ~~Count time for each task to display in the table~~
- ~~Update table hours per each task dynamically~~
- ~~Highlight the row when running a task~~
- Organize UI code in a better way, currently a mess, especially button handling.

## Reporting
- Add an HTML report.
    - We are going to only use it locally, so no need for an email.
        - Thus we can use regular `<style>` instead of inline
    - HTML report is saved to file and then opened right away with chrome browser.
    - Should be able to generate reports for 1-360 days.
    - It should still look nice for both weekly, quarterly and yearly reports.
    - Should contain bar charts
        - Time by task
        - Time*activity
    - Should have 3 scripts re-using same tool to generate weekly, quarterly and yearly reports.

## Installing
- Add desktop entry
- Add icon
- Add install.sh script
