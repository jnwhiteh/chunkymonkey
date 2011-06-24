package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"chunkymonkey/nbt"
	. "chunkymonkey/types"
	"chunkymonkey/worldstore"
)

func usage() {
	os.Stderr.WriteString("usage: " + os.Args[0] + " [options] <world path>\n")
	flag.PrintDefaults()
}

func displayNbt(indentCount int, tag nbt.ITag) {
	indent := strings.Repeat("  ", indentCount)
	switch t := tag.(type) {
	case *nbt.Compound:
		fmt.Print("Compound:\n")
		for name, subTag := range t.Tags {
			fmt.Printf("%s%#v: ", indent, name)
			displayNbt(indentCount+1, subTag)
		}
	case *nbt.List:
		fmt.Printf("List[%d]:\n", len(t.Value))
		for index, subTag := range t.Value {
			fmt.Printf("%s[%d]: ", indent, index)
			displayNbt(indentCount+1, subTag)
		}
	case *nbt.Byte:
		fmt.Printf("Byte: %d\n", t.Value)
	case *nbt.Short:
		fmt.Printf("Short: %d\n", t.Value)
	case *nbt.Int:
		fmt.Printf("Int: %d\n", t.Value)
	case *nbt.Long:
		fmt.Printf("Long: %d\n", t.Value)
	case *nbt.Float:
		fmt.Printf("Float: %f\n", t.Value)
	case *nbt.Double:
		fmt.Printf("Double: %f\n", t.Value)
	case *nbt.ByteArray:
		fmt.Printf("ByteArray: %d bytes long\n", len(t.Value))
	case *nbt.String:
		fmt.Printf("String: %#v\n", t.Value)
	case *nbt.NamedTag:
		displayNbt(indentCount, t.Tag)
	default:
		fmt.Printf("<Unknown type>\n")
	}
}

func main() {
	showWorld := flag.Bool("world", false, "Display world information")
	x := flag.Int64("x", 0, "X chunk coordinate of chunk to display")
	z := flag.Int64("z", 0, "Z chunk coordinate of chunk to display")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	worldPath := flag.Arg(0)

	worldStore, err := worldstore.LoadWorldStore(worldPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading world: %s\n", err)
		return
	}

	if *showWorld {
		fmt.Print("Level data:\n")
		displayNbt(1, worldStore.LevelData)
	}

	chunkLoc := &ChunkXz{ChunkCoord(*x), ChunkCoord(*z)}

	chunkResult := <-worldStore.ChunkStore.LoadChunk(chunkLoc)
	if chunkResult.Err != nil {
		fmt.Fprintf(os.Stderr, "Error loading chunk: %s\n", chunkResult.Err)
		return
	}

	fmt.Printf("Chunk %#v data:\n", *chunkLoc)
	displayNbt(1, chunkResult.Reader.GetRootTag())
}
