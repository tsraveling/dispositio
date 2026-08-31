Welcome to `dispositio`! This guide will help you get your bearings. You can find this same text with Gif walkthroughs at systemist.net/tools/dispositio.

If you want to just jump in now: for a full list of keybinds and functions on either the planning or detail screens, just hit `?` (shift+`/`)!

## Setting up a new project

Let's say we're making a game. "Volcano Hunter", a game about searching a vast and deep jungle for active volcanoes to study. (Note: I have my `dispositio` aliased in `.zshrc` to `dis`)

1. Run `dispositio` (or `dis` in my case) in the project root.
2. Confirm that you want to make `ROADMAP.md`.
3. Hit `e` to edit the project name.
4. Actually, we started work on Volcano Hunter a few weeks ago. You can use the left and right arrows (or `hl` for Vim users) to change the project start date, optionally holding shift to change the date by a week at a time.

## Adding some milestones

Now we want to start planning out our game.

1. Create some milestones by hitting `a` to add a new one. `o` and `O` add milestones directly above or under the currently selected one, like in Vim.
2. Hitting `enter` or `right/l` on an item goes into detail mode; more on this later.
3. Holding shift and hitting `right/left/h/l` changes the **duration** of a milestone, in weeks.
4. Holding shift + `up/down/j/k` moves the milestone up or down in the list.

Notice a few things here:

- The leftmost column shows the estimated dates of a given milestone. The `W*` column refers to the week number out of the year. If a project extends past the end of the current year, you will see a row break showing the new year.
- Because we set the project start date as a few weeks ago, that first milestone shows some alert symbols. The duration of a milestone is one week by default; because that is the first milestone in the list, and it is not completed, Dispositio knows it is several weeks overdue! And it moves the timeline back as a result. This allows you to see how much a current delay impacts the overall timeline, including when future milestones are scheduled to start (e.g. the `Movement` milestone in the example above starts 9.7; if we delay another week it and other future milestones will get pushed back a week as well)

## Tasks and subtasks

If you hit enter or `right/l` to go into "detail mode" (signified by the purple long right arrow and border), you can:

1. Hit enter to type in a description. This can be whatever text you want; I usually use it for a general overall "definition of done" for the milestone.
2. Hit `a/o/O` to add tasks, with the same rules as adding milestones above.
3. Hit shift + `up/down/j/k` to rearrange tasks.
4. Hit `shift+A` to add a subtask to a selected task.
5. Hit shift + `left/right/h/l` to move a task to be a subtask of the task above, or move a subtask out to be a regular task.
6. Hit `e` to edit the name of a task.
7. Hit `space` to complete or uncomplete a task.

Notice:

- The progress bar fills out as you accomplish tasks. In theory, once it is full, you can complete the milestone (see next section).
- There is a "tasks per weekday" readout below the progress bar. Because this milestone is overdue, it just says "do all of the tasks in a day" to get finished. But in a normal milestone, you will be able to see how many workdays each task should take on average (or tasks per workday if the ratio goes the other way) in order to meet your goals.

## Completing a Milestone

When you are ready to mark a milestone as finished, you can hit `c` in the detail view. You will see a confirmation; once accepted, the milestone is complete!

You will notice that the little character next to the milestone title becomes a checkmark -- this shows that it's complete. The next milestone, "Game Map", switches from a hollow circle to a filled one. This means it is your **current milestone**. The current milestone is always the first milestone not yet completed, and will always be highlighted when you first open Dispositio every day.


