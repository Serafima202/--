package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Player struct {
	CurrentRoom string
	Inventory   map[string]bool
	HasBackPack bool
}

type Room struct {
	Description      func() string
	EnterDescription string
	Items            map[string]bool
	Exits            map[string]bool
	LockedExits      map[string]bool
	SpecialAction    func(string, string) string
}

var player Player
var rooms map[string]*Room

func hasItems(items ...string) bool {
	for _, item := range items {
		if !player.Inventory[item] {
			return false
		}
	}
	return true
}

func initGame() {
	player = Player{
		CurrentRoom: "кухня",
		Inventory:   make(map[string]bool),
		HasBackPack: false,
	}

	rooms = make(map[string]*Room)

	rooms["кухня"] = &Room{
		Description: func() string {
			if player.HasBackPack && hasItems("ключи", "конспекты") {
				return "ты находишься на кухне, на столе: чай, надо идти в универ. можно пройти - коридор"
			}
			return "ты находишься на кухне, на столе: чай, надо собрать рюкзак и идти в универ. можно пройти - коридор"
		},
		EnterDescription: "кухня, ничего интересного. можно пройти - коридор",
		Items: map[string]bool{
			"чай": true,
		},
		Exits: map[string]bool{
			"коридор": true,
		},
	}

	rooms["коридор"] = &Room{
		Description: func() string {
			return "ничего интересного. можно пройти - кухня, комната, улица"
		},
		EnterDescription: "ничего интересного. можно пройти - кухня, комната, улица",
		Items:            make(map[string]bool),
		Exits: map[string]bool{
			"кухня":   true,
			"комната": true,
			"улица":   true,
		},
		LockedExits: map[string]bool{
			"улица": true,
		},
		SpecialAction: func(item, target string) string {
			if item == "ключи" && target == "дверь" {
				if player.Inventory["ключи"] {
					delete(rooms["коридор"].LockedExits, "улица")
					return "дверь открыта"
				}
				return "нет предмета в инвентаре - ключи"
			}
			return ""
		},
	}

	rooms["комната"] = &Room{
		Description: func() string {
			items := []string{}
			hasBackpack := false

			for item := range rooms["комната"].Items {
				if item == "рюкзак" {
					hasBackpack = true
				} else {
					items = append(items, item)
				}
			}

			if len(items) == 0 && !hasBackpack {
				return "пустая комната. можно пройти - коридор"
			}

			result := ""
			if len(items) > 0 {
				result = "на столе: " + strings.Join(items, ", ")
			}

			if hasBackpack {
				if result != "" {
					result += ", "
				}
				result += "на стуле: рюкзак"
			}

			return result + ". можно пройти - коридор"
		},
		EnterDescription: "ты в своей комнате. можно пройти - коридор",
		Items: map[string]bool{
			"ключи":     true,
			"конспекты": true,
			"рюкзак":    true,
		},
		Exits: map[string]bool{
			"коридор": true,
		},
	}

	rooms["улица"] = &Room{
		EnterDescription: "на улице весна. можно пройти - домой",
		Description: func() string {
			return "на улице весна. можно пройти - домой"
		},
		Items: make(map[string]bool),
		Exits: map[string]bool{
			"домой": true,
		},
	}
}

func handleCommand(command string) string {
	parts := strings.Split(command, " ")
	if len(parts) == 0 {
		return "неизвестная команда"
	}

	switch parts[0] {
	case "осмотреться":
		return handleLookAround()
	case "идти":
		if len(parts) < 2 {
			return "не указано куда идти"
		}
		return handleMove(parts[1])
	case "взять":
		if len(parts) < 2 {
			return "не указано что взять"
		}
		return handleTake(parts[1])
	case "надеть":
		if len(parts) < 2 {
			return "не указано что надеть"
		}
		return handleWear(parts[1])
	case "применить":
		if len(parts) < 3 {
			return "применить что и к чему?"
		}
		return handleUse(parts[1], parts[2])
	default:
		return "неизвестная команда"
	}
}

func handleLookAround() string {
	if room, exists := rooms[player.CurrentRoom]; exists {
		return room.Description()
	}
	return "ты в неизвестной комнате"
}

func handleMove(target string) string {
	currentRoom := rooms[player.CurrentRoom]

	if !currentRoom.Exits[target] {
		return "нет пути в " + target
	}

	if currentRoom.LockedExits != nil && currentRoom.LockedExits[target] {
		return "дверь закрыта"
	}

	player.CurrentRoom = target
	return rooms[target].EnterDescription
}

func handleTake(item string) string {
	currentRoom := rooms[player.CurrentRoom]

	if !currentRoom.Items[item] {
		return "нет такого"
	}

	if !player.HasBackPack {
		return "некуда класть"
	}

	player.Inventory[item] = true
	delete(currentRoom.Items, item)
	return "предмет добавлен в инвентарь: " + item
}

func handleWear(item string) string {
	if item != "рюкзак" {
		return "нельзя надеть"
	}

	currentRoom := rooms[player.CurrentRoom]
	if !currentRoom.Items[item] {
		return "нет такого"
	}

	delete(currentRoom.Items, item)
	player.HasBackPack = true
	return "вы надели: " + item
}

func handleUse(item, target string) string {
	if !player.Inventory[item] {
		return "нет предмета в инвентаре - " + item
	}

	currentRoom := rooms[player.CurrentRoom]
	if currentRoom.SpecialAction != nil {
		result := currentRoom.SpecialAction(item, target)
		if result != "" {
			return result
		}
	}

	return "не к чему применить"
}

func main() {
	initGame()
	fmt.Println(handleLookAround())
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		// Считываем ввод пользователя
		if !scanner.Scan() {
			break
		}

		// Получаем введенную команду
		command := scanner.Text()

		// Проверяем, не хочет ли пользователь выйти
		if command == "выход" {
			fmt.Println("До свидания!")
			break
		}

		// Обрабатываем команду и выводим результат
		result := handleCommand(command)
		fmt.Println(result)
	}

	// Проверяем, не произошла ли ошибка при сканировании
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения ввода: %v\n", err)
	}

}
