```
Project Name: Merchant of Venice
Project Start: Apr 28 2026
```

# Proto Map (3)
```
Finished: May 4 2026
```

- [x] I can see that I'm in port
- [x] I can pick a destination
- [x] I can embark
- [x] I can see the date
- [x] I can see time passing
- [x] I can see the progression of the journey
- [x] I can see when I arrive in the destination
- [x] I can then repeat the process

# Proto Ship
```
Finished: May 11 2026
```

This one is just about ship and crew as entities. Axing work-related stuff to stick em in the next milestone.

- [x] R: Pick a "starter" ship and research what it would take.
- [x] I can see how many crew are on my ship
- [x] I can see them as individuals, with generated names
- [x] I can open and close the ship panel in port or en route
- [x] The ship panel overrides the port panel
- [x] The ship panel shows some placeholder text above the panel "en route" or "docked in X" that will be replaced by art
- [x] I can see the cargo hold (empty for now)
- [x] Set up cargo library system for data to live
- [x] Stub out a "base materials" thing with valuables, e.g. hyperpyron -> gold
- [x] I can see a few bulk commodities: say salt and glassware.
- [x] I can see a few valuable commodities: silver and Venetian currency.

# Proto Time (2)
```
Finished: May 25 2026
```

- [x] Break the day down into 6h quarters
- [x] Implement for sailing
- [x] Drama: show anchorages in a very basic, placeholder-text way

# Loading the Ship (4)

- [x] Real time refactor -- data now deals with real time (in hours), not day-quarters.
- [ ] Refactor ship UI to match new modality
- [ ] Stub out standardized crew grid component.
- [ ] Stub out standardized cargo component.
- [ ] Stub out standardized order component
- [ ] Stub out standardized Setup Work component / UI
- [ ] Stub out standardized ongoing work component
- [ ] Standard port UI has grayboxed city shape with nodes for Port and fondaco, and a Path connecting them.
- [ ] Initial state: test cargo is in the fondaco to start with, and crew is split.
- [ ] First order: move crew.
- [ ] Crew transit along path while moving.
- [ ] State updates at end
- [ ] Time UI: tap space to advance to next completion.
- [ ] Stub basic notification system w/ job complete
- [ ] Second order: move cargo.
- [ ] Job setup: crew picker. Can quickly click to assign workers from those in area.
- [ ] Job setup: cargo picker. Can somehow select what and how much cargo to move.
- [ ] Job setup: wagon hire. Can optionally assign wagons. Not attached as yet.
- [ ] Assigned crew and wagon will ferry supplies load by load until done, transitting the Path2D.
- [ ] State updates with each load.
- [ ] After 6pm, any leg that finishes will mark the crew as "off duty" for the rest of the day.
- [ ] New button appears, "end day". Click this and we accelerate time fairly quickly to 6am, when all crew comes back on.
- [ ] Unfinished jobs (aka transport) gets resumed.
- [ ] All crew start in the fondaco in the morning, so transport jobs always resume with a trip down to the ship.
- [ ] UI Cleanup: remove old camp ui

# First Voyages (5)

