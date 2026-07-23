# changelog

updates for SEEK

## v0.3.3 (26.07.23) csv file read, diff table view

- added a rudimentary csv file read via gocsv
- very sloppy diff and diff table view to see if it works

## v0.3.2 (26.07.23) more templui, iep service crud

- swapped the user header bit for a different block
- period now uses new student multiselect
- IEP service creation & editing implemented
- IEP service uses a modified student select and some others with static data
- ran go fmt

## v0.3.1 (26.07.22) more ttf -> woff2

- switched type font to woff2

## v0.3.0 (26.07.22) no-cache static assets, view transitions, educators

- switched static asset route to "no-cache" (from "no-store") to speed up page loads
- lexend ttf -> woff2
- enabled navigation: auto for view transitions as first step
- implemented educator events for create, update, archive, and delete
- fixed many random bugs
- implemented person model with name functions to be used as embedded struct in students, educators, etc.

## v0.2.3 (26.07.22) profile bio updateable, fixed templui js bug

- implemented updating the profile bio via the 'edit profile' page
- added a role selectbox to the edit profile page
- fixed templui's js not being served correctly bug
- fixed minor layout bugs

## v0.2.2 (26.07.22) templui, tailwind, daisyui

- installed all templui components for ease of access
- installed tailwind temporarily for templui and daisyui
- installed daisyui for basic styling until seek custom styles can be created

## v0.2.1 (26.07.21) data seeding, data table implemented

- created an endpoint to seed period and student data when the db gets wiped
- data table implemented for students and periods
- profile page slightly updated

## v0.2.0 (26.07.20) user profiles, reset read models command

- added reset command to reset read models
- fixed user profile field names stuff
- user profile visible at /profile
- first draft of personal view of profile

## v0.1.3 (26.07.20) icon component

- replaced all icons with the templui icon component
- modified the component to recieve a "size" directive for variable size without tailwind classes

## v0.1.2 (26.07.20) buttons everywhere

- replaced most HTML buttons with components
- added form button components
- created a "color" property for a button (info, success, warning, error)
- unified "/{object}/list" URL pattern

## v0.1.1 (26.07.19) (re-)implementing button component, layers for css

- added layers (base, variants, components, etc.) to CSS
- working button component that needs a few more variants (including size)
- dashed lines on the table columns

## v0.1.0 (26.07.19) continued work on data tables, more templui inspiration

- implemented v2.0 of tables from templui
- added sort icons and got post to /sort working with signals
- continued work on CSS streamlining
