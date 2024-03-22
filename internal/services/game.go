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
	CreateRoom() (dtos.RoomCreationSuccess, error)
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

func (r *GameService) CreateRoom() (dtos.RoomCreationSuccess, error) {
	var roomId string = uuid.New().String()
	r.addNewRoom(roomId)

	resp := dtos.RoomCreationSuccess{
		RoomId: roomId,
	}

	return resp, nil
}

func (r *GameService) DoesRoomExists(roomId string) bool {

	_, exists := r.rooms[roomId]
	return exists
}

func (r *GameService) addNewRoom(roomId string) {

	players := []*dtos.Player{}
	newRoom := &dtos.Room{
		RoomId:     roomId,
		Players:    players,
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

	playerInfo := dtos.PlayCreationSuccess{
		PlayerId: playerId,
	}

	playerCreationMessage, err := json.Marshal(playerInfo)

	if err != nil {
		log.Printf("Unable to send message to player %s err: %#v", playerId, err)
		playerConn.Close()
		return
	}

	playerConn.WriteMessage(websocket.TextMessage, playerCreationMessage)
	fmt.Printf("New player created %s \n", playerName)

	go r.readPlayerMessages(roomId, &newPlayer)
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
			r.SendGameWordsToOtherUsers(roomId, messagePaylod.Message)
		case globals.MESSAGE_TYPE_CHAT:
			r.CheckIfGameWordMatch(roomId, player, messagePaylod.Message)

		}

	}
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

func (r *GameService) StartGame(roomId string) {
	room := r.rooms[roomId]
	drawingPlayer := r.GetPlayerToChooseWord(roomId)
	room.DrawingUser = drawingPlayer
	room.DrawnUsers[drawingPlayer.PlayerId] = true

	currentRoundWords := getRandomWordsToChoose()

	wsMessage := dtos.GameMessage{
		MessageType: globals.MESSAGE_TYPE_PLAYER_WORD_OPTIONS,
		Message:     currentRoundWords,
	}

	r.SendMessageToPlayer(drawingPlayer, wsMessage)

	for _, player := range room.Players {
		if player.PlayerId != drawingPlayer.PlayerId {
			wsMessage := dtos.GameMessage{
				MessageType: globals.MESSAGE_TYPE_PLAYER_CHOOSING_WORD,
				Message:     fmt.Sprintf("%s is choosing the word", drawingPlayer.PlayerName),
			}
			r.SendMessageToPlayer(player, wsMessage)
		}
	}
}

func (r *GameService) SendGameWordsToOtherUsers(roomId string, message interface{}) {
	room := r.rooms[roomId]

	gameWord := message.(string)

	room.CurrentRoundWord = &gameWord

	for _, player := range room.Players {
		if player.PlayerId != room.DrawingUser.PlayerId {
			wsMessage := dtos.GameMessage{
				MessageType: globals.MESSAGE_TYPE_GAME_WORD_CLUE,
				Message:     len(gameWord),
			}
			r.SendMessageToPlayer(player, wsMessage)
		}
	}

}

func (r *GameService) CheckIfGameWordMatch(roomId string, player *dtos.Player, message interface{}) {
	room := r.rooms[roomId]

	playerGuessedWord := message.(string)

	if playerGuessedWord == *room.CurrentRoundWord {
		for _, p := range room.Players {
			if p.PlayerId != player.PlayerId {
				chat := dtos.Chat{
					Sender:  globals.MESSAGE_SENDER_ADMIN,
					Message: fmt.Sprintf("Player %s guessed to word", player.PlayerName),
				}
				wsMessage := dtos.GameMessage{
					MessageType: globals.MESSAGE_TYPE_CHAT,
					Message:     chat,
				}
				r.SendMessageToPlayer(p, wsMessage)
			} else {
				chat := dtos.Chat{
					Sender:  player.PlayerName,
					Message: playerGuessedWord,
				}
				wsMessage := dtos.GameMessage{
					MessageType: globals.MESSAGE_TYPE_CHAT,
					Message:     chat,
				}
				r.SendMessageToPlayer(p, wsMessage)
			}
		}
	} else {
		for _, p := range room.Players {
			chat := dtos.Chat{
				Sender:  player.PlayerName,
				Message: playerGuessedWord,
			}
			wsMessage := dtos.GameMessage{
				MessageType: globals.MESSAGE_TYPE_CHAT,
				Message:     chat,
			}
			r.SendMessageToPlayer(p, wsMessage)
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

func (r *GameService) ProcessDrawMessage(roomId string, playerId string, messageType int, message []byte) {

	room := r.rooms[roomId]

	//Filter other players
	var otherPlayers []*dtos.Player

	for _, player := range room.Players {
		if player.PlayerId != playerId {
			otherPlayers = append(otherPlayers, player)
		}
	}

	//Broadcast messages

	for _, player := range otherPlayers {
		player.PlayerConn.WriteMessage(messageType, message)
	}

}

func (r *GameService) SendMessageToPlayer(player *dtos.Player, message dtos.GameMessage) {

	messageByte, err := json.Marshal(message)

	if err != nil {
		log.Printf("Unable to marshal message payload %#v %#v", message, err)
		return
	}

	err = player.PlayerConn.WriteMessage(websocket.TextMessage, messageByte)

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
