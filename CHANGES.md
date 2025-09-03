# v? ?

New content:

* New secondary spirit: **Fire Salamander**. It features a *fire retreat*
  ability that makes the player retreat up to two tiles while leaving a fire
  cloud on the previous position, as long as there's room for the retreat. The
  new spirit also confers +1 attack and a chance of burning foes on hit (like
  phoenixes do) and the upgraded version confers fire resistance (half duration
  of Fire effects).  As a drawback, it inflicts bad hearing on the player,
  making it harder to hear footsteps and impossible to hear creeping and
  wing-flapping noises. If the Fire Salamander is combined with the Barking
  Hound spirit, hearing becomes normal again (neither good nor bad). Fire
  Salamander is hence a spirit that offers very good offense capabilities and
  some mobility, but it can significantly impair stealth tactics.
* New unique monster: **walking mushroom**. It's less sturdy than a walking
  tree but releases lignifying spores on sight from afar, affecting other
  monsters too. Also, its melee hits are confusing.
* New unique monster: **noisy imp**. It performs music just for the player,
  attracting nearby monsters. It avoids fighting similarly to afraid monsters.
* New special monster: **warping wraith**. Its hits teleport the player away.
  It may be sent by the dungeon to investigate and protect the portal after you
  use a warping menhir.
* New portal guardian monster: **blazing golem**. A sturdy magical golem that
  explodes on death.
* New totem guardian monster: **totem wasp**. Rampaging and venomous wasps that
  like to use totems as their nest.
* New kind of item: **runic trap**. They represent *visible* traps that are
  triggered when stepped on by the player or a monster. Wandering monsters
  avoid traps, but hunting ones do not. Effects depend on the kind of rune.
  Placement is designed so that there's always at least one path that avoids
  traps.
