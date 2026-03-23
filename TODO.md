# Dispositio TODO

- [ ] Open a file from arg (and set up for testing)
- [ ] 

## Item

```
W1  2.20  ⬤  This is the item title           ≡12/14
W2  2.27  ⚬  and this is title wrapping
W3  3.3   ⚬ 
```

## Next prompt:

in project.go, set up the parser such that for the following ROADMAP.md (or other arg) input content:

```
# First Task (3)

> 2026-W12: 2026-05-14

This is the first task we're going to tackle.

- [ ] First subtask
    - This is the description of the first subtask
- [ ] Second subtask

# Second Task (2)

This is the second roadmap item.

# Third Task, without description
```

1. H1 (#) is a roadmap item (type `item`)
2. if there is a quote following the H1 title, ignore it.
3. if there is text following the title or quote, encode it as the description (including newlines).
4. if following that there are checklist items (`- [ ]`), encode those as subtasks:
5. Subtask title is whatever follows `- [ ] ` (aka trimming separation space as well)
6. If an indented bullet (`  - `, of any indentation level >= 2 spaces or one tab) follows a check mark, consume it as the subtask's description.
7. if the item title ends with `({int})`, encode that as duration (e.g. `Something (3)` -> duration = 3). if there is no addendum, set duration to 1 by default.