- [ ] Add a new Ship order "Prepare for departure"
- [ ] Once done, Ship will show new status: departing tomorrow at 6am.
- [ ] When you pass the day, only two orders will be available: Set sail (~1h) or Tie Back Down (~2h).
- [ ] Other orders disabled until tied down.
- [ ] Departing switches into Sailing Mode.
- [ ] In Sailing Mode, only the ship view is available; still with work, orders, cargo, crew.
- [ ] Graybox out a different polygonal background for this, or sketch something.
- [ ] Set up anchorages in data as points along a given route.
- [ ] Build tool: seed anchorages button on route in inspector if empty, clear with confirm if not.
- [ ] Special anchorage: smaller towns like Pula
- [ ] Build out the first few routes with this
- [ ] Graybox out the Sailing Track at the top of the screen; ship, ports, land, anchorages.
- [ ] Add land on port | starboard option for routes (when traveling in +1 direction)
- [ ] Set up raycaster to check land distance off relevant bow
- [ ] Hook that up to land depection in Sailing Track
- [ ] Show Sailing as the main piece of work while underway
- [ ] On Sailing Track, allow inspection or show in UI quality of and distance to next anchorage.
- [ ] Sailing order: pull in at next anchorage.
- [ ] Can pause / unpause at anytime with spacebar.
- [ ] Pulling in is a piece of work that takes ~2h and all of the sailors.
- [ ] When done you'll be in an "Anchored" state.
- [ ] Right UI becomes "Anchored": Ship and Camp / Town. You can move crew between them but not cargo.
- [ ] If a wild anchorage, right side is Shore with a "set up camp" order, at which point it becomes camp.
- [ ] If town, right side has "acquire lodging" order, which in future will cost money. Not a job, instant.
- [ ] If before 6pm, you have a "End day early" option that lets you release everybody.
- [ ] If shore lodgings, flavor text for camp or in.
- [ ] If not, flavor text for sleeping on board.
- [ ] Shore / Town etc is in a centralized area for supporting hunting, minor market stuff, etc in future
- [ ] Set sail is ~1h in a town, or "Break Camp 1h" then "Set Sail 1h" if camped.
- [ ] For now you can sail all night if you want to; in future this will lead to a mutiny.
- [ ] When arriving at a port you immediately get three (blocking) options: Put In, Anchor, Keep Sailing.
- [ ] If you Keep Sailing you must immediately select an outgoing route, and you immediately move onto that route.
- [ ] Stub Anchor for now -- basically treats the city like a stop town.
- [ ] Put In starts the Put In job. On completion you are docked in the new port.

# Skald Integration (2)

- [ ] I have the latest version of Skald installed
- [ ] I have the texts file structure build out
- [ ] D: Integration plan for Skald with narrative events

# Ship supplies and crew health (2)

- [ ] Sailor health and problems
- [ ] Water
- [ ] Booze
- [ ] Food
- [ ] Illness

# Expeditions (2)

Hunting, trade missions inland, etc.

# World detail (2)

Enough metadata and text is in place that cities feel unique. Many may not have this yet but we're starting to get a sense for the shape of the data.

# Detailed Sailing (2)

- [ ] Seasonal winds
- [ ] Geographic effect regions
- [ ] Day to day weather changes
- [ ] Detailed anchorages

# Whole map shallow fill (2)

Without any deep details or flavor, lay out the whole world.

- [ ] All port cities and routes in the game world are in place.

# Proto Situations (2)

A situation is a special branching narrative flow, probably powered by Skald. At key points in a Situation you have Decisions. Decisions have various Approaches, conditional on what you have brought to the situation. Setting up for a Situation involves assembling a "loadout" -- which crew to bring, what supplies, what armaments, and so on.

- [ ] Design Situation UI
- [ ] Build Situation preparation screen
- [ ] Wire up Skald to Situation-specific method interface
- [ ] Build Decision UI
- [ ] Handle the case of multiple approaches to a given requirement (e.g. Varangian vs Saracen bodyguard)
- [ ] Wire up conclusions

# Proto Major Trade (2)

- [ ] There are bases in some ports

# Proto Venice (3)

# Companies and Contracts (2)

# Robbery and defense (2)

If you don't hire enough guards, your supplies can get jumped and stolen. Dealing with authorities is also work that takes time, and may not pan out.

# Piracy (2)

# Historical change (2)

The world changes over time to match history

# World in Depth (4)

All ports, stops, etc are in place in the game world. Historical detail is fleshed out and meaningful. The world feels real and lived in.

# Cargo Variants

Cargos can have variants that are more or less prized in certain places, e.g. Sicilian Wine, Spanish Wine, and so on.

# Fourth Crusade (2)

# Opening of Constantinople (2)

# Endgames (4)

There are 2-3 endgame goals to pursue. These are interesting, rewarding, and have their own mechanics.

- [ ] To future self, from 5.4.26: are you out yet?