* New **empty totem** static item on map levels without a regular totem. Its
  purpose is to make it so the player doesn't feel the need to explore all the
  map corners on those maps (feature suggested by @mlochbaum in issue #1).
  Normal totems now become empty totems once you've picked their spirit.

Reworked and improved content:

* Make translucent walls liberate a poison cloud when destroyed. Thematically,
  they are walls that contain a special poisonous gas that makes them
  translucent and, on destruction, that gas is expelled, resulting in a poison
  cloud and ordinary rubble.
* Rework menhirs. The “mapping menhir” is now called “earth menhir” and only reveals
  the frontier formed by map walls adjacent to interior ones; its noisy echoing
  sound also destroys most nearby walls. Other menhirs remain similar, but
  offer some partial information about the map too: “poison menhirs” reveal
  translucent walls (which now contain poison), “fire menhirs” partially reveal
  foliage (that can be burnt), and “warping menhirs” (previously called
  “teleport menhirs”) reveal portals.
* When the player or a monster is both *afraid and cornered* with no possible
  escape, they now become berserk for a short while!
* A few small improvements in footstep hearing: more deterministic and better
  UI (use a different color for each monster noise class).
* Remove the Nausea status effect, because it was player-only and didn't have
  interesting interactions. It was mainly useful as a nerf for teleport: now
  teleport results in either Daze or Imbalance (chosen randomly between both,
  for some surprise and adaptability). The “disgusting rat” has been renamed
  into a “hungry rat” with a new “hunt by smell” trait that makes rats able to
  track you by smell from out of view and, as an extra trait, become berserk if
  they see you eat a comestible.
* Remove most eating restrictions. For example, a “berserking flower” can now
  be eaten even if already Berserk: it will end Berserk early and apply it
  again. That means you can use it to safely prolong Berserk when only one turn
  is left, but you still get the Poison debuff. The only remaining restriction
  is that you cannot eat a “lignification fruit” while you have the Foggy-Skin
  status (the restriction may be removed in the future if I find some
  interesting interaction).
* Make “ambrosia berries” either confuse you (existing behavior) or make you
  afraid. The idea, like for the “teleport mushroom”, is to provide some kind
  of surprise and require adaptability (a bit like identification systems do in
  some situations, but in a more restricted way). A fun interaction with the
  new afraid and cornered mechanic means it can be used to become berserk (with
  50% chance).
* Make “firebreath pepper” produce a nadre-like explosion when reaching a wall,
  providing a limited means of wall destruction for non-Boar players. Make
  explosions produce dust or smoke clouds. Also, slightly increase Fire
  duration from firebreath pepper.
* New clouds can replace existing clouds of different types. If the type is the
  same, the duration is still the maximum of the old and new ones, like before.
* Unlike plain clouds of fog, dust or smoke, dangerous clouds of fire and
  poison are now only difficult to see through, like foliage (suggested by
  @mlochbaum in issue #1). This should make menhir usage obstruct vision less,
  as well as allow for things like jumping over a fire tile with the Frog or
  Gazelle spirits. It also allows ranged attacks to reach just behind a fire
  or poison cloud (which is both good and bad for different reasons).
* Make “pushing gale” dissipate clouds, providing the first cloud
  dissipating source which, due to “recoil”, can be quite useful for Wind Fox
  players. Also, as a special interaction, “pushing gale” propagates fire
  clouds on foliage tiles instead of dissipating them.
* Also, as discussed in issue #1 about Wind Fox balance, make “pushing gale”
  always do 1 damage (instead of 1-2 damage), and make 2-tile “recoil” in melee
  produce some wind noise.
* Make “recoil” active even with the non-upgraded Wind Fox (pushing is active
  on the non-upgraded Rampaging Boar, so having recoil only on upgraded Wind
  Fox seemed a bit strange).
* Improve “poison cloud” ability of the Venomous Viper spirit so that the
  poison clouds are created *up to two free tiles away* (before a monster or
  obstacle). This creates more varied possibilities for tactical movement,
  while still restricting safe movements in a similar but less extreme way, and
  providing similar defense from ranged monsters.
* Adjust slightly down the duration of fire clouds to 6-12 turns and poison
  clouds to 5-8 turns. Adjust duration of Poison inflicted by clouds down to 3
  (experimental).
* Using “sprint” or “jump” abilities while imbalanced may now sometimes lead to
  a dazing fall.
* Space-distorting attacks now *swap clouds* too. While getting swapped into a
  cloud of fire or poison by foes was (very occasionally) fun, Temporal Cat
  players not being able to safely attack monsters on dangerous clouds was more
  annoying than fun. And in both cases, it could result in unexpected dangerous
  situations that would've required explicit examination to be avoided: when a
  monster is on a tile with a cloud, only the monster is drawn, and while that
  limitation could be solved for the tiles version with image layers,
  introducing that kind of visual clutter doesn't fit the simplicity of the
  game.
* Players with the Rampaging Boar spirit don't move into poison or fire clouds
  after pushing anymore. While sometimes funny, it had similar issues as
  swapping into unexpected clouds. Monster boars and dragons still move into
  dangerous clouds after pushing, because they're enraged and corrupted, so not
  as smart!
* Players with the Sprinting Gazelle now leave dust behind only on floor tiles
  (in particular not on foliage ones).
* The Jumping Frog spirit now has one extra jump charge initially (so 3
  initially and then still 4 when upgraded), but the jump's daze duration on
  monsters is decreased by one.
* The “sprint” and “focus” abilities now cancel each other. It's never really
  useful nor thematic to have both at the same time, so hydra players have now
  an instant sprint cancellation option (at the cost of a spirit charge).
* While sprinting, the frog's “jump” ability now unbalances foes without
  hurting them, but still dazes them, effectively combining the effects of the
  two kinds of jumps in a sensible way.
* The “dig” ability now also guarantees pushing will happen, including the
  unbalancing effect for the upgraded spirit.
* Add simple boredom clock: if you spend too much time on a single level (more
  than necessary with normal play), the dungeon core will detect your presence
  and mark you until you go through a portal.

Monster behavior and distribution improvements:

* Improve a bit monster distribution. Mostly, bias toward early-mid-late
  monsters depending on map level has been increased, making the various phases
  more distinct. Early and Mid game difficulty remains similar. Late game has
  now usually less monsters than before, but a bit more of dangerous ones, so
  it has become more stealth-friendly while still being challenging enough,
  except in case of a special “swarm” level, favoring greater numbers over
  individual monster dangerousness.
* Improve wandering monster destination selection. It's still quite random with
  some bias toward close locations, but there's some extra bias toward
  special vaults, so that interesting places are visited more often by
  monsters.
* Improvements in monster group formation. In particular, less groups of three
  monsters than before and more groups of two, so somewhat better coverage of
  the map in mid and late game, but smaller likeliness of large monster groups.
* Make the order in which monsters act within a turn after the player properly
  random. It already felt that way, but a very attentive player could sometimes
  guess that some monster seemed to always act before some other, potentially
  leading to gamey predictions.

Map generation:

* Mixed map generation algorithms in some map levels and some other smaller
  improvements in map generation.
* Malfunctioning portals are now more common (around 1-3 per game, with 2 being
  the most common). Some minor adjustments in menhir distribution (in
  particular, the same type can't appear twice on a given map level anymore).
* Improved placement of non-vault comestibles, now more often present in
  spots next to walls than before (suggested by @mlochbaum).
* Add a few new special vaults.

UI, graphics, QoL:

* Add concise information about current tile in the status bar which tells
  whether there's fog, an item, or a special kind of terrain under the player.
* Animation improvements and fixes. Monster deaths are now always animated when
  in view, independently of the cause. Hits between confused monsters are now
  animated too, making it clearer what's happening. Other minor enhancements.
* Improve various tiles (footsteps, translucent walls, food…).
* Wizard mode improvements: intended for development purposes, but now cheat
  usage is recorded in a more fine-grained way in statistics, so any frustrated
  casual players might consider using it as an *adventure mode*. Enabling
  Wizard Mode *once* with `W` only enables non-permadeath mode (each
  resurrection by the wizard will be recorded in the timeline). Extra cheats
  are enabled by pressing `W` again one or more times.
* Refine drawing of tiles with both a cloud and an item: clouds are now drawn
  over items, but if the cloud is not dangerous, the color of the item is used
  (which never conflicts with the color of dangerous clouds).
* A few improvements in game statistic's dump: ASCII map now shows monster
  noise too, the number of last messages was increased a bit.
* A few help improvements.

Fixes:

* Nadre explosions now properly blow up nearby walls too, like in Boohu and
  Harmonist.
* Make the player never appear on a vault with a totem: not really a bug, but
  it wasn't very fun.
* Fix field of view sometimes not being updated just after cloud creation but
  later at the end of a turn, which could allow monsters to attack once from a
  distance before the update.
* Fix spelling and formatting issues in a few log messages. Avoid case of
  duplicated “Saw *monster*” entries in timeline.
* Fix healing amount in description of foggy-skin onion.
* Fix application of Imbalance when jumping with the Sprinting Gazelle's sprint.
* Some confused monster actions that happened outside of view incorrectly
  appeared in the game logs.
* Improve/fix the page-down and scrolling instructions when details of all
  entities in a single examined tile don't fit.

# v1.0.0-beta.1 2025-07-22

First public release of Shamogu.
