package globals

const (
	MESSAGE_TYPE_START_GAME           = "START_GAME"
	MESSAGE_TYPE_PLAYER_WORD_OPTIONS  = "CHOOSE_WORD"
	MESSAGE_TYPE_PLAYER_CHOOSING_WORD = "PLAYER_CHOOSING_WORD"
	MESSAGE_TYPE_PLAYER_CHOOSED_WORD  = "CHOOSED_WORD"
	MESSAGE_TYPE_GAME_WORD_CLUE       = "GAME_WORD_CLUE"
	MESSAGE_TYPE_DRAW                 = "DRAW"
	MESSAGE_TYPE_CHAT                 = "CHAT"
	/*
		This message type is basically when the system(game) needs to send messages to its players
	*/
	MESSAGE_TYPE_ADMIN_INFO = "ADMIN_MESSAGE"
)

const (
	MESSAGE_SENDER_ADMIN = "ADMIN"
)
