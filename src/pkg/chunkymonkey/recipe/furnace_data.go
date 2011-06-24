package recipe

import (
	"io"
	"json"
	"os"

	. "chunkymonkey/types"
)

// FurnaceData contains data on furnace reactions.
type FurnaceData struct {
	// FuelDuration contains a map of fuel types to number of ticks that the fuel
	// lasts for.
	Fuels map[ItemTypeId]Ticks

	// Reactions contains a map of input item type to output item type and data.
	Reactions map[ItemTypeId]Reaction
}

// Reaction describes the output of a furnace reaction.
type Reaction struct {
	Output     ItemTypeId
	OutputData ItemData
}

// furnaceDataDef is used in unmarshalling data from the JSON definition of
// FurnaceData.
type furnaceDataDef struct {
	Fuels []struct {
		Id        ItemTypeId
		FuelTicks Ticks
	}
	Reactions []struct {
		Comment    string
		Input      ItemTypeId
		Output     ItemTypeId
		OutputData ItemData
	}
}

// LoadFurnaceData reads FurnaceData from the reader.
func LoadFurnaceData(reader io.Reader) (furnaceData FurnaceData, err os.Error) {
	decoder := json.NewDecoder(reader)

	var dataDef furnaceDataDef

	err = decoder.Decode(&dataDef)
	if err != nil {
		return
	}

	furnaceData.Fuels = make(map[ItemTypeId]Ticks)
	for _, fuelDef := range dataDef.Fuels {
		furnaceData.Fuels[fuelDef.Id] = fuelDef.FuelTicks
	}

	furnaceData.Reactions = make(map[ItemTypeId]Reaction)
	for _, reactionDef := range dataDef.Reactions {
		furnaceData.Reactions[reactionDef.Input] = Reaction{
			Output:     reactionDef.Output,
			OutputData: reactionDef.OutputData,
		}
	}

	return
}

// LoadFurnaceDataFromFile reads FurnaceData from the named file.
func LoadFurnaceDataFromFile(filename string) (furnaceData FurnaceData, err os.Error) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	return LoadFurnaceData(file)
}
