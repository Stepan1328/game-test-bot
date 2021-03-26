package clients

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

func UploadDatabase() {
	wg := new(sync.WaitGroup)
	wg.Add(3)

	go uploadPlayersBase(wg)
	go uploadBattlesBase(wg)
	go uploadCompatibility(wg)
	wg.Wait()

	for _, player := range Players {
		player.ParseLangMap()
	}
	log.Println("Database was successfully uploaded")
}

func uploadPlayersBase(wg *sync.WaitGroup) {
	var playersBase map[int]*UsersStatistic
	data, err := os.ReadFile("database/playersBase.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &playersBase)
	if err != nil {
		fmt.Println(err)
	}

	Players = playersBase
	wg.Done()
}

func uploadBattlesBase(wg *sync.WaitGroup) {
	var battlesBase map[string]*BattleStatistic
	data, err := os.ReadFile("database/battlesBase.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &battlesBase)
	if err != nil {
		fmt.Println(err)
	}

	Battles = battlesBase
	wg.Done()
}

func uploadCompatibility(wg *sync.WaitGroup) {
	var compatibility map[string]int
	data, err := os.ReadFile("database/compatibility.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &compatibility)
	if err != nil {
		fmt.Println(err)
	}

	Compatibility = compatibility
	wg.Done()
}

func SaveBase() {
	wg := new(sync.WaitGroup)
	wg.Add(3)

	go savePlayersBase(wg)
	go saveBattlesBase(wg)
	go saveCompatibility(wg)
	wg.Wait()
}

func savePlayersBase(wg *sync.WaitGroup) {
	data, err := json.MarshalIndent(Players, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("database/playersBase.json", data, 0600); err != nil {
		panic(err)
	}
	wg.Done()
}

func saveBattlesBase(wg *sync.WaitGroup) {
	data, err := json.MarshalIndent(Battles, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("database/battlesBase.json", data, 0600); err != nil {
		panic(err)
	}
	wg.Done()
}

func saveCompatibility(wg *sync.WaitGroup) {
	data, err := json.MarshalIndent(Compatibility, "", "  ")
	if err != nil {
		panic(err)
	}

	if err := os.WriteFile("database/compatibility.json", data, 0600); err != nil {
		panic(err)
	}
	wg.Done()
}
