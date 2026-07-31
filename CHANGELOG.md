# changelog

updates for SEEK

## v0.4.2 (26.07.31) bug fixes, style tweaks, educator refactor

- fixed bug that wasn't passing eventID in user register handler
- container queries on period card and schedule day headers
- majorly refactored educators
- includes username field
- new logic for saving and retrieving events involving rawdata to avoid multiple marshals and unmarshals
- various other bug fixes

## v0.4.2 (26.07.30) light theme tweaks, table view stuff

- changed light view variables to match dark view
- tweaked table view function and related dtos in order to accept username as target
- converted students to use username instead of id for view urls
- various other bug fixes

## v0.4.1 (26.07.30) schedule and student view page revisions

- added schedule component
- added tabs
- created URLs for each tab and handlers for each URL
- tweaked styling of classes to mimic tab trigger styling
- aligned time spans with time grid and hardcoded last time

## v0.4.0 (26.07.29) user revisions

- removed the need for a OTP for verification
- removed various fields from user registration and models
- deleted teachers (superceded by educators)
- fixed a bug with the period creation that wasn't saving service type

## v0.3.13 (26.07.29) mostly code cleanup

- removed URL /delete for delete action on objects
- refactored period edit to match period create
- added MARSS ID to student data
- other various tweaks

## v0.3.12 (26.07.29) various features and bug fixes

- style tweaks for period cards
- reverted {object}/list pattern to comply with breadcrumb structure
- reverted {object}/view pattern
- added some better logging errors in handlers
- removed every println!
- fixed up IEP services pages and blocks
- fixed data table to show grouped fields for headers
- added services to student view

## v0.3.11 (26.07.28) fixed validate field bugs, student list filter for service type

- tweaked validate so that it successfully fires and updates appropriate fields
- show period on blank schedule component view
- added DateOnly type for IEP service start and end date
- student list in create period filters based on selected service type

## v0.3.10 (26.07.28) first run of validate/{field} for period

- implemented UpdateField functions for period models
- patch new values to form via SSE
- re-implemented schedule view for period creation

## v0.3.9 (26.07.28) finished moving views folder to ui/core

- moved all components in views to ui/core
- data table styling cleanup
- student data tweaks
- removed person from educator (for now?)
- swapped period popover for neo-popover to maintain state on morph

## v0.3.8 (26.27.27) continued period work

- changed types to built in shared models
- refactored days to be more usable
- added a colored logger
- deleted lots of schedule stuff

## v0.3.7 (26.07.26) cleaning up periods

- switched period model to time.Time for start and end time
- accurately calculate end time for creation and db reads
- formatting in 24h time when displaying a view
- timepicker component for form with value (but not validation)
- added service type sharedmodel for iep service and period comparison (and eventual filtering)
- new path to shared or non-feature specific ui elements 'internal/ui/core'
- added mainheader component (still needs to be added to pages)
- added breadcrumbs component (included in mainheader component)

## v0.3.6 (26.07.25) more folder, model refactoring

- moved central views/dto to features/{feature}/dto
- consolidated form for educator
- reordered fields for educator
- added person component to create command, elsewhere
- person now includes email
- student form grade updated to templui select
- student model & view now use shared model 'grade'
- unmarshal function implemented for 'grade'
- added end time field to period model

## v0.3.5 (26.07.24) date picker component for IEP services

- added date picker component on create IEP service form
- figured out how to keep the button updated with the date
- just need to add formatting

## v0.3.4 (26.07.24) first pass at cleaning up diff stuff, refactor folder structure

- moved diff to shared models
- moved diff views to dto
- changed folder structure to match package names
- genericized many table related functions
- added helpers for creating them for specific types
- removed last remnants of domain/models folder structure

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
