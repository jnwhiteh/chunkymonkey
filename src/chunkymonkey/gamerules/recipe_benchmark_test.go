// Benchmarks for recipes.

package gamerules

import (
	"os"
	"testing"

	. "chunkymonkey/types"
)

func loadRecipesAndItems() (recipes *RecipeSet, itemTypes ItemTypeMap, err os.Error) {
	blockTypes, err := LoadBlocksFromFile("blocks.json")
	if err != nil {
		return
	}

	itemTypes, err = LoadItemTypesFromFile("items.json")
	if err != nil {
		return
	}

	blockTypes.CreateBlockItemTypes(itemTypes)

	recipes, err = LoadRecipesFromFile("recipes.json", itemTypes)
	if err != nil {
		return
	}

	return
}

func Benchmark_RecipeSet_Match_Simple2x2(b *testing.B) {
	recipes, itemTypes, err := loadRecipesAndItems()
	if err != nil {
		panic(err)
	}

	empty := *RSlot(itemTypes, ItemTypeIdNull, 0, 0)
	log := *RSlot(itemTypes, 17, 1, 0)

	inputs := Slots(log, empty, empty, empty)

	var matcher RecipeSetMatcher
	matcher.Init(recipes)

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		matcher.Match(2, 2, inputs)
	}
}

func Benchmark_RecipeSet_Match_Nothing2x2(b *testing.B) {
	recipes, itemTypes, err := loadRecipesAndItems()
	if err != nil {
		panic(err)
	}

	log := *RSlot(itemTypes, 17, 1, 0)

	inputs := Slots(log, log, log, log)

	var matcher RecipeSetMatcher
	matcher.Init(recipes)

	b.ResetTimer()
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		matcher.Match(2, 2, inputs)
	}
}
