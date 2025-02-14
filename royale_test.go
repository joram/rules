package rules

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoyaleRulesetInterface(t *testing.T) {
	var _ Ruleset = (*RoyaleRuleset)(nil)
}

func TestRoyaleDefaultSanity(t *testing.T) {
	boardState := &BoardState{}
	r := RoyaleRuleset{StandardRuleset: StandardRuleset{HazardDamagePerTurn: 1}, ShrinkEveryNTurns: 0}
	_, err := r.CreateNextBoardState(boardState, []SnakeMove{})
	require.Error(t, err)
	require.Equal(t, errors.New("royale game can't shrink more frequently than every turn"), err)

	r = RoyaleRuleset{ShrinkEveryNTurns: 1}
	_, err = r.CreateNextBoardState(boardState, []SnakeMove{})
	require.Error(t, err)
	require.Equal(t, errors.New("royale damage per turn must be greater than zero"), err)

	r = RoyaleRuleset{StandardRuleset: StandardRuleset{HazardDamagePerTurn: 1}, ShrinkEveryNTurns: 1}
	boardState, err = r.CreateNextBoardState(boardState, []SnakeMove{})
	require.NoError(t, err)
	require.Len(t, boardState.Hazards, 0)
}

func TestRoyaleName(t *testing.T) {
	r := RoyaleRuleset{}
	require.Equal(t, "royale", r.Name())
}

func TestRoyaleHazards(t *testing.T) {
	seed := int64(25543234525)
	tests := []struct {
		Width             int32
		Height            int32
		Turn              int32
		ShrinkEveryNTurns int32
		Error             error
		ExpectedHazards   []Point
	}{
		{Error: errors.New("royale game can't shrink more frequently than every turn")},
		{ShrinkEveryNTurns: 1, ExpectedHazards: []Point{}},
		{Turn: 1, ShrinkEveryNTurns: 1, ExpectedHazards: []Point{}},
		{Width: 3, Height: 3, Turn: 1, ShrinkEveryNTurns: 10, ExpectedHazards: []Point{}},
		{Width: 3, Height: 3, Turn: 9, ShrinkEveryNTurns: 10, ExpectedHazards: []Point{}},
		{
			Width: 3, Height: 3, Turn: 10, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 11, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 19, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 20, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 2}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 31, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 1}, {1, 2}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 42, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 53, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 64, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
		{
			Width: 3, Height: 3, Turn: 6987, ShrinkEveryNTurns: 10,
			ExpectedHazards: []Point{{0, 0}, {0, 1}, {0, 2}, {1, 0}, {1, 1}, {1, 2}, {2, 0}, {2, 1}, {2, 2}},
		},
	}

	for _, test := range tests {
		b := &BoardState{
			Width:  test.Width,
			Height: test.Height,
		}
		r := RoyaleRuleset{
			StandardRuleset: StandardRuleset{
				HazardDamagePerTurn: 1,
			},
			Seed:              seed,
			Turn:              test.Turn,
			ShrinkEveryNTurns: test.ShrinkEveryNTurns,
		}

		err := r.populateHazards(b, test.Turn)
		require.Equal(t, test.Error, err)
		if err == nil {
			// Obstacles should match
			require.Equal(t, test.ExpectedHazards, b.Hazards)
			for _, expectedP := range test.ExpectedHazards {
				wasFound := false
				for _, actualP := range b.Hazards {
					if expectedP == actualP {
						wasFound = true
						break
					}
				}
				require.True(t, wasFound)
			}
		}
	}
}

func TestRoyalDamageNextTurn(t *testing.T) {
	seed := int64(45897034512311)

	base := &BoardState{Width: 10, Height: 10, Snakes: []Snake{{ID: "one", Health: 100, Body: []Point{{9, 1}, {9, 1}, {9, 1}}}}}
	r := RoyaleRuleset{StandardRuleset: StandardRuleset{HazardDamagePerTurn: 30}, Seed: seed, ShrinkEveryNTurns: 10}
	m := []SnakeMove{{ID: "one", Move: "down"}}

	r.Turn = 10
	err := r.populateHazards(base, r.Turn-1)
	require.NoError(t, err)
	next, err := r.CreateNextBoardState(base, m)
	require.NoError(t, err)
	require.Equal(t, NotEliminated, next.Snakes[0].EliminatedCause)
	require.Equal(t, int32(99), next.Snakes[0].Health)
	require.Equal(t, Point{9, 0}, next.Snakes[0].Body[0])
	require.Equal(t, 10, len(next.Hazards)) // X = 0

	r.Turn = 20
	err = r.populateHazards(base, r.Turn-1)
	require.NoError(t, err)
	next, err = r.CreateNextBoardState(base, m)
	require.NoError(t, err)
	require.Equal(t, NotEliminated, next.Snakes[0].EliminatedCause)
	require.Equal(t, int32(99), next.Snakes[0].Health)
	require.Equal(t, Point{9, 0}, next.Snakes[0].Body[0])
	require.Equal(t, 20, len(next.Hazards)) // X = 9

	r.Turn = 21
	err = r.populateHazards(base, r.Turn-1)
	require.NoError(t, err)
	next, err = r.CreateNextBoardState(base, m)
	require.NoError(t, err)
	require.Equal(t, NotEliminated, next.Snakes[0].EliminatedCause)
	require.Equal(t, int32(69), next.Snakes[0].Health)
	require.Equal(t, Point{9, 0}, next.Snakes[0].Body[0])
	require.Equal(t, 20, len(next.Hazards))

	base.Snakes[0].Health = 15
	next, err = r.CreateNextBoardState(base, m)
	require.NoError(t, err)
	require.Equal(t, EliminatedByOutOfHealth, next.Snakes[0].EliminatedCause)
	require.Equal(t, int32(0), next.Snakes[0].Health)
	require.Equal(t, Point{9, 0}, next.Snakes[0].Body[0])
	require.Equal(t, 20, len(next.Hazards))

	base.Food = append(base.Food, Point{9, 0})
	next, err = r.CreateNextBoardState(base, m)
	require.NoError(t, err)
	require.Equal(t, Point{9, 0}, next.Snakes[0].Body[0])
	require.Equal(t, NotEliminated, next.Snakes[0].EliminatedCause)
	require.Equal(t, int32(100), next.Snakes[0].Health)
	require.Equal(t, Point{9, 0}, next.Snakes[0].Body[0])
	require.Equal(t, 20, len(next.Hazards))
}
