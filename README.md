# Unsure tournament

The unsure tournament is a challenge that showcases Luno's approach to building robust microservices using [shift](https://github.com/luno/shift) state machines, [reflex](https://github.com/luno/reflex) event streams and [fate](https://github.com/luno/fate) for failure injection.

The challenge is for a `team` to implement a `player` microservice(s) that interfaces with the `engine` to successfully complete a match.

It is called the **unsure tournament** since it explicitly introduces lots of errors on all IO, using the `unsure` package, resulting in a general loss of certainty and a need for designing for failure.

## Match overview

- A **team** plays against the **engine**. (A team does not directly play with other teams.)
- The goal is for a team to start a **match** and to successfully complete all the **rounds** of the match in the shortest time.
- Completed matches can be ranked by: failed rounds ascending, match duration ascending.
- A team has a name and consists of a number of **players** (> 3).
- A player has a name and is represented by a microservice instance. 
- A team is therefore multiple players/microservices connecting to the engine and collaboratively playing a match.
- The player instances can be different implementations or a single implementation with different config (names).
- A match is started by any player calling `engine.StartMatch`.
- A team can only have one active match.
- The players should subscribe and react to the engine's reflex event stream.
- The engine then starts rounds (without waiting for previous rounds to complete).
- Players must use the `unsure` IO library for `grpc`, `sql dbc`, `context`.

## Round Overview

A round consists of the following stages:
- **Join**: The round has been started and all players should join the round (`engine.JoinRound`). The engine will respond indicating if the player is **included** in this round or not.
- **Collect**: All included players should collect their rank and parts (`engine.CollectRound`). The engine responds with a rank for the player as well as a map of **parts** for each player. 
- **Submit**: In ascending order by rank, each included player should submit the total of his/her parts (`engine.SubmitRound`).
- **Success**: The team has successfully completed the round.
- **Failed**: The team has failed the round by not following the above rules.

Notes:
- Each stage has a timeout, after which the round is failed.
- Players should communicate their inclusion, ranks and parts amongst each others in order to submit the correct answer (sum of their parts) in the correct order (included by ascending rank).

## Example

This repo includes an example player implementation called `loser`. It only starts a match, but doesn't do anything else, so all the rounds timeout.

Usage:
```
# Start the engine with fresh DB
go run engine/engine/main.go --db_recreate

# In another tab, start a single loser player
go run loser/loser/main.go --db_recreate --engine_address="127.0.0.1:12048"
```

