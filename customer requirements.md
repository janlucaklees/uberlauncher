# Uberlauncher plan

## Goal

A smart launcher that can do more than just opening programs. Modular, so it's esaily extensible,
this launcher comes as a cli tool for maximum flexibility in it's usage.


## Description

UberLauncher is a smart launcher. At it's core it works like a combination of fzf and shell 
autocompletion.
The actual useful functionality is provided by skills - basically very lightbweight plugins.
Skills provide entries which are just lines of text of possible inputs.
The core accumulates all entries from all skills and displays them in a searchable list with fuzzy
matching. This way, the user can find and run the disired skill with only a few keystrokes, not
needing to type out the exact and whole entry.

There are thee different categories of skills:
1. Skills that just can be run without any additional options or the ability to enter free text.
  Examples:
  - `shutdown` (to shut down the system right now)
  - `restart` (to restart the system immediately)
  - Note: Of course these skills could offer options to control the timing of the operation. But for
    this example it is assumed they don't.
2. Skills with a fixed set of options.
  Examples:
  - `wifi`, possible options: `on`, `off`, `<SSID of the network>` (connects to or diconnects from this netowrk).
  - `bluetooth`, possible options: `on`, `off`, `<device name>` (connects or disconnects the device.).
3. Skills with free text.
  Example:
  - `todo <text of what I want to add to my todo list>`: a skill to add a todo to my todo list with freely controllable content.

The core functionality in detail:
The core manages the skills and aggregates the all the skill names for skills of category 1, 2, and 3.
For skills of category 2 it aggregates all the options that skill offers.
The core thus ends up with a long list of possible inputs the user can make, and displays this list
to the user, with an input where the user can type. When typing, the list is filtered with a fuzzy
mathing and by default the best match is selected. When the user hits enter, the selected entry is
passed as text to the skill, which then does what it needs to do to fulfill the users request.
This way, the user can type `won`, which might have the `wifi on` entry as the best match, to quickly
turn on their wifi. Under the hood, the `wifi` skill talks to the system to turn the wifi on.
But there is one exception: Skills of category 3. These offer the ability for the user to enter free
text. When the first word the user entered matches the name of a category 3 skill, the behavior of
the enter key changes. When pressed, it no longer sends the best matching entry to the coresponding
skill for execution, but sends the raw user input to the skill for execution.
This way the user can type something like `todo do laundry tomorrow` and have a "do laundry tomorrow"
todo added to their todo list.
In general, the user can always navigate the list of matching entries up and down with their arrow
keys to select another match. Hitting enter will then launch the selected entry, even on a category
3 skill. The UI reacts as follows: As long as no category 3 skill is detected, the best matching
entry is somewhow marked or highlighted. When a category 3 entry is detected, the marking or
highlighting jumps to the input field, to signal to the user that their raw input will be used.

Additional UI considerations
- Input is at the bottom by default, but can be configured to be at the top.
- List of entries fill the available space above / below the input.
- By default, the list of entries is sorted by usage with the most used entry being closest to the input.
- Matching entries to the input is done by fuzzy matching and should work similar to how fzf does it.
- When the user enters a skill name as their first word, it should get highlighted to signal to the\
  user, that their input exatcly matches a skill - just like with commands on a commandline.
- A list of aliases can be configured. For example to match `t` to the `todo` skill, or `wh` to `wifi <SSID of the users home network>`.
  Aliases can be kept simple and function similar to aliases in the shell.


## Tech-Stack

Go with Bubble-Tea for TUI
