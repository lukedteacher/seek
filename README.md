# SEEK
aka shared educational environment knowledge

a tool for educators.

## todo

- [x] basic server
- [x] figure out hot reload
- [x] index page
- [ ] students
  - [x] info page
  - [x] CRUD operations
  - [ ] better lists / tables
  - [ ] refactor grades / homerooms / etc.
- [ ] fix validation
- [ ] schedules
  - [x] CRUD operations
  - [ ] show periods
- [ ] everything else

### components todo

- [x] buttons
  - [ ] more interactivity
- [x] icons
  - [ ] remove iconify dependency
- [ ] forms
  - [x] select
  - [x] text input
  - [ ] time input
- [x] cards

## features

- scheduling
- goals
- data tracking
- behavior
- info page

## bugs

- [ ] can't delete students created by migration
- [x] pointers in student data don't translate to projections
  - fixed with stringPtr function from orisun

### reminders
- camelCase for event sourcing
- snake_case for datastar and sqlite