package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nagarajasunder/go-kirukkal/internal/dtos"
	"github.com/nagarajasunder/go-kirukkal/internal/globals"
)

var GameWords = [...]string{"Pen", "Paper", "Refrigerator", "Television", "Laptop"}

type IGameService interface {
	CreateRoom(roomName string) (dtos.RoomCreationSuccess, error)
	CreateNewPlayer(roomId string, playerName string, isAdmin bool, conn *websocket.Conn)
	DoesRoomExists(roomId string) bool
}

var roomMap = map[string]*dtos.Room{}

func NewGameService() IGameService {
	return &GameService{
		rooms: roomMap,
	}
}

type GameService struct {
	rooms map[string]*dtos.Room
}

func (r *GameService) CreateRoom(roomName string) (dtos.RoomCreationSuccess, error) {
	var roomId string = uuid.New().String()
	r.addNewRoom(roomId, roomName)

	resp := dtos.RoomCreationSuccess{
		RoomId:   roomId,
		RoomName: roomName,
	}

	message := dtos.RoomCreationMessage{
		RoomId:   roomId,
		RoomName: roomId,
	}
	SendRoomCreationUpdates(message)

	return resp, nil
}

func (r *GameService) DoesRoomExists(roomId string) bool {

	_, exists := r.rooms[roomId]
	return exists
}

func (r *GameService) addNewRoom(roomId string, roomName string) {

	players := []*dtos.Player{}
	newRoom := &dtos.Room{
		RoomId:     roomId,
		RoomName:   roomName,
		Players:    players,
		GameStatus: globals.GAME_STATUS_IDLE,
		DrawnUsers: map[string]bool{},
	}
	r.rooms[roomId] = newRoom
}

func (r *GameService) CreateNewPlayer(roomId string, playerName string, isAdmin bool, playerConn *websocket.Conn) {

	room := r.rooms[roomId]

	players := room.Players
	playerId := uuid.New().String()
	newPlayer := dtos.Player{
		PlayerId:   playerId,
		PlayerName: playerName,
		PlayerConn: playerConn,
		IsAdmin:    isAdmin,
	}

	//Todo("Need to update the logic afterwards")
	if len(players) == 0 {
		newPlayer.IsAdmin = true
	}

	players = append(players, &newPlayer)
	room.Players = players
	r.rooms[roomId] = room

	playerCreationMessage := dtos.PlayCreationSuccess{
		PlayerId:   playerId,
		PlayerName: playerName,
	}

	r.SendMessageToPlayer(&newPlayer, globals.MESSAGE_TYPE_PLAYER_CREATED, playerCreationMessage)
	fmt.Printf("New player created %s \n", playerName)

	go r.readPlayerMessages(roomId, &newPlayer)

	r.NewPlayerJoined(roomId, &newPlayer)

}

func (r *GameService) removePlayer(roomId string, playerId string) {
	room := r.rooms[roomId]
	var updatedPlayers []*dtos.Player

	for _, player := range room.Players {
		if player.PlayerId != playerId {
			updatedPlayers = append(updatedPlayers, player)
		}
	}

	room.Players = updatedPlayers

	r.rooms[roomId] = room
}

func (r *GameService) NewPlayerJoined(roomId string, newPlayer *dtos.Player) {

	/*
		When a new player joins the game
		1. Welcome him
		2. Notify other users to that he had joined
	*/

	room := r.rooms[roomId]

	for _, player := range room.Players {
		if player.PlayerId == newPlayer.PlayerId {
			chatMessage := dtos.Chat{
				Sender:  globals.MESSAGE_SENDER_ADMIN,
				Message: fmt.Sprintf("Welcome to the party %s!!", newPlayer.PlayerName),
			}
			r.SendMessageToPlayer(player, globals.MESSAGE_TYPE_CHAT, chatMessage)
		} else {
			chatMessage := dtos.Chat{
				Sender:  globals.MESSAGE_SENDER_ADMIN,
				Message: fmt.Sprintf("%s joined the party!!", newPlayer.PlayerName),
			}
			r.SendMessageToPlayer(player, globals.MESSAGE_TYPE_CHAT, chatMessage)
		}
	}
}

func (r *GameService) StartGame(roomId string) {
	room := r.rooms[roomId]
	drawingPlayer := r.GetPlayerToChooseWord(roomId)
	room.DrawingUser = drawingPlayer
	room.DrawnUsers[drawingPlayer.PlayerId] = true
	room.GameStatus = globals.GAME_STATUS_STARTED

	for _, player := range room.Players {
		r.SendMessageToPlayer(player, globals.MESSAGE_TYPE_GAME_STATUS, globals.GAME_STATUS_STARTED)
	}

	currentRoundWords := getRandomWordsToChoose()

	gameWordsMessage := dtos.GameWords{
		Words: currentRoundWords,
	}

	r.SendMessageToPlayer(drawingPlayer, globals.MESSAGE_TYPE_PLAYER_WORD_OPTIONS, gameWordsMessage)

	r.SendMessageToPlayer(drawingPlayer, globals.MESSAGE_TYPE_UPDATE_DRAWING_PLAYER, drawingPlayer.PlayerId)

	for _, player := range room.Players {
		if player.PlayerId != drawingPlayer.PlayerId {
			r.SendMessageToPlayer(player, globals.MESSAGE_TYPE_GAME_MESSAGE, fmt.Sprintf("%s is choosing the word...", drawingPlayer.PlayerName))
		}
	}
}

func (r *GameService) SendGameWordsToUsers(roomId string, messageByte json.RawMessage) {
	room := r.rooms[roomId]

	var gameWord string

	err := json.Unmarshal(messageByte, &gameWord)

	if err != nil {
		fmt.Printf("Unable to parse message err %#v", err)
		return
	}

	r.SendMessageToPlayer(room.DrawingUser, globals.MESSAGE_TYPE_GAME_MESSAGE, gameWord)

	room.CurrentRoundWord = &gameWord

	gameWordClue := ""

	for i := 0; i < len(gameWord); i++ {
		gameWordClue += " _ "
	}

	for _, player := range room.Players {
		if player.PlayerId != room.DrawingUser.PlayerId {
			r.SendMessageToPlayer(player, globals.MESSAGE_TYPE_GAME_MESSAGE, gameWordClue)
		}
	}

}

func (r *GameService) ProcessPlayerChats(roomId string, player *dtos.Player, rawMessage json.RawMessage) {
	room := r.rooms[roomId]

	var playerChat dtos.Chat

	err := json.Unmarshal(rawMessage, &playerChat)

	if err != nil {
		fmt.Printf("Unable to process message sent by %s err %#v", player.PlayerName, err)
		return
	}

	for _, p := range room.Players {
		r.SendMessageToPlayer(p, globals.MESSAGE_TYPE_CHAT, playerChat)
	}

}

func (r *GameService) CheckIfGameWordMatch(roomId string, player *dtos.Player, messageByte json.RawMessage) {
	room := r.rooms[roomId]

	var playerChat dtos.Chat

	err := json.Unmarshal(messageByte, &playerChat)

	if err != nil {
		fmt.Printf("Unable to process message sent by %s err %#v", player.PlayerName, err)
		return
	}

	if playerChat.Message == *room.CurrentRoundWord {
		for _, p := range room.Players {
			if p.PlayerId != player.PlayerId {
				chat := dtos.Chat{
					Sender:  globals.MESSAGE_SENDER_ADMIN,
					Message: fmt.Sprintf("Player %s guessed to word", player.PlayerName),
				}
				r.SendMessageToPlayer(p, globals.MESSAGE_SENDER_ADMIN, chat)
			} else {
				r.SendMessageToPlayer(p, globals.MESSAGE_TYPE_CHAT, playerChat)
			}
		}
	} else {
		for _, p := range room.Players {
			r.SendMessageToPlayer(p, globals.MESSAGE_TYPE_CHAT, playerChat)
		}
	}
}

func (r *GameService) GetPlayerToChooseWord(roomId string) *dtos.Player {
	room := r.rooms[roomId]
	for _, player := range room.Players {
		alreadyChosed := room.DrawnUsers[player.PlayerId]
		if !alreadyChosed {
			return player
		}
	}
	return nil
}

func (r *GameService) ProcessDrawMessage(roomId string, messageByte json.RawMessage) {

	room := r.rooms[roomId]

	var line dtos.Line

	err := json.Unmarshal(messageByte, &line)

	if err != nil {
		fmt.Printf("Unable to parse draw message err %#v", err)
		return
	}

	//Stream draw messages to other players
	for _, player := range room.Players {
		if player.PlayerId != room.DrawingUser.PlayerId {
			r.SendMessageToPlayer(player, globals.MESSAGE_TYPE_STREAM_DRAW, messageByte)
		}
	}

}

func (r *GameService) readPlayerMessages(roomId string, player *dtos.Player) {
	for {
		_, messageByte, err := player.PlayerConn.ReadMessage()

		defer func() {
			player.PlayerConn.Close()
			r.removePlayer(roomId, player.PlayerId)
		}()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unable to read message from player %s err %#v \\n", player.PlayerId, err)
			}
			break
		}

		var messagePaylod dtos.GameMessage

		err = json.Unmarshal(messageByte, &messagePaylod)

		if err != nil {
			log.Printf("Unable to read message from client")
		}

		switch messagePaylod.MessageType {
		case globals.MESSAGE_TYPE_START_GAME:
			r.StartGame(roomId)
		case globals.MESSAGE_TYPE_PLAYER_CHOOSED_WORD:
			r.SendGameWordsToUsers(roomId, messagePaylod.Message)
		case globals.MESSAGE_TYPE_CHAT:
			r.ProcessPlayerChats(roomId, player, messagePaylod.Message)
		case globals.MESSAGE_TYPE_DRAW:
			r.ProcessDrawMessage(roomId, messagePaylod.Message)
		case globals.MESSAGE_TYPE_GAME_CHAT:
			r.CheckIfGameWordMatch(roomId, player, messagePaylod.Message)
		}

	}
}

func (r *GameService) SendMessageToPlayer(player *dtos.Player, messageType string, message interface{}) {

	messageByte, err := json.Marshal(message)

	if err != nil {
		fmt.Printf("Unable to marshal message %#v error %#v", message, err)
		return
	}

	wsMessage := dtos.GameMessage{
		MessageType: messageType,
		Message:     messageByte,
	}

	wsMessageByte, err := json.Marshal(wsMessage)

	if err != nil {
		log.Printf("Unable to marshal message payload %#v %#v", wsMessage, err)
		return
	}

	err = player.PlayerConn.WriteMessage(websocket.TextMessage, wsMessageByte)

	if err != nil {
		log.Printf("Unable to send message to player %s err: %#v", player.PlayerId, err)
	}
}

func getRandomWordsToChoose() []string {
	currentRoundWords := []string{}
	for i := 0; i < 3; i++ {
		currentRoundWords = append(currentRoundWords, GameWords[rand.Intn(len(GameWords))])
	}

	return currentRoundWords
}
